package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// --- helper to find a subcommand of authCmd ---
func authSubcmd(name string) *cobra.Command {
	for _, c := range authCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAuthLoginWithTokenFlagEndToEnd(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	out, err := runCmd(t, cfgPath, authSubcmd("login"), nil, nil, map[string]string{
		"token": "my-test-token",
	})
	if err != nil {
		t.Fatalf("auth login --token: %v", err)
	}
	if !strings.Contains(out, "Token saved") {
		t.Errorf("expected 'Token saved' in output: %s", out)
	}

	// Verify the config was actually written
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "my-test-token") {
		t.Errorf("config should contain token, got: %s", data)
	}
}

func TestAuthLoginFromEnv(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	t.Setenv("ZENODO_TOKEN", "env-token-123")

	out, err := runCmd(t, cfgPath, authSubcmd("login"), nil, nil, nil)
	if err != nil {
		t.Fatalf("auth login with env: %v", err)
	}
	if !strings.Contains(out, "Token saved") {
		t.Errorf("expected 'Token saved' in output: %s", out)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "env-token-123") {
		t.Errorf("config should contain env token, got: %s", data)
	}
}

func TestAuthLoginNoTokenNonTerminal(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// No --token, no ZENODO_TOKEN, and stdin is not a terminal (test runner)
	_, err := runCmd(t, cfgPath, authSubcmd("login"), nil, nil, nil)
	if err == nil {
		t.Error("expected error when no token provided and not a terminal")
	}
}

func TestAuthStatusAuthenticated(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: status-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	out, err := runCmd(t, cfgPath, authSubcmd("status"), nil, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if !strings.Contains(out, "authenticated") {
		t.Errorf("expected 'authenticated' in output: %s", out)
	}
}

func TestAuthStatusNotConfigured(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// No config file
	_, err := runCmd(t, cfgPath, authSubcmd("status"), nil, nil, nil)
	if err == nil {
		t.Error("expected error when not configured")
	}
}

func TestAuthStatusProfileNotFound(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: some-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	statusCmd := authSubcmd("status")
	resetFlags(statusCmd)

	var outBuf bytes.Buffer
	statusCmd.SetOut(&outBuf)
	statusCmd.SetErr(&outBuf)

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "nonexistent",
		Timeout:    30 * time.Second,
		Retries:    0,
		StartedAt:  time.Now(),
		RequestID:  "test-request",
		JSON:       true,
	}
	ctx := WithAppContext(context.Background(), app)
	statusCmd.SetContext(ctx)

	err := statusCmd.RunE(statusCmd, nil)
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
	output := outBuf.String()
	if !strings.Contains(output, "not configured") {
		t.Errorf("expected 'not configured' in output: %s", output)
	}
}

func TestAuthStatusNoToken(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: ""
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := runCmd(t, cfgPath, authSubcmd("status"), nil, nil, nil)
	if err == nil {
		t.Error("expected error when no token configured")
	}
}

func TestAuthStatusUnsetEnvIndirection(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: "env:ZENODO_TEST_MISSING_TOKEN"
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	origToken, tokenSet := os.LookupEnv("ZENODO_TOKEN")
	_ = os.Unsetenv("ZENODO_TOKEN")
	_ = os.Unsetenv("ZENODO_TEST_MISSING_TOKEN")
	t.Cleanup(func() {
		if tokenSet {
			_ = os.Setenv("ZENODO_TOKEN", origToken)
		}
	})

	out, err := runCmd(t, cfgPath, authSubcmd("status"), nil, nil, nil)
	if err == nil {
		t.Error("expected error when env indirection is unset")
	}
	if !strings.Contains(out, "ZENODO_TEST_MISSING_TOKEN") {
		t.Errorf("expected missing var in output: %s", out)
	}
}

func TestAuthLogoutCommand(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: logout-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	out, err := runCmd(t, cfgPath, authSubcmd("logout"), nil, nil, nil)
	if err != nil {
		t.Fatalf("auth logout: %v", err)
	}
	if !strings.Contains(out, "Logged out") {
		t.Errorf("expected 'Logged out' in output: %s", out)
	}

	// Verify token was cleared
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(data), "logout-token") {
		t.Error("config should not contain old token after logout")
	}
}

func TestAuthLogoutDryRun(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: dryrun-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	out, err := runCmd(t, cfgPath, authSubcmd("logout"), nil, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("auth logout --dry-run: %v", err)
	}
	if !strings.Contains(out, "Would clear") {
		t.Errorf("expected 'Would clear' in output: %s", out)
	}

	// Token should NOT be cleared in dry-run
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "dryrun-token") {
		t.Error("config should still contain token after dry-run logout")
	}
}

func TestAuthLogoutNoProfile(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: some-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := runCmd(t, cfgPath, authSubcmd("logout"), nil, nil, nil)
	_ = err
}

// --- readFrom ---

func TestReadFrom(t *testing.T) {
	input := "  hello world  \n"
	result := readFrom(strings.NewReader(input))
	if result != "hello world" {
		t.Fatalf("readFrom = %q, want %q", result, "hello world")
	}
}

func TestReadFromEmpty(t *testing.T) {
	result := readFrom(strings.NewReader(""))
	if result != "" {
		t.Fatalf("readFrom empty = %q, want %q", result, "")
	}
}
