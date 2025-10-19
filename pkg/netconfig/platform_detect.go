// Package netconfig provides cross-platform network configuration management
package netconfig

import "runtime"

// detectCurrentPlatform detects the current operating system platform
func detectCurrentPlatform() Platform {
	switch runtime.GOOS {
	case "linux":
		return PlatformLinux
	case "windows":
		return PlatformWindows
	case "darwin":
		return PlatformDarwin
	default:
		return ""
	}
}
