package config

import (
	"os"
	"strings"
)

const (
	defaultBaseURL        = "https://zenodo.org"
	defaultSandboxBaseURL = "https://sandbox.zenodo.org"
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
			c.BaseURL = defaultSandboxBaseURL
		} else {
			c.BaseURL = defaultBaseURL
		}
	}

	return c
}

// IsAuthenticated returns true if a token is present.
func (c Credentials) IsAuthenticated() bool {
	return c.Token != ""
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
