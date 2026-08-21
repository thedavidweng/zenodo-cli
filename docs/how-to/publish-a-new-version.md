# Publishing a New Version

**Scenario:** a published record needs corrected files or updated metadata. Zenodo handles this with *versions*: one concept DOI, many dated versions. You never touch the published version itself — you create a draft version, change it, and publish that.

All outputs were captured from real runs except the two marked otherwise (this guide's example record was not actually versioned — versioning an existing record mid-guide would leave a stray draft behind).

## List existing versions

```sh
zenodo records versions 20664361
```

```text
[20664361] 全国村界数据集 / China Village Boundaries Dataset (published)


Total: 1 versions
```

Next: [open a draft version](#open-a-draft-version).

## Open a draft version

Preview the action, then run it for real:

```sh
zenodo records new-version 20664361 --dry-run
```

```text
Would create new version draft from 20664361

{"ok":true,"data":{"action":"new_version","id":"20664361","planned":true},"meta":{"command":"records.new-version","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"b2adecfb-0166-4a5d-998b-080b4e0eaae4"}}
```

> Illustrative output — not captured, to avoid leaving a draft on the example record:
>
> ```text
> Created new version 20670000 (from 20664361)
> ```

The new draft inherits metadata and files from the published version. Note its ID from the output.

Next: [update the draft](#update-the-draft).

## Update the draft

Remove the files that change, then bring in the new ones. `files delete` only removes files from a draft — published versions are immutable (captured below against a scratch draft):

```sh
zenodo files delete 22050918 sample.csv --confirm
```

```text
Deleted sample.csv
```

Upload replacements (same flow as any draft — see [Publishing a Dataset](publish-a-dataset.md#upload-files)):

```sh
zenodo files upload 22050918 ./sample.csv
```

```text
Uploaded sample.csv
```

If most files carry over unchanged, import them from the previous version instead of re-uploading:

```sh
zenodo files import 20664361 --dry-run
```

```text
Would import files from previous version into 20664361

{"ok":true,"data":{"action":"files_import","planned":true,"record_id":"20664361"},"meta":{"command":"files.import","profile":"default","duration_ms":1,"schema_version":"2026-06-11","request_id":"d1d8e74a-b926-46f7-ae12-f340687663f7"}}
```

Metadata edits go through the API escape hatch (the draft endpoint accepts a PUT with new metadata):

```sh
zenodo api put /api/records/VERSION_ID/draft --data '{"metadata":{…}}' --confirm
```

Next: [publish the version](#publish-the-version).

## Publish the version

Same gate as any publish — dry run first, then confirm:

```sh
zenodo records publish VERSION_ID --dry-run --confirm
```

> Illustrative output (shape identical to the captured dry run in [Publishing a Dataset](publish-a-dataset.md#publish)):
>
> ```text
> Would publish draft VERSION_ID (irreversible)
> ```

> Illustrative output — not captured, for the same reason as above:
>
> ```text
> Published VERSION_ID: 全国村界数据集 / China Village Boundaries Dataset
> ```

Once published, the concept DOI (`10.5281/zenodo.<concept-id>`) resolves to this new version automatically.

Next: [verify](#verify).

## Verify

```sh
zenodo records versions 20664361
```

The listing now shows the new version alongside the old one, with the newest marked as the latest.

## Where to next

- [Safety Gates](use-safety-gates.md) — rehearse this whole flow with `--dry-run`
- [Calling the API Directly](call-the-api-directly.md) — metadata fields the CLI flags don't cover
