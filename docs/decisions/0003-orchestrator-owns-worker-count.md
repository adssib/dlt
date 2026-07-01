# ADR-0003: Orchestrator owns worker count; coordinator owns only the readiness barrier

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 3, 7

## Context

Something has to decide *how many workers* run. Two candidates: the **coordinator** (it knows
about the test) or the **deployment layer** (`dlt test --workers`, `docker compose --scale`,
Kubernetes `replicas` / `kubectl scale`). Kubernetes already owns replica counts natively; if
the coordinator also tried to own it, the two would fight.

## Decision

The **orchestrator owns the worker count.** The coordinator knows only `min_workers` — a
**readiness barrier**: don't start a run until at least N workers have registered.

## Alternatives considered

- **Coordinator spawns/owns workers** — gives the coordinator exact control, but creates a
  second source of truth against Kubernetes, requires the coordinator to hold infra
  privileges, and directly fights `kubectl scale`.

## Consequences

- ✅ No conflict with Kubernetes scaling — `kubectl scale deploy/worker` "just works."
- ✅ The coordinator stays a pure control-plane component (no infra privileges).
- ⚠️ The coordinator can't guarantee an *exact* worker count; it waits for a **minimum**, then
  proceeds with whoever registered. (This is a feature: it's also what makes graceful worker
  death — F11 — natural.)
- 🔭 Revisit only if a test genuinely needs exact-N semantics the barrier can't express.
