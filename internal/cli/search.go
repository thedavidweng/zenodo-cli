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
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "search")

		client, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		query := args[0]
		resp, err := client.SearchRecords(cmd.Context(), query)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, resp.Hits, nil)
		}

		for _, rec := range resp.Hits.Hits {
			r.Human("[%s] %s\n", rec.ID, rec.Metadata.Title)
		}
		r.Human("\nTotal: %d\n", resp.Hits.Total)
		return nil
	},
}
