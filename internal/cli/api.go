package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
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
		path := zenodo.EnsureLeadingSlash(ctx.Args[0])

		var result any
		err := ctx.Client.Do(ctx.Cmd.Context(), "GET", path, nil, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, apiError(err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		return printJSON(ctx, result)
	}),
}

var apiPostCmd = newApiWriteCmd("POST")
var apiPutCmd = newApiWriteCmd("PUT")

func newApiWriteCmd(method string) *cobra.Command {
	var example string
	switch method {
	case "POST":
		example = `  zenodo api post /api/records --data '{"metadata":{"title":"Test"}}'
  zenodo api post /api/records/12345/draft/actions/publish --confirm`
	case "PUT":
		example = `  zenodo api put /api/records/12345/draft --data '{"metadata":{"title":"Updated"}}' --confirm`
	}

	return &cobra.Command{
		Use:   strings.ToLower(method),
		Short: fmt.Sprintf("Send a %s request to the Zenodo API", method),
		Long: fmt.Sprintf(`Send a %s request with a JSON body to the specified API path.

Use --data to provide the JSON request body. Without --data, sends an empty body.`, method),
		Example: example,
		Args:    cobra.ExactArgs(1),
		RunE: withAuth("api."+strings.ToLower(method), func(ctx *CmdContext) error {
			path := zenodo.EnsureLeadingSlash(ctx.Args[0])
			data, _ := ctx.Cmd.Flags().GetString("data")

			proceed, err := ctx.Gate.Allow(RiskHighWrite, Plan{
				HumanMsg:  fmt.Sprintf("Would %s to %%s\n", method),
				HumanArgs: []any{path},
				Data:      map[string]any{"method": method, "path": path},
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
				if err := json.Unmarshal([]byte(data), &body); err != nil {
					return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "invalid JSON data: %v", err))
				}
			}

			var result any
			err = ctx.Client.Do(ctx.Cmd.Context(), method, path, body, &result)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, apiError(err))
			}

			if ctx.App.JSON {
				return ctx.R.Success(ctx.Meta, result, nil)
			}
			return printJSON(ctx, result)
		}),
	}
}

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
