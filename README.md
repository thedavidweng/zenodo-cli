<p align="center">
  <h1 align="center">zenodo-cli</h1>
  <p align="center">
    <strong>Agent-friendly Zenodo CLI</strong>
  </p>
  <p align="center">
    Single-binary tool for Zenodo deposit management, file upload/download, and full InvenioRDM API access.
  </p>
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml"><img src="https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/thedavidweng/zenodo-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/zenodo-cli" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/thedavidweng/zenodo-cli"><img src="https://pkg.go.dev/badge/github.com/thedavidweng/zenodo-cli.svg" alt="Go Reference"></a>
  <a href="https://github.com/thedavidweng/zenodo-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go" alt="Go">
</p>

---

## Highlights

- **Single binary** -- no dependencies, no runtime, no containers
- **14 commands** -- auth, records, files, search, API, doctor, version, completion
- **JSON-first** -- `--json` on every command, consistent envelope, machine-parseable
- **Safety gates** -- `--read-only`, `--dry-run`, `--confirm` for destructive operations
- **Agent-ready** -- exit codes, error categories, request IDs, secret redaction
- **Cross-platform** -- Linux, macOS, Windows (amd64/arm64)
- **Sandbox support** -- test against Zenodo sandbox with `--sandbox` or `ZENODO_SANDBOX=1`

## Why zenodo-cli?

Zenodo migrated its infrastructure from the legacy deposit API to InvenioRDM in 2023. The old `/api/deposit/depositions` endpoint now returns 404. Every existing CLI tool was built against the deprecated API and no longer works for uploads or record management.

zenodo-cli is a single-binary CLI built on the new InvenioRDM API (`/api/records`) with full coverage of the Zenodo API surface.

| Tool | Language | Stars | Last Updated | API Version | Upload Works | CLI | Status |
|------|----------|-------|--------------|-------------|-------------|-----|--------|
| **zenodo-cli** | Go | -- | 2026-06 | InvenioRDM (new) | Yes | Yes | Active |
| [zenodo-client](https://github.com/cthoyt/zenodo-client) (cthoyt) | Python | 45 | 2026-04 | Mixed (old+new) | No | Yes | Upload broken |
| [zenodo](https://github.com/cheminfo/zenodo) (cheminfo) | Node.js | -- | 2026-04 | InvenioRDM (new) | Yes | No | Library only |
| [zenodraft](https://github.com/zenodraft/zenodraft) | Node.js | 8 | 2024-04 | Old (deprecated) | No | Yes | 19 open issues |
| [zotzen-lib](https://github.com/OpenDevEd/zotzen-lib) | Node.js | 3 | 2023-05 | Old (deprecated) | No | Yes | Unmaintained |
| [zenodo-lib](https://github.com/OpenDevEd/zenodo-lib) | Node.js | 2 | 2024-02 | Old (deprecated) | No | No | Unmaintained |
| [zenodo-cli](https://github.com/OpenDevEd/zenodo-cli) (opendeved) | Node.js | 1 | 2021-01 | Old (deprecated) | No | Yes | Deprecated |

**Key points:**
- `zenodo-client` (Python) references both old and new APIs but file upload uses the deprecated endpoint and fails.
- `zenodo` (cheminfo) uses the new InvenioRDM API and supports upload, but it is a library -- not a CLI tool.
- All other tools (`zenodraft`, `zotzen-lib`, `zenodo-lib`, `zenodo-cli` by opendeved) target the old `/api/deposit/depositions` which no longer exists.

## Install

### Homebrew

```sh
brew tap thedavidweng/tap
brew install zenodo
```

### Go

```sh
go install github.com/thedavidweng/zenodo-cli/cmd/zenodo@latest
```

### Binary download

Download from [Releases](https://github.com/thedavidweng/zenodo-cli/releases). Checksums and cosign signatures are provided for verification.

### Build from source

```sh
git clone https://github.com/thedavidweng/zenodo-cli.git
cd zenodo-cli
make build
```

## Quick Start

```sh
# 1. Authenticate (get a token at https://zenodo.org/account/settings/applications/tokens/)
zenodo auth login --token YOUR_TOKEN

# 2. Verify configuration
zenodo doctor

# 3. Create a draft record
zenodo records create --title "My Dataset" --description "A research dataset"

# 4. Upload files
zenodo files upload RECORD_ID ./data.csv ./metadata.json

# 5. Publish (irreversible -- use --confirm)
zenodo records publish RECORD_ID --confirm
```

## Usage

### Auth

```sh
zenodo auth login --token TOKEN        # store API token (non-interactive)
zenodo auth login                      # interactive prompt
zenodo auth status                     # check current authentication
zenodo auth logout                     # remove stored credentials
```

### Records

```sh
zenodo records list                                  # list your records
zenodo records create --title "Title"                # create a draft record
zenodo records create --metadata record.json         # create from JSON metadata file
zenodo records show RECORD_ID                        # show record details
zenodo records delete RECORD_ID --confirm            # delete a draft record
zenodo records publish RECORD_ID --confirm           # publish a draft (irreversible)
zenodo records new-version RECORD_ID                 # create a new draft version
```

### Files

```sh
zenodo files upload RECORD_ID file1.csv file2.csv    # upload files to a draft
zenodo files list RECORD_ID                          # list files in a record
zenodo files download RECORD_ID --dest ./output      # download files from a published record
```

### Search

```sh
zenodo search "machine learning"                     # search public records
zenodo search "climate data" --json | jq '.data.hits[].id'
```

### API

Direct access to any Zenodo InvenioRDM endpoint:

```sh
zenodo api get /api/records                          # GET request
zenodo api get records?q=climate                     # path prefix added automatically
zenodo api post /api/records --data '{"metadata":{"title":"Test"}}'
zenodo api put /api/records/ID/draft --data '{"metadata":{...}}'
```

### Doctor

```sh
zenodo doctor                                        # check config, profile, and token
zenodo doctor --json                                 # machine-readable diagnostics
```

## Configuration

Config file path: `~/.config/zenodo-cli/config.yaml` (respects `XDG_CONFIG_HOME`).

The config file stores named profiles, each with its own token and settings:

```yaml
current_profile: default
profiles:
  default:
    token: "your-api-token"
    sandbox: false
  staging:
    token: "sandbox-token"
    sandbox: true
```

### Profiles

Switch profiles with `--profile`:

```sh
zenodo records list --profile staging
```

### Sandbox mode

Test against the Zenodo sandbox (sandbox.zenodo.org) without touching production data:

```sh
zenodo auth login --token SANDBOX_TOKEN --sandbox
zenodo records list --sandbox
# or set globally
export ZENODO_SANDBOX=1
```

## JSON Output

Every command supports `--json` for structured output. Responses use a consistent envelope:

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

Error responses:

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

Use `--pretty` for indented JSON and `--compact` to omit empty fields. Use `--full` to include all fields even when null/empty.

## Safety Gates

Three flags protect against unintended mutations:

| Flag | Effect |
|------|--------|
| `--read-only` | Blocks all remote mutations (create, delete, publish, upload). Returns `READ_ONLY_VIOLATION`. |
| `--dry-run` | Previews what would happen without executing. Shows planned actions. |
| `--confirm` | Required for destructive operations (delete, publish). Without it, returns `CONFIRMATION_REQUIRED`. |

These can also be set via environment variables for CI/automation:

```sh
ZENODO_READ_ONLY=1 zenodo records list --json     # safe in scripts
zenodo records publish ID --confirm                # explicit opt-in
zenodo files upload ID file.csv --dry-run          # preview first
```

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

## Development

```sh
git clone https://github.com/thedavidweng/zenodo-cli.git
cd zenodo-cli
make test          # run all tests
make lint          # fmt + vet + test
make build         # build to bin/zenodo
```

### Shell completions

```sh
source <(zenodo completion bash)
source <(zenodo completion zsh)
zenodo completion fish | source
zenodo completion powershell | Out-String | Invoke-Expression
```

## License

[Apache License 2.0](LICENSE)
