package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the config file path:
// XDG_CONFIG_HOME/zenodo-cli/config.yaml if set, otherwise the platform default.
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
