package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/config"
	"github.com/thedavidweng/zenodo-cli/internal/model"
	"github.com/thedavidweng/zenodo-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Zenodo authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store a Zenodo API token",
	Long: `Authenticate with Zenodo using a personal access token.

Get yours at: https://zenodo.org/account/settings/applications/

Three ways to provide the token (in order of precedence):

  1. Flag:       zenodo auth login --token TOKEN
  2. Env var:    ZENODO_TOKEN=TOKEN zenodo auth login
  3. Interactive: zenodo auth login (prompts with a link)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "auth.login")

		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.LoadOrCreate(cfgPath)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "loading config: %v", err))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			profile = &config.Profile{}
		}

		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			token = os.Getenv("ZENODO_TOKEN")
		}
		if token == "" {
			if !isTerminal() {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "token required. Use --token or set ZENODO_TOKEN"))
			}
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Create a personal access token at:")
			fmt.Fprintln(os.Stderr)
			if app.Sandbox {
				fmt.Fprintln(os.Stderr, "  https://sandbox.zenodo.org/account/settings/applications/")
			} else {
				fmt.Fprintln(os.Stderr, "  https://zenodo.org/account/settings/applications/")
			}
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Required scopes: deposit:write, deposit:actions")
			fmt.Fprintln(os.Stderr)
			fmt.Fprint(os.Stderr, "Paste your API token: ")
			token = readLine()
		}
		if token == "" {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "token is required"))
		}

		profile.Token = token
		profile.Sandbox = app.Sandbox
		cfg.SetProfile(app.Profile, profile)

		if err := config.Save(cfgPath, cfg); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "saving config: %v", err))
		}

		r.Human("Token saved for profile %q\n", app.Profile)
		return r.Success(meta, map[string]any{
			"profile": app.Profile,
			"sandbox": app.Sandbox,
		}, nil)
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "auth.status")

		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrAuthRequired,
				"Not configured. Run 'zenodo auth login' to get started.",
				map[string]any{"profile": app.Profile},
			))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrAuthRequired,
				fmt.Sprintf("Profile %q not configured. Run 'zenodo auth login' to get started.", app.Profile),
				map[string]any{"profile": app.Profile},
			))
		}

		creds := config.CredentialsFromProfileAndEnv(profile)
		if !creds.IsAuthenticated() {
			return r.Failure(meta, output.Errorf(model.ErrAuthRequired, "no token configured"))
		}

		r.Human("Authenticated (profile: %q, sandbox: %v)\n", app.Profile, creds.Sandbox)
		return r.Success(meta, map[string]any{
			"authenticated": true,
			"profile":       app.Profile,
			"sandbox":       creds.Sandbox,
			"base_url":      creds.BaseURL,
		}, nil)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials for current profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := metaInput(app, "auth.logout")

		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "loading config: %v", err))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "no profile %q configured", app.Profile))
		}

		if app.DryRun {
			r.Human("Would clear credentials for profile %q\n", app.Profile)
			return r.Success(meta, map[string]any{
				"planned": true,
				"profile": app.Profile,
				"action":  "clear_credentials",
			}, nil)
		}

		profile.Token = ""
		cfg.SetProfile(app.Profile, profile)

		if err := config.Save(cfgPath, cfg); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "saving config: %v", err))
		}

		r.Human("Logged out from profile %q\n", app.Profile)
		return r.Success(meta, map[string]any{
			"profile": app.Profile,
			"action":  "cleared_credentials",
		}, nil)
	},
}

func init() {
	authLoginCmd.Flags().String("token", "", "Zenodo API token")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func readLine() string {
	return readFrom(os.Stdin)
}

func readFrom(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
