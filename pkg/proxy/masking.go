package proxy

import (
	"net"
	"time"
)

// DialWithMaskedIP creates a net.Conn with the given fake/masked source IP
func DialWithMaskedIP(network, address string, maskedIP net.IP) (net.Conn, error) {
	localAddr := &net.TCPAddr{IP: maskedIP}
	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   10 * time.Second,
	}
	return dialer.Dial(network, address)
}
