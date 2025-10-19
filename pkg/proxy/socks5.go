package proxy

import (
	"context"
	"net"
	"time"

	"github.com/things-go/go-socks5"

	"github.com/yourorg/gofakeip/pkg/common"
)

type socks5Proxy struct {
	listener net.Listener
	stopFunc func()
	logger   common.Logger
}

func NewSocks5Proxy(bindIP net.IP, listenAddr string, timeout time.Duration, logger common.Logger) (SOCKS5Server, error) {
	listener, stop, err := NewSocks5Server(bindIP, listenAddr, timeout)
	if err != nil {
		return nil, err
	}

	return &socks5Proxy{
		listener: listener,
		stopFunc: stop,
		logger:   logger,
	}, nil
}

func (s *socks5Proxy) Start() error {
	s.logger.Info("Starting SOCKS5 proxy on %s", s.listener.Addr().String())
	// The NewSocks5Server already starts the server in a goroutine
	return nil
}

func (s *socks5Proxy) Stop() error {
	s.logger.Info("Shutting down SOCKS5 proxy...")
	s.stopFunc()
	return nil
}

// NewSocks5Server starts a SOCKS5 proxy server with IP masking.
// Returns the listener and a stop function.
func NewSocks5Server(bindIP net.IP, listenAddr string, timeout time.Duration) (net.Listener, func(), error) {
       server := socks5.NewServer(
	       socks5.WithDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
		       				return DialWithMaskedIP(network, addr, bindIP, timeout)	       }),
	       socks5.WithBindIP(bindIP),
       )
       ln, err := net.Listen("tcp", listenAddr)
       if err != nil {
	       return nil, nil, err
       }
       stop := func() { ln.Close() }
       go server.Serve(ln)
       return ln, stop, nil
}
