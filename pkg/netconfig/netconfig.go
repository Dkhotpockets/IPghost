// Package netconfig provides cross-platform network configuration management
package netconfig

import (
	"fmt"
	"time"

	"github.com/yourorg/gofakeip/pkg/common"
)

// Platform represents the operating system platform
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformDarwin  Platform = "darwin" // macOS
)

// ConfigState represents the configuration state
type ConfigState int

const (
	ConfigStateInactive ConfigState = iota
	ConfigStateActive
)

// NetworkConfigurator defines the interface for platform-specific network configuration
// This interface ensures idempotent setup and reversible teardown (NON-NEGOTIABLE #2)
type NetworkConfigurator interface {
	// Setup configures network redirection rules
	// Must be idempotent (running twice = same state as once)
	Setup(config *ClientConfig) error

	// Teardown removes all network redirection rules
	// Must restore original network state completely
	// Must succeed even when run multiple times
	Teardown() error

	// Status returns the current configuration state
	Status() (*ConfigStatus, error)

	// CheckPrivileges verifies elevated privileges (NON-NEGOTIABLE #1)
	// Must be called BEFORE any system modification
	CheckPrivileges() error
}

// ClientConfig represents client configuration for network redirection
type ClientConfig struct {
	ProxyAddress string   // Proxy server address:port
	Ports        []int    // Specific ports to redirect (empty = all)
	Protocol     string   // tcp, udp, or both
	Interface    string   // Network interface (platform-specific)
	Force        bool     // Force setup even if rules exist
	DryRun       bool     // Show what would be executed
}

// ConfigStatus represents the current configuration status
type ConfigStatus struct {
	Platform      Platform
	State         ConfigState
	ProxyAddress  string
	SetupAt       *time.Time
	Rules         []NetworkRule
	OriginalState *NetworkState
}

// NetworkRule represents a single network redirection rule
type NetworkRule struct {
	ID         string   // Unique rule identifier
	Platform   Platform // Target platform
	RuleType   RuleType // Type of rule
	Protocol   string   // tcp, udp, or all
	TargetPort int      // Port to redirect (0 = all)
	ProxyPort  int      // Destination proxy port
	Command    string   // Actual system command executed
	RawOutput  string   // Command output for idempotency checks
}

// RuleType represents the type of network rule
type RuleType string

const (
	RuleTypeIPTablesRedirect RuleType = "iptables_redirect" // Linux
	RuleTypePortProxy        RuleType = "portproxy"         // Windows
	RuleTypePF               RuleType = "pf"                // macOS
)

// NetworkState represents a snapshot of network configuration
type NetworkState struct {
	Platform        Platform
	CapturedAt      time.Time
	ExistingRules   []string
	InterfaceStates map[string]InterfaceState
}

// InterfaceState represents the state of a network interface
type InterfaceState struct {
	Name      string
	Addresses []string
	IsUp      bool
	MTU       int
}

// NewConfigurator creates a platform-specific NetworkConfigurator
func NewConfigurator(platform Platform, config *ClientConfig) (NetworkConfigurator, error) {
	switch platform {
	case PlatformWindows:
		return newWindowsConfigurator(config)
	case PlatformLinux:
		return newLinuxConfigurator(config)
	case PlatformDarwin:
		return newDarwinConfigurator(config)
	default:
		return nil, fmt.Errorf("%w: %s", common.ErrPlatformUnsupported, platform)
	}
}

// Platform-specific constructors implemented in platform-specific files
// These are lowercase to avoid export issues across build tags

// DetectPlatform detects the current operating system platform
func DetectPlatform() Platform {
	// Implemented in platform_detect.go
	return detectCurrentPlatform()
}
