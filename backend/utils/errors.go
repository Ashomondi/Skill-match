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
	CategoryInternal   ErrorCategory = "internal"  
)

type AppError struct {
	Category   ErrorCategory
	UserMsg    string
	StatusCode int
	Err        error 
	Context    map[string]string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Category, e.UserMsg, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewStorageError(err error, context map[string]string) *AppError {
	return &AppError{
		Category:   CategoryStorage,
		UserMsg:    "We couldn't upload your file right now. Please try again shortly.",
		StatusCode: http.StatusServiceUnavailable,
		Err:        err,
		Context:    context,
	}
}

func NewDatabaseError(err error, context map[string]string) *AppError {
	return &AppError{
		Category:   CategoryDatabase,
		UserMsg:    "Something went wrong on our end. Please try again.",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		Context:    context,
	}
}

func NewValidationError(userMsg string, context map[string]string) *AppError {
	return &AppError{
		Category:   CategoryValidation,
		UserMsg:    userMsg,
		StatusCode: http.StatusBadRequest,
		Err:        errors.New(userMsg),
		Context:    context,
	}
}


func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}