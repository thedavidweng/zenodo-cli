package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestLoadOrCreateCreatesNew(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Profiles == nil {
		t.Fatal("expected non-nil profiles map")
	}

	// File should exist on disk now.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestLoadOrCreateExistingFile(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	// Write a config manually.
	yaml := `current_profile: prod
profiles:
  prod:
    token: "tok123"
    sandbox: false
    base_url: "https://zenodo.org/api"
`
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if cfg.CurrentProfile != "prod" {
		t.Errorf("current_profile = %q, want prod", cfg.CurrentProfile)
	}
	if cfg.Profiles["prod"].Token != "tok123" {
		t.Errorf("token = %q, want tok123", cfg.Profiles["prod"].Token)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("  - :\n\tbad: [}"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "sandbox",
		Profiles: map[string]*Profile{
			"sandbox": {
				Token:     "test-token",
				Sandbox:   true,
				BaseURL:   "https://sandbox.zenodo.org/api",
				Endpoints: Endpoints{API: "https://sandbox.zenodo.org/api"},
			},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.CurrentProfile != "sandbox" {
		t.Errorf("current_profile = %q, want sandbox", loaded.CurrentProfile)
	}
	if loaded.Profiles["sandbox"].Token != "test-token" {
		t.Errorf("token = %q, want test-token", loaded.Profiles["sandbox"].Token)
	}
	if loaded.Profiles["sandbox"].Endpoints.API != "https://sandbox.zenodo.org/api" {
		t.Errorf("endpoints.api = %q", loaded.Profiles["sandbox"].Endpoints.API)
	}
}

func TestSaveCreatesDirWith0700(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows")
	}
	dir := tempDir(t)
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := &Config{Profiles: map[string]*Profile{}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "subdir"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("dir perm = %o, want 0700", perm)
	}
}

func TestSaveFilePerms0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows")
	}
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{Profiles: map[string]*Profile{}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file perm = %o, want 0600", perm)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	// First save.
	cfg1 := &Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {Token: "old-token", Sandbox: false},
		},
	}
	if err := Save(path, cfg1); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// Second save — overwrites the existing file.
	cfg2 := &Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {Token: "new-token", Sandbox: true},
		},
	}
	if err := Save(path, cfg2); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	// Verify the overwrite took effect.
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Profiles["default"].Token != "new-token" {
		t.Errorf("token = %q, want new-token", loaded.Profiles["default"].Token)
	}
	if !loaded.Profiles["default"].Sandbox {
		t.Error("expected sandbox=true after overwrite")
	}
}

func TestGetProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {Token: "abc", Sandbox: false},
			"sb":      {Token: "xyz", Sandbox: true},
		},
	}

	// Named profile.
	p, err := cfg.GetProfile("sb")
	if err != nil {
		t.Fatalf("GetProfile(sb): %v", err)
	}
	if p.Token != "xyz" {
		t.Errorf("token = %q, want xyz", p.Token)
	}

	// Empty name uses current profile.
	p, err = cfg.GetProfile("")
	if err != nil {
		t.Fatalf("GetProfile(\"\"): %v", err)
	}
	if p.Token != "abc" {
		t.Errorf("token = %q, want abc", p.Token)
	}

	// Missing profile.
	_, err = cfg.GetProfile("nonexistent")
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestSetProfile(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{}}

	cfg.SetProfile("new", &Profile{Token: "tok", Sandbox: true})
	if cfg.Profiles["new"].Token != "tok" {
		t.Errorf("token = %q, want tok", cfg.Profiles["new"].Token)
	}

	// Overwrite.
	cfg.SetProfile("new", &Profile{Token: "tok2", Sandbox: false})
	if cfg.Profiles["new"].Token != "tok2" {
		t.Errorf("token = %q, want tok2", cfg.Profiles["new"].Token)
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSetProfileNilMap(t *testing.T) {
	cfg := &Config{} // Profiles is nil

	cfg.SetProfile("default", &Profile{Token: "tok", Sandbox: false})

	if cfg.Profiles == nil {
		t.Fatal("SetProfile should initialize nil Profiles map")
	}
	if cfg.Profiles["default"] == nil {
		t.Fatal("expected profile to be set")
	}
	if cfg.Profiles["default"].Token != "tok" {
		t.Errorf("token = %q, want tok", cfg.Profiles["default"].Token)
	}
}

func TestLoadOrCreateInvalidYAML(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	// Write a file that exists but contains invalid YAML.
	if err := os.WriteFile(path, []byte("  - :\n\tbad: [}"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadOrCreate(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML in LoadOrCreate")
	}
	// The error should NOT be ErrNotExist; it should be a parse/read error.
	if errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected non-ErrNotExist error, got ErrNotExist")
	}
}

func TestSaveMkdirAllFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping unix path test on Windows")
	}

	cfg := &Config{Profiles: map[string]*Profile{}}
	// /dev/null is a file, not a directory, so MkdirAll under it should fail.
	err := Save("/dev/null/impossible/config.yaml", cfg)
	if err == nil {
		t.Fatal("expected error when MkdirAll fails")
	}
}

func TestLoadNilProfilesInit(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config.yaml")

	// Write YAML that is valid but has no "profiles" key, so Profiles will be nil after unmarshal.
	yamlContent := `current_profile: default
`
	if err := os.WriteFile(path, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Profiles == nil {
		t.Fatal("Load should initialize nil Profiles to empty map")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected empty profiles map, got %d entries", len(cfg.Profiles))
	}
}
