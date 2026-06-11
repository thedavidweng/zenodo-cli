package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDoctorCommandExists(t *testing.T) {
	if doctorCmd.Use != "doctor" {
		t.Errorf("Use = %q, want doctor", doctorCmd.Use)
	}
	if doctorCmd.RunE == nil {
		t.Error("doctor should have RunE")
	}
}

func TestDoctorCheckStruct(t *testing.T) {
	c := doctorCheck{
		Name:    "config",
		OK:      true,
		Message: "loaded",
	}
	if c.Name != "config" {
		t.Errorf("Name = %q", c.Name)
	}
	if !c.OK {
		t.Error("expected OK=true")
	}
}

func TestDoctorRunNoConfig(t *testing.T) {
	app := &AppContext{
		ConfigFile: "/nonexistent/config.yaml",
		Profile:    "default",
	}

	checks := doctorRun(nil, app)
	if len(checks) == 0 {
		t.Fatal("expected at least one check")
	}
	if checks[0].Name != "config" {
		t.Errorf("first check name = %q, want config", checks[0].Name)
	}
	if checks[0].OK {
		t.Error("expected config check to fail for nonexistent file")
	}
}

func TestDoctorHumanOutput(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(WithAppContext(t.Context(), &AppContext{
		ConfigFile: "/nonexistent/config.yaml",
		Profile:    "default",
		StartedAt:  __testNow(),
		RequestID:  "test",
	}))

	err := doctorCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "FAIL") {
		t.Errorf("output should contain 'FAIL', got %q", output)
	}
}

func TestDoctorJSONOutput(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(WithAppContext(t.Context(), &AppContext{
		ConfigFile: "/nonexistent/config.yaml",
		Profile:    "default",
		JSON:       true,
		StartedAt:  __testNow(),
		RequestID:  "test",
	}))

	err := doctorCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := result["ok"]; !ok {
		t.Error("JSON should have 'ok' field")
	}
}
