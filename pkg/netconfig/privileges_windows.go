//go:build windows
// +build windows

package netconfig

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/yourorg/gofakeip/pkg/common"
)

// TOKEN_ELEVATION structure from Windows API
type tokenElevation struct {
	TokenIsElevated uint32
}

// checkWindowsAdmin verifies Administrator privileges on Windows
// Uses Windows API to check if the process has elevated privileges
func CheckPrivileges() error {
	var token windows.Token
	process := windows.CurrentProcess()

	// Open process token
	err := windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token)
	if err != nil {
		return err
	}
	defer token.Close()

	// Get token elevation status
	var elevation tokenElevation
	var returnedLen uint32

	err = windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&returnedLen,
	)
	if err != nil {
		return err
	}

	// Check if token is elevated
	if elevation.TokenIsElevated == 0 {
		return common.ErrInsufficientPrivs
	}

	return nil
}
