package apperror

import (
	"fmt"
	"net/http"
)

type AuthError struct{ *BaseAppError }
type LoginError struct{ *BaseAppError }

func NewAuthError(context string) *AuthError {
	return &AuthError{&BaseAppError{
		Code:    http.StatusForbidden,
		Message: "Unauthorized or forbidden operation",
		Context: context,
	}}
}

func NewLoginError(email string) *LoginError {
	return &LoginError{&BaseAppError{
		Code:    http.StatusUnauthorized,
		Message: fmt.Sprintf("Login failed for email: %s", email),
		Context: "authentication",
	}}
}
