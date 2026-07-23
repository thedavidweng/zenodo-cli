package cli

import (
	"encoding/json"
	"errors"

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
	R      output.Renderer
	Meta   output.RuntimeMetaInput
	Gate   *Gate
}

// CmdFunc is a command handler that receives a ready-to-use context.
type CmdFunc func(ctx *CmdContext) error

// withAuth wraps a CmdFunc: loads config, creates client, checks auth.
func withAuth(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, command)

		client, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}
		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, R: r, Meta: meta, Gate: newGate(app, &r, meta)})
	}
}

// withClient wraps a CmdFunc: loads config and creates client without requiring auth.
func withClient(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, command)

		client, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, R: r, Meta: meta, Gate: newGate(app, &r, meta)})
	}
}

// withPublicClient wraps a CmdFunc for commands that don't require any config
// (e.g. search). It creates a client using the default base URL (or sandbox URL
// if --sandbox is set) with no token. If a config/profile does exist, it uses
// those settings for the base URL.
func withPublicClient(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, command)

		cc, _ := config.ResolveClientConfig(app.ConfigFile, app.Profile, app.Sandbox, false)
		client := zenodo.NewClient(cc.BaseURL, cc.Token)
		client.Retries = app.Retries
		client.HTTPClient.Timeout = app.Timeout

		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, R: r, Meta: meta, Gate: newGate(app, &r, meta)})
	}
}

// getClient creates a Zenodo client from the current app context and config.
func getClient(app *AppContext) (*zenodo.Client, error) {
	cc, err := config.ResolveClientConfig(app.ConfigFile, app.Profile, app.Sandbox, true)
	if err != nil {
		return nil, err
	}
	client := zenodo.NewClient(cc.BaseURL, cc.Token)
	client.Retries = app.Retries
	client.HTTPClient.Timeout = app.Timeout
	return client, nil
}

// requireAuth checks that the client is authenticated.
func requireAuth(r *output.Renderer, meta output.RuntimeMetaInput, client *zenodo.Client) error {
	if client.Token == "" {
		return r.Failure(meta, output.ErrorWithDetails(
			model.ErrAuthRequired,
			"Authentication required. Run 'zenodo auth login' to authenticate.",
			map[string]any{"profile": meta.Profile},
		))
	}
	return nil
}

// parseJSON parses a JSON string into the target value.
func parseJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

// apiError translates a zenodo client error into an ErrorBody with the
// correct error code. For unwrapped APIError instances, it uses the clean
// message without the "API error (HTTP xxx):" prefix. For wrapped errors
// (e.g. "uploading data.csv: ..."), it preserves the full context.
// This centralizes the error-code selection that was previously repeated
// at 20+ call sites.
func apiError(err error) model.ErrorBody {
	var apiErr *zenodo.APIError
	if errors.As(err, &apiErr) && err.Error() == apiErr.Error() {
		return output.Errorf(model.ErrZenodoAPI, "%s", apiErr.Message)
	}
	return output.Errorf(model.ErrZenodoAPI, "%v", err)
}

// resolveConfigPath returns the config file path to use: the explicit override
// if set, otherwise the platform default.
func resolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return config.DefaultConfigPath()
}
