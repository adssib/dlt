# Architecture Decision Records (ADRs)

Each file records **one decision**: the context that forced it, the options weighed, the
choice, and what it costs. This trail is deliberate — it shows *how* the system was reasoned
about, not just how it ended up. That reasoning is the point.

**Adding one:** copy [`0000-adr-template.md`](0000-adr-template.md) to the next number, fill
it in, and add a row below. When a decision changes, write a *new* ADR and mark the old one
**Superseded** — never rewrite history.

| ADR | Decision | Status |
|---|---|---|
| [0001](0001-one-binary-two-roles.md) | One `dlt` binary with role subcommands; separate `target` program | Accepted |
| [0002](0002-merged-histograms-over-averaged-percentiles.md) | Global percentiles via merged histograms, never averaged per-worker | Accepted |
| [0003](0003-orchestrator-owns-worker-count.md) | Orchestrator owns worker count; coordinator owns only the readiness barrier | Accepted |
| [0004](0004-json-over-tcp-control-plane.md) | Newline-delimited JSON over TCP for the control plane | Accepted |
| [0005](0005-prometheus-live-histograms-authoritative.md) | Prometheus is the live view; merged histograms are authoritative | Accepted |
| [0006](0006-handwritten-load-engine.md) | Hand-write the load engine (vegeta as reference, not dependency) | Accepted |

**Proposed / not yet decided:** rate-limiter algorithm default (token bucket vs sliding
window) — to be an ADR when Phase 9 lands.
