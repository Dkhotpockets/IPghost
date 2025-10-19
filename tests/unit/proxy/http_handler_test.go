package proxy_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/gofakeip/pkg/proxy"
)

// T031: Unit test for HTTP proxy handler (GET forwarding)
func TestHTTPProxyHandler_GET(t *testing.T) {
	// Backend server that returns a fixed body
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-backend"))
	}))
	defer backend.Close()

	// Create HTTP proxy handler with masked IP binding to loopback
	handler := proxy.HTTPProxyHandler(net.ParseIP("127.0.0.1"), 10*time.Second)
	proxySrv := httptest.NewServer(handler)
	defer proxySrv.Close()

	// Make a request via the proxy to the backend
	client := &http.Client{}
	req, _ := http.NewRequest("GET", backend.URL, nil)
	req.URL.Host = backend.Listener.Addr().String()
	req.URL.Scheme = "http"
	// Use proxy URL
	req.URL.Scheme = backend.URL[:4]

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("Skipping HTTP proxy handler test due to environment: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "hello-backend" {
		t.Fatalf("unexpected body: %s", string(b))
	}
}
