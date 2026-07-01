# Distributed Load Tester (`dlt`) — Design

> Learning goals: distributed systems hands-on — membership, coordination, correct
> result aggregation, graceful failure — plus the DevOps stack around it (Docker,
> Kubernetes, Helm, Prometheus, Grafana). **Boilerplate is scaffolded; the real
> logic is written by the author.** Every `.go` file ships as types + signatures +
> `// TODO(you)` bodies. No core logic is pre-written.

---

## 1. Components & process model

Two programs, three roles:

- **`dlt`** — one binary, two role subcommands: `dlt coordinator`, `dlt worker`
  (honors "one binary, two modes"). Plus a local convenience launcher `dlt test`.
- **`target`** — a separate program: the deliberately-breakable system-under-test
  (the "defender").

```
  worker ─┐
  worker ─┼──HTTP──> target (the SUT, separate program)
  worker ─┘                │ breakable: latency model, faults, rate limiter
     │                     │
     └── register / progress / results (JSON over TCP) ──> coordinator
```

### Pod / deployment model (Kubernetes)

| Role | k8s shape | Notes |
|---|---|---|
| coordinator | 1 Deployment (1 replica) **+ Service** | Stable DNS (`coordinator:7070`) so workers can find it. |
| worker | 1 Deployment, **N replicas** | Scaled via `kubectl scale`; fungible; no Service (they dial *out*). |
| target | 1 Deployment **+ Service** | The system under test. |

**Who owns "how many workers":** the *orchestrator*, never the coordinator.
- local: `dlt test --workers N`
- compose: `docker compose up --scale worker=N`
- k8s: Deployment `replicas:` / `kubectl scale`

The coordinator only knows `min_workers` — a **readiness barrier** (don't start a run
until N workers have registered). This avoids two-sources-of-truth conflicts with k8s.

---

## 2. Dependencies (deliberately thin — it's a learning project)

- stdlib `net` / `net/http` — TCP transport + HTTP load generation.
- `github.com/HdrHistogram/hdrhistogram-go` — default mergeable histogram.
- `github.com/prometheus/client_golang` — metrics export.
- `gopkg.in/yaml.v3` — config files.
- **`tsenart/vegeta` — reference to read, NOT a dependency.** The load engine is written by hand.

---

## 3. Folder structure

```
dlt/
├── go.mod / go.sum
├── Makefile                  # build, run-coordinator, run-worker, run-target, test-local, docker, k8s targets
├── README.md                 # architecture diagram, aggregation explanation, out-of-scope
├── .gitignore / .dockerignore
│
├── docs/
│   ├── architecture.md
│   └── SPEC.md               # this design doc
│
├── configs/
│   ├── coordinator.yaml      # test definition + coordination
│   ├── worker.yaml           # the simple one: where's the coordinator + capacity
│   └── target.yaml           # the SUT: latency model, faults, rate limiter
│
├── cmd/
│   ├── dlt/main.go           # subcommand dispatch: coordinator | worker | test
│   └── target/main.go        # the breakable target server
│
├── internal/
│   ├── config/               # F10 — YAML → typed config; Duration wrapper
│   │   └── config.go
│   │
│   ├── protocol/             # F2,F6,F7 — wire types + newline-JSON codec over net.Conn
│   │   ├── messages.go       #   Register, Progress, Results, StartTest, StopTest, TestConfig, Envelope
│   │   └── codec.go          #   Conn wrapper: ReadMsg / WriteMsg (bufio, newline-delimited)
│   │
│   ├── coordinator/          # F1,F2,F3,F6,F7,F8,F11
│   │   ├── registry.go       #   membership: track registered workers + capacity
│   │   ├── planner.go        #   split total work (requests/concurrency) across workers
│   │   ├── coordinator.go    #   accept conns, broadcast StartTest (barrier), collect progress/results
│   │   └── report.go         #   merged report + render (text/json) + partial flag
│   │
│   ├── worker/               # F1,F4,F5,F6
│   │   └── worker.go         #   register → await StartTest → run loadgen → stream progress → send Results
│   │
│   ├── loadgen/              # F4,F5,F9 + 429-awareness — THE engine (author-written)
│   │   ├── engine.go         #   bounded-concurrency runner; http.Client w/ connection reuse
│   │   ├── rampup.go         #   F9 linear ramp scheduler
│   │   └── outcome.go        #   classify(resp,err) → Success | Throttled(429) | Failed
│   │
│   ├── histogram/            # F8 — THE centerpiece
│   │   ├── histogram.go      #   Histogram interface: Record/Merge/Quantile/Count/Serialize
│   │   └── hdr.go            #   HDR implementation + Deserialize
│   │
│   ├── metrics/              # F12
│   │   └── metrics.go        #   prometheus collectors + /metrics handler
│   │
│   ├── launcher/             # `dlt test` — local orchestrator (spawns coordinator + N workers)
│   │   └── launcher.go
│   │
│   └── target/               # the defender (SUT)
│       ├── server.go         #   HTTP handlers + /metrics + wiring
│       ├── behavior.go       #   config-driven latency model + fault decisions (the "realism")
│       └── ratelimit/        # Phase 9
│           ├── limiter.go    #   Limiter interface: Allow() bool, RetryAfter() time.Duration
│           ├── tokenbucket.go
│           └── slidingwindow.go
│
├── deploy/
│   ├── compose/              # Phase 6 — docker-compose.yml + Dockerfile.dlt, Dockerfile.target
│   ├── k8s/                  # Phase 7 — namespace, coordinator (Deploy+Svc), worker (Deploy), target (Deploy+Svc), configmaps
│   └── helm/dlt/             # Chart.yaml, values.yaml, templates/
│
└── observability/           # Phase 8
    ├── prometheus/prometheus.yml   # scrape config: coordinator + target (+ optional workers)
    └── grafana/dashboards/         # dashboard JSON + provisioning
```

---

## 4. Key interfaces & types (contracts; bodies are the author's)

### protocol
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

// codec — newline-delimited JSON over net.Conn
type Conn struct { /* bufio-wrapped net.Conn */ }
func (c *Conn) ReadMsg() (Envelope, error)
func (c *Conn) WriteMsg(v any) error
```

### histogram (the centerpiece)
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
Coordinator depends ONLY on the interface → swapping HDR↔t-digest never touches it.

**Aggregation rule (why this isn't a toy):** workers ship *serialized histograms*, the
coordinator `Merge`s them, then computes `Quantile` from the merged structure. Never
average per-worker percentiles — that is statistically invalid.

### loadgen (author-written engine)
```go
type Outcome int
const ( OutcomeSuccess Outcome = iota; OutcomeThrottled; OutcomeFailed )

type RawStats struct { Total, Successful, Failed, Throttled int64 }  // 429 tracked separately
type Result   struct { Stats RawStats; Hist histogram.Histogram; Duration time.Duration }

type Engine struct { /* http.Client w/ keep-alive, semaphore, config */ }
func (e *Engine) Run(ctx context.Context, cfg protocol.TestConfig, progress chan<- protocol.Progress) (Result, error)
func classify(resp *http.Response, err error) Outcome
```

### target rate limiter (Phase 9)
```go
type Limiter interface { Allow() bool; RetryAfter() time.Duration }
func NewTokenBucket(capacity int, refillPerSec float64) Limiter
func NewSlidingWindow(limit int, window time.Duration) Limiter
```

---

## 5. Configuration model (F10)

Three YAML files. CLI is just `-c <path>` per role; the launcher takes both tester configs.

### `configs/coordinator.yaml`
```yaml
coordinator:
  listen: ":7070"
  min_workers: 3           # readiness barrier before a run starts
  wait_for_workers: 15s
test:
  target_url: "http://target:8080/"
  total_requests: 100000   # planner splits across workers → RequestsPerWorker
  total_concurrency: 200   # planner splits across workers → ConcurrencyPerWorker
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
    max_inflight: 200       # hard ceiling → 503 beyond this
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

Durations (`5s`, `10ms`) parse via a small `config.Duration` wrapper type — this is
plumbing, so it gets a fuller stub than the core logic.

---

## 6. CLI / launch model

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

`dlt test` spawns 1 coordinator + N worker processes locally — the local stand-in for
a k8s Deployment. Subcommands (not flags) so they map cleanly to k8s container `args`.

---

## 7. Observability model (F12)

- **Pull-based:** components expose `/metrics`; Prometheus scrapes → Grafana visualizes.
- **Stable scrape targets** (long-lived, have Services):
  - **coordinator** — live aggregate (completed, errors, throttled, current req/s) from the `Progress` stream.
  - **target** — served / throttled(429) / 5xx rates; powers the rate-limiter "flat line at the limit" demo.
  - **workers** — optional per-worker detail via k8s service discovery (not required; coordinator aggregates).
- **Critical boundary:** Prometheus/Grafana = **live** view only. **Authoritative final
  percentiles come from the merged histograms** in the coordinator's report (F7/F8), NOT
  from Prometheus. PromQL cannot do a correct cross-worker percentile merge — that
  limitation is *why* the histogram centerpiece exists. Histograms → correctness;
  Prometheus → live observability.

---

## 8. Requirement → location traceability

| Req | Lives in |
|---|---|
| F1 membership / registration | `coordinator/registry.go`, `worker/worker.go` |
| F2 distribute config | `coordinator/planner.go` + `protocol` |
| F3 synchronized start | `coordinator/coordinator.go` (broadcast barrier) |
| F4 bounded concurrency | `loadgen/engine.go` (semaphore) |
| F5 per-request measurement | `loadgen/engine.go` |
| F6 progress stream | `worker` → `protocol.Progress` → `coordinator` |
| **F7/F8 merge + correct percentiles** | `histogram/` + `coordinator/report.go` |
| F9 ramp-up | `loadgen/rampup.go` |
| F10 configurable (YAML) | `internal/config` + `configs/` |
| F11 graceful worker death | `coordinator/coordinator.go` (survivor set + partial flag) |
| F12 metrics → Prometheus/Grafana | `internal/metrics` + `observability/` |
| Phase 9 rate limiter | `internal/target/ratelimit/` + `loadgen/outcome.go` |
| Realistic target (latency/faults) | `internal/target/behavior.go` |

---

## 9. Build order (each step a working system)

Maps to the SPEC's step list; each is independently demoable.

1. **Target server** — `cmd/target` + `internal/target/{server,behavior}.go` (latency model, faults).
2. **Single-process tester** — coordinator + 1 worker on localhost, basic stats, firing at target.
3. **Multiple workers** — 3+ `dlt worker` registering with one coordinator. Already distributed. (`dlt test` launcher helps here.)
4. **Histogram aggregation** — swap per-worker percentiles for merged histograms. THE differentiator.
5. **Ramp-up + graceful worker failure** — F9 + F11.
6. **Docker Compose** — containerize, `--scale worker=N`, DNS service discovery.
7. **Kubernetes (kind/k3d → home-lab)** — pods, `kubectl scale`, Helm chart.
8. **Observability** — Prometheus scrape + Grafana dashboards.
9. **Rate limiter on target** — token bucket + sliding window; load tester tracks 429s separately and proves the limiter trips. (Independent after step 3.)

---

## 10. Out of scope (deliberate bounds)

- Coordinator↔worker security (plaintext TCP; TLS+auth is a noted next step).
- Protocols other than HTTP (no gRPC/WebSocket targets).
- GUI (CLI + Grafana only).
- Coordinator HA (single coordinator, single point of failure — acceptable).
- Internet-scale load (single-machine resource contention caps throughput).
- Distributed rate limiting (shared limit across target replicas via Redis) — noted extension.
- Per-client / per-key / dynamic / leaky-bucket limits — later.

### Ethics / safety (non-negotiable)
Generates DDoS-shaped traffic; the difference is authorization. Targets **only the
bundled target server on owned/home-lab infrastructure**. Never pointed at third-party
systems without explicit written permission. All traffic stays inside the owner's network.

---

## 11. Boilerplate philosophy

Every `.go` file ships as: package decl, imports, type/struct definitions,
method & function signatures with doc comments, and `// TODO(you): ...` bodies
(returning zero values / `panic("not implemented")`). Infra files (compose/k8s/helm/
prometheus/grafana) ship as structured stubs. **No real logic** — the histogram merge,
the semaphore, the ramp scheduler, the limiters, the target's latency model are all the
author's to write. The scaffold provides a compiling skeleton and a map; the engineering
is the author's.

---

## 12. Definition of done (CV-worthy)

- [ ] Coordinator + multiple workers, running as k8s pods on the home-lab.
- [ ] `kubectl scale` visibly increases generated load.
- [ ] Correct global percentiles via merged histograms (can explain why naive merge is wrong).
- [ ] Deliberately-broken target whose bottleneck the tool detects.
- [ ] Live Grafana dashboard during a run.
- [ ] Rate limiter on the target; load tester proves it trips at the configured rate.
- [ ] README with architecture diagram, aggregation explanation, and out-of-scope list.
