package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/config"
)

// doctorCheck represents a single diagnostic check result.
type doctorCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check configuration and connectivity",
	Long: `Run diagnostic checks on your zenodo-cli setup.

Checks: config file exists, profile is configured, token is set, and API is reachable.`,
	Example: `  zenodo doctor
  zenodo doctor --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "doctor")

		checks := doctorRun(cmd.Context(), app)

		if app.JSON {
			return r.Success(meta, map[string]any{"checks": checks}, nil)
		}

		allOK := true
		for _, c := range checks {
			status := "PASS"
			if !c.OK {
				status = "FAIL"
				allOK = false
			}
			if c.Message != "" {
				r.Human("[%s] %s: %s\n", status, c.Name, c.Message)
			} else {
				r.Human("[%s] %s\n", status, c.Name)
			}
		}
		if allOK {
			r.Human("\nAll checks passed.\n")
		} else {
			r.Human("\nSome checks failed.\n")
		}
		return nil
	},
}

// doctorRun performs all diagnostic checks and returns the results.
func doctorRun(ctx context.Context, app *AppContext) []doctorCheck {
	var checks []doctorCheck

	// 1. Load config file
	cfgPath := app.ConfigFile
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "config",
			OK:      false,
			Message: err.Error() + "; run 'zenodo auth login' to create a profile",
		})
		return checks
	}
	checks = append(checks, doctorCheck{Name: "config", OK: true})

	// 2. Check if profile exists
	profile, err := cfg.GetProfile(app.Profile)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "profile",
			OK:      false,
			Message: err.Error() + "; run 'zenodo auth login' to create a profile",
		})
		return checks
	}
	checks = append(checks, doctorCheck{Name: "profile", OK: true})

	// 3. Check if token is configured
	creds := config.CredentialsFromProfileAndEnv(profile)
	if !creds.IsAuthenticated() {
		checks = append(checks, doctorCheck{
			Name:    "token",
			OK:      false,
			Message: "token is not configured; run 'zenodo auth login'",
		})
		return checks
	}
	checks = append(checks, doctorCheck{Name: "token", OK: true})

	// 4. Check API connectivity
	client, _, err := getClient(app)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "api",
			OK:      false,
			Message: "could not create client: " + err.Error(),
		})
		return checks
	}
	_, err = client.ListRecords(ctx)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "api",
			OK:      false,
			Message: "API unreachable: " + err.Error(),
		})
	} else {
		checks = append(checks, doctorCheck{Name: "api", OK: true})
	}

	return checks
}
