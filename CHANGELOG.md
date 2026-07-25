# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.2.0](https://github.com/thedavidweng/zenodo-cli/compare/v0.1.1...v0.2.0) (2026-07-25)


### Features

* add download --latest, enforce --read-only/--dry-run, improve CLI UX ([092b310](https://github.com/thedavidweng/zenodo-cli/commit/092b31061d9e38ca1511c4a8bb43b46c65cb0027))
* add file deletion, version listing, DOI reservation, community submission ([78581b3](https://github.com/thedavidweng/zenodo-cli/commit/78581b37662afd4e9822cd470a9c774d1ad11ec2))
* add files info/import, records requests; update CHANGELOG ([b4812cd](https://github.com/thedavidweng/zenodo-cli/commit/b4812cdbb4f2cb60b9f88125dd54413326c4c98b))
* use os.UserConfigDir for cross-platform config paths ([30f9346](https://github.com/thedavidweng/zenodo-cli/commit/30f93462badac5e48e828386c53d3e0c91297df8))


### Bug Fixes

* add force_push and fetch-depth: 0 ([dd938b2](https://github.com/thedavidweng/zenodo-cli/commit/dd938b2837e4ba0560272579191183ec89e8b605))
* add force_push and fetch-depth: 0 ([7cdd242](https://github.com/thedavidweng/zenodo-cli/commit/7cdd242fdb3d61393f7acd39a820542418717dfa))
* add missing // Output: comment to ExampleLoad ([4c6ae6a](https://github.com/thedavidweng/zenodo-cli/commit/4c6ae6a397aa1804fe8518dc0606e1330b9ada11))
* add missing internal/zenodo and cmd/zenodo packages ([33da03a](https://github.com/thedavidweng/zenodo-cli/commit/33da03a66aa69a1d41f57c0ebfa9d71c0a29d744))
* add syft installation to release workflow for SBOM generation ([fefee12](https://github.com/thedavidweng/zenodo-cli/commit/fefee12f85636e26d5b565923e28cc04d57d8ff0))
* address Go production review audit findings ([8267fce](https://github.com/thedavidweng/zenodo-cli/commit/8267fce50501eb9de78bf555c3f922bc9f347a65))
* address review feedback on test quality ([dc797b9](https://github.com/thedavidweng/zenodo-cli/commit/dc797b9de5d3f2644da8825c9e2579707de81c61))
* config Save fails on Windows when file already exists ([7a01c5f](https://github.com/thedavidweng/zenodo-cli/commit/7a01c5f14341d13a7a7ddbb676430ffcc9bec883))
* correct mirror action SHA ([9b7b81c](https://github.com/thedavidweng/zenodo-cli/commit/9b7b81c2948258515f8bf205638ac14bcf284ce8))
* correct mirror action SHA ([c06ca3b](https://github.com/thedavidweng/zenodo-cli/commit/c06ca3b2cd48bc5097c8fcf5de810dd15cbdc9e6))
* install syft via curl instead of missing GitHub Action ([b9004b9](https://github.com/thedavidweng/zenodo-cli/commit/b9004b962588338596e9eac8a253a86d70fdb8b7))
* pass raw metadata JSON, increase timeout, use InvenioRDM 5.x creators ([5fa1b98](https://github.com/thedavidweng/zenodo-cli/commit/5fa1b9806289143a20b66399cff1232d2c88ac66))
* pin action SHA, remove test.txt, add permissions ([4b2b31e](https://github.com/thedavidweng/zenodo-cli/commit/4b2b31ed55e58ffaa1d74e7f5cbc14a4968ccb31))
* remove manual completions step from release workflow ([f01b280](https://github.com/thedavidweng/zenodo-cli/commit/f01b2807c69d6b8bc814c8beb1eb5d5f3f422bae))
* resolve lint issues in zenodo package ([6830a39](https://github.com/thedavidweng/zenodo-cli/commit/6830a39ee4773fafca5b02a63f8a751c89073d52))
* separate lint job and add go mod download ([8a50316](https://github.com/thedavidweng/zenodo-cli/commit/8a503169877e82c568096418ea7156800f8de5b9))
* silence errcheck lint warnings in test files ([aecc2a6](https://github.com/thedavidweng/zenodo-cli/commit/aecc2a61de328de6a3512fdb2cca163b76a4afd3))
* skip permission tests on Windows ([0ec3b1e](https://github.com/thedavidweng/zenodo-cli/commit/0ec3b1ea362aeb61ca7803ad3615ac366df3a8c0))
* update CI to match flickr-cli pattern ([90b47bc](https://github.com/thedavidweng/zenodo-cli/commit/90b47bc1a96754efcba7c71d63ba8c9ad7d60a3e))
* update go.mod dependencies (direct vs indirect) ([dedc34d](https://github.com/thedavidweng/zenodo-cli/commit/dedc34d4f26b6274213cd8edcc32186c8652757f))


### Performance

* optimize logo to WebP ([dcd4779](https://github.com/thedavidweng/zenodo-cli/commit/dcd477930ae18fd1b23c76d7612fc86aceb4667d))


### Refactoring

* deepen shallow modules for testability and locality ([#15](https://github.com/thedavidweng/zenodo-cli/issues/15)) ([f822777](https://github.com/thedavidweng/zenodo-cli/commit/f822777d51a2979ed8682b43552d009e72b07d31))
* harden codebase to production review standards ([#14](https://github.com/thedavidweng/zenodo-cli/issues/14)) ([32e93e3](https://github.com/thedavidweng/zenodo-cli/commit/32e93e3e388cd912fe3406e8caa50dec235b3765))
* reduce cyclomatic complexity in newRootCmd and handleRecordSubpath, fix gofmt ([2fe65f9](https://github.com/thedavidweng/zenodo-cli/commit/2fe65f9b6473d79ed30fd7130254cbcdb28c750b))
* remove config migration logic ([164fb33](https://github.com/thedavidweng/zenodo-cli/commit/164fb335cc01124d159fab389d58b641c3f4eaba))
* use reusable workflows from cli-workflow-template ([48e5237](https://github.com/thedavidweng/zenodo-cli/commit/48e52376729b14ef598fe3332ad0d78f024ecbbf))


### Documentation

* add Go Report Card badge ([9c22de7](https://github.com/thedavidweng/zenodo-cli/commit/9c22de7bcd9a632e17a179464c9dc35fd7d2be37))
* add infrastructure links (CI/CD and docs) ([ce879ee](https://github.com/thedavidweng/zenodo-cli/commit/ce879ee21483f6f93714293d6c6c85f97bd3f83c))
* add real-world usage example to README ([c699ae5](https://github.com/thedavidweng/zenodo-cli/commit/c699ae549c7e0393f0618441bda574f822d3bf03))
* add root-level docs for site sync (COMMANDS.md, JSON_SCHEMA.md, CONTEXT.md) ([fe9c65b](https://github.com/thedavidweng/zenodo-cli/commit/fe9c65bef8f9c570a5e3cad58dafc8c35479fb3f))
* add verified GitHub links to comparison table ([2b54707](https://github.com/thedavidweng/zenodo-cli/commit/2b54707e9eed425ba2d2362fc29bc1d66c395896))
* add Zenodo logo to README ([162fc73](https://github.com/thedavidweng/zenodo-cli/commit/162fc73528be26da4212b04dbe70a4ec01d61b05))
* remove duplicate and stale docs ([2ee3015](https://github.com/thedavidweng/zenodo-cli/commit/2ee30150455589349e1a14073bd3699270b6b46a))
* remove inaccurate 'first CLI' claim from README ([939e34f](https://github.com/thedavidweng/zenodo-cli/commit/939e34f19c1e12241889cc27ed3ca4c0236bfd7d))
* remove self-link from comparison table ([8464f74](https://github.com/thedavidweng/zenodo-cli/commit/8464f747b3b5b0b0e244fef5fdddd2b14c19bbe3))
* restructure README to match canvas-cli pattern ([33a36f6](https://github.com/thedavidweng/zenodo-cli/commit/33a36f64dd96fef7ef7b69241378712b918426ec))
* standardize README badges ([1e47139](https://github.com/thedavidweng/zenodo-cli/commit/1e47139de28ab10b173a6bce19a778d7d87593df))

## [0.1.1] - 2026-06-12

### Added

#### Record Management
- `records versions` — list all versions of a record
- `records reserve-doi` — reserve a DOI for a draft record
- `records submit` — submit a draft for community review (`--community`)
- `records requests` — list community review requests

#### File Management
- `files delete` — delete files from a draft record
- `files info` — show metadata for a single file in a draft
- `files import` — import files from previous version into a new draft
- `files list` — now supports both draft and published records (auto-detect)
- `files download --latest` — resolve and download the latest published version

### Changed

#### CLI UX
- All commands now have `Long` descriptions and `Example` blocks
- `--read-only` enforced on all mutation commands (was dead code)
- `--dry-run` enforced on `records create/delete/publish/new-version` and `api post/put`
- `api post` and `api put` now require `--confirm`
- Flag descriptions clarified (`--compact`, `--full`, `--config`, `--profile`, etc.)
- `doctor` now checks API connectivity in addition to config/token
- `search` total output moved to stdout (was stderr)
- New environment variables: `ZENODO_JSON`, `ZENODO_READ_ONLY`, `ZENODO_DRY_RUN`, `ZENODO_CONFIRM`, `ZENODO_QUIET`

#### Build & CI
- Release workflow: removed manual completions step (GoReleaser handles via `generate_completions_from_executable`)
- Release workflow: added `syft` installation for SBOM generation

## [0.1.0] - 2026-06-12

Initial release.

### Added

#### Authentication
- API token-based authentication with Zenodo InvenioRDM
- Multi-profile credential management (`--profile`)
- Sandbox mode for testing (`--sandbox` / `ZENODO_SANDBOX=1`)
- `auth login` — store API token (interactive or `--token`)
- `auth status` — verify credentials against Zenodo API
- `auth logout` — remove stored credentials
- `doctor` — diagnostic checks for config, profile, and token
- Environment variable override for token (`ZENODO_TOKEN`)
- Secure config file storage with `0600` permissions

#### Record Management
- `records list` — list authenticated user's records
- `records create` — create a draft record (inline metadata or JSON file)
- `records show` — display record metadata
- `records delete` — delete a draft record with safety gates
- `records publish` — publish a draft (irreversible, requires `--confirm`)
- `records new-version` — create a new draft version of a published record

#### File Management
- `files upload` — upload files to a draft record
- `files list` — list files in a record
- `files download` — download files from a published record

#### Search
- `search` — search public Zenodo records with full-text query

#### API Access
- `api get` — GET request to any Zenodo InvenioRDM endpoint
- `api post` — POST request with JSON body
- `api put` — PUT request with JSON body

#### Safety
- `--read-only` flag blocks all remote mutations globally
- `--dry-run` shows planned actions without execution
- `--confirm` required for high-risk operations (delete, publish)
- Risk classification: read, medium-write, high-write

#### Output
- JSON envelope output (`--json`) on every command with consistent schema
- Pretty-print JSON (`--pretty`)
- NDJSON progress events to stderr (`--events`)
- Compact/full field modes (`--compact`, `--full`)
- Machine-readable error codes with categories and retryability flags
- Exit codes mapped to error categories

#### Build & Distribution
- Single binary, no runtime dependencies
- Cross-platform builds via GoReleaser (linux/darwin/windows, amd64/arm64)
- Version, commit, and date injected at build time via ldflags
- Cosign keyless signing of release checksums
- Homebrew tap (`thedavidweng/homebrew-tap`)
- Shell completions (bash, zsh, fish, powershell)

### Security
- Config file stored with `0600` permissions
- Parent directory created with `0700` permissions
- Tokens never printed in stdout, stderr, JSON, debug, or audit output
