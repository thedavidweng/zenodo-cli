# ADR 0004: Secrets Storage

## Status

Accepted

## Context

zenodo-cli stores an API token per profile in a YAML config file. The token is
a bearer secret: anyone holding it can act as the user against the Zenodo API.
Two questions follow: where does the secret live, and how does it stay out of
plaintext when a user prefers not to commit it to disk.

An obvious option is the operating system keychain (macOS Keychain, the
Freedesktop Secret Service, Windows Credential Manager). But this CLI is
agent-first: its primary callers are CI jobs, scripts, and unattended data
pipelines. Keychain access on those platforms triggers an interactive unlock
prompt on first use and after re-lock, which hangs a headless process with no
one to answer it. A tool that blocks waiting for a GUI prompt is unusable in
the environment it is built for.

## Decision

The secrets contract is a `0600` config file plus environment-variable
indirection. There is no keychain integration.

- The config file is written atomically with `0600` permissions (owner
  read/write only), in a `0700` config directory.
- A profile `token` may be a literal token, or the indirection form `env:NAME`.
  When it is `env:NAME`, the value is resolved from environment variable `NAME`
  at load time. If `NAME` is unset, loading fails with an explicit error naming
  the missing variable — the CLI never silently proceeds unauthenticated.

Indirection is resolved once, at the single point where a profile becomes usable
credentials, so every caller (`records`, `auth status`, `doctor`, the raw `api`
escape hatch) sees the same resolved token and the same failure.

## Consequences

### Positive

- Fully headless: no GUI prompt can ever block an unattended run.
- Secrets can be kept out of the config file entirely — the file holds only the
  reference `env:NAME`, while the real token comes from the environment or a
  secret manager that exports it (CI secrets, `direnv`, a vault agent).
- The behaviour is identical across macOS, Linux, and Windows; there is no
  per-platform keychain code path to maintain or debug.

### Negative

- A literal token left in the config file is protected only by filesystem
  permissions, not by an OS-level encrypted store.
- The indirection resolves eagerly, so a profile referencing an unset variable
  is an error even for commands that would otherwise run unauthenticated.

### Mitigations

- `env:NAME` is the documented path for users who do not want a plaintext token
  on disk; the config file stays secret-free.
- The unset-variable error names the exact variable, so the fix is obvious.
