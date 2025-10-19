package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/yourorg/gofakeip/pkg/common"
)

// hasPort checks if the host includes a port
func hasPort(host string) bool {
	_, _, err := net.SplitHostPort(host)
	return err == nil
}

type httpProxy struct {
	server   *http.Server
	listener net.Listener
	logger   common.Logger
}

func NewHTTPProxy(listenAddr string, bindIP net.IP, timeout time.Duration, logger common.Logger) (HTTPServer, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &httpProxy{
		server: &http.Server{
			Handler: HTTPProxyHandler(bindIP, timeout),
		},
		listener: listener,
		logger:   logger,
	}, nil
}

func (h *httpProxy) Start() error {
	h.logger.Info("Starting HTTP proxy on %s", h.listener.Addr().String())
	go func() {
		if err := h.server.Serve(h.listener); err != nil && err != http.ErrServerClosed {
			h.logger.Error("HTTP server error: %v", err)
		}
	}()
	return nil
}

func (h *httpProxy) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutting down HTTP proxy...")
	return h.server.Shutdown(ctx)
}

// HTTPProxyHandler handles HTTP proxy requests (CONNECT, GET, POST)
func HTTPProxyHandler(bindIP net.IP, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle CONNECT requests (for HTTPS)
		if r.Method == http.MethodConnect {
			// Handle CONNECT (tunneling)
			host := r.Host
			if !hasPort(host) {
				host += ":443"
			}
			conn, err := net.Dial("tcp", host)
			if err != nil {
				http.Error(w, "Failed to connect", http.StatusBadGateway)
				return
			}
			w.WriteHeader(http.StatusOK)
			// Hijack the connection to forward raw TCP data
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
				return
			}
			clientConn, _, err := hj.Hijack()
			if err != nil {
				http.Error(w, "Hijack failed", http.StatusInternalServerError)
				return
			}
			go proxyStream(clientConn, conn)
			go proxyStream(conn, clientConn)
			return
		}

		// Handle GET/POST requests
		// For HTTP proxy, the request URL should be absolute
		if !r.URL.IsAbs() {
			http.Error(w, "Request URL must be absolute for proxy", http.StatusBadRequest)
			return
		}

		// Create a new request with the same method, URL, and body
		outReq := r.Clone(r.Context())
		outReq.RequestURI = "" // Must clear RequestURI for client requests

		// Remove hop-by-hop headers
		removeHopByHopHeaders(outReq.Header)

		// Create a new transport with the masked IP
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return DialWithMaskedIP(network, addr, bindIP, timeout)
			},
		}

		// Send the request and get the response
		resp, err := transport.RoundTrip(outReq)
		if err != nil {
			http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy the response headers and status code to the client
		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}

func proxyStream(dst io.Writer, src io.Reader) {
	io.Copy(dst, src)
}

// Hop-by-hop headers that should not be forwarded
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

// removeHopByHopHeaders removes hop-by-hop headers from the request
func removeHopByHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

// copyHeader copies headers from src to dst
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}