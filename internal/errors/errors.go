package errors

import (
	"github.com/pkg/errors"
)

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

	// ErrInvalidAPIKey is returned when the api key is invalid
	ErrInvalidAPIKey = errors.New("API key is invalid")

	ErrChatbotNotFound           = errors.New("chatbot not found")
	ErrInvalidChatbotParameters  = errors.New("invalid chatbot parameters")
	ErrUnauthorizedChatbotAccess = errors.New("unauthorized access to chatbot")

	// Additional error definitions for repositories
	ErrUserAlreadyExists               = errors.New("user already exists")
	ErrAPIKeyAlreadyExists             = errors.New("API key already exists")
	ErrChatbotAlreadyExists            = errors.New("chatbot already exists")
	ErrFileNotFound                    = errors.New("file not found")
	ErrFileAlreadyExists               = errors.New("file already exists")
	ErrInvalidUserData                 = errors.New("invalid user data")
	ErrNotFound                        = errors.New("not found")
	ErrSharedKnowledgeBaseNotFound     = errors.New("shared knowledge base not found")
	ErrUnauthorizedKnowledgeBaseAccess = errors.New("unauthorized access to shared knowledge base")
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

// New creates a new error with the given message
func New(message string) error {
	return errors.New(message)
}
