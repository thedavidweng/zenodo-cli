package output

import (
	"testing"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

func TestCategoryForCode(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{model.ErrAuthRequired, "auth"},
		{model.ErrAuthFailed, "auth"},
		{model.ErrReadOnlyViolation, "safety"},
		{model.ErrConfirmationRequired, "safety"},
		{model.ErrZenodoAPI, "api"},
		{model.ErrValidationFailed, "validation"},
		{model.ErrConfig, "config"},
		{model.ErrNetwork, "network"},
		{model.ErrFilesystem, "filesystem"},
		{model.ErrResourceNotFound, "not_found"},
		{model.ErrInterrupted, "interrupted"},
		{model.ErrPartialSuccess, "api"},
		{"UNKNOWN_CODE", ""},
	}
	for _, tt := range tests {
		got := CategoryForCode(tt.code)
		if got != tt.want {
			t.Errorf("CategoryForCode(%q) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestErrorf(t *testing.T) {
	body := Errorf(model.ErrValidationFailed, "field %q is required", "title")
	if body.Code != model.ErrValidationFailed {
		t.Errorf("Code = %q, want %q", body.Code, model.ErrValidationFailed)
	}
	if body.Message != `field "title" is required` {
		t.Errorf("Message = %q, want %q", body.Message, `field "title" is required`)
	}
	if body.Category != "validation" {
		t.Errorf("Category = %q, want %q", body.Category, "validation")
	}
	if body.Details != nil {
		t.Errorf("Details should be nil, got %v", body.Details)
	}
}

func TestErrorfRetryable(t *testing.T) {
	tests := []struct {
		code      string
		retryable bool
	}{
		{model.ErrNetwork, true},
		{model.ErrZenodoAPI, true},
		{model.ErrValidationFailed, false},
		{model.ErrAuthFailed, false},
	}
	for _, tt := range tests {
		body := Errorf(tt.code, "test")
		if body.Retryable != tt.retryable {
			t.Errorf("Errorf(%q).Retryable = %v, want %v", tt.code, body.Retryable, tt.retryable)
		}
	}
}

func TestErrorWithDetails(t *testing.T) {
	details := map[string]any{"field": "title", "reason": "empty"}
	body := ErrorWithDetails(model.ErrValidationFailed, "validation failed", details)
	if body.Code != model.ErrValidationFailed {
		t.Errorf("Code = %q, want %q", body.Code, model.ErrValidationFailed)
	}
	if body.Message != "validation failed" {
		t.Errorf("Message = %q, want %q", body.Message, "validation failed")
	}
	if body.Category != "validation" {
		t.Errorf("Category = %q, want %q", body.Category, "validation")
	}
	if body.Details["field"] != "title" {
		t.Errorf("Details[field] = %v, want title", body.Details["field"])
	}
	if body.Details["reason"] != "empty" {
		t.Errorf("Details[reason] = %v, want empty", body.Details["reason"])
	}
}

func TestErrorWithDetailsNil(t *testing.T) {
	body := ErrorWithDetails(model.ErrConfig, "bad config", nil)
	if body.Details != nil {
		t.Errorf("Details should be nil when nil passed, got %v", body.Details)
	}
}
