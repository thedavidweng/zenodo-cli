# Calling the API Directly

**Scenario:** you need a Zenodo API endpoint that no higher-level command covers — an unusual query, a community action, a field the CLI doesn't expose. `zenodo api` is the escape hatch: raw GET/POST/PUT against any InvenioRDM path.

All outputs in this guide were captured from real runs.

## GET

Paths start with `/api`:

```sh
zenodo api get /api/records/20664361
```

```json
{
  "conceptdoi": "10.5281/zenodo.20664360",
  "conceptrecid": "20664360",
  "created": "2026-06-12T10:45:05.593141+00:00",
  "doi": "10.5281/zenodo.20664361",
  "doi_url": "https://doi.org/10.5281/zenodo.20664361",
  "files": [
    {
      "checksum": "md5:a06ef552141719c1acc8ba7dc2d2179c",
      …
```

(Trimmed — the response continues with all file entries and metadata.)

Pipe through `jq` to pull out exactly what you need:

```sh
zenodo api get /api/records/20664361 | jq '{doi, conceptdoi}'
```

```json
{
  "doi": "10.5281/zenodo.20664361",
  "conceptdoi": "10.5281/zenodo.20664360"
}
```

Query strings work like any URL:

```sh
zenodo api get '/api/records?q=village+boundaries&size=1' | jq '.hits.hits[0].id'
```

```text
"3569592"
```

Next: [the path pitfall](#always-include-the-api-prefix).

## Always include the `/api` prefix

The path is used verbatim (only a leading slash is added). A browser-style path without `/api` fetches the HTML landing page, and JSON decoding fails:

```sh
zenodo api get records/20664361
```

```text
Error [ZENODO_API_ERROR]: decode response: invalid character '<' looking for beginning of value
```

If you see `invalid character '<'`, the first thing to check is a missing `/api` prefix.

Next: [POST and PUT](#post-and-put).

## POST and PUT

Write requests are high-risk operations: they require `--confirm`, which makes drive-by mutations impossible:

```sh
zenodo api post /api/records --data '{"metadata":{"title":"x"}}'
```

```text
Error [CONFIRMATION_REQUIRED]: use --confirm to proceed
```

Rehearse with `--dry-run` (plus `--confirm`, since dry runs do not waive it):

```sh
zenodo api post /api/records --data '{"metadata":{"title":"x"}}' --confirm --dry-run
```

```text
Would POST to /api/records

{"ok":true,"data":{"method":"POST","path":"/api/records","planned":true},"meta":{"command":"api.post","profile":"default","duration_ms":0,"schema_version":"2026-06-11","request_id":"f8443baa-c03c-41ef-bb5d-5c475e4c23ae"}}
  body: {"metadata":{"title":"x"}}
```

Real calls send your stored token as a Bearer header, so authenticated endpoints work too — e.g. `zenodo api get /api/user/records`.

## Where to next

- [Agent Guide](../agent-guide.md) — scripting patterns on top of JSON mode
- [InvenioRDM REST API](https://inveniordm.docs.cern.ch/reference/rest_api/) — what endpoints exist
