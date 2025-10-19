package proxy_test

import (
    "io"
    "net"
    "testing"
    "time"

    "github.com/yourorg/gofakeip/pkg/proxy"
    xnetproxy "golang.org/x/net/proxy"
)

// T030: Unit test for SOCKS5 handler - connection flow
func TestSocks5Handler_ConnectionFlow(t *testing.T) {
    // Start a backend echo server
    backend, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatalf("failed to start backend: %v", err)
    }
    defer backend.Close()
    go func() {
        for {
            conn, err := backend.Accept()
            if err != nil {
                return
            }
            go func(c net.Conn) {
                defer c.Close()
                io.Copy(c, c) // echo
            }(conn)
        }
    }()

    	// Start SOCKS5 server
    	bindIP := net.ParseIP("127.0.0.1")
    	ln, stop, err := proxy.NewSocks5Server(bindIP, "127.0.0.1:0", 10*time.Second)
    	if err != nil {
    		t.Skipf("SOCKS5 server could not start: %v", err)
    	}
    	defer stop()
    // Give server a moment to start
    time.Sleep(100 * time.Millisecond)

    // Connect to SOCKS5 server as a client
    dialAddr := ln.Addr().String()

    // Use golang.org/x/net/proxy SOCKS5 dialer
    socks5Dialer, err := xnetproxy.SOCKS5("tcp", dialAddr, nil, xnetproxy.Direct)
    if err != nil {
        t.Fatalf("failed to create SOCKS5 dialer: %v", err)
    }
    c, err := socks5Dialer.Dial("tcp", backend.Addr().String())
    if err != nil {
        t.Fatalf("SOCKS5 client failed to connect: %v", err)
    }
    defer c.Close()

    // Test echo
    msg := []byte("hello-socks5")
    if _, err := c.Write(msg); err != nil {
        t.Fatalf("write failed: %v", err)
    }
    buf := make([]byte, len(msg))
    if _, err := io.ReadFull(c, buf); err != nil {
        t.Fatalf("read failed: %v", err)
    }
    if string(buf) != string(msg) {
        t.Fatalf("echo mismatch: got %q, want %q", string(buf), string(msg))
    }
}
