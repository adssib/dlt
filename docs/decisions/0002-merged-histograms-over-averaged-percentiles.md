# ADR-0002: Global percentiles via merged histograms, never averaged per-worker

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 4 (the centerpiece)

## Context

N workers each measure the latency of the requests they sent. We need **correct global
percentiles** (p50, p95, p99) across *all* requests. Latency distributions are heavily skewed,
and workers see different subsets of the traffic. There is no shared clock, so results can
only be combined by **counting**, not by comparing timestamps across machines.

## Decision

Each worker records into a **histogram** and ships it **serialized** to the coordinator. The
coordinator **`Merge`s** the histograms, then computes `Quantile` **from the merged
structure**. Everything depends only on a `Histogram` interface (HDR is the default impl).

## Alternatives considered

- **Average the per-worker percentiles** — trivial, and *statistically invalid*: the mean of
  each worker's p99 is **not** the global p99. This is the classic mistake this project
  exists to *not* make.
- **Ship every raw latency sample to the coordinator** — exact, but unbounded memory and
  network cost that grows with total request count.
- **Central streaming percentile over raw samples** — same transport cost as shipping samples.

## Consequences

- ✅ Correct global percentiles at **bounded, tiny** network cost — a fixed-size histogram per
  worker regardless of how many requests it sent.
- ✅ Depending on the interface (not HDR directly) means HDR ↔ t-digest is a swap the
  coordinator never sees.
- ⚠️ Histograms trade a little precision for boundedness: you must configure the value range
  `(min, max)` and significant figures. Latencies outside the range are clamped.
- 🔭 Revisit the impl (not the approach) if we need lower memory or unbounded range → t-digest.
