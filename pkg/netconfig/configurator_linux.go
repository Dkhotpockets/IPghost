//go:build linux
// +build linux

package netconfig

import "fmt"

// Stub implementations for non-Linux platforms when building on Linux

func newWindowsConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("Windows configurator not available on Linux")
}

func newDarwinConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return nil, fmt.Errorf("macOS configurator not yet implemented")
}
