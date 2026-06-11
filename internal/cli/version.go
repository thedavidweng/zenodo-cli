package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

// Version information, set via ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// VersionData is the data for the version command JSON output.
type VersionData struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)

		data := VersionData{
			Version:   Version,
			Commit:    Commit,
			Date:      Date,
			GoVersion: runtime.Version(),
		}

		if app.JSON {
			return r.Success(metaInput(app, "version"), data, nil)
		}

		r.Human("zenodo version %s (commit: %s, date: %s)\n", Version, Commit, Date)
		return nil
	},
}
