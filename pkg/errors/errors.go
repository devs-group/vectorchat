package errors

import (
	"github.com/pkg/errors"
)

// Application error types
var (
	// ErrNoDocumentsFound is returned when no documents are found for a chat ID
	ErrNoDocumentsFound = errors.New("no documents found")
	
	// ErrInvalidChatID is returned when an invalid chat ID is provided
	ErrInvalidChatID = errors.New("invalid chat ID")
	
	// ErrDocumentNotFound is returned when a specific document is not found
	ErrDocumentNotFound = errors.New("document not found")
	
	// ErrVectorizationFailed is returned when document vectorization fails
	ErrVectorizationFailed = errors.New("failed to vectorize document")
	
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// WithDetails adds context details to an error
func WithDetails(err error, details string) error {
	return errors.Wrap(err, details)
}

// Is checks if an error is of a specific type, including wrapped errors
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

// Cause returns the underlying cause of the error
func Cause(err error) error {
	return errors.Cause(err)
} 