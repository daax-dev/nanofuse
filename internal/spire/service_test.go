package spire

import (
	"context"
	"strings"
	"testing"

	"github.com/daax-dev/nanofuse/internal/config"
)

// TestBuildSPIFFEID_NormalizesTrustDomain verifies that BuildSPIFFEID emits the
// canonical lowercase trust domain so the resulting ID passes the same
// validateSPIFFEID grammar that downstream SVID validation enforces. A mixed-case
// configured trust domain must not produce an ID that later fails validation.
func TestBuildSPIFFEID_NormalizesTrustDomain(t *testing.T) {
	svc := NewService(&config.SPIREConfig{TrustDomain: "  Poley.DEV  ", WorkloadType: "microvm"})
	id := svc.BuildSPIFFEID("engineering", "jpoley", "vm-1")

	if want := "spiffe://poley.dev/g/engineering/u/jpoley/w/microvm/vm-1"; id != want {
		t.Fatalf("BuildSPIFFEID = %q, want %q", id, want)
	}
	if err := validateSPIFFEID(id); err != nil {
		t.Fatalf("built ID must pass validateSPIFFEID, got: %v", err)
	}
}

// TestValidateTrustDomain covers the source-side gate that rejects a configured
// trust domain which would otherwise build an ID failing validateSPIFFEID. A
// case-only difference is normalized and accepted; structurally invalid domains
// are rejected.
func TestValidateTrustDomain(t *testing.T) {
	cases := []struct {
		name    string
		td      string
		wantErr bool
	}{
		{"lowercase valid", "poley.dev", false},
		{"uppercase normalized", "Poley.DEV", false},
		{"whitespace trimmed", "  poley.dev  ", false},
		{"single label", "localhost", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"underscore", "poley_dev.com", true},
		{"leading hyphen label", "-poley.dev", true},
		{"empty label", "poley..dev", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&config.SPIREConfig{TrustDomain: tc.td})
			err := svc.validateTrustDomain()
			if tc.wantErr && err == nil {
				t.Fatalf("validateTrustDomain(%q) = nil, want error", tc.td)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateTrustDomain(%q) = %v, want nil", tc.td, err)
			}
		})
	}
}

// TestCreateVMWorkloadEntry_RejectsInvalidTrustDomain verifies the trust-domain
// gate fires at the source: an enabled service configured with an invalid trust
// domain returns an error before any SPIRE registration (docker exec) is
// attempted.
func TestCreateVMWorkloadEntry_RejectsInvalidTrustDomain(t *testing.T) {
	svc := NewService(&config.SPIREConfig{Enabled: true, TrustDomain: "bad_domain.com"})
	_, err := svc.CreateVMWorkloadEntry(context.Background(), "vm-1", "engineering", "jpoley")
	if err == nil {
		t.Fatal("CreateVMWorkloadEntry must reject an invalid trust domain")
	}
	if !strings.Contains(err.Error(), "trust_domain") {
		t.Fatalf("error must name trust_domain, got: %v", err)
	}
}
