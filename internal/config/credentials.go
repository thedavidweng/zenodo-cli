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

// Credentials is the resolved authentication + endpoint info for a request.
type Credentials struct {
	Token   string
	Sandbox bool
	BaseURL string
}

// CredentialsFromProfileAndEnv merges a profile with environment variable
// overrides. Env vars take precedence: ZENODO_TOKEN, ZENODO_SANDBOX, ZENODO_API_URL.
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
		c.Sandbox = parseBool(v)
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

// IsAuthenticated returns true if a token is present.
func (c Credentials) IsAuthenticated() bool {
	return c.Token != ""
}

// ClientConfig holds the resolved settings needed to construct a zenodo.Client.
// It is the result of ResolveClientConfig, which owns the base-URL/sandbox/
// endpoint selection so callers don't need to reference config defaults.
type ClientConfig struct {
	BaseURL string
	Token   string
	Sandbox bool
}

// ResolveClientConfig loads config and resolves the base URL, token, and
// sandbox setting for the given profile, applying CLI flag overrides.
//
// When requireProfile is true, returns an error if the config file or
// profile is missing. When false, falls back to the default (or sandbox)
// base URL with an empty token — used by public commands like search.
//
// Override precedence (highest first):
//   1. profile.Endpoints.API (test/self-hosted override)
//   2. CLI --sandbox flag (swaps default base URL to sandbox)
//   3. profile base_url / env ZENODO_API_URL
//   4. config defaults
func ResolveClientConfig(configFile, profileName string, sandbox, requireProfile bool) (ClientConfig, error) {
	cfgPath := configFile
	if cfgPath == "" {
		cfgPath = DefaultConfigPath()
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		if requireProfile {
			return ClientConfig{}, fmt.Errorf("not configured. Run 'zenodo auth login' to get started")
		}
		return fallbackClientConfig(sandbox), nil
	}

	profile := cfg.GetProfileOrNil(profileName)
	if profile == nil {
		if requireProfile {
			return ClientConfig{}, fmt.Errorf("not authenticated. Run 'zenodo auth login' to get started")
		}
		return fallbackClientConfig(sandbox), nil
	}

	creds := CredentialsFromProfileAndEnv(profile)

	// Apply CLI sandbox override: only swaps when base URL is the default.
	if sandbox {
		creds.Sandbox = true
		if creds.BaseURL == DefaultBaseURL {
			creds.BaseURL = DefaultSandboxBaseURL
		}
	}

	// Apply endpoint override (used for testing/self-hosted). Takes precedence
	// over sandbox swap so test servers always win.
	if profile.Endpoints.API != "" {
		creds.BaseURL = profile.Endpoints.API
	}

	return ClientConfig{BaseURL: creds.BaseURL, Token: creds.Token, Sandbox: creds.Sandbox}, nil
}

func fallbackClientConfig(sandbox bool) ClientConfig {
	if sandbox {
		return ClientConfig{BaseURL: DefaultSandboxBaseURL, Sandbox: true}
	}
	return ClientConfig{BaseURL: DefaultBaseURL}
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
