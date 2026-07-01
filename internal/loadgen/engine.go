package loadgen

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adssib/dlt/internal/protocol"
)

// Result is the tally a worker produces from one run. The latency histogram
// (Phase 4) will be added here; Phase 2 reports basic counts + throughput.
type Result struct {
	Total      int64
	Successful int64
	Failed     int64
	Throttled  int64
	Duration   time.Duration
}

// Engine generates HTTP load under bounded concurrency, reusing connections.
type Engine struct {
	client *http.Client
	// TODO(you): whatever state your generator needs (semaphore lives in Run).
}

// New builds an Engine.
//
// TODO(you): tune the http.Client / Transport — connection reuse (keep-alive),
// MaxIdleConnsPerHost, and per-request timeout matter a lot for load generation.
func New() *Engine {
	return &Engine{client: &http.Client{}}
}

// Run fires cfg.RequestsPerWorker requests at cfg.TargetURL, capping in-flight
// requests at cfg.ConcurrencyPerWorker (F4), timing each one (F5), and streaming
// live Progress on the channel (F6). It returns the aggregated Result.
//
// TODO(you): implement the bounded-concurrency generator. The semaphore
// (a buffered channel or golang.org/x/sync/semaphore) is the core of this.
func (e *Engine) Run(ctx context.Context, cfg protocol.TestConfig, progress chan<- protocol.Progress) (Result, error) {
	return Result{}, fmt.Errorf("loadgen.Engine.Run: not implemented (TODO you)")
}
