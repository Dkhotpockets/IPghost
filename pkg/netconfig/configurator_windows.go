//go:build windows
// +build windows

package netconfig

import "fmt"

// Stub implementations for non-Windows platforms when building on Windows

func newLinuxConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("Linux configurator not yet implemented")
}

func newDarwinConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("macOS configurator not yet implemented")
}
