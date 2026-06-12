package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default configuration file path.
//
// Resolution order:
//  1. XDG_CONFIG_HOME/zenodo-cli/config.yaml  (if XDG_CONFIG_HOME is set)
//  2. os.UserConfigDir()/zenodo-cli/config.yaml  (platform-appropriate:
//     Linux ~/.config, macOS ~/Library/Application Support, Windows %APPDATA%)
//  3. .config/zenodo-cli/config.yaml  (relative fallback if UserConfigDir fails)
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "zenodo-cli", "config.yaml")
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".config", "zenodo-cli", "config.yaml")
	}
	return filepath.Join(cfgDir, "zenodo-cli", "config.yaml")
}
