package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommandExists(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("Use = %q, want version", versionCmd.Use)
	}
	if versionCmd.RunE == nil {
		t.Error("version should have RunE")
	}
}

func TestVersionCommandHumanOutput(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetContext(WithAppContext(t.Context(), &AppContext{
		StartedAt: __testNow(),
		RequestID: "test",
	}))

	err := versionCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "zenodo version") {
		t.Errorf("output should contain 'zenodo version', got %q", output)
	}
}

func TestVersionCommandJSONOutput(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetContext(WithAppContext(t.Context(), &AppContext{
		JSON:      true,
		StartedAt: __testNow(),
		RequestID: "test",
		Profile:   "default",
	}))

	err := versionCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := data["ok"]; !ok {
		t.Error("JSON output should have 'ok' field")
	}
}

func TestVersionDataHasFields(t *testing.T) {
	v := VersionData{
		Version:   "1.0.0",
		Commit:    "abc123",
		Date:      "2026-01-01",
		GoVersion: "go1.26",
	}
	if v.Version != "1.0.0" {
		t.Errorf("Version = %q", v.Version)
	}
	if v.Commit != "abc123" {
		t.Errorf("Commit = %q", v.Commit)
	}
}

func TestVersionVarsHaveDefaults(t *testing.T) {
	if Version == "" {
		t.Error("Version should have default")
	}
	if Commit == "" {
		t.Error("Commit should have default")
	}
	if Date == "" {
		t.Error("Date should have default")
	}
}
