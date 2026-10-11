# zenodo-cli

Go/Cobra CLI for Zenodo/InvenioRDM records, files, search and guarded publication.

- Toolchain and checks: Go/Cobra with `mise`; run `mise run check` before every push. CI runs formatting, build, test, lint and conventions.
- The canonical agent entry point is this `AGENTS.md`; do not add IDE-specific instruction copies. CI enforces this rule.
- Read task-specific guides:
  - [Engineering and contract docs](docs/agents/engineering.md).
  - [Testing](docs/agents/testing.md).
  - [ADRs and issue workflow](docs/agents/workflow.md).
