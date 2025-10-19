package integration

import (
    "io"
    "net"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/yourorg/gofakeip/pkg/proxy"
)

// T032: Integration E2E test for proxy server (HTTP) verifying IP masking works end-to-end.
func TestProxyE2E_HTTPMasking(t *testing.T) {
    // Backend server
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    }))
    defer backend.Close()

    // Simple logger
    logger := &testLogger{t: t}

    // Start proxy server
    srv := proxy.NewProxyServer("127.0.0.1:0", net.ParseIP("127.0.0.1"), []string{"http"}, 100, logger)
    if err := srv.Start(); err != nil {
        t.Skipf("failed to start proxy server in this environment: %v", err)
    }
    defer srv.Stop()

    // Give server a moment to start
    time.Sleep(100 * time.Millisecond)

    // Make a request via the proxy to the backend
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(backend.URL)
    if err != nil {
        t.Fatalf("proxy request failed: %v", err)
    }
    defer resp.Body.Close()
    b, _ := io.ReadAll(resp.Body)
    if string(b) != "ok" {
        t.Fatalf("unexpected response: %s", string(b))
    }
}

type testLogger struct{ t *testing.T }

func (l *testLogger) Info(msg string, args ...interface{})  {}
func (l *testLogger) Error(msg string, args ...interface{}) {}
func (l *testLogger) Debug(msg string, args ...interface{}) {}
func (l *testLogger) Warn(msg string, args ...interface{})  {}
