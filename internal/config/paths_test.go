package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)
	os.Unsetenv("XDG_CONFIG_HOME")

	path := DefaultConfigPath()
	if path == "" {
		t.Fatal("DefaultConfigPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultConfigPath() = %q, want absolute path", path)
	}
	if filepath.Ext(path) != ".yaml" {
		t.Errorf("extension = %q, want .yaml", filepath.Ext(path))
	}

	// Must be based on os.UserConfigDir(), not hardcoded ~/.config.
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("cannot determine config dir: %v", err)
	}
	want := filepath.Join(cfgDir, "zenodo-cli", "config.yaml")
	if path != want {
		t.Errorf("DefaultConfigPath() = %q, want %q", path, want)
	}
}

func TestDefaultConfigPathUsesXDG(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)

	os.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg-config")
	path := DefaultConfigPath()

	want := filepath.Join("/tmp/test-xdg-config", "zenodo-cli", "config.yaml")
	if path != want {
		t.Errorf("DefaultConfigPath() = %q, want %q", path, want)
	}
}
