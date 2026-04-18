package policy

import (
	"context"
	"testing"
	"time"
)

// helpers -------------------------------------------------------------------

func mustAddPolicy(t *testing.T, e *Engine, p *WorkloadPolicy) {
	t.Helper()
	if err := e.AddPolicy(p); err != nil {
		t.Fatalf("AddPolicy(%s): %v", p.ID, err)
	}
}

func allowedPolicy(id, workload, service string) *WorkloadPolicy {
	return &WorkloadPolicy{
		ID:             id,
		Name:           id,
		WorkloadSPIFFE: workload,
		ServiceSPIFFE:  service,
		Enabled:        true,
	}
}

// TestNewEngine verifies the engine starts with no policies.
func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
}

// TestAddPolicyRequiresID verifies that a policy without an ID is rejected.
func TestAddPolicyRequiresID(t *testing.T) {
	e := NewEngine()
	err := e.AddPolicy(&WorkloadPolicy{Name: "no-id", Enabled: true})
	if err == nil {
		t.Error("expected error for missing policy ID, got nil")
	}
}

// TestEvalNoPolicy verifies that a request with no matching policy is allowed.
func TestEvalNoPolicy(t *testing.T) {
	e := NewEngine()
	result := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/workload/a",
		ServiceSPIFFE:  "spiffe://example.com/service/b",
	})
	if !result.Allowed {
		t.Errorf("expected allow with no matching policy, got denied: %s", result.Reason)
	}
	if result.PolicyID != "" {
		t.Errorf("expected empty PolicyID, got %q", result.PolicyID)
	}
}

// TestEvalDisabledPolicy verifies that disabled policies are skipped.
func TestEvalDisabledPolicy(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             "disabled-1",
		WorkloadSPIFFE: "spiffe://example.com/workload/a",
		ServiceSPIFFE:  "spiffe://example.com/service/b",
		Enabled:        false,
		Conditions: []AccessCondition{
			{
				Type: ConditionGeo,
				GeoRestriction: &GeoRestriction{
					AllowedRegions: []string{"US"},
				},
			},
		},
	})
	result := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/workload/a",
		ServiceSPIFFE:  "spiffe://example.com/service/b",
		GeoRegion:      "DE", // would be denied if policy were active
	})
	if !result.Allowed {
		t.Errorf("disabled policy should not deny: %s", result.Reason)
	}
}

// TestEvalPolicyAllowed verifies a matching policy with no conditions allows.
func TestEvalPolicyAllowed(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, allowedPolicy("p1",
		"spiffe://example.com/workload/a",
		"spiffe://example.com/service/b",
	))
	result := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/workload/a",
		ServiceSPIFFE:  "spiffe://example.com/service/b",
	})
	if !result.Allowed {
		t.Errorf("expected allow, got denied: %s", result.Reason)
	}
	if result.PolicyID != "p1" {
		t.Errorf("expected PolicyID p1, got %q", result.PolicyID)
	}
}

// TestEvalWildcardWorkload verifies wildcard service matching.
func TestEvalWildcardWorkload(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, allowedPolicy("wild", "*", "*"))
	result := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/anything",
		ServiceSPIFFE:  "spiffe://example.com/anything-else",
	})
	if !result.Allowed {
		t.Errorf("wildcard policy should allow: %s", result.Reason)
	}
}

// TestEvalTimeWindow tests time-of-day condition enforcement.
func TestEvalTimeWindow(t *testing.T) {
	e := NewEngine()
	const policyID = "tw-1"
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             policyID,
		WorkloadSPIFFE: "*",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []AccessCondition{
			{
				Type: ConditionTimeOfDay,
				TimeWindow: &TimeWindow{
					StartHour: 8,
					EndHour:   17,
				},
			},
		},
	})

	req := EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
	}

	// In window.
	req.Now = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	r := e.Evaluate(context.Background(), req)
	if !r.Allowed {
		t.Errorf("hour 12 should be in window [8–17], got denied: %s", r.Reason)
	}

	// Outside window.
	req.Now = time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	r = e.Evaluate(context.Background(), req)
	if r.Allowed {
		t.Errorf("hour 20 should be outside window [8–17], got allowed")
	}
}

// TestEvalGeoRestriction tests geo-based access control.
func TestEvalGeoRestriction(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             "geo-1",
		WorkloadSPIFFE: "*",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []AccessCondition{
			{
				Type:           ConditionGeo,
				GeoRestriction: &GeoRestriction{AllowedRegions: []string{"US", "CA"}},
			},
		},
	})

	// Allowed region.
	r := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
		GeoRegion:      "US",
	})
	if !r.Allowed {
		t.Errorf("US should be allowed: %s", r.Reason)
	}

	// Denied region.
	r = e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
		GeoRegion:      "CN",
	})
	if r.Allowed {
		t.Errorf("CN should be denied")
	}
}

// TestEvalApprovalGate tests the approval gate condition.
func TestEvalApprovalGate(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             "gate-1",
		WorkloadSPIFFE: "*",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []AccessCondition{
			{
				Type:         ConditionApprovalGate,
				ApprovalGate: &ApprovalGate{MaxValiditySeconds: 3600},
			},
		},
	})

	req := EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
	}

	// No token — denied.
	r := e.Evaluate(context.Background(), req)
	if r.Allowed {
		t.Error("no approval token should be denied")
	}

	// Grant approval.
	e.GrantApproval("my-token", 3600)
	req.ApprovalToken = "my-token"

	// With valid token — allowed.
	r = e.Evaluate(context.Background(), req)
	if !r.Allowed {
		t.Errorf("valid approval token should be allowed: %s", r.Reason)
	}
}

// TestEvalApprovalExpired verifies that expired approval tokens are rejected.
func TestEvalApprovalExpired(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             "gate-exp",
		WorkloadSPIFFE: "*",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []AccessCondition{
			{
				Type:         ConditionApprovalGate,
				ApprovalGate: &ApprovalGate{},
			},
		},
	})

	// Grant a token with 1-second validity and wait for it to expire.
	e.GrantApproval("exp-token", 1)
	time.Sleep(2 * time.Second)

	r := e.Evaluate(context.Background(), EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
		ApprovalToken:  "exp-token",
	})
	if r.Allowed {
		t.Error("expired approval token should be denied")
	}
}

// TestRemovePolicy verifies that removing a policy stops it from matching.
func TestRemovePolicy(t *testing.T) {
	e := NewEngine()
	mustAddPolicy(t, e, &WorkloadPolicy{
		ID:             "rm-1",
		WorkloadSPIFFE: "*",
		ServiceSPIFFE:  "*",
		Enabled:        true,
		Conditions: []AccessCondition{
			{
				Type:           ConditionGeo,
				GeoRestriction: &GeoRestriction{AllowedRegions: []string{"US"}},
			},
		},
	})

	req := EvalRequest{
		WorkloadSPIFFE: "spiffe://example.com/w",
		ServiceSPIFFE:  "spiffe://example.com/s",
		GeoRegion:      "DE",
	}

	// Should be denied before removal.
	r := e.Evaluate(context.Background(), req)
	if r.Allowed {
		t.Error("should be denied before removal")
	}

	e.RemovePolicy("rm-1")

	// Should be allowed after removal (no matching policy).
	r = e.Evaluate(context.Background(), req)
	if !r.Allowed {
		t.Errorf("should be allowed after removal: %s", r.Reason)
	}
}
