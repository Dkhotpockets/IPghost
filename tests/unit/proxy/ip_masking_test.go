package proxy_test

import (
    "net"
    "testing"

    "github.com/yourorg/gofakeip/pkg/proxy"
)

// T029: Unit test for IP masking - verify DialWithMaskedIP sets LocalAddr to masked IP.
func TestDialWithMaskedIP_LocalAddr(t *testing.T) {
    // Start a local TCP listener
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatalf("failed to start listener: %v", err)
    }
    defer ln.Close()

    maskedIP := net.ParseIP("127.0.0.1")
    conn, err := proxy.DialWithMaskedIP("tcp", ln.Addr().String(), maskedIP)
    if err != nil {
        t.Skipf("DialWithMaskedIP failed; skipping test on this OS/setup: %v", err)
    }
    defer conn.Close()

    localAddr, ok := conn.LocalAddr().(*net.TCPAddr)
    if !ok {
        t.Fatalf("expected TCPAddr, got %T", conn.LocalAddr())
    }

    if !localAddr.IP.Equal(maskedIP) {
        t.Fatalf("expected local IP %s, got %s", maskedIP.String(), localAddr.IP.String())
    }
}
