package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files in Zenodo records",
	Long:  "Upload, list, and download files attached to Zenodo records.",
}

var filesUploadCmd = &cobra.Command{
	Use:   "upload [ID] [FILE...]",
	Short: "Upload files to a draft record",
	Long: `Upload one or more local files to a draft record.

Files are uploaded via the three-step InvenioRDM process: init, content upload, commit.
The record must be a draft (not published).`,
	Example: `  zenodo files upload 12345 data.csv
  zenodo files upload 12345 data.csv results.json
  zenodo files upload 12345 *.csv --dry-run`,
	Args: cobra.MinimumNArgs(2),
	RunE: withAuth("files.upload", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		files := ctx.Args[1:]

		if ctx.App.DryRun {
			for _, filePath := range files {
				ctx.R.Human("Would upload %s to %s\n", filePath, id)
			}
			return ctx.R.Success(ctx.Meta, map[string]any{
				"planned":    true,
				"record_id":  id,
				"files":      files,
				"count":      len(files),
			}, nil)
		}

		for _, filePath := range files {
			if err := ctx.Client.UploadFile(ctx.Cmd.Context(), id, filePath); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "uploading %s: %v", filePath, err))
			}
			ctx.R.Human("Uploaded %s\n", filepath.Base(filePath))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"record_id": id,
				"files":     files,
				"count":     len(files),
			}, nil)
		}
		return nil
	}),
}

var filesListCmd = &cobra.Command{
	Use:   "list [ID]",
	Short: "List files in a record",
	Long: `List all files attached to a record.

Tries the draft file list first; if the record is published (no draft), falls back
to the published file list.`,
	Example: `  zenodo files list 12345
  zenodo files list 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("files.list", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		files, err := ctx.Client.ListFiles(ctx.Cmd.Context(), id)
		if err != nil {
			// Fall back to published files
			files, err = ctx.Client.ListPublishedFiles(ctx.Cmd.Context(), id)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
			}
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"record_id": id,
				"files":     files,
			}, nil)
		}
		for _, f := range files {
			ctx.R.Human("%s (%d bytes)\n", f.Key, f.Size)
		}
		return nil
	}),
}

var filesDeleteCmd = &cobra.Command{
	Use:   "delete [ID] [FILE...]",
	Short: "Delete files from a draft record",
	Long: `Delete one or more files from a draft record by filename.

Only works on draft records. Published records cannot have files removed.`,
	Example: `  zenodo files delete 12345 data.csv
  zenodo files delete 12345 data.csv results.json --confirm`,
	Args: cobra.MinimumNArgs(2),
	RunE: withAuth("files.delete", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		if err := requireConfirm(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		filenames := ctx.Args[1:]

		if ctx.App.DryRun {
			for _, name := range filenames {
				ctx.R.Human("Would delete %s from %s\n", name, id)
			}
			return ctx.R.Success(ctx.Meta, map[string]any{
				"planned":   true,
				"record_id": id,
				"files":     filenames,
			}, nil)
		}

		for _, name := range filenames {
			if err := ctx.Client.DeleteFile(ctx.Cmd.Context(), id, name); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "deleting %s: %v", name, err))
			}
			ctx.R.Human("Deleted %s\n", name)
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"record_id": id,
				"files":     filenames,
				"count":     len(filenames),
			}, nil)
		}
		return nil
	}),
}

var filesDownloadCmd = &cobra.Command{
	Use:   "download [ID]",
	Short: "Download files from a published record",
	Long: `Download all files from a published record to a local directory.

Use --dest to choose where files are saved (default: current directory).
Use --latest to resolve the latest published version before downloading.`,
	Example: `  zenodo files download 12345
  zenodo files download 12345 --dest ./data
  zenodo files download 12345 --latest --dest ./data`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("files.download", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		dest, _ := ctx.Cmd.Flags().GetString("dest")
		latest, _ := ctx.Cmd.Flags().GetBool("latest")
		if dest == "" {
			dest = "."
		}

		if latest {
			resolved, err := ctx.Client.ResolveLatest(ctx.Cmd.Context(), id)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "resolving latest version: %v", err))
			}
			if resolved != id {
				ctx.R.Human("Resolved latest version: %s -> %s\n", id, resolved)
				id = resolved
			}
		}

		if ctx.App.DryRun {
			ctx.R.Human("Would download files from %s to %s\n", id, dest)
			return ctx.R.Success(ctx.Meta, map[string]any{
				"planned":   true,
				"record_id": id,
				"dest":      dest,
			}, nil)
		}

		if err := ctx.Client.DownloadRecord(ctx.Cmd.Context(), id, dest); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"record_id": id,
				"dest":      dest,
			}, nil)
		}
		fmt.Fprintf(ctx.Cmd.ErrOrStderr(), "Downloaded files from %s to %s\n", id, dest)
		return nil
	}),
}

var filesInfoCmd = &cobra.Command{
	Use:   "info [ID] [FILE]",
	Short: "Show metadata for a single file",
	Long:  "Show metadata (name, size, checksum) for a single file in a draft record.",
	Example: `  zenodo files info 12345 data.csv
  zenodo files info 12345 data.csv --json`,
	Args: cobra.ExactArgs(2),
	RunE: withAuth("files.info", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		filename := ctx.Args[1]
		f, err := ctx.Client.GetFile(ctx.Cmd.Context(), id, filename)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, f, nil)
		}
		ctx.R.Human("Key:      %s\n", f.Key)
		ctx.R.Human("Size:     %d bytes\n", f.Size)
		ctx.R.Human("Checksum: %s\n", f.Checksum)
		return nil
	}),
}

var filesImportCmd = &cobra.Command{
	Use:   "import [ID]",
	Short: "Import files from the previous version",
	Long: `Import all files from the previous published version into a new draft.

Use this after "records new-version" to carry over files from the original record
without re-uploading them.`,
	Example: `  zenodo files import 12345
  zenodo files import 12345 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: withAuth("files.import", func(ctx *CmdContext) error {
		if err := requireReadOnly(&ctx.R, ctx.Meta, ctx.App); err != nil {
			return err
		}
		id := ctx.Args[0]
		if ctx.App.DryRun {
			ctx.R.Human("Would import files from previous version into %s\n", id)
			return ctx.R.Success(ctx.Meta, map[string]any{
				"planned":   true,
				"record_id": id,
				"action":    "files_import",
			}, nil)
		}
		if err := ctx.Client.ImportFiles(ctx.Cmd.Context(), id); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
		}
		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"record_id": id,
				"action":    "files_imported",
			}, nil)
		}
		ctx.R.Human("Imported files from previous version into %s\n", id)
		return nil
	}),
}

func init() {
	filesDownloadCmd.Flags().String("dest", ".", "destination directory")
	filesDownloadCmd.Flags().Bool("latest", false, "resolve and download the latest version")

	filesCmd.AddCommand(filesUploadCmd)
	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesDownloadCmd)
	filesCmd.AddCommand(filesDeleteCmd)
	filesCmd.AddCommand(filesInfoCmd)
	filesCmd.AddCommand(filesImportCmd)
}
