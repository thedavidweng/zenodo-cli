# Using Safety Gates

**Scenario:** you are scripting zenodo-cli — a CI job, a bot, a batch job — and need hard guarantees that nothing mutates your Zenodo account by accident, plus a way to rehearse mutations before running them for real.

Three gates layer on top of each other:

| Gate | Effect |
|------|--------|
| `--read-only` | blocks every mutation, exit code 6 |
| `--dry-run` | prints the plan, performs nothing |
| `--confirm` | required for irreversible operations (`records publish`, `records delete`) |

All outputs in this guide were captured from real runs.

## Block everything: `--read-only`

Wrap any scripted session where mutations would be a bug:

```sh
ZENODO_READ_ONLY=1 zenodo files upload 12345 x.csv
```

```text
Error [READ_ONLY_VIOLATION]: --read-only blocks this mutation
```

Exit code 6 — the same result you get passing `--read-only` explicitly. The JSON envelope carries the machine-readable code:

```sh
zenodo files upload 12345 sample.csv --read-only --json
```

```json
{"ok":false,"error":{"code":"READ_ONLY_VIOLATION","message":"--read-only blocks this mutation","category":"safety"},"meta":{"command":"files.upload","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"63b300a3-de02-461d-9445-18833a1b9238"}}
```

Next: [rehearse with `--dry-run`](#preview-with---dry-run).

## Preview with `--dry-run`

Dry-run renders exactly what would happen and stops — no network request is made:

```sh
zenodo records create --title "Doc example" --dry-run
```

```text
Would create draft record (title="Doc example", metadata=)

{"ok":true,"data":{"action":"create_record","planned":true},"meta":{"command":"records.create","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"89238450-e466-41e5-888b-72d27306d7af"}}
```

The same works for uploads, version creation, and raw API calls:

```sh
zenodo records new-version 20664361 --dry-run
```

```text
Would create new version draft from 20664361

{"ok":true,"data":{"action":"new_version","id":"20664361","planned":true},"meta":{"command":"records.new-version","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"b2adecfb-0166-4a5d-998b-080b4e0eaae4"}}
```

Next: [confirm irreversible operations](#confirm-irreversible-operations).

## Confirm irreversible operations

`records publish` and `records delete` refuse to run without `--confirm`:

```sh
zenodo records publish 20664361
```

```text
Error [CONFIRMATION_REQUIRED]: use --confirm to proceed
```

Note the interaction: **`--dry-run` does not waive the confirmation**. Rehearsing a publish still requires `--confirm`, so a dry run can never publish anything by itself:

```sh
zenodo records delete 20664361 --dry-run
```

```text
Error [CONFIRMATION_REQUIRED]: use --confirm to proceed
```

With both flags you get the plan, not the action:

```sh
zenodo records publish 12345 --confirm --dry-run
```

```text
Would publish draft 12345 (irreversible)

{"ok":true,"data":{"action":"publish_draft","id":"12345","planned":true},"meta":{"command":"records.publish","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"ea8fa4ec-05c4-4778-9b10-7aebe296cb89"}}
```

## Gate matrix

| Command | Risk | `--read-only` | `--dry-run` | `--confirm` |
|---------|------|---------------|-------------|-------------|
| `records delete` | high | blocked, exit 6 | preview (needs `--confirm` too) | required |
| `records publish` | high | blocked, exit 6 | preview (needs `--confirm` too) | required |
| `files delete` | high | blocked, exit 6 | preview (needs `--confirm` too) | required |
| `records create` | medium | blocked, exit 6 | preview | not required |
| `records new-version` | medium | blocked, exit 6 | preview | not required |
| `files upload` | medium | blocked, exit 6 | preview | not required |
| `files import` | medium | blocked, exit 6 | preview | not required |
| `api post` / `api put` | high | blocked, exit 6 | preview (needs `--confirm` too) | required |

## Where to next

- [Global Flags & Environment Variables](../flags.md) — every gate flag and its env-var form
- [Agent Guide](../agent-guide.md) — exit codes and error handling for scripts
- [Safety Model](../safety.md) — design rationale behind the gates
