package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/types"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func requestWithTLSState(state *tls.ConnectionState) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.TLS = state
	return req
}

func certWithURI(t *testing.T, rawURI string) *x509.Certificate {
	t.Helper()

	parsed, err := url.Parse(rawURI)
	if err != nil {
		t.Fatalf("failed to parse test URI: %v", err)
	}
	return &x509.Certificate{URIs: []*url.URL{parsed}}
}

func serveWithMTLSIdentityMiddleware(w *httptest.ResponseRecorder, req *http.Request, handler http.Handler) {
	MTLSIdentityMiddleware(nil, handler).ServeHTTP(w, req)
}

func assertUnauthorizedJSON(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode JSON error: %v", err)
	}
	if response.Error.Code != types.ErrUnauthorized {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrUnauthorized)
	}
	if response.Error.Message != mtlsUnauthorizedMessage {
		t.Fatalf("error message = %q, want %q", response.Error.Message, mtlsUnauthorizedMessage)
	}
	if response.Error.Details != nil {
		t.Fatalf("error details = %#v, want nil", response.Error.Details)
	}
	if strings.Contains(response.Error.Message, "mTLS is required") {
		t.Fatalf("error message leaked internal deny reason: %q", response.Error.Message)
	}
}

func TestMTLSIdentityMiddlewareAcceptsVerifiedSPIFFEURI(t *testing.T) {
	cert := certWithURI(t, "spiffe://example.com/workload/vm1")
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
		VerifiedChains:   [][]*x509.Certificate{{cert}},
	})
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with SPIFFE URI SAN, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareRejectsNoTLS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	assertUnauthorizedJSON(t, w)
}

func TestMTLSIdentityMiddlewareRejectsSpoofedHeaderWithoutTLS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("X-SPIFFE-ID", "spiffe://example.com/workload/spoofed")
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	assertUnauthorizedJSON(t, w)
}

func TestMTLSIdentityMiddlewareRejectsNoClientCert(t *testing.T) {
	req := requestWithTLSState(&tls.ConnectionState{})
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no client certificate, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareRejectsUnverifiedClientCert(t *testing.T) {
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			certWithURI(t, "spiffe://example.com/workload/vm1"),
		},
	})
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with unverified client certificate, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareRejectsNoSPIFFEURI(t *testing.T) {
	cert := &x509.Certificate{}
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
		VerifiedChains:   [][]*x509.Certificate{{cert}},
	})
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, okHandler())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no SPIFFE URI SAN, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareIgnoresSpoofedHeaderWhenTLSIsVerified(t *testing.T) {
	var gotCred *Credential
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCred = CredentialFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	cert := certWithURI(t, "spiffe://example.com/workload/real")
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
		VerifiedChains:   [][]*x509.Certificate{{cert}},
	})
	req.Header.Set("X-SPIFFE-ID", "spiffe://example.com/workload/spoofed")
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, handler)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with verified SPIFFE URI SAN, got %d", w.Code)
	}
	if gotCred == nil {
		t.Fatal("expected credential in context, got nil")
	}
	if gotCred.SpiffeID != "spiffe://example.com/workload/real" {
		t.Fatalf("credential SpiffeID = %q, want verified certificate URI", gotCred.SpiffeID)
	}
}

func TestMTLSIdentityMiddlewareStoresCredential(t *testing.T) {
	var gotCred *Credential
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCred = CredentialFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	cert := certWithURI(t, "spiffe://example.com/workload/vm2")
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
		VerifiedChains:   [][]*x509.Certificate{{cert}},
	})
	w := httptest.NewRecorder()

	serveWithMTLSIdentityMiddleware(w, req, handler)

	if gotCred == nil {
		t.Fatal("expected credential in context, got nil")
	}
	if gotCred.Kind != "mtls_spiffe" {
		t.Errorf("expected kind mtls_spiffe, got %q", gotCred.Kind)
	}
	if gotCred.SpiffeID != "spiffe://example.com/workload/vm2" {
		t.Errorf("unexpected SpiffeID: %s", gotCred.SpiffeID)
	}
}

func TestBuildAuthTLSConfigIncludesClientCAPathOnParseError(t *testing.T) {
	tmpDir := t.TempDir()
	certPath, keyPath := writeTestKeyPair(t, tmpDir)
	caPath := filepath.Join(tmpDir, "client-ca.pem")
	if err := os.WriteFile(caPath, []byte("not a certificate"), 0600); err != nil {
		t.Fatalf("write client CA: %v", err)
	}

	_, err := BuildAuthTLSConfig(&config.AuthConfig{
		TLSCertFile:  certPath,
		TLSKeyFile:   keyPath,
		ClientCAFile: caPath,
	})
	if err == nil {
		t.Fatal("BuildAuthTLSConfig returned nil error, want client CA parse error")
	}
	if !strings.Contains(err.Error(), caPath) {
		t.Fatalf("error = %q, want client CA path %q", err.Error(), caPath)
	}
}

func TestBuildAuthTLSConfigIncludesClientCAPathOnReadError(t *testing.T) {
	tmpDir := t.TempDir()
	certPath, keyPath := writeTestKeyPair(t, tmpDir)
	caPath := filepath.Join(tmpDir, "missing-client-ca.pem")

	_, err := BuildAuthTLSConfig(&config.AuthConfig{
		TLSCertFile:  certPath,
		TLSKeyFile:   keyPath,
		ClientCAFile: caPath,
	})
	if err == nil {
		t.Fatal("BuildAuthTLSConfig returned nil error, want client CA read error")
	}
	if !strings.Contains(err.Error(), caPath) {
		t.Fatalf("error = %q, want client CA path %q", err.Error(), caPath)
	}
}

func writeTestKeyPair(t *testing.T, dir string) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "nanofuse-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certPath := filepath.Join(dir, "server.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatalf("write cert: %v", err)
	}

	keyPath := filepath.Join(dir, "server.key")
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	return certPath, keyPath
}
