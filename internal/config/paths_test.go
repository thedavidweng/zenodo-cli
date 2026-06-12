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

func TestLoadOrCreateMigratesOldConfig(t *testing.T) {
	dir := t.TempDir()
	newPath := filepath.Join(dir, "new", "config.yaml")
	oldPath := filepath.Join(dir, "old", "config.yaml")

	// Create old config with content.
	if err := os.MkdirAll(filepath.Dir(oldPath), 0700); err != nil {
		t.Fatal(err)
	}
	oldContent := `current_profile: prod
profiles:
  prod:
    token: "migrated-token"
`
	if err := os.WriteFile(oldPath, []byte(oldContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Monkey-patch oldDefaultConfigPath to return our test path.
	saved := oldDefaultConfigPathFn
	oldDefaultConfigPathFn = func() string { return oldPath }
	defer func() { oldDefaultConfigPathFn = saved }()

	cfg, err := LoadOrCreate(newPath)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	// Config should have been migrated.
	if cfg.Profiles["prod"].Token != "migrated-token" {
		t.Errorf("token = %q, want migrated-token", cfg.Profiles["prod"].Token)
	}

	// Old file should be gone.
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old config file should have been removed after migration")
	}

	// New file should exist.
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("new config file should exist: %v", err)
	}
}

func TestLoadOrCreateNoMigrationWhenNewExists(t *testing.T) {
	dir := t.TempDir()
	newPath := filepath.Join(dir, "config.yaml")
	oldPath := filepath.Join(dir, "old-config.yaml")

	// Create both old and new configs.
	newContent := `current_profile: new
profiles:
  new:
    token: "new-token"
`
	if err := os.WriteFile(newPath, []byte(newContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, []byte("current_profile: old\n"), 0600); err != nil {
		t.Fatal(err)
	}

	saved := oldDefaultConfigPathFn
	oldDefaultConfigPathFn = func() string { return oldPath }
	defer func() { oldDefaultConfigPathFn = saved }()

	cfg, err := LoadOrCreate(newPath)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	// Should load the new config, not migrate old.
	if cfg.CurrentProfile != "new" {
		t.Errorf("current_profile = %q, want new", cfg.CurrentProfile)
	}

	// Old file should still exist.
	if _, err := os.Stat(oldPath); err != nil {
		t.Error("old config should not be touched when new exists")
	}
}
