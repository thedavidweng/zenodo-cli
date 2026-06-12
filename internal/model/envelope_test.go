package model

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEnvelopeJSON(t *testing.T) {
	env := Envelope{
		OK:   true,
		Data: map[string]string{"id": "123"},
		Meta: Meta{
			Command:       "deposit list",
			Profile:       "default",
			DurationMS:    42,
			SchemaVersion: SchemaVersion,
			RequestID:     "req-abc",
			Warnings:      []string{"deprecated field"},
		},
	}

	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !decoded.OK {
		t.Error("expected OK=true")
	}
	if decoded.Meta.SchemaVersion != SchemaVersion {
		t.Errorf("schema version = %q, want %q", decoded.Meta.SchemaVersion, SchemaVersion)
	}
	if decoded.Meta.Command != "deposit list" {
		t.Errorf("command = %q, want %q", decoded.Meta.Command, "deposit list")
	}
	if decoded.Meta.DurationMS != 42 {
		t.Errorf("duration = %d, want 42", decoded.Meta.DurationMS)
	}
	if len(decoded.Meta.Warnings) != 1 || decoded.Meta.Warnings[0] != "deprecated field" {
		t.Errorf("warnings = %v, want [deprecated field]", decoded.Meta.Warnings)
	}
	if decoded.Error != nil {
		t.Error("expected nil error")
	}
}

func TestEnvelopeWithError(t *testing.T) {
	env := Envelope{
		OK: false,
		Error: &ErrorBody{
			Code:      ErrValidationFailed,
			Message:   "missing title",
			Category:  "validation",
			Retryable: false,
			Details:   map[string]any{"field": "title"},
		},
		Meta: Meta{SchemaVersion: SchemaVersion},
	}

	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.OK {
		t.Error("expected OK=false")
	}
	if decoded.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if decoded.Error.Code != ErrValidationFailed {
		t.Errorf("code = %q, want %q", decoded.Error.Code, ErrValidationFailed)
	}
	if decoded.Error.Category != "validation" {
		t.Errorf("category = %q, want validation", decoded.Error.Category)
	}
	if decoded.Error.Retryable {
		t.Error("expected retryable=false")
	}
	if decoded.Error.Details["field"] != "title" {
		t.Errorf("details[field] = %v, want title", decoded.Error.Details["field"])
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []string{
		ErrValidationFailed,
		ErrAuthRequired,
		ErrAuthFailed,
		ErrZenodoAPI,
		ErrNetwork,
		ErrPartialSuccess,
		ErrReadOnlyViolation,
		ErrConfirmationRequired,
		ErrFilesystem,
		ErrConfig,
		ErrInterrupted,
		ErrResourceNotFound,
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if c == "" {
			t.Error("error code is empty string")
		}
		if seen[c] {
			t.Errorf("duplicate error code: %s", c)
		}
		seen[c] = true
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{ErrValidationFailed, 1},
		{ErrAuthRequired, 2},
		{ErrAuthFailed, 2},
		{ErrZenodoAPI, 1},
		{ErrNetwork, 1},
		{ErrPartialSuccess, 1},
		{ErrReadOnlyViolation, 3},
		{ErrConfirmationRequired, 3},
		{ErrFilesystem, 1},
		{ErrConfig, 1},
		{ErrInterrupted, 130},
		{ErrResourceNotFound, 1},
		{"unknown_code", 1},
	}
	for _, tt := range tests {
		got := ExitCode(tt.code)
		if got != tt.want {
			t.Errorf("ExitCode(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}

func TestCommandError(t *testing.T) {
	ce := &CommandError{Code: ErrAuthFailed, Message: "bad token"}
	if ce.Error() != "bad token" {
		t.Errorf("Error() = %q, want %q", ce.Error(), "bad token")
	}
	if ce.Code != ErrAuthFailed {
		t.Errorf("Code = %q, want %q", ce.Code, ErrAuthFailed)
	}
}

func TestCommandErrorAsError(t *testing.T) {
	var err error = &CommandError{Code: ErrNetwork, Message: "timeout"}
	if err.Error() != "timeout" {
		t.Errorf("Error() = %q, want timeout", err.Error())
	}
}

func TestCommandErrorIs(t *testing.T) {
	ce := &CommandError{Code: ErrAuthFailed, Message: "bad token"}
	if !errors.Is(ce, ce) {
		t.Error("CommandError should match itself with errors.Is")
	}
}
