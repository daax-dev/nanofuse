package spire

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestSVIDDocPath(t *testing.T) {
	if got := SVIDDocPath(); got != "/var/run/secrets/spiffe/svid.json" {
		t.Fatalf("SVIDDocPath() = %q", got)
	}
}

func TestRealClock(t *testing.T) {
	c := RealClock{}
	before := time.Now()
	if c.Now().Before(before.Add(-time.Second)) {
		t.Fatal("RealClock.Now returned an implausible time")
	}
	select {
	case <-c.After(time.Millisecond):
	case <-time.After(time.Second):
		t.Fatal("RealClock.After did not fire")
	}
}

func TestParseDocument_Errors(t *testing.T) {
	cases := map[string]string{
		"not json":      "{not json",
		"bad svid pem":  `{"spiffe_id":"spiffe://x/y","x509_svid":"nope","x509_svid_key":"","bundle":"","issued_at":"","expires_at":""}`,
		"empty fields":  `{}`,
		"bad issued_at": `{"spiffe_id":"spiffe://x/y","x509_svid":"-----BEGIN CERTIFICATE-----\nbad\n-----END CERTIFICATE-----","x509_svid_key":"","bundle":"","issued_at":"nope","expires_at":""}`,
	}
	for name, in := range cases {
		if _, err := ParseDocument([]byte(in)); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestDecodeHelpers_Errors(t *testing.T) {
	if _, err := decodeCertsPEM([]byte("garbage")); err == nil {
		t.Fatal("decodeCertsPEM should reject non-PEM input")
	}
	if _, err := decodeKeyPEM([]byte("garbage")); err == nil {
		t.Fatal("decodeKeyPEM should reject non-PEM input")
	}
	if _, err := decodeKeyPEM([]byte("-----BEGIN PRIVATE KEY-----\nbad\n-----END PRIVATE KEY-----")); err == nil {
		t.Fatal("decodeKeyPEM should reject malformed key bytes")
	}
}

func TestSVID_Validate_SecurityChecks(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-sec")

	t.Run("private key mismatch", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		s.PrivateKey = otherKey
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when private key does not match leaf")
		}
	})

	t.Run("expires after leaf NotAfter", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.ExpiresAt = s.Certificates[0].NotAfter.Add(time.Hour)
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when expires_at exceeds leaf NotAfter")
		}
	})

	t.Run("missing server auth EKU", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates[0].ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when leaf lacks serverAuth EKU")
		}
	})

	t.Run("missing digital signature key usage", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates[0].KeyUsage = 0
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when leaf lacks digitalSignature key usage")
		}
	})
}

func TestManager_RejectsInsecureDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics not applicable on Windows")
	}
	id := testSPIFFEID("engineering", "jpoley", "vm-perm")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	dir := filepath.Join(t.TempDir(), "spiffe")
	if err := os.MkdirAll(dir, 0o755); err != nil { // group/other readable
		t.Fatalf("mkdir: %v", err)
	}
	mgr, err := NewManager(ManagerConfig{Source: src, Path: filepath.Join(dir, "svid.json"), Clock: clk})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := mgr.Start(context.Background()); err == nil {
		mgr.Stop()
		t.Fatal("Start must refuse to write into a group/other-accessible directory")
	}
}

func TestManager_RejectsSymlinkedDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	id := testSPIFFEID("engineering", "jpoley", "vm-symlink")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	base := t.TempDir()
	realDir := filepath.Join(base, "real")
	if err := os.MkdirAll(realDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	linkDir := filepath.Join(base, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	mgr, err := NewManager(ManagerConfig{Source: src, Path: filepath.Join(linkDir, "svid.json"), Clock: clk})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := mgr.Start(context.Background()); err == nil {
		mgr.Stop()
		t.Fatal("Start must refuse to write a credential through a symlinked directory")
	}
}

func TestSVID_Validate_RejectsSigningLeafAndNonCABundle(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-leaf")

	t.Run("leaf with certSign usage", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates[0].KeyUsage |= x509.KeyUsageCertSign
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when leaf asserts certSign key usage")
		}
	})

	t.Run("non-CA bundle entry", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		// Smuggle the (non-CA) leaf in as its own trust anchor.
		s.Bundle = []*x509.Certificate{s.Certificates[0]}
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when a trust-bundle entry is not a CA")
		}
	})

	t.Run("missing basic constraints", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates[0].BasicConstraintsValid = false
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when leaf lacks Basic Constraints")
		}
	})

	t.Run("non-critical key usage", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates[0].Extensions = nil // drops the critical KeyUsage extension
		if err := s.Validate(); err == nil {
			t.Fatal("expected error when leaf KeyUsage extension is not critical/present")
		}
	})
}

// staticSource returns a fixed pre-built SVID.
type staticSource struct{ svid *SVID }

func (s staticSource) FetchSVID(_ context.Context) (*SVID, error) { return s.svid, nil }

func TestManager_RejectsStaleAdvertisedWindow(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-stale")
	t0 := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	// Mint a long-lived leaf (10h) so the certificate itself is still valid...
	mintClk := newFakeClock(t0)
	src, err := NewLocalCASource(id, 10*time.Hour, mintClk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	svid, err := src.FetchSVID(context.Background())
	if err != nil {
		t.Fatalf("FetchSVID: %v", err)
	}
	// ...but advertise an expiry already in the past relative to the manager's clock.
	svid.ExpiresAt = t0.Add(time.Minute)
	if err := svid.Validate(); err != nil {
		t.Fatalf("precondition: SVID must pass Validate, got %v", err)
	}

	mgrClk := newFakeClock(t0.Add(time.Hour)) // now is well past the advertised expiry
	mgr, err := NewManager(ManagerConfig{
		Source: staticSource{svid: svid},
		Path:   filepath.Join(t.TempDir(), "svid.json"),
		Clock:  mgrClk,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := mgr.Start(context.Background()); err == nil {
		mgr.Stop()
		t.Fatal("Start must reject an SVID whose advertised window is already expired")
	}
}

// onceThenFailSource succeeds on the first fetch, then reports SPIRE unreachable.
type onceThenFailSource struct {
	mu    sync.Mutex
	used  bool
	inner Source
}

func (s *onceThenFailSource) FetchSVID(ctx context.Context) (*SVID, error) {
	s.mu.Lock()
	first := !s.used
	s.used = true
	s.mu.Unlock()
	if first {
		return s.inner.FetchSVID(ctx)
	}
	return nil, ErrSPIREUnavailable
}

func TestManager_RemovesExpiredCredentialOnPersistentFailure(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-expire")
	t0 := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	clk := newFakeClock(t0)
	base, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	src := &onceThenFailSource{inner: base}
	mgr, path := newTestManager(t, src, clk)

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop()
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("SVID document must exist after Start: %v", statErr)
	}

	// First rotation attempt (before expiry) fails but retains the valid SVID.
	if err := clk.blockUntilWaiters(1, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	clk.Advance(DefaultSVIDTTL - DefaultRefreshBefore) // now = T0+45m, still valid
	if err := clk.blockUntilWaiters(1, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	// Advance past expiry; the next failing rotation must remove the credential.
	clk.Advance(DefaultRefreshBefore + time.Minute) // now = T0+61m, expired

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, statErr := os.Stat(path)
		if os.IsNotExist(statErr) && mgr.Current() == nil {
			return // fail-safe behavior observed
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("expired SVID document must be removed and state cleared when rotation keeps failing")
}

// nilSource returns (nil, nil) — a malformed source contract.
type nilSource struct{}

func (nilSource) FetchSVID(_ context.Context) (*SVID, error) { return nil, nil }

func TestManager_Start_NilSVIDFromSourceFailsClosed(t *testing.T) {
	mgr, path := newTestManager(t, nilSource{}, newFakeClock(time.Now()))
	if err := mgr.Start(context.Background()); err == nil {
		mgr.Stop()
		t.Fatal("Start must fail (not panic) when the source returns a nil SVID")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatal("no credential must be written on a nil-SVID failure")
	}
}

func TestManager_Start_RespectsCanceledContext(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-cancel")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	mgr, path := newTestManager(t, src, clk)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled
	if err := mgr.Start(ctx); err == nil {
		mgr.Stop()
		t.Fatal("Start must not issue a credential when the context is already canceled")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatal("no credential must be written when the context is canceled")
	}
}

func TestSVID_Validate_RejectsNilCertEntries(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-nil")

	t.Run("nil bundle entry", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Bundle = append(s.Bundle, nil)
		if err := s.Validate(); err == nil {
			t.Fatal("expected error for a nil bundle entry")
		}
	})
	t.Run("nil chain entry", func(t *testing.T) {
		s := mintTestSVID(t, id, nil)
		s.Certificates = append(s.Certificates, nil)
		if err := s.Validate(); err == nil {
			t.Fatal("expected error for a nil chain entry")
		}
	})
}

func TestSVID_Validate_AllowsOmittedEKU(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-noeku")
	s := mintTestSVID(t, id, nil)
	// SPIFFE permits an SVID with no EKU extension.
	s.Certificates[0].ExtKeyUsage = nil
	s.Certificates[0].UnknownExtKeyUsage = nil
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate must accept an SVID with no EKU extension: %v", err)
	}
}
