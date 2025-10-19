// Package config provides configuration management for GoFakeIP
package config

import (
	"fmt"
	"os"

	"github.com/yourorg/gofakeip/pkg/common"
	"github.com/yourorg/gofakeip/pkg/netconfig"
	"gopkg.in/yaml.v3"
)

// ClientConfig represents the complete client configuration
type ClientConfig struct {
	Proxy       ProxySettings       `yaml:"proxy"`
	Redirection RedirectionSettings `yaml:"redirection"`
	Platform    PlatformSettings    `yaml:"platform"`
}

// ProxySettings contains proxy server connection details
type ProxySettings struct {
	Address string `yaml:"address"` // host:port
}

// RedirectionSettings contains traffic redirection configuration
type RedirectionSettings struct {
	Ports     []int  `yaml:"ports"`     // specific ports, empty = all
	Protocol  string `yaml:"protocol"`  // tcp, udp, both
	Interface string `yaml:"interface"` // network interface (platform-specific)
}

// PlatformSettings contains platform overrides (for testing)
type PlatformSettings struct {
	Override string `yaml:"override"` // linux, windows, darwin
}

// LoadClientConfig loads client configuration from file and applies defaults
func LoadClientConfig(configPath string) (*ClientConfig, error) {
	config := &ClientConfig{
		// Defaults
		Redirection: RedirectionSettings{
			Ports:     []int{},         // empty = all ports
			Protocol:  "tcp",
			Interface: "",              // platform-specific default
		},
		Platform: PlatformSettings{
			Override: "",               // auto-detect
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
	}

	// Validate configuration
	if err := validateClientConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validateClientConfig validates the client configuration
func validateClientConfig(config *ClientConfig) error {
	// Validate proxy address
	if config.Proxy.Address == "" {
		return fmt.Errorf("%w: proxy.address is required", common.ErrInvalidConfig)
	}

	// Validate proxy address format and check for injection
	if err := netconfig.ValidateProxyAddress(config.Proxy.Address); err != nil {
		return fmt.Errorf("invalid proxy address: %w", err)
	}

	// Validate ports
	if err := netconfig.ValidatePorts(config.Redirection.Ports); err != nil {
		return fmt.Errorf("invalid ports: %w", err)
	}

	// Validate protocol
	validProtocols := map[string]bool{"tcp": true, "udp": true, "both": true}
	if !validProtocols[config.Redirection.Protocol] {
		return fmt.Errorf("%w: invalid protocol: %s (must be 'tcp', 'udp', or 'both')", 
			common.ErrInvalidConfig, config.Redirection.Protocol)
	}

	// Validate platform override if specified
	if config.Platform.Override != "" {
		validPlatforms := map[string]bool{"linux": true, "windows": true, "darwin": true}
		if !validPlatforms[config.Platform.Override] {
			return fmt.Errorf("%w: invalid platform override: %s", 
				common.ErrPlatformUnsupported, config.Platform.Override)
		}
	}

	return nil
}

// MergeWithFlags merges command-line flags into config (flags override config file)
func (c *ClientConfig) MergeWithFlags(
	proxyAddr string,
	ports []int,
	protocol string,
	iface string,
) error {
	// Override with flags if provided
	if proxyAddr != "" {
		c.Proxy.Address = proxyAddr
	}
	if len(ports) > 0 {
		c.Redirection.Ports = ports
	}
	if protocol != "" {
		c.Redirection.Protocol = protocol
	}
	if iface != "" {
		c.Redirection.Interface = iface
	}

	// Re-validate after merging
	return validateClientConfig(c)
}

// GetPlatform returns the platform to use (override or detected)
func (c *ClientConfig) GetPlatform() netconfig.Platform {
	if c.Platform.Override != "" {
		return netconfig.Platform(c.Platform.Override)
	}
	return netconfig.DetectPlatform()
}

// ToNetconfigClientConfig converts to netconfig.ClientConfig
func (c *ClientConfig) ToNetconfigClientConfig(force, dryRun bool) *netconfig.ClientConfig {
	return &netconfig.ClientConfig{
		ProxyAddress: c.Proxy.Address,
		Ports:        c.Redirection.Ports,
		Protocol:     c.Redirection.Protocol,
		Interface:    c.Redirection.Interface,
		Force:        force,
		DryRun:       dryRun,
	}
}
