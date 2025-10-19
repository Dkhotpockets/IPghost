
package proxy

import (
	"context"
	"io"
	"net"
	"net/http"
)

// hasPort checks if the host includes a port
func hasPort(host string) bool {
	_, _, err := net.SplitHostPort(host)
	return err == nil
}

// HTTPProxyHandler handles HTTP proxy requests (CONNECT, GET, POST)
func HTTPProxyHandler(bindIP net.IP) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			// Hijack connection
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
		// Handle GET/POST forwarding
		// For HTTP proxy, the request URL should be absolute
		if !r.URL.IsAbs() {
			http.Error(w, "Request URL must be absolute for proxy", http.StatusBadRequest)
			return
		}

		// Create outbound request
		outReq := r.Clone(r.Context())
		outReq.RequestURI = "" // Must clear RequestURI for client requests

		// Remove hop-by-hop headers
		removeHopByHopHeaders(outReq.Header)

		// Ensure the outbound address always has a port
		targetHost := outReq.URL.Host
		if !hasPort(targetHost) {
			if outReq.URL.Scheme == "https" {
				targetHost += ":443"
			} else {
				targetHost += ":80"
			}
		}

		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Always use the correct host:port
				return DialWithMaskedIP(network, targetHost, bindIP)
			},
		}

		resp, err := transport.RoundTrip(outReq)
		if err != nil {
			http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
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
