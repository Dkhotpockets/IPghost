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
