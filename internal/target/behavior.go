package target

import (
	"math/rand"
	"sync/atomic"
	"sync"
	"time"

	"github.com/adssib/dlt/internal/config"
)

// Behavior turns the target's config knobs into per-request decisions: how long
// a request should take, and whether it should fail. This is the "realistic SUT"
// brain — implementing these methods is what makes the target behave like a
// real, imperfect server (and gives the load tester something worth detecting).
//
// The skeleton ships with no-op implementations (zero latency, no faults) so the
// server runs immediately. Fill in the TODOs to make it interesting.
type Behavior struct {
	cfg      config.TargetConfig
	rng      *rand.Rand
	inflight atomic.Int64
	start    time.Time
	mu 		 sync.Mutex
	// TODO(you): any extra state the spike window or slowdown model needs.
	// Note: rng is not safe for concurrent use — guard it (mutex) or use a
	// per-request source when you implement the random bits.
}

func (b *Behavior) rollFloat() float64 {
    b.mu.Lock()              // 1. grab the lock NOW
    defer b.mu.Unlock()      // 2. SCHEDULE the unlock for when this function returns (doesn't run yet)
    return b.rng.Float64()   // 3. do the protected work, THEN the deferred unlock fires, THEN we return
}

// NewBehavior builds a Behavior from config. A seed of 0 means "random per run"
// (so every run is different); a fixed seed makes runs reproducible.
func NewBehavior(cfg config.TargetConfig) *Behavior {
	seed := cfg.Target.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Behavior{
		cfg:   cfg,
		rng:   rand.New(rand.NewSource(seed)),
		start: time.Now(),
	}
}

// Enter records a newly-arrived in-flight request and returns the current
// in-flight count (use it to drive concurrency-based slowdown).
func (b *Behavior) Enter() int64 { return b.inflight.Add(1) }

// Leave records an in-flight request finishing.
func (b *Behavior) Leave() { b.inflight.Add(-1) }

// LatencyFor returns how long the current request should take, given how many
// requests are in flight. Intended model:
//
//	base + jitter                          (latency.base, latency.jitter)
//	+ slowdown once inflight > capacity     (concurrency.capacity / .slowdown)
//	+ occasional fat-tail straggler         (tail.probability / tail.extra)
//
// TODO(you): implement the latency model. Returning 0 = instant response.
func (b *Behavior) LatencyFor(inflight int64) time.Duration {
	jitterDuration := b.rollFloat() * float64(b.cfg.Target.Latency.Jitter.Std())
	var concurrentSlowdown time.Duration
	var tailExtra time.Duration
	
	unit :=  b.cfg.Target.Latency.Base.Std()

	if (b.rollFloat()  < b.cfg.Target.Tail.Probability){
		tailExtra = b.cfg.Target.Tail.Extra.Std()
	}

	if (inflight > int64(b.cfg.Target.Concurrency.Capacity)){
		// if more than capacity we will add extra latency
		overload := inflight - int64(b.cfg.Target.Concurrency.Capacity)
		ratio := float64(overload) / float64(b.cfg.Target.Concurrency.Capacity)

		switch b.cfg.Target.Concurrency.Slowdown {
			case "linear":
				concurrentSlowdown = time.Duration(float64(unit) * ratio)
			case "quadratic":
				concurrentSlowdown = time.Duration(float64(unit)*float64(ratio)*float64(ratio))
		}
	}

	return b.cfg.Target.Latency.Base.Std() + time.Duration(jitterDuration) + concurrentSlowdown + tailExtra
}

// FaultStatus returns 0 when the request should succeed, or an HTTP status code
// to fail with. Intended model:
//
//	inflight > max_inflight  -> 503  (overloaded)
//	inside a fault spike     -> 5xx  (faults.spike)
//	steady-state error_rate  -> 5xx  (faults.error_rate)
//
// TODO(you): implement overload + spike + steady-state error logic.
// Returning 0 = no fault.
func (b *Behavior) FaultStatus(inflight int64, now time.Time) int {

	if (inflight > int64(b.cfg.Target.Concurrency.MaxInflight)) {
		return 503
	}

	if (b.rollFloat() < b.cfg.Target.Faults.ErrorRate) {
		return 503 // this is for random error rates
	}

	elapsed := now.Sub(b.start)

	spike_every := b.cfg.Target.Faults.Spike.Every.Std()
	if (spike_every != 0){
		positionInCycle := elapsed % spike_every

		if (positionInCycle < b.cfg.Target.Faults.Spike.Duration.Std()){
			if (b.rollFloat()  < b.cfg.Target.Faults.Spike.ErrorRate){
				return 503
			}
		}
	}

	return 0
}
