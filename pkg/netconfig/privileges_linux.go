// +build linux

package netconfig

// Stub for Linux privilege check
func checkWindowsAdmin() error {
	return nil // No-op for Linux
}

// CheckPrivileges verifies root privileges (stub for Linux)
func CheckPrivileges() error {
	return nil // No-op for Linux stub
}
