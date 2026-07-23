package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultBaseURL        = "https://zenodo.org"
	DefaultSandboxBaseURL = "https://sandbox.zenodo.org"
)

type Credentials struct {
	Token   string
	Sandbox bool
	BaseURL string
}

// CredentialsFromProfileAndEnv merges a profile with env overrides.
// Env vars take precedence: ZENODO_TOKEN, ZENODO_SANDBOX, ZENODO_API_URL.
func CredentialsFromProfileAndEnv(p *Profile) Credentials {
	c := Credentials{
		Token:   p.Token,
		Sandbox: p.Sandbox,
		BaseURL: p.BaseURL,
	}

	if v := os.Getenv("ZENODO_TOKEN"); v != "" {
		c.Token = v
	}
	if v := os.Getenv("ZENODO_SANDBOX"); v != "" {
		c.Sandbox = ParseBool(v)
	}
	if v := os.Getenv("ZENODO_API_URL"); v != "" {
		c.BaseURL = v
	}

	if c.BaseURL == "" {
		if c.Sandbox {
			c.BaseURL = DefaultSandboxBaseURL
		} else {
			c.BaseURL = DefaultBaseURL
		}
	}

	return c
}

func (c Credentials) IsAuthenticated() bool {
	return c.Token != ""
}

// ResolveClientConfig loads config and resolves credentials for the given profile,
// applying CLI flag overrides. When requireProfile is true, returns an error if
// the config file or profile is missing. When false, falls back to default base URL
// with an empty token — used by public commands like search.
func ResolveClientConfig(configFile, profileName string, sandbox, requireProfile bool) (Credentials, error) {
	cfgPath := configFile
	if cfgPath == "" {
		cfgPath = DefaultConfigPath()
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		if requireProfile {
			return Credentials{}, fmt.Errorf("not configured. Run 'zenodo auth login' to get started")
		}
		return fallbackCredentials(sandbox), nil
	}

	profile := cfg.GetProfileOrNil(profileName)
	if profile == nil {
		if requireProfile {
			return Credentials{}, fmt.Errorf("not authenticated. Run 'zenodo auth login' to get started")
		}
		return fallbackCredentials(sandbox), nil
	}

	creds := CredentialsFromProfileAndEnv(profile)

	// CLI sandbox override only swaps when base URL is the default.
	if sandbox {
		creds.Sandbox = true
		if creds.BaseURL == DefaultBaseURL {
			creds.BaseURL = DefaultSandboxBaseURL
		}
	}

	return creds, nil
}

func fallbackCredentials(sandbox bool) Credentials {
	if sandbox {
		return Credentials{BaseURL: DefaultSandboxBaseURL, Sandbox: true}
	}
	return Credentials{BaseURL: DefaultBaseURL}
}

// ParseBool returns true for "1", "true", "yes" (case-insensitive).
func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
