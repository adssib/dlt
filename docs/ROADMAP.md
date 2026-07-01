# dlt — Roadmap

> The build order. **Each phase is an independently demoable, working system** — the project
> is always in a runnable state, never a big-bang integration at the end.

## Build order

| # | Phase | Delivers | Requirements |
|---|---|---|---|
| 1 | **Target server** | `cmd/target` + latency model + fault decisions | realistic SUT |
| 2 | **Single-process tester** | coordinator + 1 worker on localhost, firing at the target | F1–F6 (single) |
| 3 | **Multiple workers** | 3+ workers registering with one coordinator — *already distributed* | F1–F6 |
| 4 | **Histogram aggregation** | swap per-worker percentiles for **merged histograms** — the differentiator | **F7/F8** |
| 5 | **Ramp-up + graceful failure** | linear ramp; coordinator finishes on survivors, flags partial | F9, F11 |
| 6 | **Docker Compose** | containerize; `--scale worker=N`; DNS service discovery | packaging |
| 7 | **Kubernetes** | pods, `kubectl scale`, Helm chart (kind/k3d → home-lab) | deployment |
| 8 | **Observability** | Prometheus scrape + Grafana dashboards during a live run | F12 |
| 9 | **Rate limiter on target** | token bucket + sliding window; tester proves it trips at the limit | F13 |

Phase 9 is independent after Phase 3.

## Definition of done (CV-worthy)

- [ ] Coordinator + multiple workers running as k8s pods on the home-lab.
- [ ] `kubectl scale` visibly increases generated load.
- [ ] Correct global percentiles via merged histograms (**can explain why the naive merge is wrong**).
- [ ] A deliberately-broken target whose bottleneck the tool detects.
- [ ] A live Grafana dashboard during a run.
- [ ] Rate limiter on the target; the load tester proves it trips at the configured rate.
- [ ] Docs: architecture diagram, aggregation explanation, out-of-scope, and an ADR trail.

## Current status

- **Phase 1 — in progress.**
  - ✅ Scaffold builds, vets, tests; CI in place.
  - ✅ `target` `FaultStatus` implemented + tested (overload / steady error / spike window /
    disabled-spike edge case all green).
  - ⏳ `target` `LatencyFor` — still a stub (`return 0`); next up via red-green TDD.
  - ⏳ Guard `behavior.rng` for concurrent use (tracked; `go test -race` will catch it).
