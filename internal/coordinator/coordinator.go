// Package coordinator is the control plane: it accepts worker connections, holds
// the registry, releases the synchronized start barrier, collects progress and
// results, and renders the report. It is a stateful singleton (ADR-0007).
package coordinator

import (
	"context"
	"fmt"

	"github.com/adssib/dlt/internal/config"
)

// Coordinator owns the run lifecycle for a single test.
type Coordinator struct {
	cfg      *config.CoordinatorConfig
	registry *Registry
}

// New builds a Coordinator from config.
func New(cfg *config.CoordinatorConfig) *Coordinator {
	return &Coordinator{cfg: cfg, registry: NewRegistry()}
}

// Run listens on cfg.Coordinator.Listen, accepts workers until the min_workers
// barrier is met (F1/F3), plans + broadcasts StartTest (F2), collects Progress
// (F6) and Results (F7), then prints the report. Honors ctx cancellation.
//
// TODO(you): implement the accept loop + coordination lifecycle. Sketch:
//  1. net.Listen; accept connections in a goroutine, each speaks protocol.Conn
//  2. handle Register -> registry.Add; wait until Count() >= MinWorkers (or timeout)
//  3. Plan the workload across registered workers; broadcast StartTest
//  4. read Progress/Results per worker; on worker death, keep survivors (F11)
//  5. Aggregate + Render the report
func (c *Coordinator) Run(ctx context.Context) error {
	return fmt.Errorf("coordinator.Run: not implemented (TODO you)")
}
