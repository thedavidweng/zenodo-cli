package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type VersionData struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Print the version, commit hash, build date, and Go version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)

		data := VersionData{
			Version:   Version,
			Commit:    Commit,
			Date:      Date,
			GoVersion: runtime.Version(),
		}

		return r.Render(metaInput(app, "version"), data, func() {
			r.Human("zenodo version %s (commit: %s, date: %s)\n", Version, Commit, Date)
		})
	},
}
