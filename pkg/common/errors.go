// Package common provides shared error definitions and utilities
package common

import "errors"

// Domain-specific errors for GoFakeIP
var (
	// Configuration errors
	ErrInvalidIP          = errors.New("invalid IP address")
	ErrInvalidPort        = errors.New("invalid port number")
	ErrInvalidAddress     = errors.New("invalid network address")
	ErrBindIPNotFound     = errors.New("bind IP not found on any interface")
	ErrConfigFileNotFound = errors.New("configuration file not found")
	ErrInvalidConfig      = errors.New("invalid configuration")

	// Security errors (NON-NEGOTIABLE constraints)
	ErrCommandInjection  = errors.New("potential command injection detected")
	ErrInsufficientPrivs = errors.New("insufficient privileges (root/admin required)")
	ErrPlainTextPassword = errors.New("plain text password not allowed (use bcrypt hash)")

	// Platform errors
	ErrPlatformUnsupported = errors.New("platform not supported")
	ErrPlatformToolMissing = errors.New("required platform tool not found")

	// Network configuration errors
	ErrRuleAlreadyExists = errors.New("network rule already exists")
	ErrRuleNotFound      = errors.New("network rule not found")
	ErrRuleCreationFailed = errors.New("failed to create network rule")
	ErrRuleRemovalFailed  = errors.New("failed to remove network rule")

	// Proxy errors
	ErrConnectionFailed   = errors.New("connection to destination failed")
	ErrAuthFailed         = errors.New("authentication failed")
	ErrServerNotRunning   = errors.New("server not running")
	ErrMaxConnectionsReached = errors.New("maximum connections reached")
	ErrProtocolNotSupported  = errors.New("protocol not supported")
)
