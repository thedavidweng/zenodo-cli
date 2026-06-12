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

			// Read flags into AppContext
			app.ConfigFile, _ = cmd.Flags().GetString("config")
			app.Profile, _ = cmd.Flags().GetString("profile")
			app.Sandbox, _ = cmd.Flags().GetBool("sandbox")
			app.JSON, _ = cmd.Flags().GetBool("json")
			app.Pretty, _ = cmd.Flags().GetBool("pretty")
			app.Compact, _ = cmd.Flags().GetBool("compact")
			app.Full, _ = cmd.Flags().GetBool("full")
			app.Quiet, _ = cmd.Flags().GetBool("quiet")
			app.Events, _ = cmd.Flags().GetBool("events")
			app.ReadOnly, _ = cmd.Flags().GetBool("read-only")
			app.DryRun, _ = cmd.Flags().GetBool("dry-run")
			app.Confirm, _ = cmd.Flags().GetBool("confirm")
			app.Timeout, _ = cmd.Flags().GetDuration("timeout")
			app.Retries, _ = cmd.Flags().GetInt("retries")
			app.NoColor, _ = cmd.Flags().GetBool("no-color")
			app.Verbose, _ = cmd.Flags().GetBool("verbose")
			app.Debug, _ = cmd.Flags().GetBool("debug")

			// Environment variable overrides (checked after flags)
			if app.ConfigFile == "" {
				if env := os.Getenv("ZENODO_CONFIG"); env != "" {
					app.ConfigFile = env
				}
			}
			if app.Profile == "default" {
				if env := os.Getenv("ZENODO_PROFILE"); env != "" {
					app.Profile = env
				}
			}
			if !app.Sandbox {
				if env := os.Getenv("ZENODO_SANDBOX"); env != "" {
					if env == "true" || env == "1" || env == "yes" {
						app.Sandbox = true
					}
				}
			}
			if app.Timeout == 30*time.Second {
				if env := os.Getenv("ZENODO_TIMEOUT"); env != "" {
					if d, err := time.ParseDuration(env); err == nil {
						app.Timeout = d
					}
				}
			}
			if app.Retries == 3 {
				if env := os.Getenv("ZENODO_RETRIES"); env != "" {
					if n, err := strconv.Atoi(env); err == nil {
						app.Retries = n
					}
				}
			}
			if !app.Debug {
				if env := os.Getenv("ZENODO_DEBUG"); env != "" {
					if env == "true" || env == "1" || env == "yes" {
						app.Debug = true
					}
				}
			}

			// Validation
			if app.Retries < 0 {
				return fmt.Errorf("--retries must be >= 0")
			}
			if app.Timeout <= 0 {
				return fmt.Errorf("--timeout must be positive")
			}

			// Full wins over compact
			if app.Full {
				app.Compact = false
			}

			// Store in command context
			ctx := WithAppContext(cmd.Context(), app)
			cmd.SetContext(ctx)
			return nil
		},
	}

	root.PersistentFlags().String("config", "", "config file path")
	root.PersistentFlags().String("profile", "default", "profile name")
	root.PersistentFlags().Bool("sandbox", false, "use Zenodo sandbox")
	root.PersistentFlags().Bool("json", false, "emit JSON envelope to stdout")
	root.PersistentFlags().Bool("pretty", false, "pretty-print JSON")
	root.PersistentFlags().Bool("compact", false, "compact output fields")
	root.PersistentFlags().Bool("full", false, "full normalized fields")
	root.PersistentFlags().Bool("quiet", false, "suppress progress output")
	root.PersistentFlags().Bool("events", false, "emit NDJSON progress events to stderr")
	root.PersistentFlags().Bool("read-only", false, "block remote mutations")
	root.PersistentFlags().Bool("dry-run", false, "plan mutations without execution")
	root.PersistentFlags().Bool("confirm", false, "confirm high-risk mutations")
	root.PersistentFlags().Duration("timeout", 30*time.Second, "command/API timeout")
	root.PersistentFlags().Int("retries", 3, "retry count for retryable failures")
	root.PersistentFlags().Bool("no-color", false, "disable ANSI color")
	root.PersistentFlags().Bool("verbose", false, "diagnostics to stderr")
	root.PersistentFlags().Bool("debug", false, "debug diagnostics to stderr with secrets redacted")

	return root
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
