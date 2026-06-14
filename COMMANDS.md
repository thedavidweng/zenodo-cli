# Command Reference

All commands accept the [global flags](#global-flags) listed at the bottom.
Use `zenodo <command> --help` for full flag details.

## Top-level Commands

| Command | Description |
|---------|-------------|
| `version` | Show version, commit, date, Go version, and schema version |
| `doctor` | Check configuration and connectivity |
| `completion` | Generate shell completion scripts (bash, zsh, fish, powershell) |

## auth

Manage Zenodo API credentials.

| Command | Usage | Description |
|---------|-------|-------------|
| `auth login` | `zenodo auth login --token TOKEN` | Store API token (non-interactive) |
| `auth login` | `zenodo auth login` | Interactive token prompt |
| `auth status` | `zenodo auth status` | Check current authentication |
| `auth logout` | `zenodo auth logout` | Remove stored credentials |

**Key flags for `auth login`:**

- `--token` — API token (non-interactive)
- `--sandbox` — store token for sandbox environment

```bash
zenodo auth login --token YOUR_TOKEN
zenodo auth login --sandbox --token SANDBOX_TOKEN
zenodo auth status --json
zenodo auth logout
```

## records

Manage Zenodo deposit records.

| Command | Usage | Description |
|---------|-------|-------------|
| `records list` | `zenodo records list` | List your records |
| `records create` | `zenodo records create --title "Title"` | Create a draft record |
| `records create` | `zenodo records create --metadata record.json` | Create from JSON metadata file |
| `records show` | `zenodo records show RECORD_ID` | Show record details |
| `records delete` | `zenodo records delete RECORD_ID --confirm` | Delete a draft record |
| `records publish` | `zenodo records publish RECORD_ID --confirm` | Publish a draft (irreversible) |
| `records new-version` | `zenodo records new-version RECORD_ID` | Create a new draft version |

```bash
zenodo records list --json
zenodo records create --title "My Dataset" --description "A research dataset"
zenodo records create --metadata record.json
zenodo records show RECORD_ID --json
zenodo records delete DRAFT_ID --confirm
zenodo records publish RECORD_ID --confirm
zenodo records new-version RECORD_ID
```

## files

Manage files attached to records.

| Command | Usage | Description |
|---------|-------|-------------|
| `files upload` | `zenodo files upload RECORD_ID file1.csv file2.csv` | Upload files to a draft |
| `files list` | `zenodo files list RECORD_ID` | List files in a record |
| `files download` | `zenodo files download RECORD_ID --dest ./output` | Download files from a published record |

```bash
zenodo files upload RECORD_ID ./data.csv ./metadata.json
zenodo files list RECORD_ID --json
zenodo files download RECORD_ID --dest ./output
```

## search

Search public Zenodo records.

```bash
zenodo search "machine learning"
zenodo search "climate data" --json | jq '.data.hits[].id'
```

## api

Direct access to any Zenodo InvenioRDM endpoint.

| Command | Usage | Description |
|---------|-------|-------------|
| `api get` | `zenodo api get /api/records` | GET request |
| `api post` | `zenodo api post /api/records --data '{...}'` | POST request |
| `api put` | `zenodo api put /api/records/ID/draft --data '{...}'` | PUT request |

Path prefix `/api` is added automatically if omitted.

```bash
zenodo api get /api/records
zenodo api get records?q=climate
zenodo api post /api/records --data '{"metadata":{"title":"Test"}}'
zenodo api put /api/records/ID/draft --data '{"metadata":{"title":"Updated"}}'
```

## doctor

Check configuration, profile, and token validity.

```bash
zenodo doctor
zenodo doctor --json
```

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
| `--timeout` | `5m` | API timeout |
| `--retries` | `3` | Retry count for retryable failures |
| `--quiet` | `false` | Suppress progress output |
