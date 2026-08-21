# Testing in the Zenodo Sandbox

**Scenario:** you want to rehearse uploads and publishing against a throwaway environment before touching production. Zenodo runs a full copy of the service at [sandbox.zenodo.org](https://sandbox.zenodo.org) with separate accounts and tokens.

**This guide's outputs are illustrative** (marked where so): no sandbox account was used while writing it, but every command is the production twin documented with real captures in the other guides — only the host differs.

## Create a sandbox profile

Sign up at <https://sandbox.zenodo.org>, then create a token there (same flow as production) and store it under its own profile:

```sh
zenodo auth login --token SANDBOX_TOKEN --profile staging --sandbox
```

> Illustrative output:
>
> ```text
> Token saved for profile "staging"
> ```

The `--profile staging` keeps both tokens side by side; `--sandbox` points that profile at `sandbox.zenodo.org`.

Next: [verify the setup](#verify-the-setup).

## Verify the setup

```sh
zenodo doctor --profile staging
```

> Illustrative output — same shape as a healthy production check, with the sandbox flag set:
>
> ```text
> [PASS] config
>
> [PASS] profile
>
> [PASS] token
>
> [PASS] api
>
>
> All checks passed.
> ```

Next: [run your rehearsal](#run-your-rehearsal).

## Run your rehearsal

Every command accepts `--profile`:

```sh
zenodo records create --title "Rehearsal" --profile staging
zenodo files upload DRAFT_ID ./data.csv --profile staging
zenodo records publish DRAFT_ID --confirm --profile staging
```

Or export the choice once for a whole script:

```sh
export ZENODO_PROFILE=staging
zenodo records list
```

`ZENODO_SANDBOX=1` switches an existing profile to the sandbox host instead:

```sh
ZENODO_SANDBOX=1 zenodo doctor
```

Publishing in the sandbox works exactly like production — irreversible *within the sandbox*, which is what makes it useful practice. Sandbox deposits are periodically cleaned out by Zenodo.

## Where to next

- [Authentication](../auth.md) — profiles, config file layout, env vars
- [Publishing a Dataset](publish-a-dataset.md) — the flow to rehearse
