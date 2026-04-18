package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jpoley/nanofuse/internal/config"
	"github.com/jpoley/nanofuse/internal/policy"
)

// ctxKeyCredential is the context key for the validated credential.
type ctxKeyCredential struct{}

// Credential holds the validated identity extracted from a request.
type Credential struct {
	// SpiffeID is the SPIFFE URI from the SVID (empty for static-key auth).
	SpiffeID string

	// Kind is "svid" or "static_key".
	Kind string

	// IssuedAt is when the SVID was issued (zero for static-key auth).
	IssuedAt time.Time
}

// CredentialFromContext retrieves the Credential stored by AuthMiddleware.
// Returns nil if the request was not authenticated (only possible when auth
// is disabled).
func CredentialFromContext(ctx context.Context) *Credential {
	v, _ := ctx.Value(ctxKeyCredential{}).(*Credential)
	return v
}

// AuthMiddleware returns an HTTP middleware that enforces credential validation
// and optional policy evaluation.
//
// Behaviour:
//   - When auth is disabled (cfg.Auth.Enabled == false) the middleware is a
//     transparent pass-through — useful for local dev without any credentials.
//   - When DEV_STATIC_KEYS=true is set in the environment AND
//     cfg.Auth.StaticAPIKeys is non-empty, requests bearing a matching
//     "Authorization: Bearer <key>" header are accepted (dev mode only).
//   - All other requests must carry a valid SPIFFE SVID header:
//     "X-SPIFFE-ID: spiffe://..." — the middleware validates the header and
//     then hands off to the policy engine if one is supplied.
//   - Every credential use (accept or deny) writes a structured audit log
//     entry with key "event"="credential_use" so it can be piped to a SIEM.
//
// NOTE: Full mTLS SVID verification (parsing the X.509 certificate from the
// TLS connection and verifying it against the SPIRE trust bundle) is wired in
// by the TLS layer.  This middleware validates the X-SPIFFE-ID assertion that
// the TLS terminator inserts after verifying the peer certificate.  In a
// production deployment the TLS terminator MUST be trusted to set this header.
func AuthMiddleware(cfg *config.AuthConfig, engine *policy.Engine) func(http.Handler) http.Handler {
	devStaticKeys := os.Getenv("DEV_STATIC_KEYS") == "true"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				// Auth disabled — pass through without any checks.
				next.ServeHTTP(w, r)
				return
			}

			cred, err := extractCredential(r, cfg, devStaticKeys)
			if err != nil {
				auditCredential(r, nil, false, err.Error())
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// If a policy engine is wired in, evaluate the request.
			if engine != nil && cred.Kind == "svid" {
				geoRegion := r.Header.Get("X-Geo-Region")
				approvalToken := r.Header.Get("X-Approval-Token")
				serviceSPIFFE := r.Header.Get("X-Service-SPIFFE")

				result := engine.Evaluate(r.Context(), policy.EvalRequest{
					WorkloadSPIFFE: cred.SpiffeID,
					ServiceSPIFFE:  serviceSPIFFE,
					GeoRegion:      geoRegion,
					ApprovalToken:  approvalToken,
				})

				if !result.Allowed {
					auditCredential(r, cred, false, result.Reason)
					http.Error(w, "Forbidden: "+result.Reason, http.StatusForbidden)
					return
				}
			}

			auditCredential(r, cred, true, "")

			ctx := context.WithValue(r.Context(), ctxKeyCredential{}, cred)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractCredential attempts to derive a Credential from the request.
// Order of precedence:
//  1. SVID header (X-SPIFFE-ID) — always tried first.
//  2. Static API key (Authorization: Bearer <key>) — only in dev mode.
func extractCredential(r *http.Request, cfg *config.AuthConfig, devStaticKeys bool) (*Credential, error) {
	// 1. SVID via X-SPIFFE-ID header (set by the mTLS terminator).
	if spiffeID := r.Header.Get("X-SPIFFE-ID"); spiffeID != "" {
		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return nil, errorf("X-SPIFFE-ID header does not look like a valid SPIFFE URI: %s", spiffeID)
		}
		return &Credential{
			SpiffeID: spiffeID,
			Kind:     "svid",
			IssuedAt: time.Now(), // actual issuance time would come from the cert
		}, nil
	}

	// 2. Static API key — dev mode only.
	if devStaticKeys && len(cfg.StaticAPIKeys) > 0 {
		bearer := r.Header.Get("Authorization")
		if strings.HasPrefix(bearer, "Bearer ") {
			provided := strings.TrimPrefix(bearer, "Bearer ")
			for _, key := range cfg.StaticAPIKeys {
				if provided == key {
					slog.Warn("static API key auth used (DEV mode)",
						slog.String("event", "dev_static_key_auth"),
						slog.String("remote_addr", r.RemoteAddr),
						slog.String("path", r.URL.Path),
					)
					return &Credential{Kind: "static_key"}, nil
				}
			}
		}
		return nil, errorf("invalid static API key")
	}

	return nil, errorf("no valid credential provided (expected X-SPIFFE-ID header or, in dev mode, Authorization: Bearer <key>)")
}

// auditCredential writes a structured audit log entry for every credential use.
func auditCredential(r *http.Request, cred *Credential, allowed bool, reason string) {
	attrs := []any{
		slog.String("event", "credential_use"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.Bool("allowed", allowed),
	}
	if cred != nil {
		attrs = append(attrs,
			slog.String("cred_kind", cred.Kind),
			slog.String("spiffe_id", cred.SpiffeID),
		)
	}
	if reason != "" {
		attrs = append(attrs, slog.String("deny_reason", reason))
	}
	if allowed {
		slog.Info("auth audit", attrs...)
	} else {
		slog.Warn("auth audit", attrs...)
	}
}

// errorf is a thin wrapper to create an error with a formatted message.
func errorf(format string, args ...any) error {
	return &authError{msg: fmt.Sprintf(format, args...)}
}

// authError is a simple error type for auth failures.
type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }
