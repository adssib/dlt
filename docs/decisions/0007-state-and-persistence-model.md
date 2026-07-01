# ADR-0007: State & persistence model — ephemeral everywhere; coordinator is a singleton

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 2 (foundational)

## Context

"Keep it stateless" conflates two separate properties: (a) **no persisted state** that must
survive a restart, and (b) **fungible / horizontally-scalable** instances. The two components
differ on the second one, so we need to be explicit — especially because it drives the
Kubernetes shapes and the failure model.

## Decision

**Nothing persists to disk or a database — all state is in-memory and ephemeral.**
- **Workers are stateless and fungible:** each holds only transient per-run counters; any
  worker is interchangeable and disposable.
- **The coordinator is a stateful *singleton* during a run:** it holds the worker registry and
  aggregates progress/results **in memory**, but persists nothing.

## Alternatives considered

- **Persist run state (DB / PersistentVolume)** so a coordinator restart can resume a run — a
  load test that's interrupted is cheap to just rerun; durable storage isn't worth the weight.
- **Multiple coordinators (HA) sharing state** — needs consensus or a shared store; out of
  scope for a home-lab test tool.

## Consequences

- ✅ Workers scale freely (`kubectl scale`) and any subset can die mid-run (F11) — no durable
  state is lost.
- ✅ The coordinator is a plain **1-replica Deployment + Service**, *not* a StatefulSet with a
  PersistentVolume — because there is nothing to persist. It's a singleton, not a store.
- ⚠️ The coordinator is a **single point of failure**: if it dies mid-run, the run is lost and
  you rerun. Accepted (test tool, not a durable service). See [ADR-0003](0003-orchestrator-owns-worker-count.md).
- 🔭 Revisit if we ever need durable run history or coordinator HA (→ external store + leader
  election).
