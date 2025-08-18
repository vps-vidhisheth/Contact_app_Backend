package web

import (
	"Contact_App/apperror"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type ContactDetailResponse struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	ContactID uint      `json:"contact_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error": "failed to encode JSON response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func RespondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(apperror.AppError); ok {
		RespondJSON(w, appErr.StatusCode(), map[string]string{"error": appErr.MessageText()})
	} else {
		RespondErrorMessage(w, http.StatusInternalServerError, "internal server error")
	}
}

func RespondErrorMessage(w http.ResponseWriter, code int, msg string) {
	RespondJSON(w, code, map[string]string{"error": msg})
}

func RespondJSONWithXTotalCount(w http.ResponseWriter, code int, count int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error": "failed to encode JSON response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SetNewHeader(w, "X-Total-Count", strconv.Itoa(count))
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func SetNewHeader(w http.ResponseWriter, headerName, value string) {
	exposedHeaders := w.Header().Get("Access-Control-Expose-Headers")
	if exposedHeaders != "" {
		exposedHeaders += ", "
	}
	exposedHeaders += headerName
	w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
	w.Header().Set(headerName, value)
}
