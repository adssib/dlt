// Package target implements the deliberately-breakable HTTP server that the load
// tester fires at — the "system under test" / defender.
package target

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/adssib/dlt/internal/config"
)

// Server is the target HTTP server. Request handling is wired up here; the
// per-request decisions (latency, faults) live in Behavior.
type Server struct {
	cfg      config.TargetConfig
	behavior *Behavior
	// TODO(you): Phase 9 — add a rate limiter here (ratelimit.Limiter).
}

// NewServer constructs a target server from config.
func NewServer(cfg config.TargetConfig) *Server {
	return &Server{
		cfg:      cfg,
		behavior: NewBehavior(cfg),
	}
}

// Run starts the HTTP server and blocks until ctx is cancelled, then shuts down
// gracefully. (Plumbing — ready to use as-is.)
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/healthz", s.handleHealth)
	// TODO(you): Phase 8 — mux.Handle("/metrics", promhttp.Handler()).

	srv := &http.Server{Addr: s.cfg.Target.Listen, Handler: mux}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// handleRoot applies the target's behavior to each request: count it in, wait the
// modeled latency, then either fail or succeed. The *decisions* live in Behavior
// (implement those); this orchestration is ready.
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	inflight := s.behavior.Enter()
	defer s.behavior.Leave()

	// TODO(you): honor r.Context() cancellation while waiting (don't sleep past
	// a client disconnect).
	if d := s.behavior.LatencyFor(inflight); d > 0 {
		time.Sleep(d)
	}

	// TODO(you): Phase 9 — check the rate limiter here; return 429 + Retry-After
	// if throttled (this is distinct from a 5xx failure).

	if status := s.behavior.FaultStatus(inflight, time.Now()); status != 0 {
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// handleHealth is a always-fast liveness endpoint (not subject to the behavior
// model), handy for container/pod health checks.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
