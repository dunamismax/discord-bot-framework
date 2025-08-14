// Package errors provides error types and handling for the Discord bot framework.
package errors

import "fmt"

// ErrorType represents different types of errors that can occur.
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeDiscord    ErrorType = "discord"
	ErrorTypeAPI        ErrorType = "api"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeAudio      ErrorType = "audio"
	ErrorTypeNetwork    ErrorType = "network"
)

// BotError represents an error that occurred in the bot framework.
type BotError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface.
func (e *BotError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error.
func (e *BotError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *BotError {
	return &BotError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewDiscordError creates a new Discord API error.
func NewDiscordError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeDiscord,
		Message: message,
		Err:     err,
	}
}

// NewAPIError creates a new API error.
func NewAPIError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeAPI,
		Message: message,
		Err:     err,
	}
}

// NewDatabaseError creates a new database error.
func NewDatabaseError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string) *BotError {
	return &BotError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Err:     err,
	}
}

// NewPermissionError creates a new permission error.
func NewPermissionError(message string) *BotError {
	return &BotError{
		Type:    ErrorTypePermission,
		Message: message,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

// NewAudioError creates a new audio-related error.
func NewAudioError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeAudio,
		Message: message,
		Err:     err,
	}
}

// NewNetworkError creates a new network-related error.
func NewNetworkError(message string, err error) *BotError {
	return &BotError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Err:     err,
	}
}

// IsErrorType checks if an error is of a specific type.
func IsErrorType(err error, errorType ErrorType) bool {
	if botErr, ok := err.(*BotError); ok {
		return botErr.Type == errorType
	}
	return false
}
