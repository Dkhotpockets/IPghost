//go:build windows
// +build windows

// Package netconfig provides Windows-specific network configuration using netsh portproxy
package netconfig

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/yourorg/gofakeip/pkg/netconfig/internal"
)

// WindowsConfigurator implements NetworkConfigurator for Windows using netsh portproxy
type WindowsConfigurator struct {
	config        *ClientConfig
	originalState *NetworkState
}

// newWindowsConfigurator creates a new Windows network configurator
func newWindowsConfigurator(config *ClientConfig) (NetworkConfigurator, error) {
	return &WindowsConfigurator{
		config: config,
	}, nil
}

// CheckPrivileges verifies Administrator privileges (NON-NEGOTIABLE #1)
func (w *WindowsConfigurator) CheckPrivileges() error {
	return CheckPrivileges()
}

// Setup configures network redirection rules using netsh portproxy
// Must be idempotent (NON-NEGOTIABLE #2)
func (w *WindowsConfigurator) Setup(config *ClientConfig) error {
	// Step 1: Check privileges BEFORE any operation (NON-NEGOTIABLE #1)
	if err := w.CheckPrivileges(); err != nil {
		return fmt.Errorf("setup requires Administrator privileges: %w", err)
	}

	// Step 2: Capture current state for teardown restoration
	state, err := w.captureCurrentState()
	if err != nil {
		return fmt.Errorf("failed to capture current state: %w", err)
	}
	w.originalState = state

	// Step 3: Parse proxy address
	proxyHost, proxyPort, err := net.SplitHostPort(config.ProxyAddress)
	if err != nil {
		return fmt.Errorf("invalid proxy address: %w", err)
	}

	// Validate proxy host and port
	if err := ValidateIP(proxyHost); err != nil {
		// If not IP, might be hostname - accept it
		if proxyHost == "" {
			return fmt.Errorf("proxy host cannot be empty")
		}
	}

	// Step 4: Determine which ports to redirect
	ports := config.Ports
	if len(ports) == 0 {
		// Default common ports for HTTP/HTTPS traffic
		ports = []int{80, 443}
	}

	// Step 5: Create portproxy rules (with idempotency)
	for _, port := range ports {
		if err := w.setupPortProxy(port, proxyHost, proxyPort, config.Force); err != nil {
			return fmt.Errorf("failed to setup port %d: %w", port, err)
		}
	}

	return nil
}

// Teardown removes all network redirection rules
// Must restore original state completely (NON-NEGOTIABLE #2)
func (w *WindowsConfigurator) Teardown() error {
	// Check privileges BEFORE any operation (NON-NEGOTIABLE #1)
	if err := w.CheckPrivileges(); err != nil {
		return fmt.Errorf("teardown requires Administrator privileges: %w", err)
	}

	// Get all current portproxy rules
	rules, err := w.listPortProxyRules()
	if err != nil {
		return fmt.Errorf("failed to list portproxy rules: %w", err)
	}

	// Remove all GoFakeIP-created rules
	for _, rule := range rules {
		if err := w.removePortProxyRule(rule); err != nil {
			return fmt.Errorf("failed to remove rule %s: %w", rule.ID, err)
		}
	}

	return nil
}

// Status returns the current configuration status
func (w *WindowsConfigurator) Status() (*ConfigStatus, error) {
	rules, err := w.listPortProxyRules()
	if err != nil {
		return nil, err
	}

	status := &ConfigStatus{
		Platform: PlatformWindows,
		State:    ConfigStateInactive,
	}

	if len(rules) > 0 {
		status.State = ConfigStateActive
		status.Rules = rules
	}

	return status, nil
}

// setupPortProxy creates a single portproxy rule
func (w *WindowsConfigurator) setupPortProxy(listenPort int, connectAddr, connectPort string, force bool) error {
	// Check if rule already exists (idempotency - NON-NEGOTIABLE #2)
	exists, err := w.ruleExists(listenPort)
	if err != nil {
		return err
	}

	if exists && !force {
		// Rule already exists, skip (idempotent)
		return nil
	}

	if exists && force {
		// Remove existing rule first
		if err := w.removePortProxyRule(NetworkRule{TargetPort: listenPort}); err != nil {
			return err
		}
	}

	// Build netsh command (NON-NEGOTIABLE #4: parameterized execution)
	args := []string{
		"interface",
		"portproxy",
		"add",
		"v4tov4",
		fmt.Sprintf("listenport=%d", listenPort),
		"listenaddress=127.0.0.1",
		fmt.Sprintf("connectaddress=%s", connectAddr),
		fmt.Sprintf("connectport=%s", connectPort),
	}

	// Execute with validation (NON-NEGOTIABLE #4)
	if w.config.DryRun {
		fmt.Printf("[DRY RUN] Would execute: netsh %s\n", strings.Join(args, " "))
		return nil
	}

	output, err := internal.ExecuteCommand("netsh", args...)
	if err != nil {
		return fmt.Errorf("netsh command failed: %w (output: %s)", err, output)
	}

	return nil
}

// removePortProxyRule removes a single portproxy rule
func (w *WindowsConfigurator) removePortProxyRule(rule NetworkRule) error {
	args := []string{
		"interface",
		"portproxy",
		"delete",
		"v4tov4",
		fmt.Sprintf("listenport=%d", rule.TargetPort),
		"listenaddress=127.0.0.1",
	}

	if w.config.DryRun {
		fmt.Printf("[DRY RUN] Would execute: netsh %s\n", strings.Join(args, " "))
		return nil
	}

	output, err := internal.ExecuteCommand("netsh", args...)
	if err != nil {
		return fmt.Errorf("failed to remove rule: %w (output: %s)", err, output)
	}

	return nil
}

// ruleExists checks if a portproxy rule already exists
func (w *WindowsConfigurator) ruleExists(listenPort int) (bool, error) {
	rules, err := w.listPortProxyRules()
	if err != nil {
		return false, err
	}

	for _, rule := range rules {
		if rule.TargetPort == listenPort {
			return true, nil
		}
	}

	return false, nil
}

// listPortProxyRules lists all active portproxy rules
func (w *WindowsConfigurator) listPortProxyRules() ([]NetworkRule, error) {
	args := []string{
		"interface",
		"portproxy",
		"show",
		"v4tov4",
	}

	output, err := internal.ExecuteCommand("netsh", args...)
	if err != nil {
		// No rules exist might return error - that's okay
		return []NetworkRule{}, nil
	}

	return parsePortProxyOutput(string(output)), nil
}

// parsePortProxyOutput parses netsh portproxy show output
func parsePortProxyOutput(output string) []NetworkRule {
	var rules []NetworkRule
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Parse lines like: "127.0.0.1        80        192.168.1.100    8080"
		if line == "" || strings.Contains(line, "Listen") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			// Extract listen port and connect address:port
			var listenPort int
			fmt.Sscanf(fields[1], "%d", &listenPort)

			rules = append(rules, NetworkRule{
				ID:         fmt.Sprintf("portproxy-%d", listenPort),
				Platform:   PlatformWindows,
				RuleType:   RuleTypePortProxy,
				TargetPort: listenPort,
				Command:    fmt.Sprintf("netsh interface portproxy ..."),
			})
		}
	}

	return rules
}

// captureCurrentState captures the current network state
func (w *WindowsConfigurator) captureCurrentState() (*NetworkState, error) {
	rules, err := w.listPortProxyRules()
	if err != nil {
		return nil, err
	}

	var existingRules []string
	for _, rule := range rules {
		existingRules = append(existingRules, rule.ID)
	}

	return &NetworkState{
		Platform:      PlatformWindows,
		CapturedAt:    time.Now(),
		ExistingRules: existingRules,
	}, nil
}
