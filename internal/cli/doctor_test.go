package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/testutil"
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

	checks := doctorRun(t.Context(), app)
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

func TestDoctorRunValidConfigNoToken(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: ""
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
	}

	checks := doctorRun(t.Context(), app)
	// Should have config=ok, profile=ok, token=fail
	if len(checks) < 3 {
		t.Fatalf("expected at least 3 checks, got %d", len(checks))
	}
	if !checks[0].OK {
		t.Errorf("config check should pass, got: %s", checks[0].Message)
	}
	if !checks[1].OK {
		t.Errorf("profile check should pass, got: %s", checks[1].Message)
	}
	if checks[2].OK {
		t.Error("token check should fail for empty token")
	}
}

func TestDoctorRunMissingProfile(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := `current_profile: test
profiles:
  test:
    token: some-token
    base_url: https://zenodo.org
`
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "nonexistent",
	}

	checks := doctorRun(t.Context(), app)
	// Should have config=ok, profile=fail
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 checks, got %d", len(checks))
	}
	if !checks[0].OK {
		t.Errorf("config check should pass")
	}
	if checks[1].OK {
		t.Error("profile check should fail for missing profile")
	}
}

func TestDoctorRunAllPass(t *testing.T) {
	token := "doctor-test-token"
	fz := testutil.NewFakeZenodo(token)
	defer fz.Close()

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`current_profile: test
profiles:
  test:
    token: %s
    base_url: https://zenodo.org
    endpoints:
      api: %s
`, token, fz.URL())
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
	}

	checks := doctorRun(t.Context(), app)
	if len(checks) != 4 {
		t.Fatalf("expected 4 checks, got %d", len(checks))
	}
	for _, c := range checks {
		if !c.OK {
			t.Errorf("check %q should pass: %s", c.Name, c.Message)
		}
	}
}

func TestDoctorRunAPIFail(t *testing.T) {
	// Server that returns 500 for all requests
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
	}))
	defer failServer.Close()

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`current_profile: test
profiles:
  test:
    token: some-token
    base_url: https://zenodo.org
    endpoints:
      api: %s
`, failServer.URL)
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
	}

	checks := doctorRun(t.Context(), app)
	if len(checks) != 4 {
		t.Fatalf("expected 4 checks, got %d", len(checks))
	}
	// config, profile, token should pass
	if !checks[0].OK {
		t.Errorf("config check should pass: %s", checks[0].Message)
	}
	if !checks[1].OK {
		t.Errorf("profile check should pass: %s", checks[1].Message)
	}
	if !checks[2].OK {
		t.Errorf("token check should pass: %s", checks[2].Message)
	}
	// api should fail
	if checks[3].OK {
		t.Error("api check should fail for 500 response")
	}
}

func TestDoctorHumanOutputPartialFail(t *testing.T) {
	// Config loads fine, token is set, but API is unreachable
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
	}))
	defer failServer.Close()

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`current_profile: test
profiles:
  test:
    token: some-token
    base_url: https://zenodo.org
    endpoints:
      api: %s
`, failServer.URL)
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(WithAppContext(t.Context(), &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
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
	if !strings.Contains(output, "PASS") {
		t.Errorf("output should contain 'PASS' for passing checks, got %q", output)
	}
	if !strings.Contains(output, "Some checks failed") {
		t.Errorf("output should contain 'Some checks failed', got %q", output)
	}
}
