package proxy

import (
	"net"
	"time"
)

// DialWithMaskedIP creates a net.Conn with the given fake/masked source IP
func DialWithMaskedIP(network, address string, maskedIP net.IP, timeout time.Duration) (net.Conn, error) {
	localAddr := &net.TCPAddr{IP: maskedIP}
	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   timeout,
	}
	return dialer.Dial(network, address)
}
