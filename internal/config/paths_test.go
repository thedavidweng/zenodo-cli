package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", orig) }()
	mustUnsetenv(t, "XDG_CONFIG_HOME")

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
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", orig) }()

	xdg := "/tmp/test-xdg-config"
	mustSetenv(t, "XDG_CONFIG_HOME", xdg)
	path := DefaultConfigPath()

	want := filepath.Join(xdg, "zenodo-cli", "config.yaml")
	if path != want {
		t.Errorf("DefaultConfigPath() = %q, want %q", path, want)
	}
}
