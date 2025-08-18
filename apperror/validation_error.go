package apperror

import (
	"fmt"
	"net/http"
)

type ValidationError struct{ *BaseAppError }

func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{&BaseAppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Invalid %s: %s", field, msg),
		Context: field,
	}}
}
