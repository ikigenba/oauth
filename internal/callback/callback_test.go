package callback

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestListenAssignsReachableIPv4Port(t *testing.T) {
	// R-11L3-FGOF
	server := &Server{CallbackPath: "/callback", State: "state"}
	port, err := server.Listen()
	if err != nil {
		t.Fatal(err)
	}
	if port == 0 {
		t.Fatal("Listen returned port zero")
	}
	connection, err := net.DialTimeout("tcp4", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if err != nil {
		t.Fatalf("connect to returned IPv4 port: %v", err)
	}
	connection.Close()
	cancelWait(t, server)
}

func TestListenServesSameHandlerOnIPv6(t *testing.T) {
	// R-12SZ-T8F4
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	response := get(t, fmt.Sprintf("http://[::1]:%d/callback?state=expected&code=ipv6", port))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "ipv6" {
		t.Fatalf("Wait returned %#v, %v", got.result, got.err)
	}
}

func TestIPv6BindFailureStillServesIPv4(t *testing.T) {
	// R-140W-705T
	server := &Server{CallbackPath: "/callback", State: "expected"}
	server.listen = func(network, address string) (net.Listener, error) {
		if network == "tcp6" {
			return nil, errors.New("simulated unavailable IPv6")
		}
		return net.Listen(network, address)
	}
	port, err := server.Listen()
	if err != nil {
		t.Fatalf("IPv6 failure made Listen fail: %v", err)
	}
	resultCh := waitAsync(server, context.Background())
	response := get(t, callbackURL(port, "expected", "ipv4"))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "ipv4" {
		t.Fatalf("IPv4 callback returned %#v, %v", got.result, got.err)
	}
}

func TestListenUsesExplicitPortAndReportsInUse(t *testing.T) {
	// R-158S-KRWI
	reservation, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := reservation.Addr().(*net.TCPAddr).Port
	reservation.Close()

	server := &Server{Port: port, CallbackPath: "/callback", State: "expected"}
	gotPort, err := server.Listen()
	if err != nil {
		t.Fatal(err)
	}
	if gotPort != port {
		t.Fatalf("Listen returned port %d, want %d", gotPort, port)
	}
	cancelWait(t, server)

	inUse, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer inUse.Close()
	blocked := &Server{Port: inUse.Addr().(*net.TCPAddr).Port}
	if _, err := blocked.Listen(); err == nil {
		t.Fatal("Listen succeeded on a port already in use")
	}
}

func TestUnknownPathDoesNotEndWait(t *testing.T) {
	// R-16GO-YJN7
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	response := get(t, fmt.Sprintf("http://127.0.0.1:%d/favicon.ico", port))
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("favicon status = %d, want 404", response.StatusCode)
	}
	select {
	case got := <-resultCh:
		t.Fatalf("Wait ended on unknown path: %v", got.err)
	case <-time.After(30 * time.Millisecond):
	}
	response = get(t, callbackURL(port, "expected", "real-code"))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "real-code" {
		t.Fatalf("valid callback returned %#v, %v", got.result, got.err)
	}
}

func TestValidCallbackReturnsCodeAndSelfContainedPage(t *testing.T) {
	// R-18WH-Q34L
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	response := get(t, callbackURL(port, "expected", "code-exact"))
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "code-exact" {
		t.Fatalf("Wait returned %#v, %v", got.result, got.err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	page := strings.ToLower(string(body))
	for _, external := range []string{" src=", " href=", "http://", "https://"} {
		if strings.Contains(page, external) {
			t.Fatalf("success page contains external reference marker %q: %s", external, page)
		}
	}
	if !strings.Contains(page, "login complete") || !strings.Contains(page, "return to your terminal") {
		t.Fatalf("success page lacks completion message: %s", page)
	}
}

func TestInvalidStateReturnsErrorPageAndNoCode(t *testing.T) {
	// R-1A4E-3UVA
	for _, test := range []struct {
		name  string
		state string
	}{
		{name: "mismatch", state: "wrong"},
		{name: "missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, port := listeningServer(t)
			resultCh := waitAsync(server, context.Background())
			response := get(t, callbackURL(port, test.state, "must-not-pass"))
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()
			got := receive(t, resultCh)
			if got.err == nil || got.result.Code != "" {
				t.Fatalf("Wait returned %#v, %v", got.result, got.err)
			}
			page := strings.ToLower(string(body))
			if response.StatusCode == http.StatusOK || strings.Contains(page, "return to your terminal") || !strings.Contains(page, "failed") {
				t.Fatalf("response was not an error page: status %d, body %s", response.StatusCode, page)
			}
		})
	}
}

func TestAuthorizationErrorWinsOverCode(t *testing.T) {
	// R-1BCA-HMLZ
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	values := url.Values{"state": {"expected"}, "code": {"ignored"}, "error": {"access_denied"}, "error_description": {"user declined"}}
	response := get(t, fmt.Sprintf("http://127.0.0.1:%d/callback?%s", port, values.Encode()))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err == nil || got.result.Code != "" {
		t.Fatalf("Wait returned %#v, %v", got.result, got.err)
	}
	if !strings.Contains(got.err.Error(), "access_denied") || !strings.Contains(got.err.Error(), "user declined") {
		t.Fatalf("error does not carry provider error and description: %v", got.err)
	}
}

func TestMissingCodeAndErrorIsFatal(t *testing.T) {
	// R-1CK6-VECO
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	response := get(t, fmt.Sprintf("http://127.0.0.1:%d/callback?state=expected", port))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err == nil || got.result.Code != "" {
		t.Fatalf("malformed callback returned %#v, %v", got.result, got.err)
	}
}

func TestFirstCallbackWinsAndListenerStops(t *testing.T) {
	// R-1DS3-963D
	server, port := listeningServer(t)
	resultCh := waitAsync(server, context.Background())
	response := get(t, callbackURL(port, "expected", "first"))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "first" {
		t.Fatalf("first callback returned %#v, %v", got.result, got.err)
	}
	connection, err := net.DialTimeout("tcp4", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
	if err == nil {
		connection.Close()
		t.Fatal("listener still accepted a connection after Wait returned")
	}
	select {
	case extra := <-resultCh:
		t.Fatalf("received second result: %#v", extra)
	default:
	}
}

func TestDeadlineReturnsBudgetErrorAndNoCode(t *testing.T) {
	// R-1EZZ-MXU2
	server, _ := listeningServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	got := <-waitAsync(server, ctx)
	if got.err == nil || got.result.Code != "" {
		t.Fatalf("deadline returned %#v, %v", got.result, got.err)
	}
	if !strings.Contains(got.err.Error(), "deadline") || !strings.Contains(got.err.Error(), "ms") {
		t.Fatalf("deadline error does not name elapsed budget: %v", got.err)
	}
}

func TestDeadlineAfterCallbackDoesNotChangeResult(t *testing.T) {
	// R-1G7W-0PKR
	server, port := listeningServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	resultCh := waitAsync(server, ctx)
	response := get(t, callbackURL(port, "expected", "on-time"))
	response.Body.Close()
	got := receive(t, resultCh)
	if got.err != nil || got.result.Code != "on-time" {
		t.Fatalf("callback returned %#v, %v", got.result, got.err)
	}
	<-ctx.Done()
	if got.err != nil || got.result.Code != "on-time" {
		t.Fatalf("deadline changed completed result to %#v, %v", got.result, got.err)
	}
}

type waitResult struct {
	result Result
	err    error
}

func listeningServer(t *testing.T) (*Server, int) {
	t.Helper()
	server := &Server{CallbackPath: "/callback", State: "expected"}
	port, err := server.Listen()
	if err != nil {
		t.Fatal(err)
	}
	return server, port
}

func waitAsync(server *Server, ctx context.Context) <-chan waitResult {
	resultCh := make(chan waitResult, 1)
	go func() {
		result, err := server.Wait(ctx)
		resultCh <- waitResult{result: result, err: err}
	}()
	return resultCh
}

func receive(t *testing.T, resultCh <-chan waitResult) waitResult {
	t.Helper()
	select {
	case got := <-resultCh:
		return got
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for callback result")
		return waitResult{}
	}
}

func get(t *testing.T, target string) *http.Response {
	t.Helper()
	client := &http.Client{
		Timeout:   time.Second,
		Transport: &http.Transport{Proxy: nil},
	}
	var response *http.Response
	var err error
	for attempt := 0; attempt < 100; attempt++ {
		response, err = client.Get(target)
		if err == nil {
			return response
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("GET %s: %v", target, err)
	return nil
}

func callbackURL(port int, state, code string) string {
	values := url.Values{"code": {code}}
	if state != "" {
		values.Set("state", state)
	}
	return fmt.Sprintf("http://127.0.0.1:%d/callback?%s", port, values.Encode())
}

func cancelWait(t *testing.T, server *Server) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	resultCh := waitAsync(server, ctx)
	cancel()
	got := receive(t, resultCh)
	if got.err == nil {
		t.Fatal("canceled Wait returned no error")
	}
}
