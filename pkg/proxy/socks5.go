package proxy

import (
	"context"
	"net"
	"github.com/things-go/go-socks5"
)

// NewSocks5Server starts a SOCKS5 proxy server with IP masking.
// Returns the listener and a stop function.
func NewSocks5Server(bindIP net.IP, listenAddr string) (net.Listener, func(), error) {
       server := socks5.NewServer(
	       socks5.WithDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
		       return DialWithMaskedIP(network, addr, bindIP)
	       }),
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
