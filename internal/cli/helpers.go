package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/config"
	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

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

func metaInput(app *AppContext, command string) output.RuntimeMetaInput {
	return output.RuntimeMetaInput{
		Command:   command,
		Profile:   app.Profile,
		RequestID: app.RequestID,
		StartedAt: app.StartedAt,
	}
}

type CmdContext struct {
	App    *AppContext
	Cmd    *cobra.Command
	Args   []string
	Client zenodo.API
	R      output.Renderer
	Meta   output.RuntimeMetaInput
	Gate   *Gate
}

type CmdFunc func(ctx *CmdContext) error

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

func withPublicClient(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, command)

		creds, _ := config.ResolveClientConfig(app.ConfigFile, app.Profile, app.Sandbox, false)
		client := zenodo.NewClient(creds.BaseURL, creds.Token)
		client.Retries = app.Retries
		client.HTTPClient.Timeout = app.Timeout

		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, R: r, Meta: meta, Gate: newGate(app, &r, meta)})
	}
}

func getClient(app *AppContext) (*zenodo.Client, error) {
	creds, err := config.ResolveClientConfig(app.ConfigFile, app.Profile, app.Sandbox, true)
	if err != nil {
		return nil, err
	}
	client := zenodo.NewClient(creds.BaseURL, creds.Token)
	client.Retries = app.Retries
	client.HTTPClient.Timeout = app.Timeout
	return client, nil
}

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

func apiError(err error) model.ErrorBody {
	var apiErr *zenodo.APIError
	if errors.As(err, &apiErr) && err.Error() == apiErr.Error() {
		return output.Errorf(model.ErrZenodoAPI, "%s", apiErr.Message)
	}
	return output.Errorf(model.ErrZenodoAPI, "%v", err)
}

func resolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return config.DefaultConfigPath()
}
