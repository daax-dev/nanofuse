package spire

import (
	"context"
	"crypto/x509"
	"strings"
	"testing"
	"time"
)

func mintTestSVID(t *testing.T, id string, clk Clock) *SVID {
	t.Helper()
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	svid, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID: %v", err)
	}
	return svid
}

func TestDocument_RoundTrip(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-rt")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	svid := mintTestSVID(t, id, clk)

	data, err := svid.MarshalDocument()
	if err != nil {
		t.Fatalf("MarshalDocument: %v", err)
	}
	if !strings.Contains(string(data), id) {
		t.Fatal("document JSON must contain the SPIFFE ID")
	}

	parsed, err := ParseDocument(data)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if parsed.ID != svid.ID {
		t.Fatalf("round-trip ID = %q, want %q", parsed.ID, svid.ID)
	}
	if parsed.Certificates[0].SerialNumber.Cmp(svid.Certificates[0].SerialNumber) != 0 {
		t.Fatal("round-trip lost the leaf certificate")
	}
	if !parsed.ExpiresAt.Equal(svid.ExpiresAt.Truncate(time.Second)) {
		t.Fatalf("round-trip ExpiresAt = %s, want %s", parsed.ExpiresAt, svid.ExpiresAt)
	}
	// Parsed SVID still verifies against its bundle.
	if err := parsed.Verify(clk.Now()); err != nil {
		t.Fatalf("parsed SVID must verify: %v", err)
	}
}

func TestSVID_Verify_RejectsExpired(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-exp")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	svid := mintTestSVID(t, id, clk)

	future := svid.ExpiresAt.Add(time.Minute)
	if err := svid.Verify(future); err == nil {
		t.Fatal("expected verification to fail after expiry")
	}
}

func TestSVID_Validate_Failures(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-val")
	base := mintTestSVID(t, id, nil)

	t.Run("empty chain", func(t *testing.T) {
		s := *base
		s.Certificates = nil
		if err := s.Validate(); err == nil {
			t.Fatal("expected error for empty chain")
		}
	})
	t.Run("nil leaf entry in non-empty chain", func(t *testing.T) {
		s := *base
		// A non-empty chain whose first entry is nil must be reported as a nil
		// entry, not a misleading "empty chain".
		s.Certificates = []*x509.Certificate{nil}
		err := s.Validate()
		if err == nil {
			t.Fatal("expected error for nil certificate entry")
		}
		if !strings.Contains(err.Error(), "entry 0 is nil") {
			t.Fatalf("expected nil-entry error, got %v", err)
		}
	})
	t.Run("missing bundle", func(t *testing.T) {
		s := *base
		s.Bundle = nil
		if err := s.Validate(); err == nil {
			t.Fatal("expected error for empty bundle")
		}
	})
	t.Run("id mismatch", func(t *testing.T) {
		s := *base
		s.ID = testSPIFFEID("engineering", "jpoley", "different-vm")
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when ID does not match leaf URI SAN")
		}
	})
	t.Run("expiry before issue", func(t *testing.T) {
		s := *base
		s.ExpiresAt = s.IssuedAt.Add(-time.Hour)
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when expires_at precedes issued_at")
		}
	})
}

func TestValidateSPIFFEID(t *testing.T) {
	valid := []string{
		"spiffe://poley.dev/g/eng/u/jp/w/microvm/vm-1",
		"spiffe://example.org/workload",
		"spiffe://my-domain.example.co/workload", // interior hyphen
		"spiffe://a.b.c/workload",                // single-char labels
	}
	for _, id := range valid {
		if err := validateSPIFFEID(id); err != nil {
			t.Errorf("validateSPIFFEID(%q) = %v, want nil", id, err)
		}
	}
	invalid := []string{
		"",
		"https://poley.dev/x",       // wrong scheme
		"spiffe://",                 // no trust domain
		"spiffe://poley.dev",        // no workload path
		"spiffe://poley.dev/",       // empty path
		"spiffe://poley.dev/x?a=b",  // query
		"spiffe://poley.dev/x#f",    // fragment
		"spiffe://poley.dev:8443/x", // port
		"spiffe://Poley.dev/x",      // uppercase trust domain
		"spiffe://poley.dev/a//b",   // doubled slash (empty segment)
		"spiffe://poley.dev/a/",     // trailing slash
		"spiffe://poley.dev/../x",   // dot-dot segment
		"spiffe://poley.dev/%2e/x",  // percent-encoding
		"spiffe://poley.dev/a b",    // space in segment
		"spiffe://my_domain.org/x",  // underscore in trust domain (not a DNS name)
		"spiffe://-poley.dev/x",     // label starts with a hyphen
		"spiffe://poley-.dev/x",     // label ends with a hyphen
		"spiffe://poley..dev/x",     // empty label (doubled dot)
	}
	for _, id := range invalid {
		if err := validateSPIFFEID(id); err == nil {
			t.Errorf("validateSPIFFEID(%q) = nil, want error", id)
		}
	}
}
