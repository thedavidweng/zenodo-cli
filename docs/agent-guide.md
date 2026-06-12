# Agent Guide

This guide is for automated agents (CI, bots, scripts) using zenodo-cli.

## JSON Mode

Always use `--json` for machine-parseable output. The envelope format:

```json
{
  "ok": true,
  "data": { ... },
  "meta": { ... }
}
```

Check `ok` field before processing `data`.

## Exit Codes

```
0    success
1    validation failed
2    auth required or auth failed
3    Zenodo API error
4    network error
5    partial success
6    safety gate blocked mutation
7    local filesystem error
8    config error
130  interrupted
```

## Error Handling

Errors include machine-readable codes:

```json
{
  "ok": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "category": "auth",
    "retryable": false
  }
}
```

## Retryable Errors

Network timeouts, HTTP 429/5xx, and temporary Zenodo errors are marked `retryable: true`.

## Events Stream

Use `--events` for NDJSON progress on stderr while getting JSON result on stdout.

## Safety

For mutations in automation:
- Use `--read-only` to test without side effects
- Use `--dry-run` to preview actions
- Use `--confirm` for high-risk operations

## Common Patterns

### Create and publish a record

```bash
# Create draft
zenodo records create --title "Dataset" --json

# Upload files
zenodo files upload RECORD_ID ./data.csv

# Publish (irreversible)
zenodo records publish RECORD_ID --confirm --json
```

### Search and iterate

```bash
zenodo search "climate data" --json | jq '.data.hits[].id'
```
