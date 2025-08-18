package apperror

import (
	"encoding/json"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error) {
	var appErr AppError

	switch e := err.(type) {
	case AppError:
		appErr = e
	default:
		appErr = &InternalError{&BaseAppError{
			Code:    http.StatusInternalServerError,
			Message: "Something went wrong",
			Context: "unexpected",
		}}
	}

	w.WriteHeader(appErr.StatusCode())
	json.NewEncoder(w).Encode(map[string]string{
		"message": appErr.MessageText(),
		"context": appErr.ErrorContext(),
	})
}

func HandleBadRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"message": msg,
		"context": "bad_request",
	})
}

func RespondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func HandleUnauthorized(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"message": msg,
		"context": "unauthorized",
	})
}
