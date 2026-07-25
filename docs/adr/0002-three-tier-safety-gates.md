# ADR 0002: Three-Tier Safety Gates for Mutations

## Status

Accepted

## Context

zenodo-cli is designed for automation, where an agent may issue commands
without a human reviewing each one. Some Zenodo operations are irreversible:
publishing a record mints a DOI that cannot be withdrawn, and deleting a draft
is permanent. A single blanket confirmation flag cannot distinguish a harmless
read from an irreversible publish, and an agent running unattended needs a way
to prove a command performs no writes before trusting it.

## Decision

Every command declares a risk tier, and mutations pass through one `Gate`
before any request is sent:

- **read** — no mutation; always allowed.
- **medium-write** — creates or modifies remote state but is not destructive
  (create draft, new version, upload files, reserve DOI, import files). Blocked
  by `--read-only`; `--dry-run` prints the plan instead of executing.
- **high-write** — irreversible (delete draft, publish, submit to community,
  and every `api post`/`api put`). Blocked by `--read-only`, and additionally
  **requires `--confirm`**; `--dry-run` prints the plan.

A blocked mutation returns a typed error: `READ_ONLY_VIOLATION` or
`CONFIRMATION_REQUIRED`, both mapping to exit code 6. `--dry-run` emits a
success envelope describing the planned action with no request sent.

## Consequences

### Positive

- An agent can prove a command is side-effect-free by running it under
  `--read-only` and checking for exit 6.
- Irreversible operations cannot happen by accident: they need an explicit
  `--confirm`.
- `--dry-run` gives a machine-readable preview of exactly what would change.

### Negative

- Command authors must classify each new command's risk tier correctly; a
  misclassified high-write command would skip the `--confirm` requirement.

### Mitigations

- The tier is passed at the single `Gate.Allow` call site in each command, so
  the classification is visible in review next to the mutating call.
