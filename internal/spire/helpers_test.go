package spire

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
)

// fakeClock is a deterministic Clock for tests. After registers a waiter that
// fires when Advance moves time to or past its deadline.
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []*fakeWaiter
}

type fakeWaiter struct {
	deadline time.Time
	ch       chan time.Time
}

func newFakeClock(start time.Time) *fakeClock { return &fakeClock{now: start} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1)
	if d <= 0 {
		ch <- c.now
		return ch
	}
	c.waiters = append(c.waiters, &fakeWaiter{deadline: c.now.Add(d), ch: ch})
	return ch
}

// Advance moves the clock forward and fires any waiters whose deadline passed.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	now := c.now
	var remaining, fire []*fakeWaiter
	for _, w := range c.waiters {
		if !w.deadline.After(now) {
			fire = append(fire, w)
		} else {
			remaining = append(remaining, w)
		}
	}
	c.waiters = remaining
	c.mu.Unlock()
	for _, w := range fire {
		w.ch <- now
	}
}

// blockUntilWaiters waits until at least n waiters are registered, or fails.
func (c *fakeClock) blockUntilWaiters(n int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		c.mu.Lock()
		count := len(c.waiters)
		c.mu.Unlock()
		if count >= n {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %d clock waiters (have %d)", n, count)
		}
		time.Sleep(time.Millisecond)
	}
}

// failingSource always reports SPIRE as unreachable.
type failingSource struct{}

func (failingSource) FetchSVID(_ context.Context) (*SVID, error) {
	return nil, fmt.Errorf("dial workload api: %w", ErrSPIREUnavailable)
}

// flakySource fails for the first failCount calls, then delegates to inner.
type flakySource struct {
	mu        sync.Mutex
	failCount int
	calls     int
	inner     Source
}

func (f *flakySource) FetchSVID(ctx context.Context) (*SVID, error) {
	f.mu.Lock()
	f.calls++
	shouldFail := f.calls <= f.failCount
	f.mu.Unlock()
	if shouldFail {
		return nil, fmt.Errorf("transient: %w", ErrSPIREUnavailable)
	}
	return f.inner.FetchSVID(ctx)
}

// countingSource records how many times FetchSVID was called.
type countingSource struct {
	mu    sync.Mutex
	calls int
	inner Source
}

func (c *countingSource) FetchSVID(ctx context.Context) (*SVID, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return c.inner.FetchSVID(ctx)
}

func (c *countingSource) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// testSPIFFEID builds a D025-format SPIFFE ID using the production builder so
// tests stay aligned with the registration path and credisolation per-VM
// identity distinctness.
func testSPIFFEID(group, user, vmID string) string {
	svc := NewService(&config.SPIREConfig{TrustDomain: "poley.dev", WorkloadType: "microvm"})
	return svc.BuildSPIFFEID(group, user, vmID)
}
