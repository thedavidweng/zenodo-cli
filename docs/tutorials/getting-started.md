# Getting Started

Create your first Zenodo record with zenodo-cli: install, authenticate, create a draft, upload a file, then clean up. Everything in this tutorial is reversible — nothing gets published.

All command outputs below were captured from real runs against zenodo.org unless marked otherwise.

## Install

macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/thedavidweng/zenodo-cli/main/install.sh | sh
```

Windows:

```sh
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/thedavidweng/zenodo-cli/main/install.ps1 | iex"
```

Verify the binary works:

```sh
zenodo version
```

```text
zenodo version dev (commit: unknown, date: unknown)
```

(The exact line depends on how the binary was built — a release install prints the release tag.)

Next: [authenticate](#authenticate).

## Authenticate

Create a token at <https://zenodo.org/account/settings/applications/tokens/> (grant `deposit:write` if you plan to upload), then store it:

```sh
zenodo auth login --token YOUR_TOKEN
```

```text
Token saved for profile "default"
```

Confirm the setup is healthy:

```sh
zenodo doctor
```

```text
[PASS] config

[PASS] profile

[PASS] token

[PASS] api


All checks passed.
```

If the token is rejected, `doctor` catches it before you start working:

```text
[FAIL] api: API unreachable: API error (HTTP 403): Permission denied.


Some checks failed.
```

Next: [create your first record](#create-a-draft).

## Create a draft

Nothing is public until you publish, so drafts are safe to experiment with:

```sh
zenodo records create --title "My First Dataset" --description "Data collected for my project"
```

```json
{
  "ok": true,
  "data": {
    "id": "22050983",
    "metadata": {
      "title": "My First Dataset",
      "description": "Data collected for my project",
      "creators": null,
      "publication_date": "2026-08-21",
      "resource_type": {
        "type": ""
      }
    },
    "created": "2026-08-21T21:49:13.698555+00:00",
    "updated": "2026-08-21T21:49:13.759250+00:00",
    "status": "draft"
  },
  "meta": {
    "command": "records.create",
    "profile": "default",
    "duration_ms": 1131,
    "schema_version": "2026-06-11",
    "request_id": "aa3a2891-350e-4ff4-93d6-5d950f47be4b"
  }
}
```

Note `"data.id"` — every later command uses it.

Next: [attach a file](#upload-a-file).

## Upload a file

Uploads only work on drafts:

```sh
zenodo files upload 22050983 ./sample.csv
```

```text
Uploaded sample.csv
```

Check what the draft now contains:

```sh
zenodo files list 22050983
```

```text
sample.csv (17 bytes)

```

Next: [review the draft](#inspect-the-draft) — or skip ahead to [Publishing a Dataset](../how-to/publish-a-dataset.md) when you are ready to go public.

## Inspect the draft

```sh
zenodo records show 22050983
```

```text
ID:          22050983

Title:       My First Dataset

Status:      draft

Created:     2026-08-21T21:49:13.698555+00:00

Updated:     2026-08-21T21:49:13.759250+00:00

Description: Data collected for my project
```

`Status: draft` means the record is still private. Publishing makes it public **forever** — a published record cannot be unpublished or deleted. That step gets its own guide: [Publishing a Dataset](../how-to/publish-a-dataset.md).

Next: [clean up your draft](#delete-the-draft).

## Delete the draft

Since this record was only for practice, remove it:

```sh
zenodo records delete 22050983 --confirm
```

```text
Deleted draft 22050983
```

`--confirm` is required because deletion cannot be undone. Your account is now exactly as it was before this tutorial.

## Search without a token

Search is public — it works before you configure anything:

```sh
zenodo search "village boundaries"
```

```text
[3569592] Advisory Board Meeting 5 Minutes. Deliverable number WP10_D10.6

[17607604] DEFEINITE-CCRI Bankable project profiles and business plans (investment ready)

[2431706] Advisory Board Meeting 4 Minutes. WP10_D10.5

…
```

(Truncated; results change as Zenodo ingests new records.)

## Where to next

- [Publishing a Dataset](../how-to/publish-a-dataset.md) — proper metadata, multiple files, and the publish step
- [Finding and Downloading Data](../how-to/find-and-download-data.md) — no token required
- [Safety Gates](../how-to/use-safety-gates.md) — rehearse mutations with `--dry-run` and `--read-only`
