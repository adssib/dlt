// Package worker is one load generator: it registers with the coordinator, waits
// for the start barrier, runs the load engine against the target, streams
// progress, and sends final results. Workers are stateless and fungible (ADR-0007).
package worker

import (
	"context"
	"fmt"

	"github.com/adssib/dlt/internal/config"
)

// Worker runs one node's share of a test.
type Worker struct {
	cfg *config.WorkerConfig
}

// New builds a Worker from config.
func New(cfg *config.WorkerConfig) *Worker {
	return &Worker{cfg: cfg}
}

// Run dials the coordinator, registers (F1), waits for StartTest (F3), runs the
// load engine (F4/F5), streams Progress (F6), and sends final Results (F7).
// Honors ctx cancellation.
//
// TODO(you): implement the worker lifecycle. Sketch:
//  1. net.Dial the coordinator; wrap in protocol.Conn
//  2. WriteMsg(Register{...})
//  3. ReadMsg loop: on StartTest, run loadgen.Engine.Run, forwarding progress
//  4. WriteMsg(Results{...}) when the run finishes
func (w *Worker) Run(ctx context.Context) error {
	return fmt.Errorf("worker.Run: not implemented (TODO you)")
}
