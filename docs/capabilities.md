# Capabilities

This document describes the high-level capabilities of zenodo-cli.

## Authentication

- API token-based authentication with Zenodo InvenioRDM
- Multiple profiles support
- Sandbox mode for testing

## Record Management

- List, create, show, delete draft records
- Publish drafts (irreversible, with safety gate)
- Create new versions of published records
- Metadata from inline flags or JSON file

## File Management

- Upload files to draft records
- List files in a record
- Download files from published records

## Search

- Full-text search of public Zenodo records
- JSON output for programmatic processing

## API Access

- Direct access to any Zenodo InvenioRDM endpoint
- GET, POST, PUT methods
- Automatic path prefix handling

## Safety

- `--read-only` blocks all remote mutations
- `--dry-run` shows planned actions without execution
- `--confirm` required for high-risk operations (delete, publish)

## Output

- JSON envelope on every command (`--json`)
- Pretty-print (`--pretty`)
- NDJSON progress events (`--events`)
- Compact/full field modes
- Machine-readable error codes
- Exit codes mapped to error categories
