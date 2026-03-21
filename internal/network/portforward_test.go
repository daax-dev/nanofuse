package network

import (
	"testing"

	"github.com/jpoley/nanofuse/internal/types"
)

func TestValidatePortForward(t *testing.T) {
	tests := []struct {
		name    string
		pf      types.PortForward
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid TCP port forward",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   80,
				Protocol: "tcp",
			},
			wantErr: false,
		},
		{
			name: "valid UDP port forward",
			pf: types.PortForward{
				HostPort: 53,
				VMPort:   53,
				Protocol: "udp",
			},
			wantErr: false,
		},
		{
			name: "valid port 1",
			pf: types.PortForward{
				HostPort: 1,
				VMPort:   1,
				Protocol: "tcp",
			},
			wantErr: false,
		},
		{
			name: "valid port 65535",
			pf: types.PortForward{
				HostPort: 65535,
				VMPort:   65535,
				Protocol: "tcp",
			},
			wantErr: false,
		},
		{
			name: "invalid host port - too low",
			pf: types.PortForward{
				HostPort: 0,
				VMPort:   80,
				Protocol: "tcp",
			},
			wantErr: true,
			errMsg:  "invalid host port 0",
		},
		{
			name: "invalid host port - too high",
			pf: types.PortForward{
				HostPort: 65536,
				VMPort:   80,
				Protocol: "tcp",
			},
			wantErr: true,
			errMsg:  "invalid host port 65536",
		},
		{
			name: "invalid host port - negative",
			pf: types.PortForward{
				HostPort: -1,
				VMPort:   80,
				Protocol: "tcp",
			},
			wantErr: true,
			errMsg:  "invalid host port -1",
		},
		{
			name: "invalid VM port - too low",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   0,
				Protocol: "tcp",
			},
			wantErr: true,
			errMsg:  "invalid VM port 0",
		},
		{
			name: "invalid VM port - too high",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   70000,
				Protocol: "tcp",
			},
			wantErr: true,
			errMsg:  "invalid VM port 70000",
		},
		{
			name: "invalid protocol - empty",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   80,
				Protocol: "",
			},
			wantErr: true,
			errMsg:  "invalid protocol",
		},
		{
			name: "invalid protocol - wrong value",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   80,
				Protocol: "sctp",
			},
			wantErr: true,
			errMsg:  "invalid protocol",
		},
		{
			name: "invalid protocol - case sensitive",
			pf: types.PortForward{
				HostPort: 8080,
				VMPort:   80,
				Protocol: "TCP",
			},
			wantErr: true,
			errMsg:  "invalid protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortForward(tt.pf)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePortForward() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePortForward() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePortForward() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidatePortForwards_MultipleErrors(t *testing.T) {
	portForwards := []types.PortForward{
		{HostPort: 8080, VMPort: 80, Protocol: "tcp"},    // Valid
		{HostPort: 0, VMPort: 80, Protocol: "tcp"},       // Invalid host port
		{HostPort: 8443, VMPort: 443, Protocol: "tcp"},   // Valid
		{HostPort: 9000, VMPort: 70000, Protocol: "tcp"}, // Invalid VM port
		{HostPort: 53, VMPort: 53, Protocol: "invalid"},  // Invalid protocol
	}

	var errors []error
	for _, pf := range portForwards {
		if err := ValidatePortForward(pf); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(errors))
		for i, err := range errors {
			t.Logf("Error %d: %v", i, err)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr ||
			(len(s) > len(substr) && contains(s[1:], substr)))))
}
