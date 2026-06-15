package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	pflag "github.com/spf13/pflag"

	"github.com/thedavidweng/zenodo-cli/internal/testutil"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

const testToken = "test-token-abc123"

// setupFakeZenodoTest starts a FakeZenodo server, writes a config file pointing
// to it, and returns the FakeZenodo instance and config path.
func setupFakeZenodoTest(t *testing.T) (*testutil.FakeZenodo, string) {
	t.Helper()
	fz := testutil.NewFakeZenodo(testToken)
	t.Cleanup(func() { fz.Close() })

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`current_profile: test
profiles:
  test:
    token: %s
    base_url: https://zenodo.org
    endpoints:
      api: %s
`, testToken, fz.URL())
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return fz, cfgPath
}

// newTestClient creates a zenodo.Client pointing at the FakeZenodo server.
func newTestClient(fz *testutil.FakeZenodo) *zenodo.Client {
	return zenodo.NewClient(fz.URL(), testToken)
}

// resetFlags resets all flags on a command to their default/zero values.
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
	})
}

// runCmd invokes a cobra command's RunE with a fresh AppContext, capturing
// output. It resets local flags on the command to avoid state leaking between
// tests. Use flagVals to set command-local flags (e.g. "title": "My Title").
func runCmd(t *testing.T, cfgPath string, cmd *cobra.Command, args []string, appFlags map[string]bool, flagVals map[string]string) (string, error) {
	t.Helper()

	// Reset local flags to defaults to avoid state leaking between tests.
	resetFlags(cmd)

	// Set command-local flags.
	for k, v := range flagVals {
		_ = cmd.Flags().Set(k, v)
	}

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
		Timeout:    30 * time.Second,
		Retries:    0,
		StartedAt:  time.Now(),
		RequestID:  "test-request",
	}
	for k, v := range appFlags {
		switch k {
		case "json":
			app.JSON = v
		case "dry-run":
			app.DryRun = v
		case "confirm":
			app.Confirm = v
		case "read-only":
			app.ReadOnly = v
		case "quiet":
			app.Quiet = v
		}
	}

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	ctx := WithAppContext(context.Background(), app)
	cmd.SetContext(ctx)

	err := cmd.RunE(cmd, args)
	return out.String(), err
}
