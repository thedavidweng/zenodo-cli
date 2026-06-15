package config

import (
	"path/filepath"
	"testing"
)

func BenchmarkLoad(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "sandbox",
		Profiles: map[string]*Profile{
			"sandbox": {
				Token:   "bench-token-abc123",
				Sandbox: true,
				BaseURL: "https://sandbox.zenodo.org/api",
				Endpoints: Endpoints{
					API: "https://sandbox.zenodo.org/api",
				},
			},
			"production": {
				Token:   "prod-token-xyz789",
				Sandbox: false,
				BaseURL: "https://zenodo.org/api",
				Endpoints: Endpoints{
					API: "https://zenodo.org/api",
				},
			},
		},
	}
	if err := Save(path, cfg); err != nil {
		b.Fatalf("setup Save: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		loaded, err := Load(path)
		if err != nil {
			b.Fatal(err)
		}
		_ = loaded
	}
}

func BenchmarkSave(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "sandbox",
		Profiles: map[string]*Profile{
			"sandbox": {
				Token:   "bench-token-abc123",
				Sandbox: true,
				BaseURL: "https://sandbox.zenodo.org/api",
				Endpoints: Endpoints{
					API: "https://sandbox.zenodo.org/api",
				},
			},
			"production": {
				Token:   "prod-token-xyz789",
				Sandbox: false,
				BaseURL: "https://zenodo.org/api",
				Endpoints: Endpoints{
					API: "https://zenodo.org/api",
				},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := Save(path, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
