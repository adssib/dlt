# ADR-0006: Hand-write the load engine (vegeta as reference, not dependency)

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 2, 4

## Context

Mature Go load-testing libraries exist (e.g. `tsenart/vegeta`, `k6`). We could depend on one
and get bounded-concurrency generation, measurement, and aggregation for free. But the
concurrency model, per-request measurement, ramp scheduling, and result aggregation **are
exactly the things this project exists to teach**.

## Decision

**Hand-write the engine** — the bounded-concurrency runner (semaphore), `http.Client` with
connection reuse, the linear ramp scheduler, and the outcome classifier (success / 429 /
fail). Read vegeta as a **reference**, don't depend on it.

## Alternatives considered

- **Depend on vegeta / k6** — faster to a working tool, but it hides the semaphore,
  measurement, ramp, and 429-classification — i.e. it hides the learning.

## Consequences

- ✅ The core distributed-load mechanics are learned and owned, not imported.
- ✅ Keeps the dependency tree thin (a stated non-functional requirement).
- ⚠️ We reimplement solved problems and may miss edge cases a mature library handles — an
  acceptable trade for a learning tool, **not** for production use.
- 🔭 For a real production need, swap in a battle-tested engine; the interfaces are designed so
  the coordinator/worker wouldn't need to change.
