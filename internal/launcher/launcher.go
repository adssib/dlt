// Package launcher implements `dlt test`: the local stand-in for a Kubernetes
// Deployment. It starts one coordinator and N workers on this machine so a full
// run can be driven with a single command.
package launcher

import (
	"context"
	"fmt"
)

// Run starts 1 coordinator + `workers` worker processes locally, waits for the
// run to finish, and surfaces the coordinator's report. This is the orchestrator
// role: it owns the worker count (ADR-0003).
//
// TODO(you): spawn the coordinator + workers. Two viable approaches:
//   - os/exec re-invoking this same binary ("dlt coordinator" / "dlt worker"),
//     which most closely mirrors the k8s multi-process model, or
//   - in-process goroutines calling coordinator.New(...).Run / worker.New(...).Run.
func Run(ctx context.Context, coordinatorConfig, workerConfig string, workers int) error {
	return fmt.Errorf("launcher.Run: not implemented (TODO you)")
}
