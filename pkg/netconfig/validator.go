//go:build windows || linux || darwin

// Package netconfig provides input validation and sanitization
package netconfig

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/yourorg/gofakeip/pkg/common"
)

// ValidateIP validates an IP address string
func ValidateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("%w: %s", common.ErrInvalidIP, ip)
	}
	return nil
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w: %d (must be 1-65535)", common.ErrInvalidPort, port)
	}
	return nil
}

// ValidateAddress validates a host:port address
func ValidateAddress(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("%w: %s (%v)", common.ErrInvalidAddress, addr, err)
	}

	// Validate host (IP or hostname)
	if net.ParseIP(host) == nil {
		// If not IP, validate as hostname
		if !isValidHostname(host) {
			return fmt.Errorf("%w: invalid hostname %s", common.ErrInvalidAddress, host)
		}
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("%w: invalid port %s", common.ErrInvalidAddress, portStr)
	}

	return ValidatePort(port)
}

// ValidateCommandArg ensures no shell injection characters (NON-NEGOTIABLE #4)
func ValidateCommandArg(arg string) error {
	// Check for null byte injection first
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("%w: null byte detected in argument", common.ErrCommandInjection)
	}

	// Blacklist approach for shell metacharacters
	forbidden := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r", "\\"}
	for _, char := range forbidden {
		if strings.Contains(arg, char) {
			return fmt.Errorf("%w: forbidden character '%s' in argument: %s",
				common.ErrCommandInjection, char, arg)
		}
	}

	// Length limit
	if len(arg) > 256 {
		return fmt.Errorf("%w: argument too long (max 256 characters)", common.ErrCommandInjection)
	}

	return nil
}

// isValidHostname performs basic hostname validation
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Basic check: contains valid characters
	// Full RFC 1123 validation would be more complex
	for _, char := range hostname {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '.') {
			return false
		}
	}

	return true
}

// ValidateProxyAddress validates a proxy server address
func ValidateProxyAddress(addr string) error {
	// First validate format
	if err := ValidateAddress(addr); err != nil {
		return err
	}

	// Then check for injection attempts
	if err := ValidateCommandArg(addr); err != nil {
		return err
	}

	return nil
}

// ValidatePorts validates a list of port numbers
func ValidatePorts(ports []int) error {
	if len(ports) == 0 {
		return nil // Empty list means all ports
	}

	for _, port := range ports {
		if err := ValidatePort(port); err != nil {
			return err
		}
	}

	return nil
}
