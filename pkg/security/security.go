// Package security provides security utilities and validation for Discord bots.
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sawyer/go-discord-bots/pkg/errors"
	"github.com/sawyer/go-discord-bots/pkg/logging"
)

// InputValidator provides comprehensive input validation.
type InputValidator struct {
	maxLength       int
	blockedPatterns []string
	rateLimiter     *RateLimiter
}

// NewInputValidator creates a new input validator.
func NewInputValidator(maxLength int) *InputValidator {
	return &InputValidator{
		maxLength: maxLength,
		blockedPatterns: []string{
			// XSS patterns
			"<script",
			"javascript:",
			"data:",
			"vbscript:",
			"onload=",
			"onerror=",
			"onclick=",
			"onmouseover=",

			// Code injection patterns
			"../",
			"..\\",
			"file://",
			"ftp://",

			// Discord mention abuse
			"@everyone",
			"@here",

			// SQL injection patterns (basic)
			"' OR ",
			"' AND ",
			"UNION SELECT",
			"DROP TABLE",
			"DELETE FROM",
			"INSERT INTO",
			"UPDATE SET",

			// Command injection
			"; rm ",
			"& rm ",
			"| rm ",
			"; del ",
			"& del ",
			"| del ",
			"$(rm",
			"`rm",

			// Path traversal
			"/etc/passwd",
			"/etc/shadow",
			"C:\\Windows\\System32",

			// Potential token patterns
			"Bot ",
			"Bearer ",
			"OAuth ",
		},
		rateLimiter: NewRateLimiter(10, time.Minute), // 10 requests per minute per user
	}
}

// ValidateInput performs comprehensive input validation.
func (iv *InputValidator) ValidateInput(input, userID string) error {
	// Rate limiting check
	if !iv.rateLimiter.Allow(userID) {
		logging.LogSecurityEvent("rate_limit_exceeded", userID, "input validation rate limit exceeded", "medium")
		return errors.NewRateLimitError("rate limit exceeded", 60)
	}

	// Basic length check
	if len(input) == 0 {
		return errors.NewValidationError("input cannot be empty")
	}

	if len(input) > iv.maxLength {
		logging.LogSecurityEvent("input_too_long", userID, "input exceeds maximum length", "low")
		return errors.NewValidationError("input too long")
	}

	// Check for suspicious patterns
	if err := iv.checkSuspiciousPatterns(input, userID); err != nil {
		return err
	}

	// Check for potential token exposure
	if err := iv.checkTokenPatterns(input, userID); err != nil {
		return err
	}

	// URL validation if input contains URLs
	if err := iv.validateURLs(input, userID); err != nil {
		return err
	}

	return nil
}

// checkSuspiciousPatterns checks for malicious patterns.
func (iv *InputValidator) checkSuspiciousPatterns(input, userID string) error {
	lower := strings.ToLower(input)

	for _, pattern := range iv.blockedPatterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			logging.LogSecurityEvent("suspicious_pattern_detected", userID,
				"blocked pattern: "+pattern, "high")
			return errors.NewSecurityError("suspicious content detected", nil)
		}
	}

	return nil
}

// checkTokenPatterns checks for potential token exposure.
func (iv *InputValidator) checkTokenPatterns(input, userID string) error {
	// Discord token patterns
	discordTokenPattern := regexp.MustCompile(`[MN][A-Za-z\d]{23}\.[\w-]{6}\.[\w-]{27}`)
	if discordTokenPattern.MatchString(input) {
		logging.LogSecurityEvent("potential_token_exposure", userID,
			"Discord token pattern detected", "critical")
		return errors.NewSecurityError("potential token exposure detected", nil)
	}

	// Generic secret patterns (long hex/base64 strings)
	secretPattern := regexp.MustCompile(`[a-fA-F0-9]{32,}|[A-Za-z0-9+/]{32,}={0,2}`)
	matches := secretPattern.FindAllString(input, -1)
	for _, match := range matches {
		if len(match) >= 32 {
			logging.LogSecurityEvent("potential_secret_exposure", userID,
				"long secret-like string detected", "high")
			return errors.NewSecurityError("potential secret exposure detected", nil)
		}
	}

	return nil
}

// validateURLs validates any URLs found in the input.
func (iv *InputValidator) validateURLs(input, userID string) error {
	// Find potential URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlPattern.FindAllString(input, -1)

	for _, urlStr := range urls {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			logging.LogSecurityEvent("malformed_url", userID,
				"malformed URL detected: "+urlStr, "medium")
			return errors.NewValidationError("invalid URL format")
		}

		// Check for suspicious domains/schemes
		if err := iv.validateURL(parsedURL, userID); err != nil {
			return err
		}
	}

	return nil
}

// validateURL validates a single URL for security issues.
func (iv *InputValidator) validateURL(parsedURL *url.URL, userID string) error {
	// Block dangerous schemes
	suspiciousSchemes := []string{"file", "ftp", "data", "javascript"}
	for _, scheme := range suspiciousSchemes {
		if strings.EqualFold(parsedURL.Scheme, scheme) {
			logging.LogSecurityEvent("dangerous_url_scheme", userID,
				"dangerous URL scheme: "+parsedURL.Scheme, "high")
			return errors.NewSecurityError("dangerous URL scheme detected", nil)
		}
	}

	// Block private/internal IPs
	if isPrivateIP(parsedURL.Hostname()) {
		logging.LogSecurityEvent("private_ip_access", userID,
			"attempt to access private IP: "+parsedURL.Hostname(), "high")
		return errors.NewSecurityError("access to private networks blocked", nil)
	}

	// Check for URL shorteners that could hide malicious links
	suspiciousDomains := []string{
		"bit.ly", "tinyurl.com", "t.co", "goo.gl", "ow.ly",
		"discord.gg", // Discord invites can be spam
	}

	for _, domain := range suspiciousDomains {
		if strings.Contains(strings.ToLower(parsedURL.Hostname()), domain) {
			logging.LogSecurityEvent("suspicious_domain", userID,
				"suspicious domain detected: "+parsedURL.Hostname(), "medium")
			// Don't block, just log for monitoring
			break
		}
	}

	return nil
}

// isPrivateIP checks if an IP address is in a private range.
func isPrivateIP(hostname string) bool {
	// Basic check for common private IP patterns
	privatePatterns := []string{
		"localhost",
		"127.",
		"192.168.",
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"0.0.0.0",
		"::1",
		"fe80:",
	}

	lower := strings.ToLower(hostname)
	for _, pattern := range privatePatterns {
		if strings.HasPrefix(lower, pattern) {
			return true
		}
	}

	return false
}

// RateLimiter provides simple rate limiting functionality.
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request should be allowed.
func (rl *RateLimiter) Allow(userID string) bool {
	now := time.Now()

	// Clean old requests
	if timestamps, exists := rl.requests[userID]; exists {
		var validTimestamps []time.Time
		for _, timestamp := range timestamps {
			if now.Sub(timestamp) <= rl.window {
				validTimestamps = append(validTimestamps, timestamp)
			}
		}
		rl.requests[userID] = validTimestamps
	}

	// Check if under limit
	currentRequests := len(rl.requests[userID])
	if currentRequests >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[userID] = append(rl.requests[userID], now)
	return true
}

// TokenValidator provides secure token validation and generation.
type TokenValidator struct {
	validTokens map[string]time.Time
}

// NewTokenValidator creates a new token validator.
func NewTokenValidator() *TokenValidator {
	return &TokenValidator{
		validTokens: make(map[string]time.Time),
	}
}

// GenerateSecureToken generates a cryptographically secure token.
func (tv *TokenValidator) GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.NewInternalError("failed to generate secure token", err)
	}

	token := hex.EncodeToString(bytes)
	tv.validTokens[token] = time.Now().Add(15 * time.Minute) // 15 minute expiry

	return token, nil
}

// ValidateToken validates a token securely.
func (tv *TokenValidator) ValidateToken(token string) bool {
	expiry, exists := tv.validTokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(tv.validTokens, token)
		return false
	}

	return true
}

// InvalidateToken securely invalidates a token.
func (tv *TokenValidator) InvalidateToken(token string) {
	delete(tv.validTokens, token)
}

// SecureCompare performs constant-time string comparison.
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SanitizeInput sanitizes input for safe display.
func SanitizeInput(input string) string {
	// Remove/escape dangerous characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")

	return input
}

// ValidateDiscordID validates Discord snowflake IDs.
func ValidateDiscordID(id string) error {
	if len(id) < 17 || len(id) > 19 {
		return errors.NewValidationError("invalid Discord ID format")
	}

	// Check if it's all digits
	for _, char := range id {
		if char < '0' || char > '9' {
			return errors.NewValidationError("Discord ID must be numeric")
		}
	}

	return nil
}

// CheckPermissions validates user permissions for sensitive operations.
func CheckPermissions(userID string, requiredPermissions []string) error {
	// This would integrate with Discord permissions or a custom permission system
	// For now, implement basic admin check

	adminUsers := []string{
		// These would come from configuration
		"123456789012345678", // Example admin ID
	}

	for _, adminID := range adminUsers {
		if SecureCompare(userID, adminID) {
			return nil // Admin has all permissions
		}
	}

	// For non-admins, would check specific permissions
	// This is a placeholder implementation
	return errors.NewPermissionError("insufficient permissions", nil)
}

// LogSecurityIncident logs a security incident with appropriate severity.
func LogSecurityIncident(incident, userID, details, severity string) {
	logging.LogSecurityEvent(incident, userID, details, severity)

	// For critical incidents, could trigger additional alerting
	if severity == "critical" {
		// Could send to external monitoring/alerting system
		logger := logging.WithComponent("security")
		logger.Error("CRITICAL SECURITY INCIDENT",
			"incident", incident,
			"user_id", userID,
			"details", details,
			"timestamp", time.Now().Format(time.RFC3339),
		)
	}
}
