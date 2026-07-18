// Package callback receives an OAuth redirect on the local loopback interface.
package callback

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Result is the successful information carried by a callback.
type Result struct {
	Code string
}

// Server binds loopback listeners and waits for one OAuth callback.
type Server struct {
	Port         int
	CallbackPath string
	State        string

	mu        sync.Mutex
	listeners []net.Listener
	listened  bool
	waited    bool
	listen    func(network, address string) (net.Listener, error)
}

// Listen binds IPv4 loopback first and then attempts to bind IPv6 loopback on
// the same port. An unavailable IPv6 loopback does not prevent IPv4 service.
func (s *Server) Listen() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listened {
		return 0, errors.New("callback server has already listened")
	}
	listen := s.listen
	if listen == nil {
		listen = net.Listen
	}

	ipv4, err := listen("tcp4", net.JoinHostPort("127.0.0.1", fmt.Sprint(s.Port)))
	if err != nil {
		return 0, fmt.Errorf("listen on IPv4 loopback port %d: %w", s.Port, err)
	}
	port := ipv4.Addr().(*net.TCPAddr).Port

	ipv6, err := listen("tcp6", net.JoinHostPort("::1", fmt.Sprint(port)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: IPv6 loopback callback listener unavailable: %v\n", err)
		s.listeners = []net.Listener{ipv4}
	} else {
		s.listeners = []net.Listener{ipv4, ipv6}
	}
	s.listened = true
	return port, nil
}

type outcome struct {
	result Result
	err    error
}

// Wait serves the bound listeners until a callback completes the flow or ctx
// expires. Listen must be called first, and Wait may be called only once.
func (s *Server) Wait(ctx context.Context) (Result, error) {
	s.mu.Lock()
	if !s.listened {
		s.mu.Unlock()
		return Result{}, errors.New("callback server is not listening")
	}
	if s.waited {
		s.mu.Unlock()
		return Result{}, errors.New("callback server has already waited")
	}
	s.waited = true
	listeners := append([]net.Listener(nil), s.listeners...)
	s.mu.Unlock()

	started := time.Now()
	resultCh := make(chan outcome, 1)
	var once sync.Once
	deliver := func(got outcome) {
		once.Do(func() { resultCh <- got })
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != s.CallbackPath {
			http.NotFound(w, r)
			return
		}

		query := r.URL.Query()
		states, hasState := query["state"]
		if !hasState || len(states) == 0 || states[0] != s.State {
			writePage(w, http.StatusBadRequest, "Login failed", "The callback state did not match.")
			deliver(outcome{err: errors.New("callback state did not match")})
			return
		}
		if oauthError := query.Get("error"); oauthError != "" {
			description := query.Get("error_description")
			message := fmt.Sprintf("authorization error %q: %s", oauthError, description)
			writePage(w, http.StatusBadRequest, "Login failed", message)
			deliver(outcome{err: errors.New(message)})
			return
		}
		code := query.Get("code")
		if code == "" {
			writePage(w, http.StatusBadRequest, "Login failed", "The callback contained neither an authorization code nor an error.")
			deliver(outcome{err: errors.New("callback contained neither code nor error")})
			return
		}

		writePage(w, http.StatusOK, "Login complete", "Login complete — return to your terminal.")
		deliver(outcome{result: Result{Code: code}})
	})

	servers := make([]*http.Server, 0, len(listeners))
	for _, listener := range listeners {
		httpServer := &http.Server{Handler: handler}
		servers = append(servers, httpServer)
		go func() {
			if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				deliver(outcome{err: fmt.Errorf("serve callback: %w", err)})
			}
		}()
	}

	var got outcome
	select {
	case got = <-resultCh:
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			elapsed := time.Since(started).Round(time.Millisecond)
			got.err = fmt.Errorf("callback deadline expired after %s: %w", elapsed, ctx.Err())
		} else {
			got.err = fmt.Errorf("callback wait canceled: %w", ctx.Err())
		}
	}

	for _, httpServer := range servers {
		httpServer.SetKeepAlivesEnabled(false)
		_ = httpServer.Close()
	}
	return got.result, got.err
}

func writePage(w http.ResponseWriter, status int, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>"+
		html.EscapeString(title)+"</title></head><body><main><h1>"+
		html.EscapeString(title)+"</h1><p>"+html.EscapeString(message)+"</p></main></body></html>")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
