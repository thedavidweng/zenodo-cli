package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Direct Zenodo API access",
}

var apiGetCmd = &cobra.Command{
	Use:   "get [PATH]",
	Short: "Send a GET request to the Zenodo API",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("api.get", func(ctx *CmdContext) error {
		path := ctx.Args[0]
		if path[0] != '/' {
			path = "/" + path
		}

		var result any
		err := ctx.Client.Do(ctx.Cmd.Context(), "GET", path, nil, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		fmt.Fprintf(ctx.Cmd.OutOrStdout(), "%v\n", result)
		return nil
	}),
}

var apiPostCmd = &cobra.Command{
	Use:   "post [PATH]",
	Short: "Send a POST request to the Zenodo API",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("api.post", func(ctx *CmdContext) error {
		path := ctx.Args[0]
		if path[0] != '/' {
			path = "/" + path
		}

		data, _ := ctx.Cmd.Flags().GetString("data")

		var body any
		if data != "" {
			if err := parseJSON(data, &body); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "invalid JSON data: %v", err))
			}
		}

		var result any
		err := ctx.Client.Do(ctx.Cmd.Context(), "POST", path, body, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		fmt.Fprintf(ctx.Cmd.OutOrStdout(), "%v\n", result)
		return nil
	}),
}

var apiPutCmd = &cobra.Command{
	Use:   "put [PATH]",
	Short: "Send a PUT request to the Zenodo API",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("api.put", func(ctx *CmdContext) error {
		path := ctx.Args[0]
		if path[0] != '/' {
			path = "/" + path
		}

		data, _ := ctx.Cmd.Flags().GetString("data")

		var body any
		if data != "" {
			if err := parseJSON(data, &body); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "invalid JSON data: %v", err))
			}
		}

		var result any
		err := ctx.Client.Do(ctx.Cmd.Context(), "PUT", path, body, &result)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, result, nil)
		}
		fmt.Fprintf(ctx.Cmd.OutOrStdout(), "%v\n", result)
		return nil
	}),
}

func init() {
	apiPostCmd.Flags().String("data", "", "JSON data to send in request body")
	apiPutCmd.Flags().String("data", "", "JSON data to send in request body")

	apiCmd.AddCommand(apiGetCmd)
	apiCmd.AddCommand(apiPostCmd)
	apiCmd.AddCommand(apiPutCmd)
}
