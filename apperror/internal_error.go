package apperror

import (
	"encoding/json"
	"net/http"
)

type InternalError struct{ *BaseAppError }

func NewInternalError(message string) *InternalError {
	return &InternalError{&BaseAppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Context: "internal",
	}}
}

func RespondWithAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*BaseAppError); ok {
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   appErr.Message,
			"context": appErr.Context,
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "An unexpected error occurred",
			"context": "internal",
		})
	}
}
