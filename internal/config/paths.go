package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default configuration file path.
// It respects XDG_CONFIG_HOME; falls back to ~/.config/zenodo-cli/config.yaml.
func DefaultConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".config", "zenodo-cli", "config.yaml")
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "zenodo-cli", "config.yaml")
}
