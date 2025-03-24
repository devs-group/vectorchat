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
	ErrVectorizationFailed = errors.New("failed to vectorize content")
	
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
	
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	
	// ErrAPIKeyNotFound is returned when an API key is not found
	ErrAPIKeyNotFound = errors.New("API key not found")
	
	// Add new errors
	ErrChatbotNotFound        = errors.New("chatbot not found")
	ErrInvalidChatbotParameters = errors.New("invalid chatbot parameters")
	ErrUnauthorizedChatbotAccess = errors.New("unauthorized access to chatbot")
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