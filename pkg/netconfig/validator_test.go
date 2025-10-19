package netconfig

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip      string
		expectErr bool
	}{
		{"192.168.1.1", false},
		{"255.255.255.255", false},
		{"0.0.0.0", false},
		{"::1", false},
		{"invalid-ip", true},
		{"192.168.1.256", true},
		{"192.168.1.", true},
	}

	for _, tt := range tests {
		err := ValidateIP(tt.ip)
		if (err != nil) != tt.expectErr {
			t.Errorf("ValidateIP(%q) error = %v, expectErr %v", tt.ip, err, tt.expectErr)
		}
	}
}
