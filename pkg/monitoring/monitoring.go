// Package monitoring provides performance monitoring and alerting capabilities.
package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/sawyer/discord-bot-framework/pkg/logging"
	"github.com/sawyer/discord-bot-framework/pkg/metrics"
)

// Monitor provides comprehensive monitoring capabilities.
type Monitor struct {
	alertManager *AlertManager
	healthCheck  *HealthChecker
	metricsExporter *MetricsExporter
	httpServer   *http.Server
	mu           sync.RWMutex
	isRunning    bool
}

// NewMonitor creates a new monitoring instance.
func NewMonitor(port int) *Monitor {
	mux := http.NewServeMux()
	
	monitor := &Monitor{
		alertManager:    NewAlertManager(),
		healthCheck:     NewHealthChecker(),
		metricsExporter: NewMetricsExporter(),
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	// Register HTTP endpoints
	mux.HandleFunc("/health", monitor.healthEndpoint)
	mux.HandleFunc("/metrics", monitor.metricsEndpoint)
	mux.HandleFunc("/status", monitor.statusEndpoint)
	mux.HandleFunc("/alerts", monitor.alertsEndpoint)

	return monitor
}

// Start starts the monitoring system.
func (m *Monitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("monitor is already running")
	}

	logger := logging.WithComponent("monitor")
	logger.Info("Starting monitoring system")

	// Start health checker
	m.healthCheck.Start()

	// Start alert manager
	m.alertManager.Start()

	// Start HTTP server
	go func() {
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	m.isRunning = true
	logger.Info("Monitoring system started", "port", m.httpServer.Addr)

	return nil
}

// Stop stops the monitoring system.
func (m *Monitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	logger := logging.WithComponent("monitor")
	logger.Info("Stopping monitoring system")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if err := m.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error stopping HTTP server", "error", err)
	}

	// Stop components
	m.healthCheck.Stop()
	m.alertManager.Stop()

	m.isRunning = false
	logger.Info("Monitoring system stopped")

	return nil
}

// healthEndpoint handles health check requests.
func (m *Monitor) healthEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	health := m.healthCheck.GetStatus()
	if health.Overall == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{"status": "%s", "checks": %d, "timestamp": "%s"}`,
		health.Overall, len(health.Checks), time.Now().Format(time.RFC3339))
}

// metricsEndpoint handles Prometheus metrics requests.
func (m *Monitor) metricsEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	
	metrics := m.metricsExporter.Export()
	fmt.Fprint(w, metrics)
}

// statusEndpoint provides detailed status information.
func (m *Monitor) statusEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := map[string]interface{}{
		"uptime":    time.Since(time.Now()).String(),
		"version":   "2.0.0",
		"go_version": runtime.Version(),
		"metrics":   metrics.GetMetricsSummary(),
		"health":    m.healthCheck.GetStatus(),
		"alerts":    m.alertManager.GetActiveAlerts(),
	}

	fmt.Fprintf(w, "%+v", status)
}

// alertsEndpoint handles alert webhook requests.
func (m *Monitor) alertsEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle incoming alert webhooks
		m.alertManager.HandleWebhook(w, r)
	} else {
		// Return active alerts
		alerts := m.alertManager.GetActiveAlerts()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%+v", alerts)
	}
}

// AlertManager handles alerts and notifications.
type AlertManager struct {
	alerts    map[string]*Alert
	rules     []*AlertRule
	notifiers []Notifier
	mu        sync.RWMutex
	ticker    *time.Ticker
	stopCh    chan struct{}
}

// Alert represents an active alert.
type Alert struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	LastSeen    time.Time `json:"last_seen"`
	Count       int       `json:"count"`
	Resolved    bool      `json:"resolved"`
	ResolvedTime *time.Time `json:"resolved_time,omitempty"`
}

// AlertRule defines conditions for triggering alerts.
type AlertRule struct {
	Name        string
	Condition   func() bool
	Severity    string
	Message     string
	Cooldown    time.Duration
	lastFired   time.Time
}

// Notifier interface for alert notifications.
type Notifier interface {
	Notify(alert *Alert) error
}

// NewAlertManager creates a new alert manager.
func NewAlertManager() *AlertManager {
	am := &AlertManager{
		alerts:    make(map[string]*Alert),
		rules:     make([]*AlertRule, 0),
		notifiers: make([]Notifier, 0),
		stopCh:    make(chan struct{}),
	}

	// Add default alert rules
	am.addDefaultRules()

	return am
}

// addDefaultRules adds default monitoring rules.
func (am *AlertManager) addDefaultRules() {
	// High memory usage alert
	am.rules = append(am.rules, &AlertRule{
		Name:     "HighMemoryUsage",
		Severity: "warning",
		Message:  "Memory usage is above 100MB",
		Cooldown: 5 * time.Minute,
		Condition: func() bool {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return m.Alloc > 100*1024*1024 // 100MB
		},
	})

	// High goroutine count alert
	am.rules = append(am.rules, &AlertRule{
		Name:     "HighGoroutineCount",
		Severity: "warning",
		Message:  "High number of goroutines detected",
		Cooldown: 5 * time.Minute,
		Condition: func() bool {
			return runtime.NumGoroutine() > 1000
		},
	})
}

// Start starts the alert manager.
func (am *AlertManager) Start() {
	am.ticker = time.NewTicker(30 * time.Second)
	go am.monitoringLoop()
}

// Stop stops the alert manager.
func (am *AlertManager) Stop() {
	if am.ticker != nil {
		am.ticker.Stop()
	}
	close(am.stopCh)
}

// monitoringLoop runs the main monitoring loop.
func (am *AlertManager) monitoringLoop() {
	logger := logging.WithComponent("alert-manager")
	
	for {
		select {
		case <-am.ticker.C:
			am.evaluateRules()
		case <-am.stopCh:
			logger.Info("Alert manager monitoring loop stopped")
			return
		}
	}
}

// evaluateRules evaluates all alert rules.
func (am *AlertManager) evaluateRules() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	
	for _, rule := range am.rules {
		// Check cooldown
		if now.Sub(rule.lastFired) < rule.Cooldown {
			continue
		}

		// Evaluate condition
		if rule.Condition() {
			rule.lastFired = now
			am.triggerAlert(rule)
		}
	}
}

// triggerAlert triggers an alert.
func (am *AlertManager) triggerAlert(rule *AlertRule) {
	alert, exists := am.alerts[rule.Name]
	if !exists {
		alert = &Alert{
			ID:        rule.Name,
			Name:      rule.Name,
			Severity:  rule.Severity,
			Message:   rule.Message,
			StartTime: time.Now(),
			LastSeen:  time.Now(),
			Count:     1,
		}
		am.alerts[rule.Name] = alert
	} else {
		alert.LastSeen = time.Now()
		alert.Count++
		alert.Resolved = false
		alert.ResolvedTime = nil
	}

	// Notify all notifiers
	for _, notifier := range am.notifiers {
		if err := notifier.Notify(alert); err != nil {
			logger := logging.WithComponent("alert-manager")
			logger.Error("Failed to send alert notification", "error", err, "alert", alert.Name)
		}
	}

	// Log the alert
	logging.LogSecurityEvent("alert_triggered", "", alert.Message, alert.Severity)
}

// GetActiveAlerts returns all active alerts.
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// HandleWebhook handles incoming alert webhooks.
func (am *AlertManager) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Placeholder for webhook handling (e.g., from Prometheus AlertManager)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// HealthChecker provides health checking capabilities.
type HealthChecker struct {
	checks   map[string]HealthCheck
	status   *HealthStatus
	mu       sync.RWMutex
	ticker   *time.Ticker
	stopCh   chan struct{}
}

// HealthCheck represents a health check function.
type HealthCheck func() error

// HealthStatus represents the overall health status.
type HealthStatus struct {
	Overall   string                    `json:"overall"`
	Checks    map[string]CheckResult    `json:"checks"`
	Timestamp time.Time                 `json:"timestamp"`
}

// CheckResult represents the result of a health check.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{
		checks: make(map[string]HealthCheck),
		status: &HealthStatus{
			Overall: "unknown",
			Checks:  make(map[string]CheckResult),
		},
		stopCh: make(chan struct{}),
	}

	// Add default health checks
	hc.addDefaultChecks()

	return hc
}

// addDefaultChecks adds default health checks.
func (hc *HealthChecker) addDefaultChecks() {
	// Memory check
	hc.checks["memory"] = func() error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.Alloc > 500*1024*1024 { // 500MB
			return fmt.Errorf("memory usage too high: %d bytes", m.Alloc)
		}
		return nil
	}

	// Goroutine check
	hc.checks["goroutines"] = func() error {
		count := runtime.NumGoroutine()
		if count > 2000 {
			return fmt.Errorf("too many goroutines: %d", count)
		}
		return nil
	}
}

// Start starts the health checker.
func (hc *HealthChecker) Start() {
	hc.ticker = time.NewTicker(10 * time.Second)
	go hc.checkLoop()
}

// Stop stops the health checker.
func (hc *HealthChecker) Stop() {
	if hc.ticker != nil {
		hc.ticker.Stop()
	}
	close(hc.stopCh)
}

// checkLoop runs the health check loop.
func (hc *HealthChecker) checkLoop() {
	for {
		select {
		case <-hc.ticker.C:
			hc.runChecks()
		case <-hc.stopCh:
			return
		}
	}
}

// runChecks runs all health checks.
func (hc *HealthChecker) runChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	results := make(map[string]CheckResult)
	overall := "healthy"

	for name, check := range hc.checks {
		if err := check(); err != nil {
			results[name] = CheckResult{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			overall = "unhealthy"
		} else {
			results[name] = CheckResult{
				Status: "healthy",
			}
		}
	}

	hc.status = &HealthStatus{
		Overall:   overall,
		Checks:    results,
		Timestamp: time.Now(),
	}
}

// GetStatus returns the current health status.
func (hc *HealthChecker) GetStatus() *HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

// MetricsExporter exports metrics in Prometheus format.
type MetricsExporter struct{}

// NewMetricsExporter creates a new metrics exporter.
func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{}
}

// Export exports metrics in Prometheus format.
func (me *MetricsExporter) Export() string {
	summary := metrics.GetMetricsSummary()
	
	output := ""
	
	// Export basic metrics
	if uptime, ok := summary["uptime_seconds"].(float64); ok {
		output += fmt.Sprintf("# HELP discord_uptime_seconds Bot uptime in seconds\n")
		output += fmt.Sprintf("# TYPE discord_uptime_seconds gauge\n")
		output += fmt.Sprintf("discord_uptime_seconds{bot_name=\"%s\",bot_type=\"%s\"} %f\n",
			summary["bot_name"], summary["bot_type"], uptime)
	}

	if commandsTotal, ok := summary["commands_total"].(float64); ok {
		output += fmt.Sprintf("# HELP discord_commands_total Total number of commands processed\n")
		output += fmt.Sprintf("# TYPE discord_commands_total counter\n")
		output += fmt.Sprintf("discord_commands_total{bot_name=\"%s\",bot_type=\"%s\"} %f\n",
			summary["bot_name"], summary["bot_type"], commandsTotal)
	}

	if successRate, ok := summary["commands_success_rate"].(float64); ok {
		output += fmt.Sprintf("# HELP discord_commands_success_rate Command success rate percentage\n")
		output += fmt.Sprintf("# TYPE discord_commands_success_rate gauge\n")
		output += fmt.Sprintf("discord_commands_success_rate{bot_name=\"%s\",bot_type=\"%s\"} %f\n",
			summary["bot_name"], summary["bot_type"], successRate)
	}

	// Add memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	output += fmt.Sprintf("# HELP discord_memory_usage_bytes Current memory usage in bytes\n")
	output += fmt.Sprintf("# TYPE discord_memory_usage_bytes gauge\n")
	output += fmt.Sprintf("discord_memory_usage_bytes{bot_name=\"%s\",bot_type=\"%s\"} %d\n",
		summary["bot_name"], summary["bot_type"], m.Alloc)

	// Add goroutine count
	output += fmt.Sprintf("# HELP discord_goroutines Current number of goroutines\n")
	output += fmt.Sprintf("# TYPE discord_goroutines gauge\n")
	output += fmt.Sprintf("discord_goroutines{bot_name=\"%s\",bot_type=\"%s\"} %d\n",
		summary["bot_name"], summary["bot_type"], runtime.NumGoroutine())

	return output
}

// LogNotifier sends alerts to structured logging.
type LogNotifier struct{}

// NewLogNotifier creates a new log notifier.
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Notify sends an alert to the logging system.
func (ln *LogNotifier) Notify(alert *Alert) error {
	logger := logging.WithComponent("alert-notifier")
	logger.Warn("Alert triggered",
		"alert_id", alert.ID,
		"alert_name", alert.Name,
		"severity", alert.Severity,
		"message", alert.Message,
		"count", alert.Count,
	)
	return nil
}