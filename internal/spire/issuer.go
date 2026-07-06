package spire

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"sync"
	"time"
)

// ErrSPIREUnavailable indicates the SPIRE endpoint that issues SVIDs could not
// be reached. A Source MUST wrap this error when the SPIRE agent/Workload API is
// unreachable so the Manager fails safe (refuses to start) rather than falling
// back to static or plaintext credentials.
var ErrSPIREUnavailable = errors.New("SPIRE agent unreachable")

// Source supplies freshly-issued SVIDs for a single workload identity.
//
// Production wires this to the SPIRE Workload API — concretely a go-spiffe/v2
// workloadapi.X509Source dialed over the existing Firecracker vsock proxy
// (internal/firecracker/vsock_proxy.go) from inside the guest. That production
// Source is deferred to runtime because it requires a live SPIRE agent and
// cannot be exercised on a dev box. LocalCASource is the in-process dev/test
// implementation used to validate the issuance/rotation lifecycle and crypto.
type Source interface {
	// FetchSVID returns a freshly-issued, currently-valid SVID. It MUST return
	// an error wrapping ErrSPIREUnavailable when the SPIRE endpoint is
	// unreachable.
	FetchSVID(ctx context.Context) (*SVID, error)
}

// LocalCASource is an in-process certificate authority that mints X.509-SVIDs
// for a fixed SPIFFE ID. It is for development and tests only; it is not a
// substitute for a real SPIRE deployment. Each FetchSVID call mints a fresh
// leaf (new serial, new validity window) so rotation produces a genuinely new,
// independently-verifiable certificate.
type LocalCASource struct {
	spiffeID string
	ttl      time.Duration
	clock    Clock

	caCert *x509.Certificate
	caKey  *ecdsa.PrivateKey

	mu     sync.Mutex
	serial int64
}

// NewLocalCASource builds a local-CA Source for the given SPIFFE ID and SVID
// TTL. The SPIFFE ID is validated; trustDomain is derived from it and used as
// the CA subject. clock may be nil (defaults to the real clock).
func NewLocalCASource(spiffeID string, ttl time.Duration, clock Clock) (*LocalCASource, error) {
	if err := validateSPIFFEID(spiffeID); err != nil {
		return nil, err
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("local CA: ttl must be positive, got %s", ttl)
	}
	if clock == nil {
		clock = RealClock{}
	}
	u, err := url.Parse(spiffeID)
	if err != nil {
		return nil, fmt.Errorf("local CA: parse SPIFFE ID: %w", err)
	}
	trustDomain := u.Host

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("local CA: generate key: %w", err)
	}
	now := clock.Now()
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: fmt.Sprintf("nanofuse-local-ca/%s", trustDomain)},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
		URIs:                  []*url.URL{{Scheme: "spiffe", Host: trustDomain}},
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("local CA: create CA certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, fmt.Errorf("local CA: parse CA certificate: %w", err)
	}
	return &LocalCASource{
		spiffeID: spiffeID,
		ttl:      ttl,
		clock:    clock,
		caCert:   caCert,
		caKey:    caKey,
		serial:   1,
	}, nil
}

// FetchSVID mints a fresh X.509-SVID signed by the local CA.
func (l *LocalCASource) FetchSVID(ctx context.Context) (*SVID, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l.mu.Lock()
	l.serial++
	serial := l.serial
	l.mu.Unlock()

	uri, err := url.Parse(l.spiffeID)
	if err != nil {
		return nil, fmt.Errorf("local CA: parse SPIFFE ID: %w", err)
	}

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("local CA: generate leaf key: %w", err)
	}
	// X.509 stores second-granular validity. Truncate so the SVID's advertised
	// IssuedAt/ExpiresAt match the parsed certificate exactly (no sub-second
	// drift that would make ExpiresAt appear to exceed the leaf NotAfter).
	now := l.clock.Now().Truncate(time.Second)
	notAfter := now.Add(l.ttl)
	leafTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		NotBefore:             now.Add(-time.Second),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		URIs:                  []*url.URL{uri},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, l.caCert, &leafKey.PublicKey, l.caKey)
	if err != nil {
		return nil, fmt.Errorf("local CA: create leaf certificate: %w", err)
	}
	leaf, err := x509.ParseCertificate(leafDER)
	if err != nil {
		return nil, fmt.Errorf("local CA: parse leaf certificate: %w", err)
	}
	return &SVID{
		ID:           l.spiffeID,
		Certificates: []*x509.Certificate{leaf},
		PrivateKey:   leafKey,
		Bundle:       []*x509.Certificate{l.caCert},
		IssuedAt:     now,
		ExpiresAt:    notAfter,
	}, nil
}
