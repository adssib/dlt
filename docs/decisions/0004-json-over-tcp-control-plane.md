# ADR-0004: Newline-delimited JSON over TCP for the control plane

- **Status:** Accepted
- **Date:** 2026-07-01
- **Phase:** 2

## Context

The coordinator and workers need a control channel: workers *register*, the coordinator
*broadcasts a start barrier*, workers *stream progress* and *send final results*. This needs a
long-lived connection with **server→client push** (the start barrier) and is internal to the
system (both ends are ours). It's also a learning project — the wire should be *visible*.

## Decision

**Newline-delimited JSON envelopes over raw TCP** (`net.Conn` wrapped in `bufio`), with a tiny
`Conn` codec exposing `ReadMsg` / `WriteMsg`.

## Alternatives considered

- **gRPC** — great for typed, polyglot RPC, but pulls in protobuf codegen and toolchain and
  *hides* the wire behind generated code — the opposite of the learning goal here.
- **HTTP / REST** — no natural long-lived server→client push for the start barrier and
  progress stream without bolting on SSE/websockets.
- **A message broker (NATS/Redis)** — extra infrastructure to run for a point-to-point channel.

## Consequences

- ✅ The transport is visible and hackable — you can literally `nc` into it and read the JSON.
- ✅ One long-lived connection gives natural server-push for the barrier and the progress stream.
- ✅ Standard library only.
- ⚠️ Hand-rolled framing: the newline delimiter means payloads can't contain raw newlines
  (JSON encoding handles that). No schema evolution / back-compat — accepted for an internal
  protocol. Plaintext — TLS is explicitly out of scope.
- 🔭 Revisit → gRPC if we ever need polyglot workers or strong, evolving schemas.
