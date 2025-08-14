// Package metrics provides metrics collection and reporting for the Clippy Bot.
package metrics

import (
	"sync"
	"time"

	"github.com/sawyer/go-discord-bots/apps/clippy/errors"
)

// Metrics holds the metrics for the bot.
type Metrics struct {
	startTime         time.Time
	commandsTotal     int64
	commandsSuccess   int64
	randomMessages    int64
	errorsByType      map[errors.ErrorType]int64
	responseTimeTotal time.Duration
	responseCount     int64
	mutex             sync.RWMutex
}

// defaultMetrics is the global metrics instance.
var defaultMetrics *Metrics

// Initialize initializes the global metrics instance.
func Initialize() {
	defaultMetrics = &Metrics{
		startTime:    time.Now(),
		errorsByType: make(map[errors.ErrorType]int64),
	}
}

// Get returns the global metrics instance.
func Get() *Metrics {
	return defaultMetrics
}

// RecordCommand records a command execution.
func RecordCommand(success bool) {
	if defaultMetrics == nil {
		return
	}

	defaultMetrics.mutex.Lock()
	defer defaultMetrics.mutex.Unlock()

	defaultMetrics.commandsTotal++
	if success {
		defaultMetrics.commandsSuccess++
	}
}

// RecordRandomMessage records a random message sent.
func RecordRandomMessage() {
	if defaultMetrics == nil {
		return
	}

	defaultMetrics.mutex.Lock()
	defer defaultMetrics.mutex.Unlock()

	defaultMetrics.randomMessages++
}

// RecordError records an error by type.
func RecordError(err error) {
	if defaultMetrics == nil {
		return
	}

	defaultMetrics.mutex.Lock()
	defer defaultMetrics.mutex.Unlock()

	var clippyErr *errors.ClippyError
	if errors.IsErrorType(err, errors.ErrorTypeDiscord) {
		defaultMetrics.errorsByType[errors.ErrorTypeDiscord]++
	} else if errors.IsErrorType(err, errors.ErrorTypeValidation) {
		defaultMetrics.errorsByType[errors.ErrorTypeValidation]++
	} else if errors.IsErrorType(err, errors.ErrorTypeRateLimit) {
		defaultMetrics.errorsByType[errors.ErrorTypeRateLimit]++
	} else if clippyErr != nil {
		defaultMetrics.errorsByType[clippyErr.Type]++
	} else {
		defaultMetrics.errorsByType[errors.ErrorTypeInternal]++
	}
}

// RecordResponseTime records a response time.
func RecordResponseTime(duration time.Duration) {
	if defaultMetrics == nil {
		return
	}

	defaultMetrics.mutex.Lock()
	defer defaultMetrics.mutex.Unlock()

	defaultMetrics.responseTimeTotal += duration
	defaultMetrics.responseCount++
}

// Summary represents a summary of metrics.
type Summary struct {
	UptimeSeconds         float64
	CommandsTotal         int64
	CommandsSuccessful    int64
	CommandSuccessRate    float64
	CommandsPerSecond     float64
	RandomMessages        int64
	RandomMessagesPerHour float64
	AverageResponseTime   float64
	ErrorsByType          map[errors.ErrorType]int64
}

// GetSummary returns a summary of all metrics.
func (m *Metrics) GetSummary() Summary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	uptime := time.Since(m.startTime)
	uptimeSeconds := uptime.Seconds()

	var commandSuccessRate float64
	if m.commandsTotal > 0 {
		commandSuccessRate = float64(m.commandsSuccess) / float64(m.commandsTotal) * 100
	}

	var commandsPerSecond float64
	if uptimeSeconds > 0 {
		commandsPerSecond = float64(m.commandsTotal) / uptimeSeconds
	}

	var randomMessagesPerHour float64
	if uptimeSeconds > 0 {
		randomMessagesPerHour = float64(m.randomMessages) / (uptimeSeconds / 3600)
	}

	var averageResponseTime float64
	if m.responseCount > 0 {
		averageResponseTime = float64(m.responseTimeTotal.Milliseconds()) / float64(m.responseCount)
	}

	// Copy error map to prevent race conditions
	errorsByType := make(map[errors.ErrorType]int64)
	for k, v := range m.errorsByType {
		errorsByType[k] = v
	}

	return Summary{
		UptimeSeconds:         uptimeSeconds,
		CommandsTotal:         m.commandsTotal,
		CommandsSuccessful:    m.commandsSuccess,
		CommandSuccessRate:    commandSuccessRate,
		CommandsPerSecond:     commandsPerSecond,
		RandomMessages:        m.randomMessages,
		RandomMessagesPerHour: randomMessagesPerHour,
		AverageResponseTime:   averageResponseTime,
		ErrorsByType:          errorsByType,
	}
}
