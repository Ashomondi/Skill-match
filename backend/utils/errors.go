package utils

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCategory string

const (
	CategoryValidation ErrorCategory = "validation"
	CategoryStorage    ErrorCategory = "storage"
	CategoryDatabase   ErrorCategory = "database"
	CategoryAuth       ErrorCategory = "auth"
	CategoryNotFound   ErrorCategory = "not_found"
	CategoryConflict   ErrorCategory = "conflict"
	CategoryInternal   ErrorCategory = "internal"
	CategoryUpstream   ErrorCategory = "upstream"
)

// AppError separates the message returned to clients from the internal cause.
// Context must contain identifiers and operation names only, never credentials,
// tokens, request bodies, file contents, or presigned URLs.
type AppError struct {
	Category   ErrorCategory
	UserMsg    string
	StatusCode int
	Err        error
	Context    map[string]string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Category, e.UserMsg)
}

func (e *AppError) Unwrap() error { return e.Err }

func NewStorageError(err error, context map[string]string) *AppError {
	return newAppError(CategoryStorage, "File storage is temporarily unavailable. Please try again.", http.StatusServiceUnavailable, err, context)
}

func NewDatabaseError(err error, context map[string]string) *AppError {
	return newAppError(CategoryDatabase, "Something went wrong on our end. Please try again.", http.StatusInternalServerError, err, context)
}

func NewUpstreamError(userMsg string, err error, context map[string]string) *AppError {
	return newAppError(CategoryUpstream, userMsg, http.StatusServiceUnavailable, err, context)
}

func NewValidationError(userMsg string, context map[string]string) *AppError {
	return newAppError(CategoryValidation, userMsg, http.StatusBadRequest, errors.New(userMsg), context)
}

func NewAuthError(userMsg string, status int) *AppError {
	return newAppError(CategoryAuth, userMsg, status, errors.New(userMsg), nil)
}

func NewNotFoundError(userMsg string) *AppError {
	return newAppError(CategoryNotFound, userMsg, http.StatusNotFound, errors.New(userMsg), nil)
}

func NewConflictError(userMsg string) *AppError {
	return newAppError(CategoryConflict, userMsg, http.StatusConflict, errors.New(userMsg), nil)
}

func NewInternalError(err error, context map[string]string) *AppError {
	return newAppError(CategoryInternal, "Something went wrong on our end. Please try again.", http.StatusInternalServerError, err, context)
}

func newAppError(category ErrorCategory, userMsg string, status int, err error, context map[string]string) *AppError {
	return &AppError{Category: category, UserMsg: userMsg, StatusCode: status, Err: err, Context: context}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
