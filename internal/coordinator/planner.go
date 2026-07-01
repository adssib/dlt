package coordinator

import (
	"github.com/adssib/dlt/internal/config"
	"github.com/adssib/dlt/internal/protocol"
)

// Plan splits the coordinator's total workload (total_requests, total_concurrency)
// across `workers` registered workers, producing the per-worker TestConfig each
// will run (F2). For Phase 2 (workers == 1) it's the whole workload; the split
// (and handling remainders fairly) is what matters once there are multiple workers.
//
// TODO(you): implement the split. Watch: requests and concurrency that don't
// divide evenly — where does the remainder go?
func Plan(cfg *config.CoordinatorConfig, workers int) []protocol.TestConfig {
	return nil // TODO(you)
}
