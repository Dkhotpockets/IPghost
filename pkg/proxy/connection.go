package proxy

import (
	"net"
	"sync"
)

// Connection represents a proxied client connection with IP masking.
// Fields per data-model.md: ID, ClientAddr, DestinationAddr, MaskedSourceIP, state tracking
type Connection struct {
	ID              string   // Unique connection ID
	ClientAddr      net.Addr // Original client address
	DestinationAddr net.Addr // Target destination address
	MaskedSourceIP  net.IP   // The fake/masked source IP used for outbound traffic

	mu    sync.RWMutex
	state string // Connection state (e.g., "new", "active", "closed")
}

// SetState updates the connection state in a thread-safe manner.
func (c *Connection) SetState(s string) {
	c.mu.Lock()
	c.state = s
	c.mu.Unlock()
}

// GetState returns the current connection state.
func (c *Connection) GetState() string {
	c.mu.RLock()
	s := c.state
	c.mu.RUnlock()
	return s
}

// RemoteIPString returns the client IP address as a string, or empty string if unavailable.
func (c *Connection) RemoteIPString() string {
	if c == nil || c.ClientAddr == nil {
		return ""
	}
	return c.ClientAddr.String()
}

// MaskedSourceIPString returns the masked source IP as a string, or empty string if nil.
func (c *Connection) MaskedSourceIPString() string {
	if c == nil || c.MaskedSourceIP == nil {
		return ""
	}
	return c.MaskedSourceIP.String()
}
