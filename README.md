# zenodo-cli

[![CI](https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/thedavidweng/zenodo-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/thedavidweng/zenodo-cli)](https://github.com/thedavidweng/zenodo-cli/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/thedavidweng/zenodo-cli.svg)](https://pkg.go.dev/github.com/thedavidweng/zenodo-cli)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](https://github.com/thedavidweng/zenodo-cli/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go)](https://go.dev/)

Agent-friendly Zenodo CLI. Single-binary tool for deposit management, file upload/download, and full InvenioRDM API access.

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

**Homebrew (macOS/Linux):**

```shell
brew tap thedavidweng/tap
brew install zenodo
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
# Homebrew
brew uninstall zenodo

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
