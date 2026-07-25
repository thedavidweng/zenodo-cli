# ADR 0003: Raw API Escape Hatch

## Status

Accepted

## Context

Zenodo runs on InvenioRDM, whose REST API surface is large and still evolving.
The CLI provides typed commands (`records`, `files`, `search`) for the common
workflows, but those cover only a subset of endpoints. Users automating less
common operations — editing arbitrary metadata fields, calling a new endpoint,
inspecting a resource the typed commands do not model — would otherwise be
blocked until a typed command is added.

## Decision

Ship a raw escape hatch: `zenodo api get|post|put <path>` sends a request to
any InvenioRDM endpoint and returns the JSON response. `post`/`put` take a
`--data` JSON body and are classified as high-write, so they pass through the
same safety gate as typed mutations (`--read-only` blocks them, `--confirm` is
required, `--dry-run` previews). Leading `/api` is added automatically so both
`records` and `/api/records` work.

The escape hatch reuses the shared client, authentication, and output
envelope; it is not a separate code path with its own auth or retry rules.

## Consequences

### Positive

- The CLI is never a hard blocker: any InvenioRDM endpoint is reachable today.
- Raw writes inherit the same safety guarantees as typed commands.
- New endpoints can be used immediately and promoted to typed commands later
  based on real demand.

### Negative

- Raw requests bypass the typed request/response validation, so malformed
  bodies surface as API errors rather than local validation errors.

### Mitigations

- `--data` is parsed as JSON before sending, so syntactically invalid bodies
  fail locally with a `VALIDATION_FAILED` error instead of a wasted request.
