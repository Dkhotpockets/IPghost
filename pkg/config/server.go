// Package config provides configuration management for GoFakeIP
package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/yourorg/gofakeip/pkg/common"
	"gopkg.in/yaml.v3"
)

// ServerConfig represents the complete server configuration
type ServerConfig struct {
	Server      ServerSettings     `yaml:"server"`
	Auth        AuthSettings       `yaml:"auth"`
	HTTPHeaders HTTPHeaderSettings `yaml:"http_headers"`
	Logging     LoggingSettings    `yaml:"logging"`
}

// ServerSettings contains server-specific settings
type ServerSettings struct {
	ListenAddress     string        `yaml:"listen_address"`
	BindIP            string        `yaml:"bind_ip"`
	Protocols         []string      `yaml:"protocols"`
	MaxConnections    int           `yaml:"max_connections"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

// AuthSettings contains authentication configuration
type AuthSettings struct {
	Enabled      bool   `yaml:"enabled"`
	Method       string `yaml:"method"`
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"` // bcrypt only, NEVER plain text
}

// HTTPHeaderSettings contains HTTP header injection configuration
type HTTPHeaderSettings struct {
	InjectXForwardedFor string `yaml:"inject_x_forwarded_for"`
	InjectXRealIP       string `yaml:"inject_x_real_ip"`
	InjectVia           string `yaml:"inject_via"`
	RemoveExisting      bool   `yaml:"remove_existing"`
}

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // text, json
	File   string `yaml:"file"`   // empty = stdout
}

// LoadServerConfig loads server configuration from file and applies defaults
func LoadServerConfig(configPath string) (*ServerConfig, error) {
	config := &ServerConfig{
		// Defaults
		Server: ServerSettings{
			Protocols:         []string{"socks5", "http"},
			MaxConnections:    1000,
			ConnectionTimeout: 30 * time.Second,
		},
		Auth: AuthSettings{
			Enabled: false,
			Method:  "none",
		},
		HTTPHeaders: HTTPHeaderSettings{
			RemoveExisting: false,
		},
		Logging: LoggingSettings{
			Level:  "info",
			Format: "text",
			File:   "", // stdout
		},
	}

	// If config file provided, load it
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		// Validate configuration from file
		if err := validateServerConfig(config); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return config, nil
}

// ValidateServerConfig validates the server configuration (called after CLI flag overrides)
func ValidateServerConfig(config *ServerConfig) error {
	return validateServerConfig(config)
}

// validateServerConfig validates the server configuration
func validateServerConfig(config *ServerConfig) error {
	// Validate listen address
	if config.Server.ListenAddress == "" {
		return fmt.Errorf("%w: listen_address is required", common.ErrInvalidConfig)
	}
	if _, _, err := net.SplitHostPort(config.Server.ListenAddress); err != nil {
		return fmt.Errorf("%w: invalid listen_address format: %v", common.ErrInvalidAddress, err)
	}

	// Validate bind IP
	if config.Server.BindIP == "" {
		return fmt.Errorf("%w: bind_ip is required", common.ErrInvalidConfig)
	}
	if net.ParseIP(config.Server.BindIP) == nil {
		return fmt.Errorf("%w: invalid bind_ip: %s", common.ErrInvalidIP, config.Server.BindIP)
	}

	// CRITICAL: Validate bind IP is assigned to an interface
	if !isIPAssignedToInterface(config.Server.BindIP) {
		return fmt.Errorf("%w: %s", common.ErrBindIPNotFound, config.Server.BindIP)
	}

	// Validate protocols
	if len(config.Server.Protocols) == 0 {
		return fmt.Errorf("%w: at least one protocol must be enabled", common.ErrInvalidConfig)
	}
	for _, proto := range config.Server.Protocols {
		if proto != "socks5" && proto != "http" {
			return fmt.Errorf("%w: invalid protocol: %s (must be 'socks5' or 'http')", common.ErrInvalidConfig, proto)
		}
	}

	// Validate max connections
	if config.Server.MaxConnections <= 0 {
		return fmt.Errorf("%w: max_connections must be > 0", common.ErrInvalidConfig)
	}

	// Validate connection timeout
	if config.Server.ConnectionTimeout <= 0 {
		return fmt.Errorf("%w: connection_timeout must be > 0", common.ErrInvalidConfig)
	}

	// Validate auth settings
	if config.Auth.Enabled {
		if config.Auth.Method == "" || config.Auth.Method == "none" {
			return fmt.Errorf("%w: auth.method must be set when auth.enabled is true", common.ErrInvalidConfig)
		}
		if config.Auth.Username == "" {
			return fmt.Errorf("%w: auth.username is required when auth is enabled", common.ErrInvalidConfig)
		}
		if config.Auth.PasswordHash == "" {
			return fmt.Errorf("%w: auth.password_hash is required when auth is enabled", common.ErrInvalidConfig)
		}

		// Validate bcrypt hash format (NON-NEGOTIABLE #4)
		if !strings.HasPrefix(config.Auth.PasswordHash, "$2a$") &&
			!strings.HasPrefix(config.Auth.PasswordHash, "$2b$") &&
			!strings.HasPrefix(config.Auth.PasswordHash, "$2y$") {
			return fmt.Errorf("%w: password_hash must be bcrypt format (starts with $2a$, $2b$, or $2y$)", common.ErrPlainTextPassword)
		}
	}

	// Validate logging settings
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("%w: invalid log level: %s", common.ErrInvalidConfig, config.Logging.Level)
	}

	validLogFormats := map[string]bool{"text": true, "json": true}
	if !validLogFormats[config.Logging.Format] {
		return fmt.Errorf("%w: invalid log format: %s", common.ErrInvalidConfig, config.Logging.Format)
	}

	return nil
}

// isIPAssignedToInterface checks if the IP is assigned to a network interface
func isIPAssignedToInterface(ipStr string) bool {
	targetIP := net.ParseIP(ipStr)
	if targetIP == nil {
		return false
	}

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	// Check each interface for the IP
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.Equal(targetIP) {
				return true
			}
		}
	}

	return false
}

// ListAvailableIPs returns a list of IPs assigned to interfaces (for error messages)
func ListAvailableIPs() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil {
				ips = append(ips, fmt.Sprintf("%s (%s)", ip.String(), iface.Name))
			}
		}
	}

	return ips
}
