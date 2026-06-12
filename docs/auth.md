# Authentication

## API Token

zenodo-cli uses Zenodo API tokens for authentication.

### First-time Setup

```bash
zenodo auth login --token YOUR_TOKEN
```

Or interactive prompt:

```bash
zenodo auth login
```

Get a token at: https://zenodo.org/account/settings/applications/tokens/

### Environment Variables

```
ZENODO_TOKEN        API token (overrides config file)
ZENODO_SANDBOX      Set 1/true/yes to use sandbox
ZENODO_API_URL      Override API base URL
```

### Check Status

```bash
zenodo auth status --json
```

### Logout

```bash
zenodo auth logout
```

## Config Schema

Credentials are stored in:

```text
$XDG_CONFIG_HOME/zenodo-cli/config.yaml
~/.config/zenodo-cli/config.yaml
```

```yaml
current_profile: default
profiles:
  default:
    token: "REDACTED"
    sandbox: false
  staging:
    token: "REDACTED"
    sandbox: true
```

### Profiles

Multiple accounts or environments via profiles:

```bash
zenodo auth login --token TOKEN --profile staging --sandbox
zenodo records list --profile staging
```

### Sandbox Mode

Test against the Zenodo sandbox (sandbox.zenodo.org) without touching production data:

```bash
zenodo auth login --token SANDBOX_TOKEN --sandbox
zenodo records list --sandbox
# or set globally
export ZENODO_SANDBOX=1
```
