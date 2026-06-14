# Command Reference

## Quick Examples

```sh
# Auth
zenodo auth login --token TOKEN        # store API token (non-interactive)
zenodo auth login                      # interactive prompt
zenodo auth status                     # check current authentication
zenodo auth logout                     # remove stored credentials

# Records
zenodo records list                                  # list your records
zenodo records create --title "Title"                # create a draft record
zenodo records create --metadata record.json         # create from JSON metadata file
zenodo records show RECORD_ID                        # show record details
zenodo records delete RECORD_ID --confirm            # delete a draft record
zenodo records publish RECORD_ID --confirm           # publish a draft (irreversible)
zenodo records new-version RECORD_ID                 # create a new draft version

# Files
zenodo files upload RECORD_ID file1.csv file2.csv    # upload files to a draft
zenodo files list RECORD_ID                          # list files in a record
zenodo files download RECORD_ID --dest ./output      # download files from a published record

# Search
zenodo search "machine learning"                     # search public records
zenodo search "climate data" --json | jq '.data.hits[].id'

# API (direct access to any Zenodo InvenioRDM endpoint)
zenodo api get /api/records                          # GET request
zenodo api get records?q=climate                     # path prefix added automatically
zenodo api post /api/records --data '{"metadata":{"title":"Test"}}'
zenodo api put /api/records/ID/draft --data '{"metadata":{...}}'

# Doctor
zenodo doctor                                        # check config, profile, and token
zenodo doctor --json                                 # machine-readable diagnostics

# Shell completions
source <(zenodo completion bash)
source <(zenodo completion zsh)
zenodo completion fish | source
zenodo completion powershell | Out-String | Invoke-Expression
```

## Configuration

Config file path: `~/.config/zenodo-cli/config.yaml` (respects `XDG_CONFIG_HOME`).

The config file stores named profiles, each with its own token and settings:

```yaml
current_profile: default
profiles:
  default:
    token: "your-api-token"
    sandbox: false
  staging:
    token: "sandbox-token"
    sandbox: true
```

### Profiles

Switch profiles with `--profile`:

```sh
zenodo records list --profile staging
```

### Sandbox mode

Test against the Zenodo sandbox (sandbox.zenodo.org) without touching production data:

```sh
zenodo auth login --token SANDBOX_TOKEN --sandbox
zenodo records list --sandbox
# or set globally
export ZENODO_SANDBOX=1
```

See [Authentication](auth.md) for full auth details.
