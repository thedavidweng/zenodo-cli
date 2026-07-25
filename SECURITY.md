# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `zenodo-cli`, please report it responsibly:

1. **Do not open a public GitHub issue.**
2. Use [GitHub's private vulnerability reporting](https://github.com/thedavidweng/zenodo-cli/security/advisories/new).
3. Include a description of the vulnerability, steps to reproduce, and the potential impact.
4. You should receive an acknowledgement within 7 days.

## Scope

`zenodo-cli` handles Zenodo API tokens and can perform mutations including record creation, file upload/deletion, and metadata updates. The following are in scope:

- Token leakage (API tokens in logs, stderr, JSON output, or shell history)
- Bypass of safety gates (`--read-only`, `--dry-run`, `--confirm`)
- Unauthorized data access through CLI commands
- Secrets written to plaintext (config files, logs, stderr)

## Design Decisions

- **Token storage.** API tokens are stored in the OS config dir (`zenodo-cli/config.yaml`) with `0600` permissions.
- **Token input.** `zenodo auth login` accepts a token via the `--token` flag, the `ZENODO_TOKEN` environment variable, or an interactive prompt (used when neither is set and stdin is a terminal). Prefer `ZENODO_TOKEN` in automation to keep the token out of shell history.
- **Secret redaction.** The stored token is never echoed to stdout, stderr, or the JSON envelope in any output mode.
- **Safety gates.** `--read-only` blocks all write operations. `--dry-run` previews mutations without sending requests. `--confirm` is required for destructive operations.
- **No telemetry.** `zenodo-cli` does not phone home, embed analytics, or send data to any server other than Zenodo/InvenioRDM.
