# Global Flags and Environment Variables

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | | Config file path |
| `--profile` | `default` | Credential profile |
| `--sandbox` | `false` | Use Zenodo sandbox |
| `--json` | `false` | JSON envelope to stdout |
| `--pretty` | `false` | Pretty-print JSON |
| `--compact` | `false` | Compact output fields |
| `--full` | `false` | Full normalized fields (overrides `--compact`) |
| `--read-only` | `false` | Block all remote mutations |
| `--dry-run` | `false` | Preview without execution |
| `--confirm` | `false` | Confirm high-risk operations |
| `--timeout` | `30s` | API timeout |
| `--retries` | `3` | Retry count for retryable failures |
| `--events` | `false` | NDJSON progress events to stderr |
| `--no-color` | `false` | Disable ANSI color |
| `--verbose` | `false` | Diagnostics to stderr |
| `--debug` | `false` | Debug diagnostics with secrets redacted |
| `--quiet` | `false` | Suppress progress output |

Run `zenodo --help` or `zenodo <command> --help` for full flag details.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ZENODO_TOKEN` | API token (overrides config file) |
| `ZENODO_CONFIG` | Config file path |
| `ZENODO_PROFILE` | Active profile name |
| `ZENODO_SANDBOX` | Set `1`/`true`/`yes` to use sandbox |
| `ZENODO_API_URL` | Override API base URL |
| `ZENODO_TIMEOUT` | Command timeout (e.g. `60s`) |
| `ZENODO_RETRIES` | Retry count for retryable failures |
| `ZENODO_READ_ONLY` | Set `1` to block mutations |
| `ZENODO_DEBUG` | Set `1` to enable debug diagnostics |
