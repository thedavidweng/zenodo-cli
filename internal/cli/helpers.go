package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/config"
	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

// newRenderer creates a Renderer from AppContext and cobra.Command.
func newRenderer(app *AppContext, cmd *cobra.Command) output.Renderer {
	return output.Renderer{
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
		JSON:    app.JSON,
		Pretty:  app.Pretty,
		Compact: app.Compact,
		Full:    app.Full,
		Quiet:   app.Quiet,
		NoColor: app.NoColor,
		Verbose: app.Verbose,
	}
}

// metaInput builds a RuntimeMetaInput from AppContext.
func metaInput(app *AppContext, command string) output.RuntimeMetaInput {
	return output.RuntimeMetaInput{
		Command:   command,
		Profile:   app.Profile,
		RequestID: app.RequestID,
		StartedAt: app.StartedAt,
	}
}

// CmdContext bundles everything a command handler needs.
type CmdContext struct {
	App    *AppContext
	Cmd    *cobra.Command
	Args   []string
	Client *zenodo.Client
	Config *config.Config
	R      output.Renderer
	Meta   output.RuntimeMetaInput
}

// CmdFunc is a command handler that receives a ready-to-use context.
type CmdFunc func(ctx *CmdContext) error

// withAuth wraps a CmdFunc: loads config, creates client, checks auth.
func withAuth(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, command)

		client, cfg, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}
		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, Config: cfg, R: r, Meta: meta})
	}
}

// getClient creates a Zenodo client from the current app context and config.
func getClient(app *AppContext) (*zenodo.Client, *config.Config, error) {
	cfgPath := app.ConfigFile
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("not configured. Run 'zenodo auth login' to get started")
	}

	profile, err := cfg.GetProfile(app.Profile)
	if err != nil {
		return nil, cfg, fmt.Errorf("not authenticated. Run 'zenodo auth login' to get started")
	}

	creds := config.CredentialsFromProfileAndEnv(profile)
	// CLI --sandbox flag overrides profile setting
	if app.Sandbox {
		creds.Sandbox = true
		if creds.BaseURL == "https://zenodo.org" {
			creds.BaseURL = "https://sandbox.zenodo.org"
		}
	}
	client := zenodo.NewClient(creds.BaseURL, creds.Token)
	client.Retries = app.Retries
	client.HTTPClient.Timeout = app.Timeout

	// Apply endpoint overrides from config (used for testing)
	if profile.Endpoints.API != "" {
		client.BaseURL = profile.Endpoints.API
	}

	return client, cfg, nil
}

// requireAuth checks that the client is authenticated.
func requireAuth(r *output.Renderer, meta output.RuntimeMetaInput, client *zenodo.Client) error {
	if client.Token == "" {
		return r.Failure(meta, output.ErrorWithDetails(
			"AUTH_REQUIRED",
			"Authentication required. Run 'zenodo auth login' to authenticate.",
			map[string]any{"profile": meta.Profile},
		))
	}
	return nil
}

// requireConfirm checks that --confirm was passed for destructive operations.
func requireConfirm(r *output.Renderer, meta output.RuntimeMetaInput, app *AppContext) error {
	if !app.Confirm {
		return r.Failure(meta, output.Errorf("CONFIRMATION_REQUIRED", "use --confirm to proceed"))
	}
	return nil
}

// requireReadOnly blocks mutations when --read-only is set.
func requireReadOnly(r *output.Renderer, meta output.RuntimeMetaInput, app *AppContext) error {
	if app.ReadOnly {
		return r.Failure(meta, output.Errorf(model.ErrReadOnlyViolation, "--read-only blocks this mutation"))
	}
	return nil
}

// parseJSON parses a JSON string into the target value.
func parseJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}
