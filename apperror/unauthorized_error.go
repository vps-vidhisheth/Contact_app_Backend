package apperror

import "net/http"

type UnauthorizedError struct{ *BaseAppError }

func NewUnauthorized(context string) *UnauthorizedError {
	return &UnauthorizedError{&BaseAppError{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized access",
		Context: context,
	}}
}
