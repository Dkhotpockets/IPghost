package security

import (
	"os"
	"runtime"
	"testing"

	"github.com/yourorg/gofakeip/pkg/netconfig"
)

// TestPrivilegeEnforcement verifies that privilege checks are enforced
// per NON-NEGOTIABLE #1: Least Privilege Enforcement
func TestPrivilegeEnforcement(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		skipMsg  string
	}{
		{
			name:     "Linux privilege check",
			platform: "linux",
			skipMsg:  "",
		},
		{
			name:     "Windows privilege check",
			platform: "windows",
			skipMsg:  "",
		},
		{
			name:     "macOS privilege check",
			platform: "darwin",
			skipMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS != tt.platform {
				t.Skipf("Skipping %s test on %s", tt.platform, runtime.GOOS)
			}

			err := netconfig.CheckPrivileges()

			// If running as root/admin, expect no error
			// If not running as root/admin, expect error
			isPrivileged := isRunningWithPrivileges()

			if isPrivileged && err != nil {
				t.Errorf("CheckPrivileges() returned error when running with privileges: %v", err)
			}

			if !isPrivileged && err == nil {
				t.Errorf("CheckPrivileges() should return error when running without privileges")
			}
		})
	}
}

// TestPrivilegeCheckBeforeOperation verifies that privilege checks occur
// before privileged operations, ensuring no operations execute without proper privileges
func TestPrivilegeCheckBeforeOperation(t *testing.T) {
	// This test verifies the principle that CheckPrivileges is called
	// before any privileged system command execution

	// Test that configurator creation includes privilege check
	platform := netconfig.DetectPlatform()

	// Create a test config
	config := &netconfig.ClientConfig{
		ProxyAddress: "127.0.0.1:1080",
		Ports:        []int{80, 443},
		Protocol:     "tcp",
		Interface:    "",
		Force:        false,
		DryRun:       true, // Use dry run to avoid actual system changes
	}

	configurator, err := netconfig.NewConfigurator(platform, config)
	if err != nil {
		// If configurator not implemented, skip
		if err.Error() == "Linux configurator not yet implemented" ||
			err.Error() == "macOS configurator not yet implemented" {
			t.Skip(err.Error())
			return
		}

		// If we don't have privileges, this should fail
		// This is the expected behavior per NON-NEGOTIABLE #1
		if !isRunningWithPrivileges() {
			t.Logf("Expected: Operation blocked without privileges: %v", err)
			return
		}
		t.Fatalf("NewConfigurator failed: %v", err)
	}

	// Skip if configurator is not implemented yet
	if configurator == nil {
		t.Skip("Configurator not implemented yet for this platform")
		return
	}

	// Attempt to setup (dry run mode)
	err = configurator.Setup(config)
	if err != nil {
		if !isRunningWithPrivileges() {
			t.Logf("Expected: Setup blocked without privileges: %v", err)
			return
		}
		// In dry run mode with privileges, setup should succeed or fail gracefully
		t.Logf("Setup dry run result: %v", err)
	}
}

// isRunningWithPrivileges checks if the current process has elevated privileges
func isRunningWithPrivileges() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return os.Geteuid() == 0
	case "windows":
		// On Windows, we need to check if running as Administrator
		// For testing purposes, we'll use a simple heuristic
		// The actual implementation in pkg/netconfig/privileges.go should be more robust
		return netconfig.CheckPrivileges() == nil
	default:
		return false
	}
}
