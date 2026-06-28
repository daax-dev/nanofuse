package credisolation

import (
	"fmt"
	"log/slog"
	"time"
)

// AccessAttempt describes a detected attempt by one VM to reach another VM's
// credential store. It is the input to the fail-safe response.
type AccessAttempt struct {
	RequestingVMID string
	TargetVMID     string
	Path           string
	When           time.Time
}

// Terminator terminates the offending VM. In production this is wired to the
// Firecracker manager's Kill; tests inject a recorder. A nil Terminator means
// termination is unavailable (the attempt is still audited).
type Terminator func(vmID string) error

// Monitor implements the fail-safe response to cross-VM credential access
// attempts: it emits a structured audit record and terminates the offending VM.
type Monitor struct {
	log       *slog.Logger
	terminate Terminator
}

// NewMonitor builds a Monitor. A nil logger falls back to slog.Default(). A nil
// terminate disables termination (the attempt is still audited) — callers that
// require fail-safe termination must supply one.
func NewMonitor(logger *slog.Logger, terminate Terminator) *Monitor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Monitor{log: logger, terminate: terminate}
}

// HandleCrossVMAttempt is the fail-safe response invoked by the runtime LSM/
// audit hook when a cross-VM credential access is detected. It records an audit
// event with both VM identifiers and a timestamp, terminates the offending VM,
// and returns an error wrapping ErrCrossVMAccess so callers can react.
//
// Input is validated at this boundary: a malformed identifier is audited as a
// violation rather than trusted.
func (mon *Monitor) HandleCrossVMAttempt(a AccessAttempt) error {
	when := a.When
	if when.IsZero() {
		when = time.Now()
	}

	mon.log.Warn("credential isolation violation detected",
		slog.String("event", "cred_isolation.cross_vm_access_attempt"),
		slog.String("requesting_vm", a.RequestingVMID),
		slog.String("target_vm", a.TargetVMID),
		slog.String("path", a.Path),
		slog.Time("timestamp", when),
	)

	var termErr error
	terminated := false
	if mon.terminate != nil {
		if termErr = mon.terminate(a.RequestingVMID); termErr == nil {
			terminated = true
		}
	}

	mon.log.Error("credential isolation fail-safe response",
		slog.String("event", "cred_isolation.vm_terminated"),
		slog.String("requesting_vm", a.RequestingVMID),
		slog.Bool("terminated", terminated),
	)

	if termErr != nil {
		return fmt.Errorf("%w: terminate offending VM %q: %v",
			ErrCrossVMAccess, a.RequestingVMID, termErr)
	}
	if mon.terminate == nil {
		return fmt.Errorf("%w: VM %q attempted to access VM %q credentials (no terminator configured)",
			ErrCrossVMAccess, a.RequestingVMID, a.TargetVMID)
	}
	return fmt.Errorf("%w: VM %q terminated after attempting to access VM %q credentials",
		ErrCrossVMAccess, a.RequestingVMID, a.TargetVMID)
}
