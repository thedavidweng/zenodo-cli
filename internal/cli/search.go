package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search [QUERY]",
	Short: "Search Zenodo records",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "search")

		client, _, err := getClient(app)
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
		fmt.Fprintf(cmd.ErrOrStderr(), "\nTotal: %d\n", resp.Hits.Total)
		return nil
	},
}

func init() {
	searchCmd.Flags().Int("page", 1, "page number")
	searchCmd.Flags().Int("size", 10, "results per page")
}
