# Global Flags and Environment Variables

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | | Config file path (YAML, default: `~/.config/zenodo-cli/config.yaml`) |
| `--profile` | `default` | Credential profile |
| `--sandbox` | `false` | Use Zenodo sandbox |
| `--json` | `false` | JSON envelope to stdout |
| `--pretty` | `false` | Pretty-print JSON |
| `--compact` | `false` | Omit null/empty fields from JSON |
| `--full` | `false` | Include all fields in JSON (overrides `--compact`) |
| `--quiet` | `false` | Suppress progress messages |
| `--read-only` | `false` | Block all remote mutations |
| `--dry-run` | `false` | Preview without execution |
| `--confirm` | `false` | Confirm irreversible operations |
| `--timeout` | `5m0s` | Command/API timeout |
| `--retries` | `3` | Retry count for retryable failures |

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
| `ZENODO_DRY_RUN` | Set `1` to preview without execution |
| `ZENODO_CONFIRM` | Set `1` to confirm irreversible operations |
| `ZENODO_JSON` | Set `1` to emit the JSON envelope |
| `ZENODO_QUIET` | Set `1` to suppress progress messages |
