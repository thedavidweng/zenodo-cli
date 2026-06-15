package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/output"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

// __testNow returns a fixed time for testing.
func __testNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestNewRenderer(t *testing.T) {
	app := &AppContext{
		JSON:    true,
		Pretty:  true,
		Compact: false,
		Full:    false,
		Quiet:   true,
	}
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	r := newRenderer(app, cmd)
	if !r.JSON {
		t.Error("expected JSON=true")
	}
	if !r.Pretty {
		t.Error("expected Pretty=true")
	}
	if !r.Quiet {
		t.Error("expected Quiet=true")
	}
}

func TestMetaInput(t *testing.T) {
	app := &AppContext{
		Profile:   "sandbox",
		RequestID: "req-123",
	}
	meta := metaInput(app, "records.list")
	if meta.Command != "records.list" {
		t.Errorf("Command = %q, want records.list", meta.Command)
	}
	if meta.Profile != "sandbox" {
		t.Errorf("Profile = %q, want sandbox", meta.Profile)
	}
	if meta.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want req-123", meta.RequestID)
	}
}

func TestRequireAuthWithToken(t *testing.T) {
	var out bytes.Buffer
	r := output.Renderer{Out: &out, Err: &out}
	meta := output.RuntimeMetaInput{Command: "test"}

	client := &zenodo.Client{Token: "valid-token"}
	err := requireAuth(&r, meta, client)
	if err != nil {
		t.Errorf("expected no error with valid token, got: %v", err)
	}
}

func TestRequireAuthWithoutToken(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := output.Renderer{Out: &out, Err: &errBuf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test"}

	client := &zenodo.Client{Token: ""}
	err := requireAuth(&r, meta, client)
	if err == nil {
		t.Error("expected error without token")
	}
}

func TestRequireConfirmWithFlag(t *testing.T) {
	var out bytes.Buffer
	r := output.Renderer{Out: &out, Err: &out}
	meta := output.RuntimeMetaInput{Command: "test"}
	app := &AppContext{Confirm: true}

	err := requireConfirm(&r, meta, app)
	if err != nil {
		t.Errorf("expected no error with confirm, got: %v", err)
	}
}

func TestRequireConfirmWithoutFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := output.Renderer{Out: &out, Err: &errBuf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test"}
	app := &AppContext{Confirm: false}

	err := requireConfirm(&r, meta, app)
	if err == nil {
		t.Error("expected error without confirm")
	}
}

func TestRequireReadOnlyWithFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := output.Renderer{Out: &out, Err: &errBuf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test"}
	app := &AppContext{ReadOnly: true}

	err := requireReadOnly(&r, meta, app)
	if err == nil {
		t.Error("expected error when read-only is set")
	}
}

func TestRequireReadOnlyWithoutFlag(t *testing.T) {
	var out bytes.Buffer
	r := output.Renderer{Out: &out, Err: &out}
	meta := output.RuntimeMetaInput{Command: "test"}
	app := &AppContext{ReadOnly: false}

	err := requireReadOnly(&r, meta, app)
	if err != nil {
		t.Errorf("expected no error when read-only is not set, got: %v", err)
	}
}

func TestParseJSONValid(t *testing.T) {
	var result map[string]any
	err := parseJSON(`{"key":"value","num":42}`, &result)
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("key = %v, want value", result["key"])
	}
	if result["num"] != float64(42) {
		t.Errorf("num = %v, want 42", result["num"])
	}
}

func TestParseJSONInvalid(t *testing.T) {
	var result map[string]any
	err := parseJSON("not-json{{{", &result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetClientWithValidConfig(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
		Timeout:    30 * time.Second,
		Retries:    0,
	}

	client, err := getClient(app)
	if err != nil {
		t.Fatalf("getClient: %v", err)
	}
	if client.Token != testToken {
		t.Errorf("Token = %q, want %q", client.Token, testToken)
	}
}

func TestGetClientMissingConfig(t *testing.T) {
	app := &AppContext{
		ConfigFile: "/nonexistent/config.yaml",
		Profile:    "test",
		Timeout:    30 * time.Second,
		Retries:    0,
	}

	_, err := getClient(app)
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestGetClientMissingProfile(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "nonexistent",
		Timeout:    30 * time.Second,
		Retries:    0,
	}

	_, err := getClient(app)
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestGetClientWithSandboxOverride(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	_ = fz

	app := &AppContext{
		ConfigFile: cfgPath,
		Profile:    "test",
		Sandbox:    true,
		Timeout:    30 * time.Second,
		Retries:    0,
	}

	client, err := getClient(app)
	if err != nil {
		t.Fatalf("getClient: %v", err)
	}
	// Sandbox override should set BaseURL to sandbox, but Endpoints.API overrides it
	// so the client should still use the fake server URL
	if client.BaseURL != fz.URL() {
		t.Errorf("BaseURL = %q, want %q (endpoint override)", client.BaseURL, fz.URL())
	}
}
