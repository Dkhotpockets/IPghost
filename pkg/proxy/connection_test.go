package proxy

import (
	"net"
	"testing"
)

func TestConnectionStateAndStrings(t *testing.T) {
	c := &Connection{ID: "conn-1"}

	if c.GetState() != "" {
		t.Fatalf("expected empty initial state, got %q", c.GetState())
	}

	c.SetState("active")
	if got := c.GetState(); got != "active" {
		t.Fatalf("expected state active, got %q", got)
	}

	// Test RemoteIPString when ClientAddr is nil
	if s := c.RemoteIPString(); s != "" {
		t.Fatalf("expected empty RemoteIPString, got %q", s)
	}

	// create a dummy UDPAddr
	addr := &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 12345}
	c.ClientAddr = addr
	if s := c.RemoteIPString(); s == "" {
		t.Fatalf("expected non-empty RemoteIPString, got empty")
	}

	// MaskedSourceIP
	if s := c.MaskedSourceIPString(); s != "" {
		t.Fatalf("expected empty MaskedSourceIPString, got %q", s)
	}
	c.MaskedSourceIP = net.ParseIP("198.51.100.2")
	if s := c.MaskedSourceIPString(); s == "" {
		t.Fatalf("expected non-empty MaskedSourceIPString, got empty")
	}
}
