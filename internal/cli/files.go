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
}

var filesUploadCmd = &cobra.Command{
	Use:   "upload [ID] [FILE...]",
	Short: "Upload files to a draft record",
	Args:  cobra.MinimumNArgs(2),
	RunE: withAuth("files.upload", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		files := ctx.Args[1:]

		for _, filePath := range files {
			if ctx.App.DryRun {
				ctx.R.Human("Would upload %s to %s\n", filePath, id)
				continue
			}
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
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("files.list", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		files, err := ctx.Client.ListFiles(ctx.Cmd.Context(), id)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrZenodoAPI, "%v", err))
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

var filesDownloadCmd = &cobra.Command{
	Use:   "download [ID]",
	Short: "Download files from a published record",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("files.download", func(ctx *CmdContext) error {
		id := ctx.Args[0]
		dest, _ := ctx.Cmd.Flags().GetString("dest")
		if dest == "" {
			dest = "."
		}

		if ctx.App.DryRun {
			ctx.R.Human("Would download files from %s to %s\n", id, dest)
			return nil
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

func init() {
	filesDownloadCmd.Flags().String("dest", ".", "destination directory")

	filesCmd.AddCommand(filesUploadCmd)
	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesDownloadCmd)
}
