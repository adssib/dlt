# dlt — Specification

> The **contract**: what the system must do, the shapes it must speak, and the bounds it
> must respect. *How* it's built lives in [ARCHITECTURE.md](ARCHITECTURE.md); *why* choices
> were made lives in [decisions/](decisions/); *in what order* lives in [ROADMAP.md](ROADMAP.md).

## Functional requirements

| ID | Requirement |
|---|---|
| **F1** | **Membership** — workers self-register with the coordinator and report their capacity. |
| **F2** | **Distribute config** — the coordinator splits total work (requests, concurrency) across registered workers. |
| **F3** | **Synchronized start** — a run begins only after a broadcast start barrier (no worker fires early). |
| **F4** | **Bounded concurrency** — each worker generates load under a hard in-flight ceiling (semaphore). |
| **F5** | **Per-request measurement** — every request's latency and outcome is recorded. |
| **F6** | **Progress stream** — workers stream live progress (completed, errors) to the coordinator during a run. |
| **F7** | **Merge results** — the coordinator merges per-worker results into one global result. |
| **F8** | **Correct percentiles** — global percentiles come from **merged histograms**, never from averaging per-worker percentiles. |
| **F9** | **Ramp-up** — load can ramp linearly to target concurrency over a configured window. |
| **F10** | **Configurable** — all behavior is driven by YAML config, one file per role. |
| **F11** | **Graceful worker death** — if a worker dies mid-run, the coordinator finishes with the survivors and flags the result *partial*. |
| **F12** | **Metrics** — components expose Prometheus metrics; Grafana visualizes a live run. |
| **F13** | **Rate limiter (target)** — the target can enforce a request-rate limit; the tester detects and reports throttling (429) separately from failure. |

## Non-functional requirements

- **Thin dependencies** — stdlib-first; the load engine and aggregation are hand-written (see [ADR-0006](decisions/0006-handwritten-load-engine.md)).
- **Each phase is a working, demoable system** (see [ROADMAP.md](ROADMAP.md)).
- **Deployable** as local processes, Docker Compose, and Kubernetes (Helm).
- **Reproducible runs** — the target accepts a seed so a "broken" run can be replayed.

## Contracts

These are the fixed interfaces between components. Bodies are the author's; the *shapes* are the spec.

### Control-plane protocol (newline-delimited JSON over TCP — [ADR-0004](decisions/0004-json-over-tcp-control-plane.md))

```go
type MsgType string
const (
    MsgRegister  MsgType = "register"
    MsgProgress  MsgType = "progress"
    MsgResults   MsgType = "results"
    MsgStartTest MsgType = "start_test"
    MsgStopTest  MsgType = "stop_test"
)

type Envelope struct { Type MsgType; Payload json.RawMessage }

// Worker -> Coordinator
type Register struct { WorkerID string; MaxConcurrency int }
type Progress struct { TestID string; Completed, Errors int64 }
type Results  struct { TestID string; Histogram []byte;   // serialized, NOT percentiles
                       Total, Successful, Failed, Throttled int64; DurationMS int64 }
// Coordinator -> Worker
type StartTest struct { TestID string; Config TestConfig }
type StopTest  struct { TestID string }

type TestConfig struct {
    TargetURL            string
    RequestsPerWorker    int
    ConcurrencyPerWorker int
    TimeoutMS            int
    RampUpSeconds        int
}

// codec
type Conn struct { /* bufio-wrapped net.Conn */ }
func (c *Conn) ReadMsg() (Envelope, error)
func (c *Conn) WriteMsg(v any) error
```

### Histogram (the centerpiece — [ADR-0002](decisions/0002-merged-histograms-over-averaged-percentiles.md))

```go
type Histogram interface {
    Record(latency time.Duration)
    Merge(other Histogram) error
    Quantile(q float64) time.Duration   // q in [0,1]
    Count() int64
    Serialize() ([]byte, error)
}
func NewHDR(min, max time.Duration, sigfigs int) Histogram
func Deserialize(b []byte) (Histogram, error)
```

The coordinator depends **only** on this interface, so swapping HDR ↔ t-digest never touches it.
Workers ship *serialized histograms*; the coordinator `Merge`s them, then computes `Quantile`
from the merged structure. **Averaging per-worker percentiles is statistically invalid** — this
is the requirement that separates this project from a toy.

### Load engine (author-written)

```go
type Outcome int
const ( OutcomeSuccess Outcome = iota; OutcomeThrottled; OutcomeFailed )

type RawStats struct { Total, Successful, Failed, Throttled int64 }  // 429 tracked separately
type Result   struct { Stats RawStats; Hist histogram.Histogram; Duration time.Duration }

type Engine struct { /* http.Client w/ keep-alive, semaphore, config */ }
func (e *Engine) Run(ctx context.Context, cfg protocol.TestConfig, progress chan<- protocol.Progress) (Result, error)
func classify(resp *http.Response, err error) Outcome
```

### Target rate limiter (F13)

```go
type Limiter interface { Allow() bool; RetryAfter() time.Duration }
func NewTokenBucket(capacity int, refillPerSec float64) Limiter
func NewSlidingWindow(limit int, window time.Duration) Limiter
```

## Configuration model (F10)

Three YAML files, one per role. CLI is just `-c <path>`; the local launcher takes both tester configs.

### `configs/coordinator.yaml`
```yaml
coordinator:
  listen: ":7070"
  min_workers: 3           # readiness barrier before a run starts
  wait_for_workers: 15s
test:
  target_url: "http://target:8080/"
  total_requests: 100000   # planner splits across workers -> RequestsPerWorker
  total_concurrency: 200   # planner splits across workers -> ConcurrencyPerWorker
  timeout: 5s
  ramp_up: 10s
report:
  format: text             # text | json
```

### `configs/worker.yaml` (the simple one)
```yaml
worker:
  coordinator: "coordinator:7070"
  max_concurrency: 100
  metrics_listen: ":9100"
  # id: optional, defaults to hostname
```

### `configs/target.yaml` (the defender / SUT)
```yaml
target:
  listen: ":8080"
  metrics_listen: ":9101"
  seed: 0                   # 0 = random each run; set for reproducible runs
  latency:
    base: 5ms
    jitter: 3ms
    distribution: normal    # normal | exponential
  concurrency:
    capacity: 50            # comfortable in-flight load
    slowdown: quadratic     # none | linear | quadratic (latency grows past capacity)
    max_inflight: 200       # hard ceiling -> 503 beyond this
  tail:
    probability: 0.01       # fraction of slow stragglers...
    extra: 250ms            # ...extra latency (fattens p99)
  faults:
    error_rate: 0.0         # steady-state 5xx probability
    spike: { every: 30s, duration: 2s, error_rate: 0.3 }  # periodic "incident" window
  ratelimit:
    enabled: false
    algorithm: token_bucket # token_bucket | sliding_window
    rate: 100               # req/s
    burst: 100              # token-bucket capacity
```

Durations (`5s`, `10ms`) parse via a small `config.Duration` wrapper (plumbing).

## CLI / launch model

```
# pod entrypoints (what compose/k8s run):
dlt coordinator -c configs/coordinator.yaml
dlt worker      -c configs/worker.yaml
target          -c configs/target.yaml

# local convenience (plays the "Deployment" role on your machine):
dlt test --coordinator-config configs/coordinator.yaml \
         --worker-config      configs/worker.yaml \
         --workers 3
```

Role subcommands (not flags) so they map cleanly to container `args` in k8s.

## Out of scope (deliberate bounds)

- Coordinator ↔ worker security (plaintext TCP; TLS + auth is a noted next step).
- Non-HTTP targets (no gRPC / WebSocket).
- GUI (CLI + Grafana only).
- Coordinator HA (single coordinator = single point of failure — accepted).
- Internet-scale load (single-machine resource contention caps throughput).
- Distributed rate limiting (shared limit across target replicas via Redis) — noted extension.
- Per-client / per-key / dynamic / leaky-bucket limits — later.

## Ethics & safety (non-negotiable)

This tool generates DDoS-shaped traffic; the only difference between a load test and an
attack is **authorization**. It targets **only the bundled `target` server on owned /
home-lab infrastructure**, is **never** pointed at third-party systems without explicit
written permission, and all traffic **stays inside the owner's network**.
