package cli

import (
	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search [QUERY]",
	Short: "Search Zenodo records",
	Long: `Search publicly available Zenodo records using a full-text query.

This command does not require authentication.`,
	Example: `  zenodo search "machine learning"
  zenodo search "climate" --json`,
	Args: cobra.ExactArgs(1),
	RunE: withPublicClient("search", func(ctx *CmdContext) error {
		query := ctx.Args[0]
		resp, err := ctx.Client.SearchRecords(ctx.Cmd.Context(), query)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		return ctx.R.Render(ctx.Meta, resp.Hits, func() {
			for _, rec := range resp.Hits.Hits {
				ctx.R.Human("[%s] %s\n", rec.ID, rec.Metadata.Title)
			}
			ctx.R.Human("\nTotal: %d\n", resp.Hits.Total)
		})
	}),
}
