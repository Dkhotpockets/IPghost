// Package netconfig provides privilege checking (NON-NEGOTIABLE #1)
package netconfig

// CheckPrivileges verifies elevated privileges before system modifications
// This MUST be called before any privileged operation (NON-NEGOTIABLE #1)
// Platform-specific implementations in privileges_*.go files
