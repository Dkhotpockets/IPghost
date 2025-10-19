// Package internal provides internal utilities for netconfig
package internal

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourorg/gofakeip/pkg/common"
)

// ValidatedCommand creates a safe exec.Cmd with validated inputs
// This ensures no command injection (NON-NEGOTIABLE #4)
func ValidatedCommand(name string, args ...string) (*exec.Cmd, error) {
	// Validate command name (whitelist approach)
	if err := validateCommandName(name); err != nil {
		return nil, err
	}

	// Validate each argument
	for i, arg := range args {
		if err := validateCommandArg(arg); err != nil {
			return nil, fmt.Errorf("invalid argument at position %d: %w", i, err)
		}
	}

	// Use exec.Command with validated inputs (parameterized execution)
	cmd := exec.Command(name, args...)
	return cmd, nil
}

// validateCommandName ensures the command is in the allowed list
func validateCommandName(name string) error {
	// Whitelist of allowed system commands
	allowedCommands := map[string]bool{
		"iptables":   true, // Linux
		"ip6tables":  true, // Linux IPv6
		"netsh":      true, // Windows
		"pfctl":      true, // macOS
		"sysctl":     true, // macOS/Linux
		"route":      true, // Network routing
		"ipconfig":   true, // Windows
		"ifconfig":   true, // Unix-like
		"powershell": true, // Windows (for specific operations)
		"cmd":        true, // Windows (for specific operations)
	}

	if !allowedCommands[name] {
		return fmt.Errorf("%w: command not in whitelist: %s", common.ErrCommandInjection, name)
	}

	return nil
}

// validateCommandArg ensures no shell injection characters in argument
func validateCommandArg(arg string) error {
	// Blacklist approach for shell metacharacters
	forbiddenChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r", "\\"}

	for _, char := range forbiddenChars {
		if strings.Contains(arg, char) {
			return fmt.Errorf("%w: forbidden character '%s' in argument: %s",
				common.ErrCommandInjection, char, arg)
		}
	}

	// Additional checks for command substitution patterns
	if strings.Contains(arg, "$(") || strings.Contains(arg, "${") {
		return fmt.Errorf("%w: command substitution pattern detected", common.ErrCommandInjection)
	}

	// Length limit to prevent buffer overflow
	if len(arg) > 1024 {
		return fmt.Errorf("%w: argument too long (max 1024 characters)", common.ErrCommandInjection)
	}

	return nil
}

// ExecuteCommand runs a validated command and returns output
func ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd, err := ValidatedCommand(name, args...)
	if err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command execution failed: %w (output: %s)", err, string(output))
	}

	return output, nil
}

// ExecuteCommandWithInput runs a validated command with stdin input
func ExecuteCommandWithInput(input string, name string, args ...string) ([]byte, error) {
	cmd, err := ValidatedCommand(name, args...)
	if err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Validate input as well
	if err := validateCommandInput(input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Set stdin
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command execution failed: %w (output: %s)", err, string(output))
	}

	return output, nil
}

// validateCommandInput validates stdin input for dangerous patterns
func validateCommandInput(input string) error {
	// Check for command injection patterns in input
	dangerousPatterns := []string{
		";", "&", "|", "`", "$(",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			return fmt.Errorf("%w: dangerous pattern in input: %s", common.ErrCommandInjection, pattern)
		}
	}

	// Length limit
	if len(input) > 10240 {
		return fmt.Errorf("%w: input too long (max 10240 characters)", common.ErrCommandInjection)
	}

	return nil
}
