// Package errors provides unified error handling for all Discord bot applications.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of error that occurred.
type ErrorType string

const (
	// ErrorTypeAPI represents an API-related error.
	ErrorTypeAPI ErrorType = "api_error"
	// ErrorTypeConfig represents a configuration error.
	ErrorTypeConfig ErrorType = "config_error"
	// ErrorTypeDiscord represents a Discord-related error.
	ErrorTypeDiscord ErrorType = "discord_error"
	// ErrorTypeValidation represents a validation error.
	ErrorTypeValidation ErrorType = "validation_error"
	// ErrorTypeNotFound represents a not found error.
	ErrorTypeNotFound ErrorType = "not_found_error"
	// ErrorTypeRateLimit represents a rate limit error.
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeNetwork represents a network error.
	ErrorTypeNetwork ErrorType = "network_error"
	// ErrorTypeInternal represents an internal error.
	ErrorTypeInternal ErrorType = "internal_error"
	// ErrorTypeCache represents a cache error.
	ErrorTypeCache ErrorType = "cache_error"
	// ErrorTypeSecurity represents a security-related error.
	ErrorTypeSecurity ErrorType = "security_error"
	// ErrorTypeDatabase represents a database error.
	ErrorTypeDatabase ErrorType = "database_error"
	// ErrorTypeAudio represents an audio processing error.
	ErrorTypeAudio ErrorType = "audio_error"
	// ErrorTypePermission represents a permission error.
	ErrorTypePermission ErrorType = "permission_error"
)

// BotError represents a categorized error with additional context.
type BotError struct {
	ErrorType  ErrorType
	Message    string
	Cause      error
	StatusCode int
	Context    map[string]interface{}
}

// Error implements the error interface.
func (e *BotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.ErrorType, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.ErrorType, e.Message)
}

// Unwrap implements the error unwrapping interface.
func (e *BotError) Unwrap() error {
	return e.Cause
}

// Type returns the error type for logging purposes.
func (e *BotError) Type() string {
	return string(e.ErrorType)
}

// Context returns the error context for logging purposes.
func (e *BotError) Context() map[string]interface{} {
	if e.Context == nil {
		return make(map[string]interface{})
	}
	return e.Context
}

// NewAPIError creates a new API-related error.
func NewAPIError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeAPI,
		Message:   message,
		Cause:     cause,
	}
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeConfig,
		Message:   message,
		Cause:     cause,
	}
}

// NewDiscordError creates a new Discord-related error.
func NewDiscordError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeDiscord,
		Message:   message,
		Cause:     cause,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *BotError {
	return &BotError{
		ErrorType: ErrorTypeValidation,
		Message:   message,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string) *BotError {
	return &BotError{
		ErrorType: ErrorTypeNotFound,
		Message:   message,
	}
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(message string, retryAfter int) *BotError {
	return &BotError{
		ErrorType: ErrorTypeRateLimit,
		Message:   message,
		Context: map[string]interface{}{
			"retry_after": retryAfter,
		},
	}
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeNetwork,
		Message:   message,
		Cause:     cause,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeInternal,
		Message:   message,
		Cause:     cause,
	}
}

// NewCacheError creates a new cache-related error.
func NewCacheError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeCache,
		Message:   message,
		Cause:     cause,
	}
}

// NewSecurityError creates a new security-related error.
func NewSecurityError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeSecurity,
		Message:   message,
		Cause:     cause,
	}
}

// NewDatabaseError creates a new database error.
func NewDatabaseError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeDatabase,
		Message:   message,
		Cause:     cause,
	}
}

// NewAudioError creates a new audio processing error.
func NewAudioError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypeAudio,
		Message:   message,
		Cause:     cause,
	}
}

// NewPermissionError creates a new permission error.
func NewPermissionError(message string, cause error) *BotError {
	return &BotError{
		ErrorType: ErrorTypePermission,
		Message:   message,
		Cause:     cause,
	}
}

// IsErrorType checks if an error is of a specific type.
func IsErrorType(err error, errorType ErrorType) bool {
	var botErr *BotError
	if errors.As(err, &botErr) {
		return botErr.ErrorType == errorType
	}
	return false
}

// FromHTTPStatus creates an appropriate error based on HTTP status code.
func FromHTTPStatus(statusCode int, message string) *BotError {
	botErr := &BotError{
		Message:    message,
		StatusCode: statusCode,
	}

	switch {
	case statusCode == http.StatusNotFound:
		botErr.ErrorType = ErrorTypeNotFound
	case statusCode == http.StatusTooManyRequests:
		botErr.ErrorType = ErrorTypeRateLimit
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		botErr.ErrorType = ErrorTypePermission
	case statusCode >= 400 && statusCode < 500:
		botErr.ErrorType = ErrorTypeValidation
	case statusCode >= 500:
		botErr.ErrorType = ErrorTypeAPI
	default:
		botErr.ErrorType = ErrorTypeInternal
	}

	return botErr
}

// WithContext adds context to an existing error.
func WithContext(err error, key string, value interface{}) error {
	var botErr *BotError
	if !errors.As(err, &botErr) {
		// Convert regular error to BotError
		botErr = NewInternalError(err.Error(), err)
	}

	if botErr.Context == nil {
		botErr.Context = make(map[string]interface{})
	}
	botErr.Context[key] = value

	return botErr
}

// WithContextMap adds multiple context values to an existing error.
func WithContextMap(err error, context map[string]interface{}) error {
	var botErr *BotError
	if !errors.As(err, &botErr) {
		// Convert regular error to BotError
		botErr = NewInternalError(err.Error(), err)
	}

	if botErr.Context == nil {
		botErr.Context = make(map[string]interface{})
	}

	for key, value := range context {
		botErr.Context[key] = value
	}

	return botErr
}

// IsRetryable determines if an error indicates a retryable condition.
func IsRetryable(err error) bool {
	var botErr *BotError
	if errors.As(err, &botErr) {
		switch botErr.ErrorType {
		case ErrorTypeNetwork, ErrorTypeAPI, ErrorTypeRateLimit:
			return true
		case ErrorTypeInternal:
			// Some internal errors might be retryable
			return botErr.StatusCode == 0 || botErr.StatusCode >= 500
		default:
			return false
		}
	}
	return false
}

// GetSeverity returns the severity level of an error.
func GetSeverity(err error) string {
	var botErr *BotError
	if !errors.As(err, &botErr) {
		return "medium"
	}

	switch botErr.ErrorType {
	case ErrorTypeSecurity, ErrorTypePermission:
		return "high"
	case ErrorTypeAPI, ErrorTypeDatabase, ErrorTypeInternal:
		return "medium"
	case ErrorTypeNetwork, ErrorTypeRateLimit, ErrorTypeCache:
		return "low"
	case ErrorTypeValidation, ErrorTypeNotFound:
		return "info"
	default:
		return "medium"
	}
}