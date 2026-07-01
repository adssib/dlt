package coordinator

import "github.com/adssib/dlt/internal/protocol"

// Registry tracks the workers that have registered for the current run. It is
// touched from multiple connection goroutines, so it must be concurrency-safe.
//
// TODO(you): choose the internal representation (map keyed by WorkerID?) and the
// guard (sync.Mutex). Remember the rng-race lesson from the target.
type Registry struct {
	// TODO(you)
}

// NewRegistry builds an empty Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Add records a newly-registered worker and returns the current worker count.
//
// TODO(you)
func (r *Registry) Add(reg protocol.Register) int {
	return 0 // TODO(you)
}

// Count returns how many workers are currently registered.
//
// TODO(you)
func (r *Registry) Count() int {
	return 0 // TODO(you)
}
