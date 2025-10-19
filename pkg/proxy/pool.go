package proxy

import (
	"sync"
)

// ConnectionPool manages concurrent connections and enforces a maximum limit

type ConnectionPool struct {
	connections map[string]*Connection // Active connections by ID
	maxConns    int                    // Maximum allowed concurrent connections
	mu          sync.Mutex             // Mutex for thread-safe access
}

// NewConnectionPool creates a new pool with a max connection limit
func NewConnectionPool(maxConns int) *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*Connection),
		maxConns:    maxConns,
	}
}

// Add adds a connection to the pool if under limit
func (p *ConnectionPool) Add(conn *Connection) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.connections) >= p.maxConns {
		return false
	}
	p.connections[conn.ID] = conn
	return true
}

// Remove deletes a connection from the pool
func (p *ConnectionPool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.connections, id)
}

// Count returns the number of active connections
func (p *ConnectionPool) Count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.connections)
}
