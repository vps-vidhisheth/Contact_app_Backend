package web

import (
	"Contact_App/apperror"
	"encoding/json"
	"io"
	"net/http"
)

type CreateDetailRequest struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	IsActive bool   `json:"is_active"`
}

type UpdateDetailRequest struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

func UnmarshalJSON(r *http.Request, out interface{}) error {
	if r.Body == nil {
		return apperror.NewValidationError("body", "request body is empty")
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperror.NewValidationError("body", "unable to read request body")
	}

	if len(body) == 0 {
		return apperror.NewValidationError("body", "request body is empty")
	}

	if err := json.Unmarshal(body, out); err != nil {
		return apperror.NewValidationError("body", "invalid JSON format")
	}

	return nil
}
