// Package policy implements workload IAM policies compatible with the Aembit
// access-policy schema.  It evaluates time-of-day, geographic, and approval-gate
// conditions and emits structured audit events for every policy decision.
//
// # Aembit access-policy schema compatibility
//
// The WorkloadPolicy struct mirrors the Aembit API schema documented at
// https://docs.aembit.io/reference/access-policies.  Key field mapping:
//
//	WorkloadPolicy.ID             → policy.id
//	WorkloadPolicy.Name           → policy.name
//	WorkloadPolicy.WorkloadSPIFFE → policy.workload_client_identity (SPIFFE URI)
//	WorkloadPolicy.ServiceSPIFFE  → policy.server_workload_id (SPIFFE URI)
//	WorkloadPolicy.Conditions     → policy.access_conditions[]
//	  TimeWindow                  → condition type "time_of_day"
//	  GeoRestriction              → condition type "geolocation"
//	  ApprovalGate                → condition type "just_in_time_approval"
//	WorkloadPolicy.Enabled        → policy.is_enabled
//
// Emit: every evaluation writes a structured audit log via log/slog with
// key "event"="policy_evaluation" so consumers (SIEM, log pipelines) can
// filter on it.
package policy

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ConditionType enumerates supported access condition kinds.
type ConditionType string

const (
	// ConditionTimeOfDay restricts access to specific hours of the day (UTC).
	ConditionTimeOfDay ConditionType = "time_of_day"

	// ConditionGeo restricts access to callers matching an allowed geographic flag.
	ConditionGeo ConditionType = "geolocation"

	// ConditionApprovalGate requires an explicit human approval before access is
	// granted (just-in-time / break-glass pattern).
	ConditionApprovalGate ConditionType = "just_in_time_approval"
)

// TimeWindow defines the allowed hours within a day (UTC).
// StartHour and EndHour are in [0, 23].  If EndHour < StartHour the window
// wraps midnight (e.g. 22–06 means 10pm through 6am).
type TimeWindow struct {
	// StartHour is the first permitted hour (inclusive), 0–23, UTC.
	StartHour int `json:"start_hour" yaml:"start_hour"`
	// EndHour is the last permitted hour (inclusive), 0–23, UTC.
	EndHour int `json:"end_hour" yaml:"end_hour"`
	// AllowedDays lists ISO weekday numbers (1=Mon … 7=Sun).  Empty = all days.
	AllowedDays []time.Weekday `json:"allowed_days,omitempty" yaml:"allowed_days,omitempty"`
}

// GeoRestriction permits access only when the caller claims one of the listed
// region codes.  Region codes follow ISO 3166-1 alpha-2 (e.g. "US", "DE").
// Verification of the caller's actual location is out-of-scope for this layer;
// the restriction is enforced against the geo_flag claim in the SVID JWT or
// request context.
type GeoRestriction struct {
	// AllowedRegions lists ISO 3166-1 alpha-2 region codes that are permitted.
	AllowedRegions []string `json:"allowed_regions" yaml:"allowed_regions"`
}

// ApprovalGate requires that a valid approval token has been recorded before
// access is granted.  Approvals are tracked in-memory by the Engine.
type ApprovalGate struct {
	// ApproverGroups lists the SPIFFE groups authorised to issue approvals.
	ApproverGroups []string `json:"approver_groups" yaml:"approver_groups"`
	// MaxValiditySeconds is how long an approval remains valid (default 3600).
	MaxValiditySeconds int `json:"max_validity_seconds" yaml:"max_validity_seconds"`
}

// AccessCondition is a single named condition that must be satisfied.
type AccessCondition struct {
	// Type identifies the kind of condition.
	Type ConditionType `json:"type" yaml:"type"`

	// TimeWindow is populated when Type == ConditionTimeOfDay.
	TimeWindow *TimeWindow `json:"time_window,omitempty" yaml:"time_window,omitempty"`

	// GeoRestriction is populated when Type == ConditionGeo.
	GeoRestriction *GeoRestriction `json:"geo_restriction,omitempty" yaml:"geo_restriction,omitempty"`

	// ApprovalGate is populated when Type == ConditionApprovalGate.
	ApprovalGate *ApprovalGate `json:"approval_gate,omitempty" yaml:"approval_gate,omitempty"`
}

// WorkloadPolicy is an access policy governing one workload (identified by its
// SPIFFE ID) reaching a service (identified by its SPIFFE ID).
//
// Schema is intentionally compatible with the Aembit access-policy API — see
// package-level doc comment for field mapping details.
type WorkloadPolicy struct {
	// ID is the policy's unique identifier.
	ID string `json:"id" yaml:"id"`

	// Name is a human-readable label.
	Name string `json:"name" yaml:"name"`

	// WorkloadSPIFFE is the SPIFFE URI of the calling workload (client identity).
	WorkloadSPIFFE string `json:"workload_spiffe" yaml:"workload_spiffe"`

	// ServiceSPIFFE is the SPIFFE URI of the server workload being accessed.
	ServiceSPIFFE string `json:"service_spiffe" yaml:"service_spiffe"`

	// Conditions is the list of access conditions that must ALL be satisfied.
	Conditions []AccessCondition `json:"conditions" yaml:"conditions"`

	// Enabled controls whether this policy is enforced.  Disabled policies are
	// stored but skipped during evaluation.
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// EvalRequest carries the context needed to evaluate a policy.
type EvalRequest struct {
	// WorkloadSPIFFE is the SPIFFE URI presented by the caller.
	WorkloadSPIFFE string

	// ServiceSPIFFE is the SPIFFE URI of the resource being accessed.
	ServiceSPIFFE string

	// GeoRegion is the ISO 3166-1 alpha-2 region code claimed by the caller.
	// May be empty if not available.
	GeoRegion string

	// ApprovalToken is a previously issued approval token (for ApprovalGate).
	// May be empty.
	ApprovalToken string

	// Now is the reference time for time-of-day checks.  Defaults to time.Now()
	// if zero.
	Now time.Time
}

// EvalResult captures the outcome of a policy evaluation.
type EvalResult struct {
	// Allowed is true when all conditions are satisfied (or no policy matched).
	Allowed bool

	// PolicyID is the ID of the matching policy, empty if no policy matched.
	PolicyID string

	// Reason describes why access was denied, empty when allowed.
	Reason string
}

// approvalRecord stores a granted approval and its expiry.
type approvalRecord struct {
	token     string
	expiresAt time.Time
}

// Engine stores WorkloadPolicies and evaluates them against EvalRequests.
type Engine struct {
	mu        sync.RWMutex
	policies  map[string]*WorkloadPolicy // keyed by policy ID
	approvals map[string]approvalRecord  // keyed by approval token
}

// NewEngine creates an empty policy engine.
func NewEngine() *Engine {
	return &Engine{
		policies:  make(map[string]*WorkloadPolicy),
		approvals: make(map[string]approvalRecord),
	}
}

// AddPolicy registers a policy.  Replaces any existing policy with the same ID.
func (e *Engine) AddPolicy(p *WorkloadPolicy) error {
	if p.ID == "" {
		return fmt.Errorf("policy ID must not be empty")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[p.ID] = p
	slog.Info("policy registered",
		slog.String("policy_id", p.ID),
		slog.String("workload", p.WorkloadSPIFFE),
		slog.String("service", p.ServiceSPIFFE),
	)
	return nil
}

// RemovePolicy removes a policy by ID.  A no-op if the policy does not exist.
func (e *Engine) RemovePolicy(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.policies, id)
}

// GrantApproval records a human approval token valid for validitySeconds.
// This satisfies ApprovalGate conditions that match the token.
func (e *Engine) GrantApproval(token string, validitySeconds int) {
	if validitySeconds <= 0 {
		validitySeconds = 3600
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.approvals[token] = approvalRecord{
		token:     token,
		expiresAt: time.Now().Add(time.Duration(validitySeconds) * time.Second),
	}
	slog.Info("approval granted",
		slog.String("token_prefix", safePrefix(token, 8)),
		slog.Time("expires_at", e.approvals[token].expiresAt),
	)
}

// Evaluate checks whether the request satisfies all applicable policies.
// If no policy matches the workload/service pair the default is to allow
// (fail-open is explicit here; callers that want fail-closed should check
// EvalResult.PolicyID == "").
func (e *Engine) Evaluate(ctx context.Context, req EvalRequest) EvalResult {
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, p := range e.policies {
		if !p.Enabled {
			continue
		}
		if !spiffeMatches(p.WorkloadSPIFFE, req.WorkloadSPIFFE) {
			continue
		}
		if !spiffeMatches(p.ServiceSPIFFE, req.ServiceSPIFFE) {
			continue
		}

		// Policy matched — evaluate all conditions.
		for _, cond := range p.Conditions {
			if deny, reason := e.evalCondition(cond, req, now); deny {
				result := EvalResult{
					Allowed:  false,
					PolicyID: p.ID,
					Reason:   reason,
				}
				e.audit(ctx, req, result, p.Name)
				return result
			}
		}

		// All conditions passed.
		result := EvalResult{Allowed: true, PolicyID: p.ID}
		e.audit(ctx, req, result, p.Name)
		return result
	}

	// No matching policy — allow by default.
	result := EvalResult{Allowed: true}
	e.audit(ctx, req, result, "")
	return result
}

// evalCondition evaluates a single AccessCondition.  Returns (true, reason) on
// denial, (false, "") on pass.
func (e *Engine) evalCondition(cond AccessCondition, req EvalRequest, now time.Time) (bool, string) {
	switch cond.Type {
	case ConditionTimeOfDay:
		if cond.TimeWindow == nil {
			return false, "" // misconfigured but non-blocking
		}
		return evalTimeWindow(cond.TimeWindow, now)

	case ConditionGeo:
		if cond.GeoRestriction == nil {
			return false, ""
		}
		return evalGeo(cond.GeoRestriction, req.GeoRegion)

	case ConditionApprovalGate:
		if cond.ApprovalGate == nil {
			return false, ""
		}
		return e.evalApproval(req.ApprovalToken)

	default:
		// Unknown condition types are treated as non-blocking to allow forward
		// compatibility.
		slog.Warn("unknown policy condition type, skipping",
			slog.String("type", string(cond.Type)))
		return false, ""
	}
}

// evalTimeWindow checks the request time against the allowed window.
func evalTimeWindow(tw *TimeWindow, now time.Time) (bool, string) {
	hour := now.UTC().Hour()
	day := now.UTC().Weekday()

	// Check allowed days.
	if len(tw.AllowedDays) > 0 {
		allowed := false
		for _, d := range tw.AllowedDays {
			if d == day {
				allowed = true
				break
			}
		}
		if !allowed {
			return true, fmt.Sprintf("access denied: day %s not in allowed days", day)
		}
	}

	// Check hour window (handles midnight wrap).
	inWindow := false
	if tw.StartHour <= tw.EndHour {
		inWindow = hour >= tw.StartHour && hour <= tw.EndHour
	} else {
		// Wraps midnight.
		inWindow = hour >= tw.StartHour || hour <= tw.EndHour
	}

	if !inWindow {
		return true, fmt.Sprintf("access denied: hour %d outside allowed window [%d–%d]",
			hour, tw.StartHour, tw.EndHour)
	}
	return false, ""
}

// evalGeo checks the caller's geo region against the allowed list.
func evalGeo(gr *GeoRestriction, callerRegion string) (bool, string) {
	if len(gr.AllowedRegions) == 0 {
		return false, "" // empty allow-list = unrestricted
	}
	for _, r := range gr.AllowedRegions {
		if r == callerRegion {
			return false, ""
		}
	}
	return true, fmt.Sprintf("access denied: region %q not in allowed regions %v",
		callerRegion, gr.AllowedRegions)
}

// evalApproval checks for a valid, non-expired approval token.
func (e *Engine) evalApproval(token string) (bool, string) {
	if token == "" {
		return true, "access denied: approval gate requires a valid approval token"
	}
	rec, ok := e.approvals[token]
	if !ok {
		return true, "access denied: approval token not found"
	}
	if time.Now().After(rec.expiresAt) {
		return true, "access denied: approval token has expired"
	}
	return false, ""
}

// audit emits a structured audit log entry for every policy evaluation.
func (e *Engine) audit(_ context.Context, req EvalRequest, result EvalResult, policyName string) {
	attrs := []any{
		slog.String("event", "policy_evaluation"),
		slog.String("workload_spiffe", req.WorkloadSPIFFE),
		slog.String("service_spiffe", req.ServiceSPIFFE),
		slog.Bool("allowed", result.Allowed),
		slog.String("policy_id", result.PolicyID),
		slog.String("policy_name", policyName),
	}
	if !result.Allowed {
		attrs = append(attrs, slog.String("deny_reason", result.Reason))
	}
	slog.Info("policy audit", attrs...)
}

// spiffeMatches checks whether candidate equals pattern, or pattern is "*"
// (wildcard).
func spiffeMatches(pattern, candidate string) bool {
	return pattern == "*" || pattern == candidate
}

// safePrefix returns the first n characters of s (or all of s if shorter).
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
