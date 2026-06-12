# Safety

## Safety Gates

### --read-only

Blocks all remote mutations. Commands that would modify Zenodo state return exit code 6.

```bash
zenodo files upload ID file.csv --read-only
# Error: read-only mode blocks mutation
```

### --dry-run

Shows planned actions without executing remote mutations.

```bash
zenodo records create --title "Test" --dry-run --json
# Output shows planned creation but performs 0 mutations
```

### --confirm

Required for high-risk operations:

- `records delete`
- `records publish`

```bash
zenodo records publish ID --confirm
```

## Mutation Commands and Risk Classification

### High Risk (require --confirm)

These operations are destructive and irreversible. They are blocked unless
`--confirm` is passed. `--read-only` also blocks them (exit code 6).
`--dry-run` shows what would happen without executing.

| Command | Description |
|---------|-------------|
| `records delete` | Permanently delete a draft record |
| `records publish` | Publish a draft (irreversible) |

```bash
# Blocked without --confirm
zenodo records delete ID
# Error: high-risk operation requires --confirm

# Explicit confirmation
zenodo records delete ID --confirm
```

### Medium Risk (blocked by --read-only)

These operations modify remote state but are not inherently destructive.
They are blocked by `--read-only` (exit code 6). `--dry-run` shows planned
actions without executing. `--confirm` is **not** required.

| Command | Description |
|---------|-------------|
| `records create` | Create a new draft record |
| `records new-version` | Create a new version of a published record |
| `files upload` | Upload files to a draft record |

### Safety Gate Summary

| Gate | High Risk | Medium Risk |
|------|-----------|-------------|
| `--read-only` | Blocked (exit 6) | Blocked (exit 6) |
| `--dry-run` | Preview only | Preview only |
| `--confirm` | **Required** | Not required |

## Secret Redaction

The following fields are automatically redacted in output:
- `token`
- `ZENODO_TOKEN`
