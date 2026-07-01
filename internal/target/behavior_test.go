package target

import (
	"testing"
	"time"

	"github.com/adssib/dlt/internal/config"
)

// TestBehavior_LatencyFor is our first real TDD test.
//
// Simplest slice of the latency model: with jitter turned off, no fat-tail, and
// the server NOT overloaded (inflight <= capacity), LatencyFor should return
// exactly the configured base latency — nothing added, nothing random.
//
// It fails right now on purpose (LatencyFor returns 0). Your job is to make it
// pass by implementing the base case. Then we add the next failing test.
func TestBehavior_LatencyFor(t *testing.T) {
	// ---- Arrange: a config with every random/extra knob switched off ----
	var cfg config.TargetConfig
	cfg.Target.Seed = 1 // fixed seed = reproducible (no rolls happen here anyway)
	cfg.Target.Latency.Base = config.Duration(100 * time.Millisecond)
	cfg.Target.Latency.Jitter = 0   // no random jitter
	cfg.Target.Concurrency.Capacity = 100
	cfg.Target.Tail.Probability = 0 // never add a fat-tail straggler

	b := NewBehavior(cfg)

	// ---- Act: one request, comfortably under capacity ----
	got := b.LatencyFor(1)

	// ---- Assert: exactly base, because everything else is disabled ----
	want := 100 * time.Millisecond
	if got != want {
		t.Errorf("LatencyFor(1) = %v, want %v", got, want)
	}
}

// TestBehavior_FaultStatus: you already implemented FaultStatus, so a test now
// would pass instantly (tests-after) and prove little. We'll add real regression
// tests for it separately — see the note at the end of this turn.
func TestBehavior_FaultStatus(t *testing.T) {
	t.Skip("regression tests coming after we finish LatencyFor via TDD")
}
