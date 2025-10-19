//go:build linux
// +build linux

// Package netconfig provides Linux-specific network configuration (stub)
package netconfig

import (
	"fmt"
)

// LinuxConfigurator implements NetworkConfigurator for Linux (stub implementation)
type LinuxConfigurator struct {
	config *ClientConfig
}

// newLinuxConfigurator creates a new Linux network configurator
func newLinuxConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return &LinuxConfigurator{
		config: config,
	}, nil
}

// CheckPrivileges verifies root privileges (NON-NEGOTIABLE #1)
func (l *LinuxConfigurator) CheckPrivileges() error {
	return CheckPrivileges()
}

// Setup configures network redirection rules (stub)
func (l *LinuxConfigurator) Setup(config *ClientConfig) error {
	return fmt.Errorf("Linux configurator not yet implemented")
}

// Teardown removes all network redirection rules (stub)
func (l *LinuxConfigurator) Teardown() error {
	return fmt.Errorf("Linux configurator not yet implemented")
}

// Status returns the current configuration status (stub)
func (l *LinuxConfigurator) Status() (*ConfigStatus, error) {
	return &ConfigStatus{
		Platform: PlatformLinux,
		State:    ConfigStateInactive,
	}, nil
}
