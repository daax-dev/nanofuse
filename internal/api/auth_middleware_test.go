package api

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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

func TestMTLSIdentityMiddlewareAcceptsVerifiedSPIFFEURI(t *testing.T) {
	cert := certWithURI(t, "spiffe://example.com/workload/vm1")
	req := requestWithTLSState(&tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
		VerifiedChains:   [][]*x509.Certificate{{cert}},
	})
	w := httptest.NewRecorder()

	MTLSIdentityMiddleware(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with SPIFFE URI SAN, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareRejectsNoTLS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	w := httptest.NewRecorder()

	MTLSIdentityMiddleware(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no TLS state, got %d", w.Code)
	}
}

func TestMTLSIdentityMiddlewareRejectsNoClientCert(t *testing.T) {
	req := requestWithTLSState(&tls.ConnectionState{})
	w := httptest.NewRecorder()

	MTLSIdentityMiddleware(okHandler()).ServeHTTP(w, req)

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

	MTLSIdentityMiddleware(okHandler()).ServeHTTP(w, req)

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

	MTLSIdentityMiddleware(okHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no SPIFFE URI SAN, got %d", w.Code)
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

	MTLSIdentityMiddleware(handler).ServeHTTP(w, req)

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
