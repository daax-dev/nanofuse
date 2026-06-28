package spire

import (
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// oidKeyUsage is the X.509 Key Usage extension OID (2.5.29.15).
var oidKeyUsage = asn1.ObjectIdentifier{2, 5, 29, 15}

// svidDocPath is the canonical guest mount path for the issued SVID document.
// The Manager writes this file mode 0400 (owner read-only). Consumers inside
// the microVM read their identity from here instead of static credentials.
const svidDocPath = "/var/run/secrets/spiffe/svid.json"

// SVIDDocPath returns the canonical mount path for the SVID document.
func SVIDDocPath() string { return svidDocPath }

// SVID is an issued SPIFFE X.509-SVID: an identity document binding a single
// SPIFFE ID to a short-lived certificate plus private key, verifiable against a
// trust bundle. It mirrors the SPIFFE Workload API X509-SVID so a go-spiffe
// based guest client can consume the persisted form without translation.
type SVID struct {
	// ID is the SPIFFE ID (spiffe://<trust-domain>/<path>). It MUST equal the
	// single URI SAN of the leaf certificate.
	ID string
	// Certificates is the certificate chain, leaf first, then intermediates.
	Certificates []*x509.Certificate
	// PrivateKey is the private key matching the leaf certificate.
	PrivateKey crypto.Signer
	// Bundle is the set of trust-domain CA certificates (roots) used to verify
	// this and peer SVIDs.
	Bundle []*x509.Certificate
	// IssuedAt / ExpiresAt bound the validity window.
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Document is the on-disk JSON representation mounted at SVIDDocPath. The schema
// mirrors a SPIFFE Workload API X509-SVID response so a future go-spiffe based
// guest client can read it directly. All certificate material is PEM-encoded.
type Document struct {
	SpiffeID    string `json:"spiffe_id"`
	X509SVID    string `json:"x509_svid"`     // PEM leaf + intermediates
	X509SVIDKey string `json:"x509_svid_key"` // PEM PKCS#8 private key
	Bundle      string `json:"bundle"`        // PEM trust bundle (CA roots)
	IssuedAt    string `json:"issued_at"`     // RFC3339
	ExpiresAt   string `json:"expires_at"`    // RFC3339
}

// leaf returns the leaf certificate, or nil if the chain is empty.
func (s *SVID) leaf() *x509.Certificate {
	if len(s.Certificates) == 0 {
		return nil
	}
	return s.Certificates[0]
}

// Validate checks structural integrity and SPIFFE ID/URI-SAN binding. It does
// not perform cryptographic chain verification (see Verify).
func (s *SVID) Validate() error {
	if err := validateSPIFFEID(s.ID); err != nil {
		return err
	}
	leaf := s.leaf()
	if leaf == nil {
		return fmt.Errorf("svid: certificate chain is empty")
	}
	if len(s.Bundle) == 0 {
		return fmt.Errorf("svid: trust bundle is empty")
	}
	if s.PrivateKey == nil {
		return fmt.Errorf("svid: private key is missing")
	}
	// Guard against nil certificate entries from a malformed Source so the rest
	// of validation (and Verify) fail closed instead of panicking.
	if i := firstNilCert(s.Certificates); i >= 0 {
		return fmt.Errorf("svid: certificate chain entry %d is nil", i)
	}
	if i := firstNilCert(s.Bundle); i >= 0 {
		return fmt.Errorf("svid: trust bundle entry %d is nil", i)
	}
	if err := s.validateWindow(leaf); err != nil {
		return err
	}
	if err := s.validateBundle(); err != nil {
		return err
	}
	if err := validateLeafConstraints(leaf, s.ID); err != nil {
		return err
	}
	// The private key must correspond to the leaf certificate, otherwise the
	// mounted credential cannot complete a handshake (and may bind to a key the
	// holder does not control).
	return keyMatchesLeaf(s.PrivateKey, leaf)
}

// validateWindow checks the advertised IssuedAt/ExpiresAt against each other and
// against the EFFECTIVE validity of the verification path: the earliest NotAfter
// across the leaf, intermediates, and trust bundle. Rotation is scheduled from
// ExpiresAt; if any chain/bundle certificate expires earlier, the SVID would
// become unverifiable before the scheduled rotation/removal point.
func (s *SVID) validateWindow(leaf *x509.Certificate) error {
	if !s.ExpiresAt.After(s.IssuedAt) {
		return fmt.Errorf("svid: expires_at %s is not after issued_at %s",
			s.ExpiresAt.Format(time.RFC3339), s.IssuedAt.Format(time.RFC3339))
	}
	effExpiry := leaf.NotAfter
	for _, c := range s.Certificates {
		if c.NotAfter.Before(effExpiry) {
			effExpiry = c.NotAfter
		}
	}
	for _, c := range s.Bundle {
		if c.NotAfter.Before(effExpiry) {
			effExpiry = c.NotAfter
		}
	}
	if s.ExpiresAt.After(effExpiry) {
		return fmt.Errorf("svid: expires_at %s is after the earliest chain/bundle NotAfter %s",
			s.ExpiresAt.Format(time.RFC3339), effExpiry.Format(time.RFC3339))
	}
	if s.IssuedAt.Before(leaf.NotBefore) {
		return fmt.Errorf("svid: issued_at %s is before leaf NotBefore %s",
			s.IssuedAt.Format(time.RFC3339), leaf.NotBefore.Format(time.RFC3339))
	}
	return nil
}

// validateBundle requires every trust-bundle entry to be a CA. Bundle entries
// are anchors used as Roots during verification; this also prevents a self
// signed leaf from being smuggled in as its own trust root.
func (s *SVID) validateBundle() error {
	for i, c := range s.Bundle {
		if !c.IsCA {
			return fmt.Errorf("svid: trust bundle entry %d is not a CA certificate", i)
		}
	}
	return nil
}

// validateLeafConstraints enforces the SPIFFE X509-SVID leaf requirements:
// exactly one URI SAN equal to the SPIFFE ID, Basic Constraints with cA=false,
// a critical Key Usage extension with digitalSignature (and no cert/CRL signing),
// and both clientAuth and serverAuth EKUs.
func validateLeafConstraints(leaf *x509.Certificate, id string) error {
	if len(leaf.URIs) != 1 || leaf.URIs[0].String() != id {
		return fmt.Errorf("svid: leaf URI SAN %v does not match SPIFFE ID %q", leaf.URIs, id)
	}
	if !leaf.BasicConstraintsValid {
		return fmt.Errorf("svid: leaf certificate is missing the Basic Constraints extension (cA=false required)")
	}
	if leaf.IsCA {
		return fmt.Errorf("svid: leaf certificate must not be a CA")
	}
	if !hasCriticalExtension(leaf, oidKeyUsage) {
		return fmt.Errorf("svid: leaf certificate must carry a critical Key Usage extension")
	}
	if leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("svid: leaf certificate lacks digitalSignature key usage")
	}
	if leaf.KeyUsage&(x509.KeyUsageCertSign|x509.KeyUsageCRLSign) != 0 {
		return fmt.Errorf("svid: leaf certificate must not assert certSign/crlSign key usage")
	}
	// The SPIFFE X509-SVID spec permits omitting the EKU extension, but if it is
	// present an SVID used for mTLS must carry both clientAuth and serverAuth.
	if len(leaf.ExtKeyUsage) > 0 || len(leaf.UnknownExtKeyUsage) > 0 {
		if !hasExtKeyUsage(leaf, x509.ExtKeyUsageClientAuth) || !hasExtKeyUsage(leaf, x509.ExtKeyUsageServerAuth) {
			return fmt.Errorf("svid: leaf certificate has an EKU extension but does not assert both clientAuth and serverAuth")
		}
	}
	return nil
}

// firstNilCert returns the index of the first nil certificate, or -1 if none.
func firstNilCert(certs []*x509.Certificate) int {
	for i, c := range certs {
		if c == nil {
			return i
		}
	}
	return -1
}

// keyMatchesLeaf confirms the private key's public half equals the leaf's public
// key. All stdlib key types implement Equal(crypto.PublicKey) bool.
func keyMatchesLeaf(key crypto.Signer, leaf *x509.Certificate) error {
	if key == nil {
		return fmt.Errorf("svid: private key is missing")
	}
	pub, ok := key.Public().(interface{ Equal(x crypto.PublicKey) bool })
	if !ok {
		return fmt.Errorf("svid: private key type does not support public-key comparison")
	}
	if !pub.Equal(leaf.PublicKey) {
		return fmt.Errorf("svid: private key does not match the leaf certificate")
	}
	return nil
}

func hasExtKeyUsage(cert *x509.Certificate, want x509.ExtKeyUsage) bool {
	for _, eku := range cert.ExtKeyUsage {
		if eku == want {
			return true
		}
	}
	return false
}

// hasCriticalExtension reports whether cert carries the given extension OID
// marked critical.
func hasCriticalExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oid) {
			return ext.Critical
		}
	}
	return false
}

// Verify confirms the leaf chains to the trust bundle and that both the
// advertised SVID window and the leaf certificate are valid at now. The
// advertised window is checked explicitly because rotation is scheduled from it
// and it may be narrower than the leaf's own validity.
func (s *SVID) Verify(now time.Time) error {
	leaf := s.leaf()
	if leaf == nil {
		return fmt.Errorf("svid: certificate chain is empty")
	}
	if now.Before(s.IssuedAt) {
		return fmt.Errorf("svid: not yet valid (issued_at %s is in the future)", s.IssuedAt.Format(time.RFC3339))
	}
	if !now.Before(s.ExpiresAt) {
		return fmt.Errorf("svid: expired (expires_at %s)", s.ExpiresAt.Format(time.RFC3339))
	}
	roots := x509.NewCertPool()
	for _, c := range s.Bundle {
		roots.AddCert(c)
	}
	intermediates := x509.NewCertPool()
	for _, c := range s.Certificates[1:] {
		intermediates.AddCert(c)
	}
	// An X509-SVID must be usable for both sides of mTLS. x509.Verify accepts a
	// chain if ANY listed EKU is permitted, so verify each role separately to
	// reject a chain constrained (via EKU nesting) to only one role.
	for _, eku := range []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth} {
		if _, err := leaf.Verify(x509.VerifyOptions{
			Roots:         roots,
			Intermediates: intermediates,
			CurrentTime:   now,
			KeyUsages:     []x509.ExtKeyUsage{eku},
		}); err != nil {
			return fmt.Errorf("svid: chain verification failed for EKU %d: %w", eku, err)
		}
	}
	return nil
}

// Document renders the SVID to its persisted JSON form.
func (s *SVID) Document() (*Document, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(s.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("svid: marshal private key: %w", err)
	}
	return &Document{
		SpiffeID:    s.ID,
		X509SVID:    encodeCertsPEM(s.Certificates),
		X509SVIDKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})),
		Bundle:      encodeCertsPEM(s.Bundle),
		IssuedAt:    s.IssuedAt.UTC().Format(time.RFC3339),
		ExpiresAt:   s.ExpiresAt.UTC().Format(time.RFC3339),
	}, nil
}

// MarshalDocument renders the SVID to indented JSON bytes for persistence.
func (s *SVID) MarshalDocument() ([]byte, error) {
	doc, err := s.Document()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(doc, "", "  ")
}

// ParseDocument reconstructs an SVID from its persisted JSON form. It is the
// inverse of MarshalDocument and models how a guest client reads the mount.
func ParseDocument(data []byte) (*SVID, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("svid: decode document: %w", err)
	}
	certs, err := decodeCertsPEM([]byte(doc.X509SVID))
	if err != nil {
		return nil, fmt.Errorf("svid: decode x509_svid: %w", err)
	}
	bundle, err := decodeCertsPEM([]byte(doc.Bundle))
	if err != nil {
		return nil, fmt.Errorf("svid: decode bundle: %w", err)
	}
	key, err := decodeKeyPEM([]byte(doc.X509SVIDKey))
	if err != nil {
		return nil, fmt.Errorf("svid: decode key: %w", err)
	}
	issuedAt, err := time.Parse(time.RFC3339, doc.IssuedAt)
	if err != nil {
		return nil, fmt.Errorf("svid: parse issued_at: %w", err)
	}
	expiresAt, err := time.Parse(time.RFC3339, doc.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("svid: parse expires_at: %w", err)
	}
	svid := &SVID{
		ID:           doc.SpiffeID,
		Certificates: certs,
		PrivateKey:   key,
		Bundle:       bundle,
		IssuedAt:     issuedAt,
		ExpiresAt:    expiresAt,
	}
	if err := svid.Validate(); err != nil {
		return nil, err
	}
	return svid, nil
}

// trustDomainPattern is the SPIFFE trust-domain grammar: lowercase letters,
// digits, and the special chars dot, hyphen, underscore.
var trustDomainPattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

// pathSegmentPattern is the SPIFFE path-segment grammar: letters, digits, and
// the special chars dot, hyphen, underscore.
var pathSegmentPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// validateSPIFFEID enforces the SPIFFE ID grammar (scheme, trust domain, and
// path). This is a boundary check on data that becomes a filesystem-mounted
// credential and a policy subject, so canonicalization ambiguity is rejected.
func validateSPIFFEID(id string) error {
	if id == "" {
		return fmt.Errorf("svid: SPIFFE ID is required")
	}
	if len(id) > 2048 {
		return fmt.Errorf("svid: SPIFFE ID exceeds 2048 bytes")
	}
	// Percent-encoding is not permitted in a SPIFFE ID; reject before url.Parse
	// can canonicalize escapes into otherwise-disallowed characters.
	if strings.Contains(id, "%") {
		return fmt.Errorf("svid: SPIFFE ID must not contain percent-encoding")
	}
	u, err := url.Parse(id)
	if err != nil {
		return fmt.Errorf("svid: SPIFFE ID is not a valid URI: %w", err)
	}
	if u.Scheme != "spiffe" {
		return fmt.Errorf("svid: SPIFFE ID scheme must be \"spiffe\", got %q", u.Scheme)
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("svid: SPIFFE ID must not contain userinfo, query, or fragment")
	}
	if u.Port() != "" {
		return fmt.Errorf("svid: SPIFFE ID trust domain must not contain a port")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("svid: SPIFFE ID is missing a trust domain")
	}
	if !trustDomainPattern.MatchString(host) {
		return fmt.Errorf("svid: SPIFFE ID trust domain %q is invalid (lowercase letters, digits, .-_ only)", host)
	}
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return fmt.Errorf("svid: SPIFFE ID is missing a workload path")
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == "" {
			return fmt.Errorf("svid: SPIFFE ID has an empty path segment (no trailing or doubled slashes)")
		}
		if seg == "." || seg == ".." {
			return fmt.Errorf("svid: SPIFFE ID path segment %q is not permitted", seg)
		}
		if !pathSegmentPattern.MatchString(seg) {
			return fmt.Errorf("svid: SPIFFE ID path segment %q is invalid (letters, digits, .-_ only)", seg)
		}
	}
	return nil
}

func encodeCertsPEM(certs []*x509.Certificate) string {
	var b strings.Builder
	for _, c := range certs {
		_ = pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	}
	return b.String()
}

func decodeCertsPEM(data []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse certificate: %w", err)
		}
		certs = append(certs, c)
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM block")
	}
	return certs, nil
}

func decodeKeyPEM(data []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 key: %w", err)
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("key does not implement crypto.Signer")
	}
	return signer, nil
}
