//go:build darwin
// +build darwin

// Package netconfig provides macOS-specific network configuration (stub)
package netconfig

import (
	"fmt"
)

// DarwinConfigurator implements NetworkConfigurator for macOS (stub implementation)
type DarwinConfigurator struct {
	config *ClientConfig
}

// newDarwinConfigurator creates a new macOS network configurator
func newDarwinConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return &DarwinConfigurator{
		config: config,
	}, nil
}

// CheckPrivileges verifies root privileges (NON-NEGOTIABLE #1)
func (d *DarwinConfigurator) CheckPrivileges() error {
	return CheckPrivileges()
}

// Setup configures network redirection rules (stub)
func (d *DarwinConfigurator) Setup(config *ClientConfig) error {
	return fmt.Errorf("macOS configurator not yet implemented")
}

// Teardown removes all network redirection rules (stub)
func (d *DarwinConfigurator) Teardown() error {
	return fmt.Errorf("macOS configurator not yet implemented")
}

// Status returns the current configuration status (stub)
func (d *DarwinConfigurator) Status() (*ConfigStatus, error) {
	return &ConfigStatus{
		Platform: PlatformDarwin,
		State:    ConfigStateInactive,
	}, nil
}
