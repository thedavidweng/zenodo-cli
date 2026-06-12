package cli

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Manage Zenodo records",
	Long:  "Create, list, view, publish, and manage Zenodo deposit records.",
}

var recordsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List your records",
	Long:    "List all records (draft and published) owned by the authenticated user.",
	Example: "  zenodo records list\n  zenodo records list --json",
	RunE: withAuth("records.list", func(ctx *CmdContext) error {
		resp, err := ctx.Client.ListRecords(ctx.Cmd.Context())
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, resp.Hits, nil)
		}
		for _, rec := range resp.Hits.Hits {
			ctx.R.Human("[%s] %s (%s)\n", rec.ID, rec.Metadata.Title, rec.Status)
		}
		ctx.R.Human("\nTotal: %d\n", resp.Hits.Total)
		return nil
	}),
}

var recordsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new draft record",
	Long: `Create a new draft record on Zenodo.

Use --title and --description for a quick record, or --metadata for full control
(--metadata overrides --title and --description). The new record is a draft;
use "records publish" to make it public.`,
	Example: `  zenodo records create --title "My Dataset" --description "Research data"
  zenodo records create --metadata meta.json
  zenodo records create --title "Test" --dry-run`,
	RunE: withAuth("records.create", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		if ctx.App.DryRun {
			title, _ := ctx.Cmd.Flags().GetString("title")
			metadataFile, _ := ctx.Cmd.Flags().GetString("metadata")
			ctx.R.Human("Would create draft record (title=%q, metadata=%s)\n", title, metadataFile)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "action": "create_record"}, nil)
		}

		title, _ := ctx.Cmd.Flags().GetString("title")
		description, _ := ctx.Cmd.Flags().GetString("description")
		metadataFile, _ := ctx.Cmd.Flags().GetString("metadata")

		var meta zenodo.RecordMetadata
		if metadataFile != "" {
			data, err := os.ReadFile(metadataFile)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFilesystem, "reading metadata file: %v", err))
			}
			if err := json.Unmarshal(data, &meta); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "parsing metadata JSON: %v", err))
			}
		} else {
			if title == "" {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "title is required (use --title or --metadata)"))
			}
			meta = zenodo.RecordMetadata{
				Title:           title,
				Description:     description,
				PublicationDate: time.Now().Format("2006-01-02"),
				ResourceType:    zenodo.ResourceType{Type: "dataset"},
			}
		}

		rec, err := ctx.Client.CreateRecord(ctx.Cmd.Context(), meta)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, rec, nil)
		}
		ctx.R.Human("Created draft %s: %s\n", rec.ID, rec.Metadata.Title)
		return nil
	}),
}

var recordsShowCmd = &cobra.Command{
	Use:   "show [ID]",
	Short: "Show record details",
	Long: `Show metadata for a record by its ID.

Tries to fetch the draft first; if none exists, falls back to the published record.`,
	Example: `  zenodo records show 12345
  zenodo records show 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("records.show", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		// Try draft first, fall back to published
		rec, err := ctx.Client.GetDraft(ctx.Cmd.Context(), id)
		if err != nil {
			rec, err = ctx.Client.GetRecord(ctx.Cmd.Context(), id)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
			}
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, rec, nil)
		}
		ctx.R.Human("ID:          %s\n", rec.ID)
		ctx.R.Human("Title:       %s\n", rec.Metadata.Title)
		ctx.R.Human("Status:      %s\n", rec.Status)
		ctx.R.Human("Created:     %s\n", rec.CreatedAt)
		ctx.R.Human("Updated:     %s\n", rec.UpdatedAt)
		ctx.R.Human("Description: %s\n", rec.Metadata.Description)
		return nil
	}),
}

var recordsDeleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Delete a draft record",
	Long: `Delete a draft record. Published records cannot be deleted.

Requires --confirm because this operation is irreversible.`,
	Example: "  zenodo records delete 12345 --confirm",
	Args:    cobra.ExactArgs(1),
	RunE: withAuth("records.delete", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		if err := requireConfirm(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		if ctx.App.DryRun {
			ctx.R.Human("Would delete draft %s\n", id)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "id": id, "action": "delete_draft"}, nil)
		}
		if err := ctx.Client.DeleteDraft(ctx.Cmd.Context(), id); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{"deleted": id}, nil)
		}
		ctx.R.Human("Deleted draft %s\n", id)
		return nil
	}),
}

var recordsPublishCmd = &cobra.Command{
	Use:   "publish [ID]",
	Short: "Publish a draft record",
	Long: `Publish a draft record, making it publicly accessible with a DOI.

This is irreversible. Once published, a record cannot be unpublished or deleted.`,
	Example: "  zenodo records publish 12345 --confirm",
	Args:    cobra.ExactArgs(1),
	RunE: withAuth("records.publish", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		if err := requireConfirm(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		if ctx.App.DryRun {
			ctx.R.Human("Would publish draft %s (irreversible)\n", id)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "id": id, "action": "publish_draft"}, nil)
		}
		rec, err := ctx.Client.PublishDraft(ctx.Cmd.Context(), id)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, rec, nil)
		}
		ctx.R.Human("Published %s: %s\n", rec.ID, rec.Metadata.Title)
		return nil
	}),
}

var recordsNewVersionCmd = &cobra.Command{
	Use:   "new-version [ID]",
	Short: "Create a new draft version of a record",
	Long: `Create a new editable draft from an existing published record.

The new draft inherits metadata and files from the original. You can then
modify it and publish it as a new version.`,
	Example: "  zenodo records new-version 12345",
	Args:    cobra.ExactArgs(1),
	RunE: withAuth("records.new-version", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		if ctx.App.DryRun {
			ctx.R.Human("Would create new version draft from %s\n", id)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "id": id, "action": "new_version"}, nil)
		}
		rec, err := ctx.Client.NewVersion(ctx.Cmd.Context(), id)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, rec, nil)
		}
		ctx.R.Human("Created new version %s (from %s)\n", rec.ID, id)
		return nil
	}),
}

func init() {
	recordsCreateCmd.Flags().String("title", "", "record title")
	recordsCreateCmd.Flags().String("description", "", "record description")
	recordsCreateCmd.Flags().String("metadata", "", "path to JSON metadata file")

	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsCreateCmd)
	recordsCmd.AddCommand(recordsShowCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)
	recordsCmd.AddCommand(recordsPublishCmd)
	recordsCmd.AddCommand(recordsNewVersionCmd)
}
