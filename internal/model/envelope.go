package model

// SchemaVersion is the current envelope schema version.
const SchemaVersion = "2026-06-11"

// Error codes.
const (
	ErrValidationFailed      = "VALIDATION_FAILED"
	ErrAuthRequired          = "AUTH_REQUIRED"
	ErrAuthFailed            = "AUTH_FAILED"
	ErrZenodoAPI             = "ZENODO_API_ERROR"
	ErrNetwork               = "NETWORK_ERROR"
	ErrPartialSuccess        = "PARTIAL_SUCCESS"
	ErrReadOnlyViolation     = "READ_ONLY_VIOLATION"
	ErrConfirmationRequired  = "CONFIRMATION_REQUIRED"
	ErrFilesystem            = "FILESYSTEM_ERROR"
	ErrConfig                = "CONFIG_ERROR"
	ErrInterrupted           = "INTERRUPTED"
	ErrResourceNotFound      = "RESOURCE_NOT_FOUND"
)

// Envelope is the standard JSON output wrapper for all commands.
type Envelope struct {
	OK    bool       `json:"ok"`
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  Meta       `json:"meta"`
}

// ErrorBody contains structured error information.
type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  string         `json:"category,omitempty"`
	Retryable bool           `json:"retryable,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// Meta contains request metadata.
type Meta struct {
	Command      string   `json:"command,omitempty"`
	Profile      string   `json:"profile,omitempty"`
	DurationMS   int64    `json:"duration_ms,omitempty"`
	SchemaVersion string  `json:"schema_version"`
	RequestID    string   `json:"request_id,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// ExitCode maps an error code to a process exit code.
func ExitCode(code string) int {
	switch code {
	case ErrAuthRequired, ErrAuthFailed:
		return 2
	case ErrReadOnlyViolation, ErrConfirmationRequired:
		return 3
	case ErrInterrupted:
		return 130
	default:
		return 1
	}
}

// CommandError is a simple error with a code and message.
type CommandError struct {
	Code    string
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}
