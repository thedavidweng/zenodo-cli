package config

import (
	"os"
	"testing"
)

func mustSetenv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set %s: %v", key, err)
	}
}

func mustUnsetenv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

func TestCredentialsFromProfile(t *testing.T) {
	p := &Profile{
		Token:   "profile-token",
		Sandbox: true,
		BaseURL: "https://sandbox.zenodo.org/api",
	}

	c := CredentialsFromProfileAndEnv(p)
	if c.Token != "profile-token" {
		t.Errorf("token = %q, want profile-token", c.Token)
	}
	if !c.Sandbox {
		t.Error("expected sandbox=true")
	}
	if c.BaseURL != "https://sandbox.zenodo.org/api" {
		t.Errorf("base_url = %q, want https://sandbox.zenodo.org/api", c.BaseURL)
	}
}

func TestCredentialsEnvOverridesProfile(t *testing.T) {
	origToken := os.Getenv("ZENODO_TOKEN")
	origSandbox := os.Getenv("ZENODO_SANDBOX")
	origURL := os.Getenv("ZENODO_API_URL")
	defer func() {
		_ = os.Setenv("ZENODO_TOKEN", origToken)
		_ = os.Setenv("ZENODO_SANDBOX", origSandbox)
		_ = os.Setenv("ZENODO_API_URL", origURL)
	}()

	mustSetenv(t, "ZENODO_TOKEN", "env-token")
	mustSetenv(t, "ZENODO_SANDBOX", "true")
	mustSetenv(t, "ZENODO_API_URL", "https://env.example.com/api")

	p := &Profile{
		Token:   "profile-token",
		Sandbox: false,
		BaseURL: "https://zenodo.org/api",
	}

	c := CredentialsFromProfileAndEnv(p)
	if c.Token != "env-token" {
		t.Errorf("token = %q, want env-token", c.Token)
	}
	if !c.Sandbox {
		t.Error("expected sandbox=true from env")
	}
	if c.BaseURL != "https://env.example.com/api" {
		t.Errorf("base_url = %q, want https://env.example.com/api", c.BaseURL)
	}
}

func TestCredentialsEnvOnly(t *testing.T) {
	origToken := os.Getenv("ZENODO_TOKEN")
	origSandbox := os.Getenv("ZENODO_SANDBOX")
	origURL := os.Getenv("ZENODO_API_URL")
	defer func() {
		_ = os.Setenv("ZENODO_TOKEN", origToken)
		_ = os.Setenv("ZENODO_SANDBOX", origSandbox)
		_ = os.Setenv("ZENODO_API_URL", origURL)
	}()

	mustSetenv(t, "ZENODO_TOKEN", "env-token")
	mustUnsetenv(t, "ZENODO_SANDBOX")
	mustUnsetenv(t, "ZENODO_API_URL")

	p := &Profile{}

	c := CredentialsFromProfileAndEnv(p)
	if c.Token != "env-token" {
		t.Errorf("token = %q, want env-token", c.Token)
	}
	if c.Sandbox {
		t.Error("expected sandbox=false (default)")
	}
	if c.BaseURL != "https://zenodo.org" {
		t.Errorf("base_url = %q, want https://zenodo.org", c.BaseURL)
	}
}

func TestCredentialsDefaults(t *testing.T) {
	origToken := os.Getenv("ZENODO_TOKEN")
	origSandbox := os.Getenv("ZENODO_SANDBOX")
	origURL := os.Getenv("ZENODO_API_URL")
	defer func() {
		_ = os.Setenv("ZENODO_TOKEN", origToken)
		_ = os.Setenv("ZENODO_SANDBOX", origSandbox)
		_ = os.Setenv("ZENODO_API_URL", origURL)
	}()

	mustUnsetenv(t, "ZENODO_TOKEN")
	mustUnsetenv(t, "ZENODO_SANDBOX")
	mustUnsetenv(t, "ZENODO_API_URL")

	p := &Profile{}
	c := CredentialsFromProfileAndEnv(p)
	if c.Sandbox {
		t.Error("expected sandbox=false (default)")
	}
	if c.BaseURL != "https://zenodo.org" {
		t.Errorf("base_url = %q, want https://zenodo.org", c.BaseURL)
	}
}

func TestCredentialsSandboxDefault(t *testing.T) {
	origSandbox := os.Getenv("ZENODO_SANDBOX")
	origURL := os.Getenv("ZENODO_API_URL")
	defer func() {
		_ = os.Setenv("ZENODO_SANDBOX", origSandbox)
		_ = os.Setenv("ZENODO_API_URL", origURL)
	}()

	mustUnsetenv(t, "ZENODO_SANDBOX")
	mustUnsetenv(t, "ZENODO_API_URL")

	p := &Profile{Sandbox: true}
	c := CredentialsFromProfileAndEnv(p)
	if !c.Sandbox {
		t.Error("expected sandbox=true")
	}
	if c.BaseURL != "https://sandbox.zenodo.org" {
		t.Errorf("base_url = %q, want https://sandbox.zenodo.org", c.BaseURL)
	}
}

func TestIsAuthenticated(t *testing.T) {
	c := Credentials{Token: "abc"}
	if !c.IsAuthenticated() {
		t.Error("expected authenticated with token")
	}

	c = Credentials{Token: ""}
	if c.IsAuthenticated() {
		t.Error("expected not authenticated without token")
	}
}

func TestCredentialsSandboxEnvBoolParsing(t *testing.T) {
	origSandbox := os.Getenv("ZENODO_SANDBOX")
	defer func() { _ = os.Setenv("ZENODO_SANDBOX", origSandbox) }()

	tests := []struct {
		envVal string
		want   bool
	}{
		{"true", true},
		{"1", true},
		{"TRUE", true},
		{"false", false},
		{"0", false},
		{"", false},
	}
	for _, tt := range tests {
		mustSetenv(t, "ZENODO_SANDBOX", tt.envVal)
		c := CredentialsFromProfileAndEnv(&Profile{})
		if c.Sandbox != tt.want {
			t.Errorf("ZENODO_SANDBOX=%q: got sandbox=%v, want %v", tt.envVal, c.Sandbox, tt.want)
		}
	}
}
