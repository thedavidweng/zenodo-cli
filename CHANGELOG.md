# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] - 2026-06-11

Initial release.

### Added

#### Authentication
- API token-based authentication with Zenodo InvenioRDM
- Multi-profile credential management (`--profile`)
- Sandbox mode for testing (`--sandbox` / `ZENODO_SANDBOX=1`)
- `auth login` — store API token (interactive or `--token`)
- `auth status` — verify credentials against Zenodo API
- `auth logout` — remove stored credentials
- `doctor` — diagnostic checks for config, profile, and token
- Environment variable override for token (`ZENODO_TOKEN`)
- Secure config file storage with `0600` permissions

#### Record Management
- `records list` — list authenticated user's records
- `records create` — create a draft record (inline metadata or JSON file)
- `records show` — display record metadata
- `records delete` — delete a draft record with safety gates
- `records publish` — publish a draft (irreversible, requires `--confirm`)
- `records new-version` — create a new draft version of a published record

#### File Management
- `files upload` — upload files to a draft record
- `files list` — list files in a record
- `files download` — download files from a published record

#### Search
- `search` — search public Zenodo records with full-text query

#### API Access
- `api get` — GET request to any Zenodo InvenioRDM endpoint
- `api post` — POST request with JSON body
- `api put` — PUT request with JSON body

#### Safety
- `--read-only` flag blocks all remote mutations globally
- `--dry-run` shows planned actions without execution
- `--confirm` required for high-risk operations (delete, publish)
- Risk classification: read, medium-write, high-write

#### Output
- JSON envelope output (`--json`) on every command with consistent schema
- Pretty-print JSON (`--pretty`)
- NDJSON progress events to stderr (`--events`)
- Compact/full field modes (`--compact`, `--full`)
- Machine-readable error codes with categories and retryability flags
- Exit codes mapped to error categories

#### Build & Distribution
- Single binary, no runtime dependencies
- Cross-platform builds via GoReleaser (linux/darwin/windows, amd64/arm64)
- Version, commit, and date injected at build time via ldflags
- Cosign keyless signing of release checksums
- Homebrew tap (`thedavidweng/homebrew-tap`)
- Shell completions (bash, zsh, fish, powershell)

### Security
- Config file stored with `0600` permissions
- Parent directory created with `0700` permissions
- Tokens never printed in stdout, stderr, JSON, debug, or audit output
