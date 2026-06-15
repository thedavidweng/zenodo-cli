<p align="center">
  <img src="public/icon.png" alt="zenodo-cli" width="160" />
</p>

<h1 align="center">zenodo-cli</h1>

<p align="center">
  Agent-friendly Zenodo CLI for records, files, search, and InvenioRDM API access.
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml"><img src="https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/thedavidweng/zenodo-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/zenodo-cli" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/thedavidweng/zenodo-cli"><img src="https://pkg.go.dev/badge/github.com/thedavidweng/zenodo-cli.svg" alt="Go Reference"></a>
  <a href="https://github.com/thedavidweng/zenodo-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go" alt="Go">
</p>

`zenodo-cli` is a single-binary CLI for creating Zenodo records, uploading files, publishing drafts, searching public records, and calling the current InvenioRDM API directly.

## Highlights

- Single binary: no runtime, containers, or Python environment required
- JSON-first: `--json` on every command with a consistent envelope
- Safety gates: `--read-only`, `--dry-run`, and `--confirm` for remote mutations
- Agent-ready: predictable exit codes, error categories, request IDs, and secret redaction
- Sandbox support: test against `sandbox.zenodo.org` with `--sandbox` or `ZENODO_SANDBOX=1`

## Why

Zenodo moved from the legacy deposit API to InvenioRDM in 2023. Tools built around `/api/deposit/depositions` can still install cleanly while failing on uploads and record management. `zenodo-cli` is built directly on the current `/api/records` API surface.

**Used in production:** [全国村界数据集 / China Village Boundaries Dataset](https://zenodo.org/records/20664361) was uploaded and published with `zenodo-cli`: 875,140 records across 58 Shapefile datasets.

<details>
<summary>Comparison with older Zenodo tools</summary>

| Tool | Language | API Version | Upload Works | CLI | Status |
|------|----------|-------------|---------------|-----|--------|
| **zenodo-cli** | Go | InvenioRDM | Yes | Yes | Active |
| [zenodo-client](https://github.com/cthoyt/zenodo-client) | Python | Mixed old/new | No | Yes | Upload broken |
| [zenodo](https://github.com/cheminfo/zenodo) | Node.js | InvenioRDM | Yes | No | Library only |
| [zenodraft](https://github.com/zenodraft/zenodraft) | Node.js | Old deposit API | No | Yes | Deprecated API |
| [zotzen-lib](https://github.com/OpenDevEd/zotzen-lib) | Node.js | Old deposit API | No | Yes | Unmaintained |
| [zenodo-cli](https://github.com/OpenDevEd/zenodo-cli) | Node.js | Old deposit API | No | Yes | Deprecated |

</details>

## Quickstart

### Install

Run the following on macOS or Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/thedavidweng/zenodo-cli/main/install.sh | sh
```

Run the following on Windows:

```shell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/thedavidweng/zenodo-cli/main/install.ps1 | iex"
```

The installer detects Homebrew automatically and uses it when available (recommended for easy upgrades). Otherwise it downloads the binary to `~/.local/bin`.

<details>
<summary>Other installation methods</summary>

**Homebrew Cask (macOS/Linux):**

```shell
brew tap thedavidweng/tap
brew install --cask zenodo
```

**Go:**

```shell
go install github.com/thedavidweng/zenodo-cli/cmd/zenodo@latest
```

**Manual download:** grab the archive for your platform from the [latest GitHub Release](https://github.com/thedavidweng/zenodo-cli/releases/latest), extract it, and place the `zenodo` binary on your `PATH`.

**Build from source:**

```shell
git clone https://github.com/thedavidweng/zenodo-cli.git
cd zenodo-cli
make build
```

</details>

### Set up

```shell
zenodo auth login --token YOUR_TOKEN
zenodo doctor
```

Get a token at https://zenodo.org/account/settings/applications/tokens/

Then try it:

```shell
zenodo records create --title "My Dataset" --description "A research dataset"
zenodo files upload RECORD_ID ./data.csv
zenodo records publish RECORD_ID --confirm
```

### Uninstall

```shell
# Homebrew Cask
brew uninstall --cask zenodo

# install.sh
curl -fsSL https://raw.githubusercontent.com/thedavidweng/zenodo-cli/main/install.sh | sh -s uninstall

# Go
rm "$(go env GOPATH)/bin/zenodo"
```

Remove config if desired: `rm -rf ~/.config/zenodo-cli`

## Documentation

- [Command Reference](docs/commands.md) — all commands with flags, examples, and configuration
- [Authentication](docs/auth.md) — token setup, profiles, sandbox mode
- [Safety Model](docs/safety.md) — `--read-only`, `--dry-run`, `--confirm` gates
- [JSON Contract](docs/json-contract.md) — envelope schema, output modifiers
- [Global Flags & Environment Variables](docs/flags.md) — all CLI flags and env vars
- [Capabilities](docs/capabilities.md) — high-level feature overview
- [Architecture](docs/ARCHITECTURE.md) — codebase structure and design decisions
- [Agent Guide](docs/agent-guide.md) — scripting, JSON mode, exit codes

## Infrastructure

- **CI/CD:** [cli-workflow-template](https://github.com/thedavidweng/cli-workflow-template) — reusable GitHub Actions workflows
- **Docs:** [site](https://github.com/thedavidweng/site) — landing page and documentation

## License

[Apache License 2.0](LICENSE)
