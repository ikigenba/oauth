//go:build live

// Run the live login smoke test with:
//
//	go test -tags live ./cmd/oauth/ -run TestLiveLogin -v
//
// Set OAUTH_LIVE_AUTH_URL, OAUTH_LIVE_TOKEN_URL,
// OAUTH_LIVE_CLIENT_ID, OAUTH_LIVE_SCOPE,
// OAUTH_LIVE_PORT, and OAUTH_LIVE_CALLBACK_PATH before running it.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiveLogin(t *testing.T) {
	// R-2262-WKX9
	const (
		authURLEnv      = "OAUTH_LIVE_AUTH_URL"
		tokenURLEnv     = "OAUTH_LIVE_TOKEN_URL"
		clientIDEnv     = "OAUTH_LIVE_CLIENT_ID"
		scopeEnv        = "OAUTH_LIVE_SCOPE"
		portEnv         = "OAUTH_LIVE_PORT"
		callbackPathEnv = "OAUTH_LIVE_CALLBACK_PATH"
	)

	required := []string{authURLEnv, tokenURLEnv, clientIDEnv, scopeEnv, portEnv, callbackPathEnv}
	values := make(map[string]string, len(required))
	var missing []string
	for _, name := range required {
		values[name] = os.Getenv(name)
		if values[name] == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		t.Skipf("live oauth requires these environment variables: %s", strings.Join(missing, ", "))
	}

	binary := filepath.Join(t.TempDir(), "oauth")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build oauth: %v\n%s", err, output)
	}

	command := exec.Command(binary,
		"--auth-url", values[authURLEnv],
		"--token-url", values[tokenURLEnv],
		"--client-id", values[clientIDEnv],
		"--scope", values[scopeEnv],
		"--port", values[portEnv],
		"--callback-path", values[callbackPathEnv],
		"--no-browser",
	)
	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("oauth exited unsuccessfully: %v", err)
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not a JSON token response: %v; stdout=%q", err, stdout.Bytes())
	}
	if response.AccessToken == "" {
		t.Fatalf("stdout token response has an empty access_token: %s", stdout.Bytes())
	}
}
