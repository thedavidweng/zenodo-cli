package output

import (
	"fmt"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

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

var retryableCodes = map[string]bool{
	model.ErrNetwork:   true,
	model.ErrZenodoAPI: true,
}

func Errorf(code, format string, args ...any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Category:  CategoryForCode(code),
		Retryable: retryableCodes[code],
	}
}

func ErrorWithDetails(code, message string, details map[string]any) model.ErrorBody {
	return model.ErrorBody{
		Code:      code,
		Message:   message,
		Category:  CategoryForCode(code),
		Retryable: retryableCodes[code],
		Details:   details,
	}
}
