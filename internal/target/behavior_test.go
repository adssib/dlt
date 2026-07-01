package target

import (
	"sync"
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

// latencyConfig returns a config with a fixed base and every source of extra
// latency (jitter, slowdown, tail) switched off. Each sub-test turns on exactly
// one knob. Fixed seed = reproducible.
func latencyConfig() config.TargetConfig {
	var cfg config.TargetConfig
	cfg.Target.Seed = 1
	cfg.Target.Latency.Base = config.Duration(100 * time.Millisecond)
	cfg.Target.Latency.Jitter = 0
	cfg.Target.Latency.Distribution = "normal"
	cfg.Target.Concurrency.Capacity = 100
	cfg.Target.Concurrency.Slowdown = "linear"
	// Tail.Probability / Tail.Extra default to 0.
	return cfg
}

// TestBehavior_LatencyFor pins the documented latency model:
//
//	base + jitter  +  slowdown once inflight > capacity  +  occasional tail
//
// Most assertions are robust *properties*. A few (marked "ASSUMES ...") encode the
// simplest reading of the model — if your implementation differs (e.g. symmetric
// jitter, or a different slowdown curve), adjust that one assertion; the sub-test
// name says exactly what it pins down.
func TestBehavior_LatencyFor(t *testing.T) {
	const base = 100 * time.Millisecond

	t.Run("base only: jitter off, tail off, under capacity -> exactly base", func(t *testing.T) {
		b := NewBehavior(latencyConfig())
		if got := b.LatencyFor(1); got != base {
			t.Errorf("LatencyFor(1) = %v, want %v", got, base)
		}
	})

	t.Run("under capacity, jitter stays within [base, base+jitter]", func(t *testing.T) {
		// ASSUMES additive jitter in [0, jitter]. If you model +/- jitter, widen the lower bound.
		cfg := latencyConfig()
		cfg.Target.Latency.Jitter = config.Duration(20 * time.Millisecond)
		b := NewBehavior(cfg)
		lo, hi := base, base+20*time.Millisecond
		for i := 0; i < 1000; i++ {
			got := b.LatencyFor(1)
			if got < lo || got > hi {
				t.Fatalf("sample %d = %v, want within [%v, %v]", i, got, lo, hi)
			}
		}
	})

	t.Run("jitter produces variation (not constant)", func(t *testing.T) {
		cfg := latencyConfig()
		cfg.Target.Latency.Jitter = config.Duration(20 * time.Millisecond)
		b := NewBehavior(cfg)
		seen := map[time.Duration]bool{}
		for i := 0; i < 200; i++ {
			seen[b.LatencyFor(1)] = true
		}
		if len(seen) < 2 {
			t.Errorf("expected jitter to vary latency, got %d distinct value(s)", len(seen))
		}
	})

	t.Run("same seed -> same sequence (reproducible)", func(t *testing.T) {
		cfg := latencyConfig()
		cfg.Target.Latency.Jitter = config.Duration(20 * time.Millisecond)
		a, b := NewBehavior(cfg), NewBehavior(cfg)
		for i := 0; i < 50; i++ {
			if x, y := a.LatencyFor(1), b.LatencyFor(1); x != y {
				t.Fatalf("call %d: %v != %v — same seed must reproduce", i, x, y)
			}
		}
	})

	t.Run("over capacity costs more than under capacity", func(t *testing.T) {
		b := NewBehavior(latencyConfig()) // jitter 0, slowdown linear, capacity 100
		cap := int64(latencyConfig().Target.Concurrency.Capacity)
		under, over := b.LatencyFor(1), b.LatencyFor(cap*2)
		if over <= under {
			t.Errorf("over capacity (%v) should exceed under capacity (%v)", over, under)
		}
	})

	t.Run("slowdown is monotonic in inflight past capacity", func(t *testing.T) {
		b := NewBehavior(latencyConfig())
		cap := int64(latencyConfig().Target.Concurrency.Capacity)
		less, more := b.LatencyFor(cap+10), b.LatencyFor(cap+200)
		if more < less {
			t.Errorf("more inflight must not reduce latency: LatencyFor(%d)=%v < LatencyFor(%d)=%v",
				cap+200, more, cap+10, less)
		}
	})

	t.Run("slowdown=none ignores capacity -> base", func(t *testing.T) {
		cfg := latencyConfig()
		cfg.Target.Concurrency.Slowdown = "none"
		b := NewBehavior(cfg)
		cap := int64(cfg.Target.Concurrency.Capacity)
		if got := b.LatencyFor(cap * 10); got != base {
			t.Errorf("slowdown=none: LatencyFor(%d) = %v, want base %v", cap*10, got, base)
		}
	})

	t.Run("quadratic slows down more than linear past capacity", func(t *testing.T) {
		// ASSUMES the documented ordering none < linear < quadratic. Adjust if your model differs.
		lin, quad := latencyConfig(), latencyConfig()
		quad.Target.Concurrency.Slowdown = "quadratic"
		bl, bq := NewBehavior(lin), NewBehavior(quad)
		cap := int64(lin.Target.Concurrency.Capacity)
		inflight := cap * 3
		if bq.LatencyFor(inflight) <= bl.LatencyFor(inflight) {
			t.Errorf("quadratic (%v) should exceed linear (%v) at inflight=%d",
				bq.LatencyFor(inflight), bl.LatencyFor(inflight), inflight)
		}
	})

	t.Run("tail extra applies when probability=1", func(t *testing.T) {
		// ASSUMES tail adds up to Extra. Bounds hold for both "fixed extra" and "up to extra".
		cfg := latencyConfig() // jitter 0
		cfg.Target.Tail.Probability = 1.0
		cfg.Target.Tail.Extra = config.Duration(200 * time.Millisecond)
		b := NewBehavior(cfg)
		got := b.LatencyFor(1)
		if got <= base || got > base+200*time.Millisecond {
			t.Errorf("tail prob=1: LatencyFor(1) = %v, want within (%v, %v]", got, base, base+200*time.Millisecond)
		}
	})

	t.Run("no tail when probability=0 -> base", func(t *testing.T) {
		b := NewBehavior(latencyConfig()) // tail prob 0, jitter 0
		for i := 0; i < 200; i++ {
			if got := b.LatencyFor(1); got != base {
				t.Fatalf("tail prob=0 must never add extra: sample %d = %v, want base", i, got)
			}
		}
	})

	t.Run("never returns negative latency", func(t *testing.T) {
		cfg := latencyConfig()
		cfg.Target.Latency.Jitter = config.Duration(20 * time.Millisecond)
		cfg.Target.Tail.Probability = 0.5
		cfg.Target.Tail.Extra = config.Duration(200 * time.Millisecond)
		b := NewBehavior(cfg)
		cap := int64(cfg.Target.Concurrency.Capacity)
		for _, inflight := range []int64{0, 1, cap, cap * 5} {
			for i := 0; i < 200; i++ {
				if got := b.LatencyFor(inflight); got < 0 {
					t.Fatalf("negative latency %v at inflight=%d", got, inflight)
				}
			}
		}
	})
}

// TestBehavior_ConcurrentAccess hammers LatencyFor and FaultStatus from many
// goroutines at once — exactly how the target's HTTP server calls them.
//
// It asserts nothing itself; its whole job is to give the RACE DETECTOR something
// to catch. math/rand.Rand is NOT safe for concurrent use, so until b.rng is
// guarded this fails under:
//
//	go test -race -run TestBehavior_ConcurrentAccess ./internal/target/
func TestBehavior_ConcurrentAccess(t *testing.T) {
	cfg := latencyConfig()
	cfg.Target.Latency.Jitter = config.Duration(5 * time.Millisecond) // LatencyFor rolls rng
	cfg.Target.Concurrency.MaxInflight = 1000                         // so inflight=1 doesn't short-circuit to 503...
	cfg.Target.Faults.ErrorRate = 0.5                                 // ...and FaultStatus reaches its rng roll
	b := NewBehavior(cfg)

	const goroutines, iterations = 64, 250

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			now := time.Now()
			for i := 0; i < iterations; i++ {
				_ = b.LatencyFor(1)
				_ = b.FaultStatus(1, now)
			}
		}()
	}
	wg.Wait()
}
