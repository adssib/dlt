# CLAUDE.md — dlt (Distributed Load Tester)

Guidance for Claude Code when working in this repo. For the documentation map, read `docs/README.md`.

## What this project is

A distributed HTTP load tester in Go, built as a **hands-on learning + portfolio project**.
It exists to demonstrate distributed-systems and DevOps competence — membership,
coordination, correct result aggregation, graceful failure, and the Docker / Kubernetes /
Helm / Prometheus / Grafana stack around it — and, just as importantly, to **read like a
staff engineer's work**: every non-trivial choice is on the record with its tradeoff.

- Module: `github.com/adssib/dlt`
- Spec: `docs/SPEC.md` · Architecture + diagrams: `docs/ARCHITECTURE.md` · Decisions: `docs/decisions/`

## Working agreement (IMPORTANT — overrides default behavior)

This is a **learning project**. The division of labor is strict and explicit:

| Who | Writes |
|---|---|
| **Claude** | **Boilerplate + plumbing** (package layout, types, signatures, doc comments, config loading, server/CLI wiring, infra stubs) **and the tests**. |
| **Adib** | The **business logic and behavior** — the algorithms: histogram merge, the concurrency semaphore, the ramp scheduler, the rate limiters, the target's latency/fault model. |

- Core logic ships from Claude as compiling stubs (`// TODO(you)` returning zero values);
  **Adib fills them in himself.**
- **Never write the business logic unless Adib explicitly asks.** The normal working mode is
  **TDD-pairing: Claude writes the failing test, Adib writes the code that makes it pass.**
- Coaching and explaining are always welcome. If unsure whether something counts as
  "business logic," ask before writing it.

## Commit attribution (faithful commits)

Git history must honestly reflect **who wrote what**:

- **Code Adib writes** (business logic, behavior) → committed as **Adib alone**, no co-author trailer.
- **Code Claude writes** (boilerplate, plumbing, tests) → committed with a **`Co-Authored-By`
  trailer for Claude**, so the work shows as shared in the history:

  ```
  Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
  ```

- **Keep the two kinds of change in separate commits** so every commit carries the correct
  attribution — never mix Claude-written scaffolding and Adib-written logic in one commit.

*(The trailer is the honest record either way; GitHub only links a co-author to a profile when
the email maps to a real account.)*

## System-design literacy (the point of this repo)

This repo is meant to *signal architectural thinking*. Throughout the project:

- **Every non-trivial decision gets an ADR** in `docs/decisions/` — context → options
  considered → the tradeoff → consequences. Copy `0000-adr-template.md`.
- **Document the assumption and the tradeoff, not just the outcome.** Prefer
  "chose X over Y because Z, accepting cost W" over "uses X."
- **Keep the diagrams current** (mermaid lives in `docs/ARCHITECTURE.md`).
- When a decision changes, **supersede** its ADR (add a new one, mark the old one
  superseded) — don't silently rewrite history. That trail *is* the portfolio.

Self-check before finishing work: *would a senior reviewer see the reasoning, or just the result?*

## Commands

| Task | Command |
|---|---|
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Test (race detector) | `go test -race ./...` |
| Run the target | `go run ./cmd/target -c configs/target.yaml` |
| Format before commit | `gofmt -w ./...` |

## Conventions

- Go idioms: `camelCase` locals, `gofmt`-clean, table-driven tests / subtests.
- Tests co-located as `_test.go`; use package-internal tests when they need unexported access.
- CI: `.github/workflows/ci.yml` — build / vet / test / lint as isolated, parallel jobs.

## Ethics (non-negotiable)

Generates DDoS-shaped traffic. Point it **only at the bundled `target` on owned / home-lab
infrastructure**. Never at third-party systems without explicit written permission. All
traffic stays inside the owner's network. See `docs/SPEC.md` (Ethics & safety).
