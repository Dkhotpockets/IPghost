// Package internal provides internal utilities for netconfig
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ClientState represents the persistent client configuration state
// Used for teardown restoration (NON-NEGOTIABLE #2: Reversibility)
type ClientState struct {
	Version       string      `json:"version"`
	Platform      string      `json:"platform"`
	ProxyAddress  string      `json:"proxy_address"`
	SetupAt       time.Time   `json:"setup_at"`
	Rules         []RuleState `json:"rules"`
	OriginalState StateSnapshot `json:"original_state"`
}

// RuleState represents a single network rule in the state
type RuleState struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	RuleType string `json:"rule_type"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Command  string `json:"command"`
}

// StateSnapshot captures the original network state before setup
type StateSnapshot struct {
	CapturedAt    time.Time         `json:"captured_at"`
	ExistingRules []string          `json:"existing_rules"`
	Interfaces    map[string]string `json:"interfaces,omitempty"`
}

// GetStateFilePath returns the platform-specific path for state file
func GetStateFilePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "GoFakeIP", "client-state.json")
	case "darwin", "linux":
		return "/var/lib/gofakeip/client-state.json"
	default:
		return "./client-state.json"
	}
}

// SaveState persists the client state to disk
func SaveState(state *ClientState) error {
	statePath := GetStateFilePath()

	// Ensure directory exists
	dir := filepath.Dir(statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(statePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads the client state from disk
func LoadState() (*ClientState, error) {
	statePath := GetStateFilePath()

	// Read file
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file = no active configuration
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var state ClientState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// DeleteState removes the state file (called after successful teardown)
func DeleteState() error {
	statePath := GetStateFilePath()

	if err := os.Remove(statePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted, success
		}
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

// StateExists checks if a state file exists
func StateExists() bool {
	statePath := GetStateFilePath()
	_, err := os.Stat(statePath)
	return err == nil
}

// NewClientState creates a new ClientState instance
func NewClientState(platform, proxyAddress string) *ClientState {
	return &ClientState{
		Version:      "1.0",
		Platform:     platform,
		ProxyAddress: proxyAddress,
		SetupAt:      time.Now(),
		Rules:        []RuleState{},
		OriginalState: StateSnapshot{
			CapturedAt:    time.Now(),
			ExistingRules: []string{},
			Interfaces:    make(map[string]string),
		},
	}
}
