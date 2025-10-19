package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/yourorg/gofakeip/pkg/common"
)

// ProxyServer represents the GoFakeIP proxy server instance
// Fields per data-model.md: ListenAddress, BindIP, Protocols, state management
type ProxyServer struct {
	ListenAddress string        // Address to listen on (host:port)
	BindIP        net.IP        // IP to bind for outbound connections (masked IP)
	Protocols     []string      // Supported protocols (e.g., ["socks5", "http"])
	MaxConns      int           // Maximum concurrent connections
	Timeout       time.Duration // Timeout for outbound connections

	pool          *ConnectionPool
	httpServer    HTTPServer
	socks5Server  SOCKS5Server
	mu            sync.RWMutex
	state         string        // Server state (e.g., "starting", "running", "stopped")
	shutdownCh    chan struct{}
	logger        common.Logger
}

// HTTPServer interface for HTTP proxy server operations
type HTTPServer interface {
	Start() error
	Shutdown(ctx context.Context) error
}

// SOCKS5Server interface for SOCKS5 proxy server operations
type SOCKS5Server interface {
	Start() error
	Stop() error
}

// NewProxyServer creates a new ProxyServer instance
func NewProxyServer(listenAddr string, bindIP net.IP, protocols []string, maxConns int, timeout time.Duration, logger common.Logger) *ProxyServer {
	return &ProxyServer{
		ListenAddress: listenAddr,
		BindIP:        bindIP,
		Protocols:     protocols,
		MaxConns:      maxConns,
		Timeout:       timeout,
		pool:          NewConnectionPool(maxConns),
		state:         "stopped",
		shutdownCh:    make(chan struct{}),
		logger:        logger,
	}
}

// Start initializes and starts the proxy server
func (s *ProxyServer) Start() error {
	s.mu.Lock()
	if s.state != "stopped" {
		s.mu.Unlock()
		return fmt.Errorf("server is not stopped")
	}
	s.state = "starting"
	s.mu.Unlock()

	s.logger.Info("Starting GoFakeIP proxy server on %s with bind IP %s", s.ListenAddress, s.BindIP.String())

	// Start HTTP proxy if enabled
	if s.hasProtocol("http") {
		httpProxy, err := NewHTTPProxy(s.ListenAddress, s.BindIP, s.Timeout, s.logger)
		if err != nil {
			s.setState("stopped")
			return fmt.Errorf("failed to create HTTP proxy: %w", err)
		}
		s.httpServer = httpProxy
		if err := s.httpServer.Start(); err != nil {
			s.setState("stopped")
			return fmt.Errorf("failed to start HTTP proxy: %w", err)
		}
		s.logger.Info("HTTP proxy started")
	}

	// Start SOCKS5 proxy if enabled
	if s.hasProtocol("socks5") {
		socks5Proxy, err := NewSocks5Proxy(s.BindIP, s.ListenAddress, s.Timeout, s.logger)
		if err != nil {
			s.setState("stopped")
			return fmt.Errorf("failed to create SOCKS5 proxy: %w", err)
		}
		s.socks5Server = socks5Proxy
		if err := s.socks5Server.Start(); err != nil {
			s.setState("stopped")
			return fmt.Errorf("failed to start SOCKS5 proxy: %w", err)
		}
		s.logger.Info("SOCKS5 proxy started")
	}

	s.setState("running")
	s.logger.Info("Proxy server started successfully")
	return nil
}

// Stop gracefully shuts down the proxy server
func (s *ProxyServer) Stop() error {
	s.mu.Lock()
	if s.state != "running" {
		s.mu.Unlock()
		return nil
	}
	s.state = "stopping"
	s.mu.Unlock()

	s.logger.Info("Stopping proxy server...")

	// Signal shutdown
	close(s.shutdownCh)

	// Stop HTTP server
	if s.httpServer != nil {
		ctx := context.Background()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server shutdown error: %v", err)
		}
	}

	// Stop SOCKS5 server
	if s.socks5Server != nil {
		if err := s.socks5Server.Stop(); err != nil {
			s.logger.Error("SOCKS5 server shutdown error: %v", err)
		}
	}

	s.setState("stopped")
	s.logger.Info("Proxy server stopped")
	return nil
}

// GetState returns the current server state
func (s *ProxyServer) GetState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// setState updates the server state thread-safely
func (s *ProxyServer) setState(state string) {
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()
}

// hasProtocol checks if a protocol is enabled
func (s *ProxyServer) hasProtocol(protocol string) bool {
	for _, p := range s.Protocols {
		if p == protocol {
			return true
		}
	}
	return false
}

// GetConnectionCount returns the current number of active connections
func (s *ProxyServer) GetConnectionCount() int {
	return s.pool.Count()
}

// WaitForShutdown blocks until the server receives a shutdown signal
func (s *ProxyServer) WaitForShutdown() {
	<-s.shutdownCh
}
