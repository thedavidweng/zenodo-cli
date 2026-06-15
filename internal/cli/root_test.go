package cli

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestRootCommandExists(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Use != "zenodo" {
		t.Errorf("Use = %q, want %q", cmd.Use, "zenodo")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestRootCommandHasGlobalFlags(t *testing.T) {
	cmd := newRootCmd()

	flags := []string{
		"config", "profile", "sandbox", "json", "pretty", "compact", "full",
		"quiet", "read-only", "dry-run", "confirm",
		"timeout", "retries",
	}

	for _, name := range flags {
		f := cmd.PersistentFlags().Lookup(name)
		if f == nil {
			t.Errorf("missing persistent flag: --%s", name)
		}
	}
}

func TestRootCommandSilenceFlags(t *testing.T) {
	cmd := newRootCmd()
	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage=true")
	}
	if !cmd.SilenceErrors {
		t.Error("expected SilenceErrors=true")
	}
}

func TestRootCommandPersistentPreRunEReadsFlags(t *testing.T) {
	cmd := newRootCmd()

	// Set some flags
	cmd.SetArgs([]string{
		"--config", "/tmp/test.yaml",
		"--profile", "sb",
		"--sandbox",
		"--json",
		"--timeout", "10s",
		"--retries", "5",
	})

	// Add a dummy subcommand to capture context
	var capturedApp *AppContext
	dummy := &cobra.Command{
		Use: "dummy",
		RunE: func(c *cobra.Command, args []string) error {
			capturedApp = GetAppContext(c.Context())
			return nil
		},
	}
	cmd.AddCommand(dummy)
	cmd.SetArgs([]string{
		"--config", "/tmp/test.yaml",
		"--profile", "sb",
		"--sandbox",
		"--json",
		"--timeout", "10s",
		"--retries", "5",
		"dummy",
	})

	_ = cmd.Execute()

	if capturedApp == nil {
		t.Fatal("expected AppContext to be set")
	}
	if capturedApp.ConfigFile != "/tmp/test.yaml" {
		t.Errorf("ConfigFile = %q, want /tmp/test.yaml", capturedApp.ConfigFile)
	}
	if capturedApp.Profile != "sb" {
		t.Errorf("Profile = %q, want sb", capturedApp.Profile)
	}
	if !capturedApp.Sandbox {
		t.Error("expected Sandbox=true")
	}
	if !capturedApp.JSON {
		t.Error("expected JSON=true")
	}
	if capturedApp.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", capturedApp.Timeout)
	}
	if capturedApp.Retries != 5 {
		t.Errorf("Retries = %d, want 5", capturedApp.Retries)
	}
}

func TestRootCommandEnvOverrides(t *testing.T) {
	cmd := newRootCmd()

	t.Setenv("ZENODO_TOKEN", "env-tok")
	t.Setenv("ZENODO_SANDBOX", "true")
	t.Setenv("ZENODO_TIMEOUT", "45s")
	t.Setenv("ZENODO_RETRIES", "7")

	var capturedApp *AppContext
	dummy := &cobra.Command{
		Use: "dummy",
		RunE: func(c *cobra.Command, args []string) error {
			capturedApp = GetAppContext(c.Context())
			return nil
		},
	}
	cmd.AddCommand(dummy)
	cmd.SetArgs([]string{"dummy"})

	_ = cmd.Execute()

	if capturedApp == nil {
		t.Fatal("expected AppContext to be set")
	}
	if !capturedApp.Sandbox {
		t.Error("expected Sandbox=true from ZENODO_SANDBOX")
	}
	if capturedApp.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", capturedApp.Timeout)
	}
	if capturedApp.Retries != 7 {
		t.Errorf("Retries = %d, want 7", capturedApp.Retries)
	}
}

func TestRegisterSubcommands(t *testing.T) {
	cmd := newRootCmd()
	registerSubcommands(cmd)

	expectedNames := []string{
		"version", "auth", "records", "files", "search",
		"doctor", "completion", "api",
	}

	names := make(map[string]bool)
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}

	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestValidateAppContextNegativeRetries(t *testing.T) {
	app := &AppContext{Retries: -1, Timeout: 5 * time.Second}
	err := validateAppContext(app)
	if err == nil {
		t.Error("expected error for negative retries")
	}
}

func TestValidateAppContextZeroTimeout(t *testing.T) {
	app := &AppContext{Retries: 0, Timeout: 0}
	err := validateAppContext(app)
	if err == nil {
		t.Error("expected error for zero timeout")
	}
}

func TestValidateAppContextFullOverridesCompact(t *testing.T) {
	app := &AppContext{Retries: 0, Timeout: 5 * time.Second, Full: true, Compact: true}
	err := validateAppContext(app)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Compact {
		t.Error("expected Compact=false when Full=true")
	}
}

func TestSilenceAllCommands(t *testing.T) {
	parent := &cobra.Command{Use: "parent"}
	child := &cobra.Command{Use: "child"}
	grandchild := &cobra.Command{Use: "grandchild"}
	child.AddCommand(grandchild)
	parent.AddCommand(child)

	silenceAllCommands(parent)

	if !parent.SilenceUsage {
		t.Error("parent SilenceUsage")
	}
	if !child.SilenceUsage {
		t.Error("child SilenceUsage")
	}
	if !grandchild.SilenceUsage {
		t.Error("grandchild SilenceUsage")
	}
	if !parent.SilenceErrors {
		t.Error("parent SilenceErrors")
	}
	if !child.SilenceErrors {
		t.Error("child SilenceErrors")
	}
	if !grandchild.SilenceErrors {
		t.Error("grandchild SilenceErrors")
	}
}

func TestEnvOrWithSetValue(t *testing.T) {
	t.Setenv("TEST_ENV_OR_KEY", "from-env")
	v := envOr("TEST_ENV_OR_KEY", "fallback")
	if v != "from-env" {
		t.Fatalf("envOr = %q, want %q", v, "from-env")
	}
}

func TestEnvOrWithFallback(t *testing.T) {
	t.Setenv("TEST_ENV_OR_MISSING", "")
	v := envOr("TEST_ENV_OR_MISSING", "fallback")
	if v != "fallback" {
		t.Fatalf("envOr = %q, want %q", v, "fallback")
	}
}

func TestEnvDurationInvalidValue(t *testing.T) {
	t.Setenv("TEST_ENV_DURATION_BAD", "not-a-duration")
	v := envDuration("TEST_ENV_DURATION_BAD", 5*time.Second, 10*time.Second)
	// Should fall back to current value since the env is not a valid duration
	if v != 5*time.Second {
		t.Fatalf("envDuration = %v, want 5s", v)
	}
}

func TestEnvIntInvalidValue(t *testing.T) {
	t.Setenv("TEST_ENV_INT_BAD", "not-an-int")
	v := envInt("TEST_ENV_INT_BAD", 3, 5)
	// Should fall back to current value since the env is not a valid int
	if v != 3 {
		t.Fatalf("envInt = %d, want 3", v)
	}
}

func TestEnvOrDefaultCurrentNotDefault(t *testing.T) {
	// When current != default, don't look at env
	t.Setenv("TEST_ENV_OR_DEFAULT", "env-val")
	v := envOrDefault("custom", "default", "TEST_ENV_OR_DEFAULT")
	if v != "custom" {
		t.Fatalf("envOrDefault = %q, want %q", v, "custom")
	}
}
