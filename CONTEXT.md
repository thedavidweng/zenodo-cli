# zenodo-cli Domain Glossary

## Core Creed

**zenodo-cli is a replacement for the Zenodo web interface**, additionally providing automation convenience and agent-friendliness.

## Core Concepts

**zenodo-cli** — A local CLI tool for Zenodo/InvenioRDM, used to manage deposition records, upload and download files, and access the full InvenioRDM API. The installed binary is `zenodo`.

## Users

**Researcher** — Someone who needs to manage Zenodo deposition records and upload datasets.

**Agent** — An automation agent (CI, scripts, data pipelines). Requires deterministic behavior.

## Command Design Decisions

**Safety gates** — A three-tier safety mechanism: `--read-only`, `--dry-run`, `--confirm`.

**Raw API escape hatch** — `zenodo api` allows calling any InvenioRDM endpoint directly.
