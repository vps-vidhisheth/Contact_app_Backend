package apperror

import "fmt"

type AppError interface {
	error
	StatusCode() int
	MessageText() string
	ErrorContext() string
}

type BaseAppError struct {
	Code    int
	Message string
	Context string
}

func (e *BaseAppError) Error() string {
	return fmt.Sprintf("error [%d]: %s - %s", e.Code, e.Context, e.Message)
}

func (e *BaseAppError) StatusCode() int {
	return e.Code
}

func (e *BaseAppError) MessageText() string {
	return e.Message
}

func (e *BaseAppError) ErrorContext() string {
	return e.Context
}
