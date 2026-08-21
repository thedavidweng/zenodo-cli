# Finding and Downloading Data

**Scenario:** you want to find a public Zenodo record and get its files onto your machine. No token or account needed — every command here works without authentication.

## Search

```sh
zenodo search "village boundaries"
```

```text
[3569592] Advisory Board Meeting 5 Minutes. Deliverable number WP10_D10.6

[17607604] DEFEINITE-CCRI Bankable project profiles and business plans (investment ready)

[2431706] Advisory Board Meeting 4 Minutes. WP10_D10.5

…
```

(Truncated — search returns up to 25 hits per page.)

For scripts, use JSON mode and pull out what you need:

```sh
zenodo search "village boundaries" --json | jq -r '.data.hits[].id'
```

```text
1476510
2431706
1473404
3569592
17607604
…
```

(Truncated — 25 IDs in the full output.)

The full hit objects carry titles, dates, creators, and file metadata — see [JSON Contract](../../JSON_SCHEMA.md) for the envelope.

Next: [inspect a record](#inspect-a-record).

## Inspect a record

Check what a result contains before downloading it:

```sh
zenodo records show 17607604
```

```text
ID:          17607604

Title:       DEFEINITE-CCRI Bankable project profiles and business plans (investment ready)

Status:      published

Created:     2025-11-14T09:51:23.931679+00:00

Updated:     2025-11-14T09:51:25.236024+00:00

Description: <p>The purpose of this report is to provide a structured and comprehensive overview of four key projects&mdash;<br>Tissel (Roubaix, France), Circular Library Network (Reykjavik, Iceland), Tehdassaari (Nokia, Finland),<br>and Return2Sender (Ghent, Belgium) &mdash; that progressed to the due diligence phase of the Circular Cities<br>and Regions Initiative (CCRI).</p>
<p>The report outlines the support provided by the CCRI consortium to prepare these projects for<br>investment readiness.</p>…
```

(The description is trimmed here; `records show` prints it in full.)

List its files with sizes:

```sh
zenodo files list 17607604
```

```text
DEFINITE-CCRI Bankable Project Profiles and Business Plans.pdf (4262817 bytes)

```

Next: [download](#download).

## Download

Downloads all files of a published record into a directory:

```sh
zenodo files download 17607604 --dest ./data
```

```text
Downloaded files from 17607604 to ./data
```

```sh
ls -la ./data
```

```text
-rw-r--r--@ 1 david  wheel  4262817 Aug 21 14:52 DEFINITE-CCRI Bankable Project Profiles and Business Plans.pdf
```

Records with many versions may have newer ones — `--latest` resolves the newest published version before downloading:

```sh
zenodo files download 20664361 --latest --dest ./boundaries
```

> Illustrative output — not run, because that record holds ~10 GB of Shapefiles:
>
> ```text
> Downloaded files from 20664361 to ./boundaries
> ```

## Where to next

- [Publishing a Dataset](publish-a-dataset.md) — put your own data on Zenodo
- [Calling the API Directly](call-the-api-directly.md) — query fields the CLI commands don't expose
