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
		NoColor: true,
		Verbose: true,
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
	if !r.NoColor {
		t.Error("expected NoColor=true")
	}
	if !r.Verbose {
		t.Error("expected Verbose=true")
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
