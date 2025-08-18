package apperror

import (
	"fmt"
	"net/http"
)

type NotFoundError struct{ *BaseAppError }

func NewNotFoundError(resource string, id int) *NotFoundError {
	return &NotFoundError{&BaseAppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s with ID %d not found", resource, id),
		Context: resource,
	}}
}
