package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jpoley/nanofuse/internal/config"
	"github.com/jpoley/nanofuse/internal/policy"
)

func authCfg(enabled bool, keys ...string) *config.AuthConfig {
	return &config.AuthConfig{
		Enabled:       enabled,
		StaticAPIKeys: keys,
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// TestAuthDisabled verifies that all requests pass when auth is disabled.
func TestAuthDisabled(t *testing.T) {
	mw := AuthMiddleware(authCfg(false), nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when auth disabled, got %d", w.Code)
	}
}

// TestAuthSVIDHeaderAccepted verifies that a valid X-SPIFFE-ID is accepted.
func TestAuthSVIDHeaderAccepted(t *testing.T) {
	mw := AuthMiddleware(authCfg(true), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("X-SPIFFE-ID", "spiffe://example.com/workload/vm1")
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid SVID header, got %d", w.Code)
	}
}

// TestAuthSVIDInvalidPrefix verifies that a malformed X-SPIFFE-ID is rejected.
func TestAuthSVIDInvalidPrefix(t *testing.T) {
	mw := AuthMiddleware(authCfg(true), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("X-SPIFFE-ID", "not-a-spiffe-uri")
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for malformed SVID, got %d", w.Code)
	}
}

// TestAuthNoCredential verifies that a request with no credential is rejected.
func TestAuthNoCredential(t *testing.T) {
	mw := AuthMiddleware(authCfg(true), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for no credential, got %d", w.Code)
	}
}

// TestAuthStaticKeyDevMode verifies static API key auth when DEV_STATIC_KEYS=true.
func TestAuthStaticKeyDevMode(t *testing.T) {
	os.Setenv("DEV_STATIC_KEYS", "true")
	defer os.Unsetenv("DEV_STATIC_KEYS")

	mw := AuthMiddleware(authCfg(true, "secret-key"), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid static key in dev mode, got %d", w.Code)
	}
}

// TestAuthStaticKeyProductionMode verifies static API key is NOT accepted when
// DEV_STATIC_KEYS is not set.
func TestAuthStaticKeyProductionMode(t *testing.T) {
	os.Unsetenv("DEV_STATIC_KEYS")
	mw := AuthMiddleware(authCfg(true, "secret-key"), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for static key in production mode, got %d", w.Code)
	}
}

// TestAuthStaticKeyWrongKey verifies that an incorrect static key is rejected.
func TestAuthStaticKeyWrongKey(t *testing.T) {
	os.Setenv("DEV_STATIC_KEYS", "true")
	defer os.Unsetenv("DEV_STATIC_KEYS")

	mw := AuthMiddleware(authCfg(true, "secret-key"), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong static key, got %d", w.Code)
	}
}

// TestAuthPolicyDeniesGeo verifies that the policy engine can deny a request.
func TestAuthPolicyDeniesGeo(t *testing.T) {
	eng := policy.NewEngine()
	_ = eng.AddPolicy(&policy.WorkloadPolicy{
		ID:             "geo-deny",
		WorkloadSPIFFE: "spiffe://example.com/workload/vm1",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []policy.AccessCondition{
			{
				Type:           policy.ConditionGeo,
				GeoRestriction: &policy.GeoRestriction{AllowedRegions: []string{"US"}},
			},
		},
	})

	mw := AuthMiddleware(authCfg(true), eng)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("X-SPIFFE-ID", "spiffe://example.com/workload/vm1")
	req.Header.Set("X-Geo-Region", "CN") // not in US
	w := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for geo-denied request, got %d", w.Code)
	}
}

// TestCredentialFromContext verifies that the credential is available in the
// handler context after successful auth.
func TestCredentialFromContext(t *testing.T) {
	var gotCred *Credential

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCred = CredentialFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := AuthMiddleware(authCfg(true), nil)
	req := httptest.NewRequest(http.MethodGet, "/vms", nil)
	req.Header.Set("X-SPIFFE-ID", "spiffe://example.com/workload/vm2")
	w := httptest.NewRecorder()
	mw(handler).ServeHTTP(w, req)

	if gotCred == nil {
		t.Fatal("expected credential in context, got nil")
	}
	if gotCred.Kind != "svid" {
		t.Errorf("expected kind 'svid', got %q", gotCred.Kind)
	}
	if gotCred.SpiffeID != "spiffe://example.com/workload/vm2" {
		t.Errorf("unexpected SpiffeID: %s", gotCred.SpiffeID)
	}
}
