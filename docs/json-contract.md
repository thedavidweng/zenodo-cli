# JSON Output Contract

Every command supports `--json` for structured output. Responses use a consistent envelope:

## Success Response

```json
{
  "ok": true,
  "data": {
    "hits": [],
    "total": 0
  },
  "meta": {
    "command": "records.list",
    "profile": "default",
    "duration_ms": 234,
    "schema_version": "2026-06-11",
    "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

## Error Response

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

## Output Modifiers

- `--pretty` — indented JSON
- `--compact` — omit empty fields
- `--full` — include all fields even when null/empty (overrides `--compact`)
