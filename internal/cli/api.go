package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Direct Zenodo API access",
	Long: `Send raw requests to any Zenodo InvenioRDM API endpoint.

This is an escape hatch for operations not covered by higher-level commands.
Paths are relative to the API base (e.g. /api/records, /api/records/12345/draft).`,
}

var apiGetCmd = &cobra.Command{
	Use:   "get [PATH]",
	Short: "Send a GET request to the Zenodo API",
	Long:  "Send a GET request to the specified API path and return the JSON response.",
	Example: `  zenodo api get /api/records/12345
  zenodo api get /api/user/records --json`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("api.get", func(ctx *CmdContext) error {
		path := ensurePath(ctx.Args[0])

		var result any
		err := ctx.Client.Do(ctx.Cmd.Context(), "GET", path, nil, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		return printJSON(ctx, result)
	}),
}

var apiPostCmd = &cobra.Command{
	Use:   "post [PATH]",
	Short: "Send a POST request to the Zenodo API",
	Long: `Send a POST request with a JSON body to the specified API path.

Use --data to provide the JSON request body. Without --data, sends an empty body.`,
	Example: `  zenodo api post /api/records --data '{"metadata":{"title":"Test"}}'
  zenodo api post /api/records/12345/draft/actions/publish --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("api.post", func(ctx *CmdContext) error {
		path := ensurePath(ctx.Args[0])
		data, _ := ctx.Cmd.Flags().GetString("data")

		proceed, err := ctx.Gate.Allow(RiskHighWrite, Plan{
			HumanMsg:  "Would POST to %s\n",
			HumanArgs: []any{path},
			Data:      map[string]any{"method": "POST", "path": path},
		})
		if err != nil {
			return err
		}
		if !proceed {
			if data != "" {
				ctx.R.Human("  body: %s\n", data)
			}
			return nil
		}

		var body any
		if data != "" {
			if err := parseJSON(data, &body); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "invalid JSON data: %v", err))
			}
		}

		var result any
		err = ctx.Client.Do(ctx.Cmd.Context(), "POST", path, body, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		return printJSON(ctx, result)
	}),
}

var apiPutCmd = &cobra.Command{
	Use:   "put [PATH]",
	Short: "Send a PUT request to the Zenodo API",
	Long: `Send a PUT request with a JSON body to the specified API path.

Use --data to provide the JSON request body. Without --data, sends an empty body.`,
	Example: `  zenodo api put /api/records/12345/draft --data '{"metadata":{"title":"Updated"}}'
  zenodo api put /api/records/12345/draft --data @meta.json --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("api.put", func(ctx *CmdContext) error {
		path := ensurePath(ctx.Args[0])
		data, _ := ctx.Cmd.Flags().GetString("data")

		proceed, err := ctx.Gate.Allow(RiskHighWrite, Plan{
			HumanMsg:  "Would PUT to %s\n",
			HumanArgs: []any{path},
			Data:      map[string]any{"method": "PUT", "path": path},
		})
		if err != nil {
			return err
		}
		if !proceed {
			if data != "" {
				ctx.R.Human("  body: %s\n", data)
			}
			return nil
		}

		var body any
		if data != "" {
			if err := parseJSON(data, &body); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "invalid JSON data: %v", err))
			}
		}

		var result any
		err = ctx.Client.Do(ctx.Cmd.Context(), "PUT", path, body, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		return printJSON(ctx, result)
	}),
}

// ensurePath ensures the given path starts with "/".
func ensurePath(path string) string {
	if path == "" || path[0] != '/' {
		return "/" + path
	}
	return path
}

// printJSON writes a value as indented JSON to the command's stdout.
func printJSON(ctx *CmdContext, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = ctx.Cmd.OutOrStdout().Write(append(b, '\n'))
	return err
}

func init() {
	apiPostCmd.Flags().String("data", "", "JSON data to send in request body")
	apiPutCmd.Flags().String("data", "", "JSON data to send in request body")

	apiCmd.AddCommand(apiGetCmd)
	apiCmd.AddCommand(apiPostCmd)
	apiCmd.AddCommand(apiPutCmd)
}
