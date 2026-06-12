package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

func TestRendererSuccessJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:  &out,
		Err:  &errBuf,
		JSON: true,
	}
	started := time.Now()
	meta := RuntimeMetaInput{
		Command:   "deposit list",
		Profile:   "default",
		RequestID: "req-123",
		StartedAt: started,
	}
	data := map[string]any{"id": "12345"}

	if err := r.Success(meta, data, nil); err != nil {
		t.Fatalf("Success: %v", err)
	}

	var env model.Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Error != nil {
		t.Errorf("expected nil error, got %v", env.Error)
	}
	if env.Meta.Command != "deposit list" {
		t.Errorf("command = %q, want %q", env.Meta.Command, "deposit list")
	}
	if env.Meta.Profile != "default" {
		t.Errorf("profile = %q, want %q", env.Meta.Profile, "default")
	}
	if env.Meta.RequestID != "req-123" {
		t.Errorf("request_id = %q, want %q", env.Meta.RequestID, "req-123")
	}
	if env.Meta.SchemaVersion != model.SchemaVersion {
		t.Errorf("schema_version = %q, want %q", env.Meta.SchemaVersion, model.SchemaVersion)
	}
	if env.Meta.DurationMS < 0 {
		t.Errorf("duration_ms = %d, want >= 0", env.Meta.DurationMS)
	}
}

func TestRendererSuccessWithWarnings(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{Out: &out, JSON: true}
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}
	warnings := []string{"deprecated field", "rate limit approaching"}

	if err := r.Success(meta, nil, warnings); err != nil {
		t.Fatalf("Success: %v", err)
	}

	var env model.Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(env.Meta.Warnings) != 2 {
		t.Errorf("warnings count = %d, want 2", len(env.Meta.Warnings))
	}
	if env.Meta.Warnings[0] != "deprecated field" {
		t.Errorf("warnings[0] = %q, want %q", env.Meta.Warnings[0], "deprecated field")
	}
}

func TestRendererFailureJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:  &out,
		Err:  &errBuf,
		JSON: true,
	}
	meta := RuntimeMetaInput{Command: "deposit create", StartedAt: time.Now()}
	errBody := Errorf(model.ErrValidationFailed, "title is required")

	err := r.Failure(meta, errBody)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	cmdErr, ok := err.(*model.CommandError)
	if !ok {
		t.Fatalf("expected *model.CommandError, got %T", err)
	}
	if cmdErr.Code != model.ErrValidationFailed {
		t.Errorf("CommandError.Code = %q, want %q", cmdErr.Code, model.ErrValidationFailed)
	}

	var env model.Envelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.OK {
		t.Error("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if env.Error.Code != model.ErrValidationFailed {
		t.Errorf("error.code = %q, want %q", env.Error.Code, model.ErrValidationFailed)
	}
	if env.Error.Message != "title is required" {
		t.Errorf("error.message = %q, want %q", env.Error.Message, "title is required")
	}
}

func TestRendererFailureHuman(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:  &out,
		Err:  &errBuf,
		JSON: false,
	}
	meta := RuntimeMetaInput{Command: "deposit create", StartedAt: time.Now()}
	errBody := Errorf(model.ErrAuthFailed, "invalid token")

	err := r.Failure(meta, errBody)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	errOut := errBuf.String()
	if !strings.Contains(errOut, "invalid token") {
		t.Errorf("stderr = %q, should contain 'invalid token'", errOut)
	}
	if !strings.Contains(errOut, "AUTH_FAILED") {
		t.Errorf("stderr = %q, should contain 'AUTH_FAILED'", errOut)
	}
	// JSON should not be written in human mode
	if out.Len() > 0 {
		t.Errorf("stdout should be empty in human failure mode, got %q", out.String())
	}
}

func TestRendererHumanSuppressedQuiet(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:   &out,
		Err:   &errBuf,
		Quiet: true,
	}
	r.Human("hello %s", "world")
	if out.Len() > 0 {
		t.Errorf("Human output should be suppressed in quiet mode, got %q", out.String())
	}
}

func TestRendererHumanNotSuppressed(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{Out: &out}
	r.Human("hello %s", "world")
	if got := out.String(); got != "hello world\n" {
		t.Errorf("Human output = %q, want %q", got, "hello world\n")
	}
}

func TestRendererDiagnosticsOnlyVerbose(t *testing.T) {
	var out, errBuf bytes.Buffer

	// Not verbose: should not print
	r := Renderer{Out: &out, Err: &errBuf, Verbose: false}
	r.Diagnostics("debug info %d", 42)
	if out.Len() > 0 || errBuf.Len() > 0 {
		t.Error("Diagnostics should not write when Verbose=false")
	}

	// Verbose: should print to stderr
	r2 := Renderer{Out: &out, Err: &errBuf, Verbose: true}
	r2.Diagnostics("debug info %d", 42)
	if !strings.Contains(errBuf.String(), "debug info 42") {
		t.Errorf("Diagnostics output = %q, should contain 'debug info 42'", errBuf.String())
	}
}

func TestCompactJSONRemovesEmpty(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{
		Out:     &out,
		JSON:    true,
		Compact: true,
	}
	started := time.Now()
	meta := RuntimeMetaInput{Command: "test", StartedAt: started}

	if err := r.Success(meta, nil, nil); err != nil {
		t.Fatalf("Success: %v", err)
	}

	raw := out.String()
	// Empty data should be omitted
	if strings.Contains(raw, `"data"`) {
		t.Errorf("compact JSON should omit empty data, got %q", raw)
	}
	// Empty warnings should be omitted
	if strings.Contains(raw, `"warnings"`) {
		t.Errorf("compact JSON should omit empty warnings, got %q", raw)
	}
}

func TestCompactJSONRemovesEmptyStrings(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{
		Out:     &out,
		JSON:    true,
		Compact: true,
	}
	// Only set Command, leave Profile and RequestID empty
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	if err := r.Success(meta, nil, nil); err != nil {
		t.Fatalf("Success: %v", err)
	}

	raw := out.String()
	// Empty strings should be omitted
	if strings.Contains(raw, `"profile"`) {
		t.Errorf("compact JSON should omit empty profile string, got %q", raw)
	}
	if strings.Contains(raw, `"request_id"`) {
		t.Errorf("compact JSON should omit empty request_id string, got %q", raw)
	}
}

func TestCompactJSONRemovesEmptyMap(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:     &out,
		Err:     &errBuf,
		JSON:    true,
		Compact: true,
	}
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}
	errBody := Errorf(model.ErrConfig, "bad config")

	r.Failure(meta, errBody)

	raw := out.String()
	if strings.Contains(raw, `"details"`) {
		t.Errorf("compact JSON should omit nil details, got %q", raw)
	}
}

func TestCompactJSONWithDetails(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:     &out,
		Err:     &errBuf,
		JSON:    true,
		Compact: true,
	}
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}
	errBody := ErrorWithDetails(model.ErrValidationFailed, "bad", map[string]any{"field": "title"})

	r.Failure(meta, errBody)

	raw := out.String()
	if !strings.Contains(raw, `"details"`) {
		t.Errorf("compact JSON should keep non-empty details, got %q", raw)
	}
}

func TestFullJSONKeepsAllFields(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{
		Out:  &out,
		JSON: true,
		Full: true,
	}
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	if err := r.Success(meta, nil, nil); err != nil {
		t.Fatalf("Success: %v", err)
	}

	raw := out.String()
	// Full mode should include empty fields
	if !strings.Contains(raw, `"data"`) {
		t.Errorf("full JSON should include data field, got %q", raw)
	}
}

func TestPrettyJSON(t *testing.T) {
	var out bytes.Buffer
	r := Renderer{
		Out:    &out,
		JSON:   true,
		Pretty: true,
	}
	meta := RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	if err := r.Success(meta, map[string]string{"id": "1"}, nil); err != nil {
		t.Fatalf("Success: %v", err)
	}

	raw := out.String()
	// Pretty JSON should have newlines/indentation
	if !strings.Contains(raw, "\n") {
		t.Errorf("pretty JSON should contain newlines, got %q", raw)
	}
}
