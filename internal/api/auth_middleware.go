package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/types"
)

// ctxKeyCredential is the context key for the authenticated credential.
type ctxKeyCredential struct{}

// Credential holds the mTLS identity extracted from a verified client
// certificate.
type Credential struct {
	SpiffeID string
	Kind     string
}

// CredentialFromContext retrieves the Credential stored by MTLSIdentityMiddleware.
// It returns nil when no mTLS identity middleware ran for the request.
func CredentialFromContext(ctx context.Context) *Credential {
	v, _ := ctx.Value(ctxKeyCredential{}).(*Credential)
	return v
}

// BuildAuthTLSConfig builds the TCP listener TLS config used when auth.enabled
// is true. Client certificates are verified by the TLS stack before requests
// reach handlers.
func BuildAuthTLSConfig(cfg *config.AuthConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("auth config is required")
	}

	cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load API TLS key pair: %w", err)
	}

	caPEM, err := os.ReadFile(cfg.ClientCAFile) //nolint:gosec // operator-supplied daemon config path
	if err != nil {
		return nil, fmt.Errorf("failed to read API client CA file: %w", err)
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse API client CA file %q", cfg.ClientCAFile)
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
	}, nil
}

// MTLSIdentityMiddleware requires a verified client certificate containing a
// SPIFFE URI SAN. It assumes the TCP listener was configured with
// tls.RequireAndVerifyClientCert and defensively requires VerifiedChains to be
// populated before trusting PeerCertificates. It does not trust
// caller-controlled identity headers.
func MTLSIdentityMiddleware(logger *logging.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cred, err := credentialFromTLS(r)
		if err != nil {
			auditCredential(logger, r, nil, false, err.Error())
			types.WriteError(w, http.StatusUnauthorized, types.ErrUnauthorized,
				"mTLS client certificate with SPIFFE identity is required", nil)
			return
		}

		auditCredential(logger, r, cred, true, "")
		ctx := context.WithValue(r.Context(), ctxKeyCredential{}, cred)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func credentialFromTLS(r *http.Request) (*Credential, error) {
	if r.TLS == nil {
		return nil, errorf("mTLS is required")
	}
	if len(r.TLS.PeerCertificates) == 0 {
		return nil, errorf("client certificate is required")
	}
	if len(r.TLS.VerifiedChains) == 0 {
		return nil, errorf("verified client certificate chain is required")
	}

	for _, uri := range r.TLS.PeerCertificates[0].URIs {
		if uri == nil || uri.Scheme != "spiffe" {
			continue
		}
		spiffeID := uri.String()
		if err := validateSPIFFEID(spiffeID); err != nil {
			return nil, err
		}
		return &Credential{
			SpiffeID: spiffeID,
			Kind:     "mtls_spiffe",
		}, nil
	}

	return nil, errorf("client certificate does not contain a SPIFFE URI SAN")
}

func validateSPIFFEID(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return errorf("invalid SPIFFE URI: %w", err)
	}
	if parsed.Scheme != "spiffe" {
		return errorf("SPIFFE URI must use spiffe scheme")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errorf("SPIFFE URI must not include userinfo, query, or fragment")
	}
	if parsed.Host == "" {
		return errorf("SPIFFE URI must include a trust domain")
	}
	if strings.Contains(parsed.Host, ":") {
		return errorf("SPIFFE URI trust domain must not include a port")
	}
	if net.ParseIP(parsed.Host) != nil {
		return errorf("SPIFFE URI trust domain must be a DNS name, not an IP address")
	}
	if parsed.EscapedPath() == "" || parsed.EscapedPath() == "/" {
		return errorf("SPIFFE URI must include a workload path")
	}
	return nil
}

func auditCredential(logger *logging.Logger, r *http.Request, cred *Credential, allowed bool, reason string) {
	if logger == nil {
		return
	}

	if cred != nil {
		logger.Info("auth audit event=credential_use method=%s path=%s remote_addr=%s allowed=%t cred_kind=%s spiffe_id=%s",
			r.Method, r.URL.Path, r.RemoteAddr, allowed, cred.Kind, cred.SpiffeID)
		return
	}
	logger.Warn("auth audit event=credential_use method=%s path=%s remote_addr=%s allowed=%t deny_reason=%q",
		r.Method, r.URL.Path, r.RemoteAddr, allowed, reason)
}

func errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
