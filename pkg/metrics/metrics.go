// Package metrics provides unified metrics collection and monitoring for all Discord bots.
package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sawyer/go-discord-bots/pkg/logging"
)

// MetricsCollector provides centralized metrics collection.
type MetricsCollector struct {
	// Command metrics
	commandsExecuted  int64
	commandsSucceeded int64
	commandsFailed    int64

	// API metrics
	apiRequestsTotal   int64
	apiRequestsSuccess int64
	apiRequestsFailed  int64
	apiResponseTime    int64 // in milliseconds

	// Cache metrics
	cacheHits   int64
	cacheMisses int64
	cacheWrites int64

	// Performance metrics
	averageResponseTime int64 // in milliseconds

	// Bot-specific metrics
	botSpecific sync.Map // map[string]interface{} for custom bot metrics

	// Configuration
	botName   string
	botType   string
	startTime time.Time

	// Channels for real-time metrics
	commandChan chan CommandMetric
	apiChan     chan APIMetric
	cacheChan   chan CacheMetric
	perfChan    chan PerformanceMetric

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// CommandMetric represents a command execution metric.
type CommandMetric struct {
	Command   string
	UserID    string
	Success   bool
	Duration  time.Duration
	Timestamp time.Time
}

// APIMetric represents an API request metric.
type APIMetric struct {
	Service   string
	Endpoint  string
	Success   bool
	Duration  time.Duration
	Timestamp time.Time
}

// CacheMetric represents a cache operation metric.
type CacheMetric struct {
	Operation string
	Key       string
	Hit       bool
	Duration  time.Duration
	Timestamp time.Time
}

// PerformanceMetric represents a performance measurement.
type PerformanceMetric struct {
	Component string
	Metric    string
	Value     float64
	Unit      string
	Timestamp time.Time
}

// Global metrics collector instance
var globalCollector *MetricsCollector
var once sync.Once

// Initialize initializes the global metrics collector.
func Initialize(botName, botType string) {
	once.Do(func() {
		globalCollector = NewMetricsCollector(botName, botType)
		globalCollector.Start()
	})
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector(botName, botType string) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsCollector{
		botName:     botName,
		botType:     botType,
		startTime:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
		commandChan: make(chan CommandMetric, 100),
		apiChan:     make(chan APIMetric, 100),
		cacheChan:   make(chan CacheMetric, 100),
		perfChan:    make(chan PerformanceMetric, 100),
	}
}

// Start starts the metrics collection background processes.
func (m *MetricsCollector) Start() {
	m.wg.Add(1)
	go m.processMetrics()

	// Start periodic metrics collection
	m.wg.Add(1)
	go m.collectPeriodicMetrics()
}

// Stop stops the metrics collection.
func (m *MetricsCollector) Stop() {
	m.cancel()
	close(m.commandChan)
	close(m.apiChan)
	close(m.cacheChan)
	close(m.perfChan)
	m.wg.Wait()
}

// processMetrics processes incoming metrics in real-time.
func (m *MetricsCollector) processMetrics() {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case metric := <-m.commandChan:
			m.processCommandMetric(metric)
		case metric := <-m.apiChan:
			m.processAPIMetric(metric)
		case metric := <-m.cacheChan:
			m.processCacheMetric(metric)
		case metric := <-m.perfChan:
			m.processPerformanceMetric(metric)
		case <-ticker.C:
			m.logAggregatedMetrics()
		}
	}
}

// collectPeriodicMetrics collects system metrics periodically.
func (m *MetricsCollector) collectPeriodicMetrics() {
	defer m.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectSystemMetrics()
		}
	}
}

// processCommandMetric processes a command execution metric.
func (m *MetricsCollector) processCommandMetric(metric CommandMetric) {
	atomic.AddInt64(&m.commandsExecuted, 1)

	if metric.Success {
		atomic.AddInt64(&m.commandsSucceeded, 1)
	} else {
		atomic.AddInt64(&m.commandsFailed, 1)
	}

	// Update average response time
	duration := metric.Duration.Milliseconds()
	atomic.StoreInt64(&m.averageResponseTime, duration)

	logging.LogPerformanceMetric("commands", "execution", duration, "ms")
}

// processAPIMetric processes an API request metric.
func (m *MetricsCollector) processAPIMetric(metric APIMetric) {
	atomic.AddInt64(&m.apiRequestsTotal, 1)

	if metric.Success {
		atomic.AddInt64(&m.apiRequestsSuccess, 1)
	} else {
		atomic.AddInt64(&m.apiRequestsFailed, 1)
	}

	// Update API response time
	duration := metric.Duration.Milliseconds()
	atomic.StoreInt64(&m.apiResponseTime, duration)

	logging.LogPerformanceMetric("api", metric.Service, duration, "ms")
}

// processCacheMetric processes a cache operation metric.
func (m *MetricsCollector) processCacheMetric(metric CacheMetric) {
	if metric.Hit {
		atomic.AddInt64(&m.cacheHits, 1)
	} else {
		atomic.AddInt64(&m.cacheMisses, 1)
	}

	if metric.Operation == "set" || metric.Operation == "write" {
		atomic.AddInt64(&m.cacheWrites, 1)
	}

	logging.LogCacheOperation(metric.Operation, metric.Key, metric.Hit, metric.Duration)
}

// processPerformanceMetric processes a general performance metric.
func (m *MetricsCollector) processPerformanceMetric(metric PerformanceMetric) {
	logging.LogPerformanceMetric(metric.Component, metric.Metric, metric.Value, metric.Unit)

	// Store bot-specific metrics
	key := metric.Component + "." + metric.Metric
	m.botSpecific.Store(key, metric.Value)
}

// collectSystemMetrics collects system-level metrics.
func (m *MetricsCollector) collectSystemMetrics() {
	// This would typically collect runtime metrics, memory usage, etc.
	// For now, we'll log basic runtime information
	uptime := time.Since(m.startTime)

	logging.LogPerformanceMetric("system", "uptime", uptime.Seconds(), "seconds")
	logging.LogPerformanceMetric("system", "commands_total", float64(atomic.LoadInt64(&m.commandsExecuted)), "count")
	logging.LogPerformanceMetric("system", "api_requests_total", float64(atomic.LoadInt64(&m.apiRequestsTotal)), "count")
}

// logAggregatedMetrics logs aggregated metrics summary.
func (m *MetricsCollector) logAggregatedMetrics() {
	commandsTotal := atomic.LoadInt64(&m.commandsExecuted)
	commandsSuccess := atomic.LoadInt64(&m.commandsSucceeded)
	_ = atomic.LoadInt64(&m.commandsFailed) // Read but don't use to avoid compiler error

	cacheHits := atomic.LoadInt64(&m.cacheHits)
	cacheMisses := atomic.LoadInt64(&m.cacheMisses)

	var successRate float64
	if commandsTotal > 0 {
		successRate = float64(commandsSuccess) / float64(commandsTotal) * 100
	}

	var cacheHitRate float64
	totalCacheOps := cacheHits + cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCacheOps) * 100
	}

	logger := logging.WithBot(m.botName, m.botType)
	logger.Info("Metrics summary",
		"commands_total", commandsTotal,
		"commands_success_rate", successRate,
		"cache_hit_rate", cacheHitRate,
		"avg_response_time_ms", atomic.LoadInt64(&m.averageResponseTime),
		"uptime_seconds", time.Since(m.startTime).Seconds(),
	)
}

// Public API functions that use the global collector

// RecordCommand records a command execution.
func RecordCommand(command, userID string, success bool, duration time.Duration) {
	if globalCollector == nil {
		return
	}

	select {
	case globalCollector.commandChan <- CommandMetric{
		Command:   command,
		UserID:    userID,
		Success:   success,
		Duration:  duration,
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, drop the metric
	}
}

// RecordAPIRequest records an API request.
func RecordAPIRequest(service, endpoint string, success bool, duration time.Duration) {
	if globalCollector == nil {
		return
	}

	select {
	case globalCollector.apiChan <- APIMetric{
		Service:   service,
		Endpoint:  endpoint,
		Success:   success,
		Duration:  duration,
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, drop the metric
	}
}

// RecordCacheOperation records a cache operation.
func RecordCacheOperation(operation, key string, hit bool, duration time.Duration) {
	if globalCollector == nil {
		return
	}

	select {
	case globalCollector.cacheChan <- CacheMetric{
		Operation: operation,
		Key:       key,
		Hit:       hit,
		Duration:  duration,
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, drop the metric
	}
}

// RecordPerformanceMetric records a general performance metric.
func RecordPerformanceMetric(component, metric string, value float64, unit string) {
	if globalCollector == nil {
		return
	}

	select {
	case globalCollector.perfChan <- PerformanceMetric{
		Component: component,
		Metric:    metric,
		Value:     value,
		Unit:      unit,
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, drop the metric
	}
}

// GetMetricsSummary returns a summary of current metrics.
func GetMetricsSummary() map[string]interface{} {
	if globalCollector == nil {
		return make(map[string]interface{})
	}

	commandsTotal := atomic.LoadInt64(&globalCollector.commandsExecuted)
	commandsSuccess := atomic.LoadInt64(&globalCollector.commandsSucceeded)

	var successRate float64
	if commandsTotal > 0 {
		successRate = float64(commandsSuccess) / float64(commandsTotal) * 100
	}

	cacheHits := atomic.LoadInt64(&globalCollector.cacheHits)
	cacheMisses := atomic.LoadInt64(&globalCollector.cacheMisses)

	var cacheHitRate float64
	totalCacheOps := cacheHits + cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCacheOps) * 100
	}

	return map[string]interface{}{
		"bot_name":              globalCollector.botName,
		"bot_type":              globalCollector.botType,
		"uptime_seconds":        time.Since(globalCollector.startTime).Seconds(),
		"commands_total":        commandsTotal,
		"commands_success_rate": successRate,
		"api_requests_total":    atomic.LoadInt64(&globalCollector.apiRequestsTotal),
		"cache_hit_rate":        cacheHitRate,
		"avg_response_time_ms":  atomic.LoadInt64(&globalCollector.averageResponseTime),
	}
}

// Shutdown gracefully shuts down the metrics collector.
func Shutdown() {
	if globalCollector != nil {
		globalCollector.Stop()
	}
}
