# ADR-0005: Prometheus is the live view; merged histograms are authoritative

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 8

## Context

There are two ways to surface results: **Prometheus/Grafana** (scraped, live, great for
watching a run unfold) and the **coordinator's merged report** (the final, correct numbers
from [ADR-0002](0002-merged-histograms-over-averaged-percentiles.md)). We need to be explicit
about which one is the source of truth, because they can disagree.

## Decision

**Prometheus/Grafana is the live view only.** The **authoritative final percentiles come from
the merged histograms** in the coordinator's report.

## Alternatives considered

- **Use Prometheus histograms for the final percentiles too** — `histogram_quantile()` over
  Prometheus buckets can't correctly merge across workers without shared, pre-agreed buckets,
  and even then it *approximates* and can't reproduce the exact merge the coordinator does.
- **Push the final numbers into Prometheus** — conflates "live" and "authoritative" into one
  surface and hides the distinction this project is meant to demonstrate.

## Consequences

- ✅ Clean split of concerns: **correctness** (merged histograms) vs **live observability**
  (Prometheus) — each tool does what it's best at.
- ✅ Demonstrates *why* PromQL cross-worker percentile merging is lossy — a real
  systems-literacy talking point.
- ⚠️ Two result surfaces exist; the docs and report must state plainly which is authoritative.
- 🔭 Revisit if a future need makes a single unified surface worth the loss of precision.
