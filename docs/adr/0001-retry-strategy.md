# ADR 0001: HTTP Retry Strategy with Retry-After Cap

## Status

Accepted

## Context

Zenodo InvenioRDM API calls fail transiently: HTTP 429 (Too Many Requests),
5xx server errors, and network timeouts. Long-running operations (record
creation followed by file upload and publish) must survive a single transient
failure rather than aborting the whole flow.

Two request paths existed with near-identical retry loops: one for JSON bodies
buffered in memory, one for streaming file uploads that must reopen the file
on each attempt (the transport closes the body after a failed request). Both
used a fixed 500ms wait between attempts, which neither backs off under load
nor honors the `Retry-After` header Zenodo sends on 429.

## Decision

A single `doRequest` loop serves both paths. It takes an `openBody` provider
that returns a fresh `io.ReadCloser` per attempt, so buffered JSON bodies and
reopenable file handles share one code path.

Retries run up to `Client.Retries` times (default 3, set from `--retries`).
Retryable failures are HTTP 429, HTTP 5xx, and network errors; any other 4xx
returns immediately. Between attempts the client backs off exponentially
(`100ms * 2^(attempt-1)`). On HTTP 429, if a `Retry-After` header is present
and parseable, the client additionally waits that duration **capped at 60
seconds** before the next attempt. The cap bounds worst-case latency when
Zenodo requests an impractically long wait.

Context cancellation is checked during every wait, so `Ctrl+C` interrupts a
retry immediately.

## Consequences

### Positive

- Transient failures no longer abort multi-step operations.
- Exponential backoff avoids retry storms against a struggling server.
- `Retry-After` is respected for typical rate-limit windows, capped so the CLI
  never appears hung.
- One retry loop instead of two removes the duplicated logic that let the JSON
  and streaming paths drift apart.

### Negative

- Retries add latency to genuinely failing operations.
- The 60-second cap can cause repeated 429s if Zenodo genuinely needs longer.

### Mitigations

- `--retries 0` disables retrying for callers that want fail-fast behavior.
- The `retryable` field in the JSON error envelope lets agents decide whether
  to retry at a higher level.
