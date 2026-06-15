package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "zenodo",
		Short:         "Agent-friendly Zenodo CLI",
		Long:          `A single-binary CLI tool for Zenodo deposit management, file upload/download, and API access.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			app := &AppContext{
				StartedAt: time.Now(),
				RequestID: uuid.New().String(),
			}

			readFlags(cmd, app)
			applyEnvOverrides(app)

			if err := validateAppContext(app); err != nil {
				return err
			}

			ctx := WithAppContext(cmd.Context(), app)
			cmd.SetContext(ctx)
			return nil
		},
	}

	registerFlags(root)
	return root
}

func readFlags(cmd *cobra.Command, app *AppContext) {
	app.ConfigFile, _ = cmd.Flags().GetString("config")
	app.Profile, _ = cmd.Flags().GetString("profile")
	app.Sandbox, _ = cmd.Flags().GetBool("sandbox")
	app.JSON, _ = cmd.Flags().GetBool("json")
	app.Pretty, _ = cmd.Flags().GetBool("pretty")
	app.Compact, _ = cmd.Flags().GetBool("compact")
	app.Full, _ = cmd.Flags().GetBool("full")
	app.Quiet, _ = cmd.Flags().GetBool("quiet")
	app.ReadOnly, _ = cmd.Flags().GetBool("read-only")
	app.DryRun, _ = cmd.Flags().GetBool("dry-run")
	app.Confirm, _ = cmd.Flags().GetBool("confirm")
	app.Timeout, _ = cmd.Flags().GetDuration("timeout")
	app.Retries, _ = cmd.Flags().GetInt("retries")
}

func applyEnvOverrides(app *AppContext) {
	if app.ConfigFile == "" {
		app.ConfigFile = envOr("ZENODO_CONFIG", "")
	}
	app.Profile = envOrDefault(app.Profile, "default", "ZENODO_PROFILE")
	app.Sandbox = app.Sandbox || envBool("ZENODO_SANDBOX")
	app.Timeout = envDuration("ZENODO_TIMEOUT", app.Timeout, 5*time.Minute)
	app.Retries = envInt("ZENODO_RETRIES", app.Retries, 3)
	app.JSON = app.JSON || envBool("ZENODO_JSON")
	app.ReadOnly = app.ReadOnly || envBool("ZENODO_READ_ONLY")
	app.DryRun = app.DryRun || envBool("ZENODO_DRY_RUN")
	app.Confirm = app.Confirm || envBool("ZENODO_CONFIRM")
	app.Quiet = app.Quiet || envBool("ZENODO_QUIET")
}

func validateAppContext(app *AppContext) error {
	if app.Retries < 0 {
		return fmt.Errorf("--retries must be >= 0")
	}
	if app.Timeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	if app.Full {
		app.Compact = false
	}
	return nil
}

func registerFlags(root *cobra.Command) {
	root.PersistentFlags().String("config", "", "config file path (YAML, default: ~/.config/zenodo-cli/config.yaml)")
	root.PersistentFlags().String("profile", "default", "credential profile name")
	root.PersistentFlags().Bool("sandbox", false, "use Zenodo sandbox (sandbox.zenodo.org)")
	root.PersistentFlags().Bool("json", false, "emit JSON envelope to stdout")
	root.PersistentFlags().Bool("pretty", false, "pretty-print JSON output")
	root.PersistentFlags().Bool("compact", false, "omit null/empty fields from JSON output")
	root.PersistentFlags().Bool("full", false, "include all fields in JSON output (overrides --compact)")
	root.PersistentFlags().Bool("quiet", false, "suppress progress messages")
	root.PersistentFlags().Bool("read-only", false, "block all remote mutations")
	root.PersistentFlags().Bool("dry-run", false, "show what would happen without executing")
	root.PersistentFlags().Bool("confirm", false, "confirm irreversible operations")
	root.PersistentFlags().Duration("timeout", 5*time.Minute, "command/API timeout")
	root.PersistentFlags().Int("retries", 3, "retry count for retryable failures")
}

// Execute runs the root command with signal handling.
func Execute() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		_ = sig
		cancel()

		sig = <-sigCh
		_ = sig
		fmt.Fprintf(os.Stderr, "\ninterrupted\n")
		os.Exit(130)
	}()

	silenceAllCommands(rootCmd)
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

// silenceAllCommands recursively propagates SilenceUsage and SilenceErrors
// to every command in the tree.
func silenceAllCommands(cmd *cobra.Command) {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	for _, child := range cmd.Commands() {
		silenceAllCommands(child)
	}
}

func init() {
	registerSubcommands(rootCmd)
}

// registerSubcommands attaches all subcommands to the given root.
func registerSubcommands(root *cobra.Command) {
	root.AddCommand(versionCmd)
	root.AddCommand(authCmd)
	root.AddCommand(recordsCmd)
	root.AddCommand(filesCmd)
	root.AddCommand(searchCmd)
	root.AddCommand(doctorCmd)
	root.AddCommand(completionCmd)
	root.AddCommand(apiCmd)
}

// --- env helpers ---

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	v := os.Getenv(key)
	return v == "true" || v == "1" || v == "yes"
}

func envOrDefault(current, defaultVal, key string) string {
	if current != defaultVal {
		return current
	}
	return envOr(key, current)
}

func envDuration(key string, current, defaultVal time.Duration) time.Duration {
	if current != defaultVal {
		return current
	}
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return current
}

func envInt(key string, current, defaultVal int) int {
	if current != defaultVal {
		return current
	}
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return current
}
