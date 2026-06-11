package output

import (
	"fmt"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

// CategoryForCode maps an error code to a human-readable category.
func CategoryForCode(code string) string {
	switch code {
	case model.ErrAuthRequired, model.ErrAuthFailed:
		return "auth"
	case model.ErrReadOnlyViolation, model.ErrConfirmationRequired:
		return "safety"
	case model.ErrZenodoAPI, model.ErrPartialSuccess:
		return "api"
	case model.ErrValidationFailed:
		return "validation"
	case model.ErrConfig:
		return "config"
	case model.ErrNetwork:
		return "network"
	case model.ErrFilesystem:
		return "filesystem"
	case model.ErrResourceNotFound:
		return "not_found"
	case model.ErrInterrupted:
		return "interrupted"
	default:
		return ""
	}
}

// retryableCodes lists error codes that are retryable.
var retryableCodes = map[string]bool{
	model.ErrNetwork:    true,
	model.ErrZenodoAPI:  true,
}

// Errorf creates an ErrorBody with the given code and formatted message.
func Errorf(code, format string, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Category:  CategoryForCode(code),
		Retryable: retryableCodes[code],
	}
}

// ErrorWithDetails creates an ErrorBody with the given code, formatted message, and details.
func ErrorWithDetails(code, format string, details map[string]any, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Category:  CategoryForCode(code),
		Retryable: retryableCodes[code],
		Details:   details,
	}
}
