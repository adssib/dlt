# dlt — Documentation

Start here. This project is documented like a real system, split by **purpose** so a
reviewer can find what they need fast:

| Doc | Answers |
|---|---|
| [SPEC.md](SPEC.md) | **What** must be true — requirements, wire/API contracts, config model, scope, ethics. |
| [ARCHITECTURE.md](ARCHITECTURE.md) | **How** it's shaped — components, process/pod model, diagrams, observability. |
| [ROADMAP.md](ROADMAP.md) | **In what order** — the 9-phase build plan and the definition of done. |
| [decisions/](decisions/) | **Why** — Architecture Decision Records: every tradeoff, on the record. |

Reading path for someone new: **SPEC → ARCHITECTURE → skim the ADRs.** The ADRs are where
the engineering judgment lives — start with
[ADR-0002 (merged histograms)](decisions/0002-merged-histograms-over-averaged-percentiles.md),
the decision the whole project is built around.

> Working agreement and Claude Code guidance live in [`/CLAUDE.md`](../CLAUDE.md).
