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
- **No token flags.** Tokens are never passed via `--token` flag on the command line. Use `--token-stdin` or `--token-env` for non-interactive use.
- **Secret redaction.** Debug output and error messages redact tokens and credentials.
- **Safety gates.** `--read-only` blocks all write operations. `--dry-run` previews mutations without sending requests. `--confirm` is required for destructive operations.
- **No telemetry.** `zenodo-cli` does not phone home, embed analytics, or send data to any server other than Zenodo/InvenioRDM.
