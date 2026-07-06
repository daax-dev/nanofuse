package spire

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLocalCASource_IssuesValidSVID(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-abc123")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}

	svid, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID: %v", err)
	}
	if err := svid.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	// URI SAN must equal the SPIFFE ID.
	leaf := svid.Certificates[0]
	if len(leaf.URIs) != 1 || leaf.URIs[0].String() != id {
		t.Fatalf("leaf URI SAN = %v, want exactly [%s]", leaf.URIs, id)
	}
	if leaf.IsCA {
		t.Fatal("leaf must not be a CA certificate")
	}
	// Signature must verify against the trust bundle.
	if err := svid.Verify(clk.Now()); err != nil {
		t.Fatalf("Verify against bundle: %v", err)
	}
	// TTL window honored.
	if got := svid.ExpiresAt.Sub(svid.IssuedAt); got != DefaultSVIDTTL {
		t.Fatalf("TTL = %s, want %s", got, DefaultSVIDTTL)
	}
}

func TestLocalCASource_PerVMDistinctIdentity(t *testing.T) {
	idA := testSPIFFEID("engineering", "jpoley", "vm-aaa")
	idB := testSPIFFEID("engineering", "jpoley", "vm-bbb")
	if idA == idB {
		t.Fatal("expected distinct SPIFFE IDs per VM")
	}
	srcA, err := NewLocalCASource(idA, DefaultSVIDTTL, nil)
	if err != nil {
		t.Fatalf("NewLocalCASource A: %v", err)
	}
	srcB, err := NewLocalCASource(idB, DefaultSVIDTTL, nil)
	if err != nil {
		t.Fatalf("NewLocalCASource B: %v", err)
	}
	svidA, err := srcA.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID A: %v", err)
	}
	svidB, err := srcB.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID B: %v", err)
	}
	if svidA.ID == svidB.ID {
		t.Fatalf("expected distinct SVID identities, both = %s", svidA.ID)
	}
	if svidA.Certificates[0].URIs[0].String() == svidB.Certificates[0].URIs[0].String() {
		t.Fatal("expected distinct leaf URI SANs per VM")
	}
	// Each SVID verifies against its own bundle but NOT the other's: the two
	// VMs hold cryptographically independent identities.
	if err := svidA.Verify(time.Now()); err != nil {
		t.Fatalf("svidA must verify against its own bundle: %v", err)
	}
	crossA := &SVID{ID: svidA.ID, Certificates: svidA.Certificates, PrivateKey: svidA.PrivateKey, Bundle: svidB.Bundle, IssuedAt: svidA.IssuedAt, ExpiresAt: svidA.ExpiresAt}
	if err := crossA.Verify(time.Now()); err == nil {
		t.Fatal("svidA must NOT verify against svidB's trust bundle (distinct identities)")
	}
}

func TestLocalCASource_RotationProducesFreshCert(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-rot")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	first, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("first FetchSVID: %v", err)
	}
	clk.Advance(DefaultSVIDTTL - DefaultRefreshBefore)
	second, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("second FetchSVID: %v", err)
	}
	if first.Certificates[0].SerialNumber.Cmp(second.Certificates[0].SerialNumber) == 0 {
		t.Fatal("rotation must mint a new certificate (distinct serial)")
	}
	// The new SVID must be valid before the old one expires (overlap window).
	if !second.IssuedAt.Before(first.ExpiresAt) {
		t.Fatal("new SVID must be issued before the old SVID expires")
	}
	if err := second.Verify(clk.Now()); err != nil {
		t.Fatalf("rotated SVID must verify: %v", err)
	}
}

func TestLocalCASource_LongTTLLeafDoesNotOutliveCA(t *testing.T) {
	// A caller requesting ttl > 24h must still receive a currently-valid SVID:
	// the CA validity has to cover the leaf's NotAfter, otherwise Validate/Verify
	// reject the leaf for outliving its issuing CA.
	id := testSPIFFEID("engineering", "jpoley", "vm-longttl")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	const ttl = 72 * time.Hour
	src, err := NewLocalCASource(id, ttl, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	svid, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID: %v", err)
	}
	if got := svid.ExpiresAt.Sub(svid.IssuedAt); got != ttl {
		t.Fatalf("TTL = %s, want %s", got, ttl)
	}
	// The CA (trust bundle root) must not expire before the leaf.
	leaf := svid.Certificates[0]
	ca := svid.Bundle[0]
	if ca.NotAfter.Before(leaf.NotAfter) {
		t.Fatalf("CA NotAfter %s precedes leaf NotAfter %s; leaf outlives its CA",
			ca.NotAfter.Format(time.RFC3339), leaf.NotAfter.Format(time.RFC3339))
	}
	if err := svid.Validate(); err != nil {
		t.Fatalf("Validate with ttl > 24h: %v", err)
	}
	if err := svid.Verify(clk.Now()); err != nil {
		t.Fatalf("Verify with ttl > 24h: %v", err)
	}
}

func TestLocalCASource_LeafNeverOutlivesCA(t *testing.T) {
	// For a long-running Source, a leaf minted late in the CA's life must be
	// clamped so it never outlives the CA (an unclampled leaf would fail Verify).
	id := testSPIFFEID("engineering", "jpoley", "vm-latefetch")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	const ttl = time.Hour
	src, err := NewLocalCASource(id, ttl, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	// CA lifetime is ttl + 24h. Advance to just inside it so a full-ttl leaf
	// would otherwise outlive the CA, forcing the clamp.
	clk.Advance(24*time.Hour + 30*time.Minute)
	svid, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID: %v", err)
	}
	leaf := svid.Certificates[0]
	ca := svid.Bundle[0]
	if leaf.NotAfter.After(ca.NotAfter) {
		t.Fatalf("leaf NotAfter %s outlives CA NotAfter %s",
			leaf.NotAfter.Format(time.RFC3339), ca.NotAfter.Format(time.RFC3339))
	}
	if err := svid.Verify(clk.Now()); err != nil {
		t.Fatalf("clamped SVID must still verify: %v", err)
	}
}

func TestNewLocalCASource_RejectsBadInput(t *testing.T) {
	if _, err := NewLocalCASource("not-a-spiffe-id", DefaultSVIDTTL, nil); err == nil {
		t.Fatal("expected error for non-SPIFFE ID")
	}
	if _, err := NewLocalCASource(testSPIFFEID("g", "u", "w"), 0, nil); err == nil {
		t.Fatal("expected error for non-positive TTL")
	}
}

func TestFailingSource_WrapsErrSPIREUnavailable(t *testing.T) {
	_, err := failingSource{}.FetchSVID(context.Background())
	if !errors.Is(err, ErrSPIREUnavailable) {
		t.Fatalf("expected ErrSPIREUnavailable, got %v", err)
	}
}
