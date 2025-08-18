package apperror

import "net/http"

type ContactDetailError struct{ *BaseAppError }

func NewContactDetailError(context, message string) *ContactDetailError {
	return &ContactDetailError{&BaseAppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Context: context,
	}}
}
