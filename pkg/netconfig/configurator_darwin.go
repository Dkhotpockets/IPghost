//go:build darwin
// +build darwin

package netconfig

import "fmt"

// Stub implementations for non-macOS platforms when building on macOS

func newWindowsConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("Windows configurator not available on macOS")
}

func newLinuxConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("Linux configurator not yet implemented")
}

// newDarwinConfigurator creates a new Darwin network configurator (stub)
// TODO: Implement this function
func newDarwinConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("macOS configurator not yet implemented")
}
