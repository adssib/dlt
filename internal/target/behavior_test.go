package target

import (
	"testing"
	"time"

	"github.com/adssib/dlt/internal/config"
)

// benignConfig returns a target config where every fault knob is OFF, so any
// fault a test sees comes only from the knob that test deliberately turns on.
// Fixed seed = reproducible (matters once we rely on the random rolls).
func benignConfig() config.TargetConfig {
	var cfg config.TargetConfig
	cfg.Target.Seed = 1
	cfg.Target.Concurrency.MaxInflight = 10
	cfg.Target.Faults.ErrorRate = 0
	cfg.Target.Faults.Spike.Every = config.Duration(30 * time.Second)
	cfg.Target.Faults.Spike.Duration = config.Duration(2 * time.Second)
	cfg.Target.Faults.Spike.ErrorRate = 0
	return cfg
}

// TestBehavior_FaultStatus probes each branch of the fault logic you wrote.
// Each sub-test turns on exactly one knob so the result is deterministic
// (rates of 0 never fire, rates of 1 always fire — no luck involved).
func TestBehavior_FaultStatus(t *testing.T) {
	t.Run("inflight over max_inflight -> 503", func(t *testing.T) {
		cfg := benignConfig() // MaxInflight = 10
		b := NewBehavior(cfg)
		if got := b.FaultStatus(11, time.Now()); got != 503 {
			t.Errorf("FaultStatus(11) = %d, want 503 (overloaded)", got)
		}
	})

	t.Run("exactly at capacity, no faults -> 0", func(t *testing.T) {
		cfg := benignConfig() // MaxInflight = 10, all rates 0
		b := NewBehavior(cfg)
		if got := b.FaultStatus(10, time.Now()); got != 0 {
			t.Errorf("FaultStatus(10) = %d, want 0 (== capacity is allowed)", got)
		}
	})

	t.Run("steady error_rate = 1 -> 503", func(t *testing.T) {
		cfg := benignConfig()
		cfg.Target.Faults.ErrorRate = 1.0 // always fails
		b := NewBehavior(cfg)
		if got := b.FaultStatus(1, time.Now()); got != 503 {
			t.Errorf("FaultStatus(1) = %d, want 503 (steady error)", got)
		}
	})

	t.Run("inside spike window, spike rate = 1 -> 503", func(t *testing.T) {
		cfg := benignConfig()
		cfg.Target.Faults.Spike.ErrorRate = 1.0 // always fails during spike
		b := NewBehavior(cfg)
		now := b.start // elapsed = 0, position 0 < duration => inside the window
		if got := b.FaultStatus(1, now); got != 503 {
			t.Errorf("FaultStatus inside spike = %d, want 503", got)
		}
	})

	t.Run("outside spike window -> 0", func(t *testing.T) {
		cfg := benignConfig()
		cfg.Target.Faults.Spike.ErrorRate = 1.0 // would fail, but we're outside it
		b := NewBehavior(cfg)
		now := b.start.Add(15 * time.Second) // 15s % 30s = 15s, not < 2s => outside
		if got := b.FaultStatus(1, now); got != 0 {
			t.Errorf("FaultStatus outside spike = %d, want 0", got)
		}
	})

	// Edge case: what if the spike is disabled? A config with no spike leaves
	// Spike.Every at its zero value (0). recover() catches a panic so this test
	// reports it instead of crashing the whole run.
	t.Run("spike disabled (every = 0) must not panic", func(t *testing.T) {
		cfg := benignConfig()
		cfg.Target.Faults.Spike.Every = 0
		b := NewBehavior(cfg)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FaultStatus panicked with spike.every=0: %v", r)
			}
		}()
		b.FaultStatus(1, time.Now())
	})
}

// TestBehavior_LatencyFor: parked until you implement LatencyFor (still returns 0).
// We'll drive it with real red-green TDD when you're ready.
func TestBehavior_LatencyFor(t *testing.T) {
	t.Skip("TODO: TDD LatencyFor next")
}
