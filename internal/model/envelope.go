package model

const SchemaVersion = "2026-06-11"

const (
	ErrValidationFailed     = "VALIDATION_FAILED"
	ErrAuthRequired         = "AUTH_REQUIRED"
	ErrAuthFailed           = "AUTH_FAILED"
	ErrZenodoAPI            = "ZENODO_API_ERROR"
	ErrNetwork              = "NETWORK_ERROR"
	ErrPartialSuccess       = "PARTIAL_SUCCESS"
	ErrReadOnlyViolation    = "READ_ONLY_VIOLATION"
	ErrConfirmationRequired = "CONFIRMATION_REQUIRED"
	ErrFilesystem           = "FILESYSTEM_ERROR"
	ErrConfig               = "CONFIG_ERROR"
	ErrInterrupted          = "INTERRUPTED"
	ErrResourceNotFound     = "RESOURCE_NOT_FOUND"
)

type Envelope struct {
	OK    bool       `json:"ok"`
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  Meta       `json:"meta"`
}

type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  string         `json:"category,omitempty"`
	Retryable bool           `json:"retryable,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

type Meta struct {
	Command       string   `json:"command,omitempty"`
	Profile       string   `json:"profile,omitempty"`
	DurationMS    int64    `json:"duration_ms"`
	SchemaVersion string   `json:"schema_version"`
	RequestID     string   `json:"request_id,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

func ExitCode(code string) int {
	switch code {
	case ErrValidationFailed:
		return 1
	case ErrAuthRequired, ErrAuthFailed:
		return 2
	case ErrZenodoAPI, ErrResourceNotFound:
		return 3
	case ErrNetwork:
		return 4
	case ErrPartialSuccess:
		return 5
	case ErrReadOnlyViolation, ErrConfirmationRequired:
		return 6
	case ErrFilesystem:
		return 7
	case ErrConfig:
		return 8
	case ErrInterrupted:
		return 130
	default:
		return 1
	}
}

type CommandError struct {
	Code    string
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}
