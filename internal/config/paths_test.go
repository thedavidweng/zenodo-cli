package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
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

func TestDefaultConfigPathFallsBackToHome(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)

	os.Unsetenv("XDG_CONFIG_HOME")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	path := DefaultConfigPath()
	want := filepath.Join(home, ".config", "zenodo-cli", "config.yaml")
	if path != want {
		t.Errorf("DefaultConfigPath() = %q, want %q", path, want)
	}
}
