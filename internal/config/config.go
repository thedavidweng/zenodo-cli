package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration file structure.
type Config struct {
	CurrentProfile string              `yaml:"current_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
}

// Profile holds settings for a named configuration profile.
type Profile struct {
	Token     string    `yaml:"token"`
	Sandbox   bool      `yaml:"sandbox"`
	BaseURL   string    `yaml:"base_url"`
	Endpoints Endpoints `yaml:"endpoints"`
}

// Endpoints allows overriding API endpoints (useful for testing/self-hosted).
type Endpoints struct {
	API string `yaml:"api"`
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	return &cfg, nil
}

// Save writes the config to path atomically with 0600 permissions.
// Parent directories are created with 0700 permissions.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	// Atomic write: write to temp file in same dir, then rename.
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }() // clean up on error

	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	// Remove existing file before rename — required on Windows where
	// os.Rename fails if the destination already exists.
	_ = os.Remove(path)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

// LoadOrCreate loads the config from path, or creates a new default config
// if the file does not exist.
func LoadOrCreate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	// File doesn't exist; create a default config and save it.
	cfg = &Config{Profiles: map[string]*Profile{}}
	if err := Save(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetProfile returns the named profile, or the current profile if name is empty.
func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == "" {
		name = c.CurrentProfile
	}
	p, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

// SetProfile adds or replaces a named profile.
func (c *Config) SetProfile(name string, p *Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	c.Profiles[name] = p
}
