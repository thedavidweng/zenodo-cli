# JSON Envelope Schema

All commands emit a standard JSON envelope when invoked with `--json`.
The current schema version is `2026-06-11`.

## Success Envelope

```json
{
  "ok": true,
  "data": { ... },
  "meta": {
    "command": "records.list",
    "profile": "default",
    "duration_ms": 234,
    "schema_version": "2026-06-11",
    "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

### Fields

- `ok` — always `true` on success.
- `data` — command-specific payload. May be an object, array, or null.
- `meta` — request metadata (always present).
- `meta.command` — dot-separated command name (e.g. `records.list`, `files.upload`).
- `meta.profile` — profile used for the request.
- `meta.duration_ms` — wall-clock duration in milliseconds.
- `meta.schema_version` — JSON schema version string.
- `meta.request_id` — UUID for correlation and audit log lookup.

## Error Envelope

```json
{
  "ok": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Authentication required. Run 'zenodo auth login' to authenticate.",
    "category": "auth",
    "retryable": false
  },
  "meta": {
    "command": "records.list",
    "profile": "default",
    "duration_ms": 12,
    "schema_version": "2026-06-11",
    "request_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
  }
}
```

### Error Fields

- `ok` — always `false` on error.
- `error.code` — machine-readable error code (see table below).
- `error.message` — human-readable description.
- `error.category` — high-level grouping (`auth`, `api`, `network`, `validation`, `safety`).
- `error.retryable` — `true` if the operation can be safely retried.

## Error Codes

| Code | Category | Description |
|------|----------|-------------|
| `AUTH_REQUIRED` | auth | Not logged in or token expired |
| `AUTH_FAILED` | auth | Token rejected by Zenodo |
| `READ_ONLY_VIOLATION` | safety | Mutation attempted with `--read-only` |
| `CONFIRMATION_REQUIRED` | safety | Destructive operation without `--confirm` |
| `ZENODO_API_ERROR` | api | Zenodo API returned an error or unexpected status |
| `NETWORK_ERROR` | network | Could not connect to Zenodo servers |
| `VALIDATION_FAILED` | validation | Input arguments or flag values are invalid |
| `RESOURCE_NOT_FOUND` | not_found | Specified record or file ID does not exist |
| `PARTIAL_SUCCESS` | api | Some operations in a bulk request failed |
| `FILESYSTEM_ERROR` | filesystem | Local file operation failed |
| `CONFIG_ERROR` | config | Configuration error |
| `INTERRUPTED` | — | Process interrupted (SIGINT) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (API, network, filesystem, validation, config) |
| 2 | Auth required or auth failed (`AUTH_REQUIRED`, `AUTH_FAILED`) |
| 3 | Safety gate blocked (`READ_ONLY_VIOLATION`, `CONFIRMATION_REQUIRED`) |
| 130 | Interrupted |

## Output Modifiers

- `--pretty` — indented JSON for human reading.
- `--compact` — omit empty/null fields.
- `--full` — include all fields even when null/empty (overrides `--compact`).
