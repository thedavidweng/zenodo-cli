# Architecture

This document describes the internal structure of zenodo-cli.

## Package Layout

```
cmd/zenodo/           Entry point (main.go)
internal/
  cli/                Cobra command definitions and CLI logic
  zenodo/             Zenodo InvenioRDM API client (HTTP, auth, retries)
  config/             Configuration loading, saving, profiles, credentials
  model/              Domain types (envelope, errors)
  output/             JSON/human rendering, error formatting
  testutil/           Fake Zenodo server and test helpers
```

## Request Flow

```
User input
  │
  ▼
Cobra command (internal/cli/)
  │
  ├─ Reads flags → AppContext
  ├─ Loads config → profile + credentials
  ├─ Creates zenodo.Client
  │
  ├─ [Safety gate] → Check(read-only, dry-run, confirm, risk)
  │     │
  │     ├─ Blocked → error envelope (exit 6)
  │     ├─ Dry-run → plan output, no execution
  │     └─ Allowed → continue
  │
  ├─ Calls Zenodo API via client.Get() / client.Post() / client.Put()
  │     │
  │     ├─ Bearer token auth
  │     ├─ Retry on 429/5xx (exponential backoff)
  │     └─ Response parsing + error normalization
  │
  ├─ Renders output
  │     ├─ --json → JSON envelope to stdout
  │     └─ human → table/text to stdout
  │
  └─ Returns error → exit code mapping
```

## Zenodo Client (`internal/zenodo/`)

| File | Responsibility |
|------|---------------|
| `client.go` | Client struct, constructor, HTTP methods (Get/Post/Put/Delete), retry logic |
| `types.go` | API response types: Record, RecordMetadata, Creator, ResourceType, RecordFile, SearchResponse |

Key design decisions:
- All API calls go through typed HTTP methods with automatic JSON handling
- Automatic retry with exponential backoff on 429 and 5xx responses
- Bearer token authentication via `Authorization` header
- API base URL is configurable (sandbox vs production)
- Tokens are never logged or printed

## Configuration (`internal/config/`)

Config is stored as YAML at `~/.config/zenodo-cli/config.yaml` (XDG-compliant).

```yaml
current_profile: default
profiles:
  default:
    token: "REDACTED"
    sandbox: false
  staging:
    token: "REDACTED"
    sandbox: true
```

Credential resolution priority:
1. Explicit flags (`--token`)
2. Environment variables (`ZENODO_TOKEN`)
3. Profile config
4. Interactive prompt

## Safety System

Every mutation command passes through a safety gate before execution.

Risk levels:
- **read** — no mutation, always allowed
- **medium-write** — blocked by `--read-only`, supports `--dry-run`
- **high-write** — blocked by `--read-only`, requires `--confirm`

## Output System (`internal/output/`)

Every command produces output through the `Renderer`:

- **JSON mode** (`--json`): writes a standard envelope to stdout
- **Human mode**: writes formatted text to stdout
- **Events** (`--events`): writes NDJSON progress events to stderr
- **Errors**: always written to stderr in human mode; included in JSON envelope

## Testing

- Unit tests in every package using standard `testing`
- Fake Zenodo server (`internal/testutil/`) for integration tests
- Table-driven tests preferred
- `go test -race` passes across all packages

## Dependencies

| Dependency | Purpose |
|-----------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | Config file parsing |
| `github.com/google/uuid` | Request IDs |
