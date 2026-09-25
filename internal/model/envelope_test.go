package model

import (
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{ErrValidationFailed, 1},
		{ErrAuthRequired, 2},
		{ErrAuthFailed, 2},
		{ErrZenodoAPI, 3},
		{ErrNetwork, 4},
		{ErrPartialSuccess, 5},
		{ErrReadOnlyViolation, 6},
		{ErrConfirmationRequired, 6},
		{ErrFilesystem, 7},
		{ErrConfig, 8},
		{ErrInterrupted, 130},
		{ErrResourceNotFound, 3},
		{"unknown_code", 1},
	}
	for _, tt := range tests {
		got := ExitCode(tt.code)
		if got != tt.want {
			t.Errorf("ExitCode(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}
