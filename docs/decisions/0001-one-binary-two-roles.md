# ADR-0001: One `dlt` binary with role subcommands; separate `target` program

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 1–3

## Context

The system has three roles: a coordinator, N workers, and a system-under-test (SUT). We need
to decide how to package them. The coordinator and worker share a lot of code (protocol,
config, metrics); the SUT is conceptually different — it's the *thing being attacked*, and it
must be independently deployable and deliberately breakable.

## Decision

Ship **one `dlt` binary** with role **subcommands** (`dlt coordinator`, `dlt worker`) plus a
local launcher (`dlt test`). Ship the **`target` as a separate program**.

## Alternatives considered

- **Three separate binaries** — clean separation, but the coordinator and worker duplicate
  build/release wiring despite sharing most of their code.
- **One binary that also includes the target** — simplest to build, but it blurs the
  attacker/defender boundary and couples the SUT's lifecycle and scaling to the tester's.

## Consequences

- ✅ Coordinator and worker share code and a single image; subcommands map cleanly to
  Kubernetes container `args`.
- ✅ The `target` has a clean boundary: its own image, its own Deployment + Service, scaled
  and broken independently of the tester.
- ⚠️ One binary does two jobs, so it needs subcommand dispatch and role-specific config.
- 🔭 Revisit if the roles diverge enough that sharing a binary causes more friction than it saves.
