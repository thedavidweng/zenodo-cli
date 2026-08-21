# Publishing a Dataset

**Scenario:** you have data files ready for public release and want them on Zenodo as a citable record with a DOI.

Prerequisites: [installed](../../README.md#install) and [authenticated](../tutorials/getting-started.md#authenticate) CLI, `deposit:write` token scope.

## Write the metadata file

Quick records can be created with `--title`/`--description` alone, but a published dataset deserves full metadata. Put it in a JSON file:

```json
{
  "title": "Survey measurements 2024-2026",
  "description": "<p>Measurements from 42 stations.</p><p>Instrument: X, calibration: Y.</p>",
  "publication_date": "2026-08-21",
  "resource_type": { "id": "dataset" },
  "creators": [{ "name": "Doe, Jane" }, { "name": "Roe, John" }]
}
```

`description` accepts HTML. See Zenodo's [metadata reference](https://developers.zenodo.org/#representation) for all fields — the file is passed to the API as-is.

Next: [create the draft](#create-the-draft).

## Create the draft

Preview first if you like (`--dry-run` contacts nothing):

```sh
zenodo records create --title "Test" --dry-run
```

```text
Would create draft record (title="Test", metadata=)
```

Then create from your metadata file:

```sh
zenodo records create --metadata meta.json
```

```text
Created draft 22050993: Survey measurements 2024-2026
```

(`22050993` is the draft ID in this example run — use the ID your command prints.)

Next: [upload the files](#upload-files).

## Upload files

Pass one or more files — globs work too:

```sh
zenodo files upload 22050993 ./measurements.csv ./stations.geojson
```

```text
Uploaded measurements.csv

Uploaded stations.geojson

```

Verify what the draft holds:

```sh
zenodo files list 22050993
```

```text
stations.geojson (37 bytes)

measurements.csv (21 bytes)

```

Wrong file? Remove it again before publishing:

```sh
zenodo files delete 22050993 stations.geojson --confirm
```

```text
Deleted stations.geojson
```

Next: [publish](#publish).

## Publish

Publishing is **irreversible**: the record gets a DOI and can never be unpublished or deleted. Rehearse with a dry run first — it makes no network request, so it is safe on any draft (a throwaway one is shown here):

```sh
zenodo records publish 22051055 --dry-run --confirm
```

```text
Would publish draft 22051055 (irreversible)

{"ok":true,"data":{"action":"publish_draft","id":"22051055","planned":true},"meta":{"command":"records.publish","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"243e5a52-6e75-45f2-9b38-86ec2688187a"}}
```

When everything checks out:

```sh
zenodo records publish 22050993 --confirm
```

> Illustrative output — not captured, because publishing is irreversible and this guide's example record was never really published:
>
> ```text
> Published 22050993: Survey measurements 2024-2026
> ```

Without `--confirm`, the command refuses to run — captured against a real draft:

```text
Error [CONFIRMATION_REQUIRED]: use --confirm to proceed
```

Next: [verify the published record](#verify).

## Verify

Show the record — status should read `published` (real record shown here):

```sh
zenodo records show 20664361
```

```text
ID:          20664361

Title:       全国村界数据集 / China Village Boundaries Dataset

Status:      published

Created:     2026-06-12T10:45:05.593141+00:00

Updated:     2026-06-12T10:45:06.922225+00:00

Description: <p>中国全国范围的村级行政边界空间数据(Shapefile 格式),覆盖 33 个省级行政区划单位,58 个数据集,875,140 条记录。</p>…
```

The DOI appears in the raw API view under `doi` / `conceptdoi`:

```sh
zenodo api get /api/records/20664361 | jq '{doi, conceptdoi}'
```

```json
{
  "doi": "10.5281/zenodo.20664361",
  "conceptdoi": "10.5281/zenodo.20664360"
}
```

The concept DOI stays stable across versions; cite it and readers always land on the latest version.

## Where to next

- [Publishing a New Version](publish-a-new-version.md) — correct or extend a published record
- [Safety Gates](use-safety-gates.md) — rehearse every step of this guide without touching Zenodo
