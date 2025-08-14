// Package errors provides custom error types and utilities for the Clippy Bot.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of error that occurred.
type ErrorType string

const (
	// ErrorTypeDiscord represents a Discord-related error.
	ErrorTypeDiscord ErrorType = "discord_error"
	// ErrorTypeConfig represents a configuration error.
	ErrorTypeConfig ErrorType = "config_error"
	// ErrorTypeValidation represents a validation error.
	ErrorTypeValidation ErrorType = "validation_error"
	// ErrorTypeRateLimit represents a rate limit error.
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeNetwork represents a network error.
	ErrorTypeNetwork ErrorType = "network_error"
	// ErrorTypeInternal represents an internal error.
	ErrorTypeInternal ErrorType = "internal_error"
	// ErrorTypeResponse represents a response generation error.
	ErrorTypeResponse ErrorType = "response_error"
)

// ClippyError represents a categorized error with additional context.
type ClippyError struct {
	Type       ErrorType
	Message    string
	Cause      error
	StatusCode int
	Context    map[string]interface{}
}

func (e *ClippyError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *ClippyError) Unwrap() error {
	return e.Cause
}

// NewDiscordError creates a new Discord-related error.
func NewDiscordError(message string, cause error) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeDiscord,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string, cause error) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeConfig,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(message string, retryAfter int) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Context: map[string]interface{}{
			"retry_after": retryAfter,
		},
	}
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, cause error) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, cause error) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewResponseError creates a new response generation error.
func NewResponseError(message string, cause error) *ClippyError {
	return &ClippyError{
		Type:    ErrorTypeResponse,
		Message: message,
		Cause:   cause,
	}
}

// IsErrorType checks if an error is of a specific type.
func IsErrorType(err error, errorType ErrorType) bool {
	var clippyErr *ClippyError
	if errors.As(err, &clippyErr) {
		return clippyErr.Type == errorType
	}

	return false
}

// FromHTTPStatus creates an appropriate error based on HTTP status code.
func FromHTTPStatus(statusCode int, message string) *ClippyError {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return NewRateLimitError(message, 0)
	case statusCode >= 400 && statusCode < 500:
		return NewValidationError(message)
	case statusCode >= 500:
		return NewDiscordError(message, nil)
	default:
		return NewInternalError(message, nil)
	}
}