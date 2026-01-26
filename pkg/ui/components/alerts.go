package components

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity int

const (
	Info AlertSeverity = iota
	Warning
	Critical
)

// Alert represents a single alert
type Alert struct {
	Severity    AlertSeverity
	Message     string
	Timestamp   time.Time
	TriggerTime time.Time
	Value       float64
	Threshold   float64
	Metric      string
}

// AlertManager manages active alerts
type AlertManager struct {
	mu           sync.RWMutex
	alerts       map[string]*Alert
	thresholds   map[string]ThresholdConfig
	history      []Alert
	maxHistory   int
	enabled      bool
}

// ThresholdConfig defines alert thresholds
type ThresholdConfig struct {
	Warning  float64
	Critical float64
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:     make(map[string]*Alert),
		thresholds: make(map[string]ThresholdConfig),
		history:    make([]Alert, 0, 100),
		maxHistory: 100,
		enabled:    true,
	}
}

// SetThreshold sets a threshold for a metric
func (a *AlertManager) SetThreshold(metric string, warning, critical float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.thresholds[metric] = ThresholdConfig{
		Warning:  warning,
		Critical: critical,
	}
}

// SetEnabled enables or disables alerting
func (a *AlertManager) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

// CheckValue checks a value against thresholds and generates alerts
func (a *AlertManager) CheckValue(metric string, value float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.enabled {
		return
	}

	threshold, ok := a.thresholds[metric]
	if !ok {
		return
	}

	key := metric
	severity := Info
	alertMsg := ""

	if value >= threshold.Critical {
		severity = Critical
		alertMsg = fmt.Sprintf("%s critical: %.1f%% (threshold: %.1f%%)", metric, value, threshold.Critical)
	} else if value >= threshold.Warning {
		severity = Warning
		alertMsg = fmt.Sprintf("%s warning: %.1f%% (threshold: %.1f%%)", metric, value, threshold.Warning)
	}

	if alertMsg != "" {
		// Check if we already have an alert for this metric
		if existing, ok := a.alerts[key]; !ok || existing.Severity != severity {
			alert := &Alert{
				Severity:    severity,
				Message:     alertMsg,
				Timestamp:   time.Now(),
				TriggerTime: time.Now(),
				Value:       value,
				Threshold:   threshold.Warning,
				Metric:      metric,
			}
			a.alerts[key] = alert
			a.history = append(a.history, *alert)

			// Trim history
			if len(a.history) > a.maxHistory {
				a.history = a.history[1:]
			}
		}
	} else {
		// Value returned to normal, clear the alert
		if _, ok := a.alerts[key]; ok && value < threshold.Warning {
			delete(a.alerts, key)
		}
	}
}

// GetActiveAlerts returns all active alerts
func (a *AlertManager) GetActiveAlerts() []Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alerts := make([]Alert, 0, len(a.alerts))
	for _, alert := range a.alerts {
		alerts = append(alerts, *alert)
	}
	return alerts
}

// GetHistory returns alert history
func (a *AlertManager) GetHistory() []Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	history := make([]Alert, len(a.history))
	copy(history, a.history)
	return history
}

// ClearAll clears all active alerts
func (a *AlertManager) ClearAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.alerts = make(map[string]*Alert)
}

// AlertBar displays active alerts
type AlertBar struct {
	manager      *AlertManager
	style        lipgloss.Style
	warningStyle lipgloss.Style
	criticalStyle lipgloss.Style
	width        int
	visible      bool
}

// NewAlertBar creates a new alert bar
func NewAlertBar(manager *AlertManager) *AlertBar {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &AlertBar{
		manager:      manager,
		style:        lipgloss.NewStyle().Foreground(colorForeground),
		warningStyle: lipgloss.NewStyle().Foreground(colorOrange).Bold(true),
		criticalStyle: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
		visible:      false,
	}
}

// SetWidth sets the width
func (a *AlertBar) SetWidth(w int) {
	a.width = w
}

// Show shows the alert bar
func (a *AlertBar) Show() {
	a.visible = true
}

// Hide hides the alert bar
func (a *AlertBar) Hide() {
	a.visible = false
}

// Render returns the rendered alert bar
func (a *AlertBar) Render() string {
	if !a.visible {
		return ""
	}

	alerts := a.manager.GetActiveAlerts()
	if len(alerts) == 0 {
		return ""
	}

	var b strings.Builder

	for i, alert := range alerts {
		if i > 0 {
			b.WriteString(" | ")
		}

		var style lipgloss.Style
		if alert.Severity == Critical {
			style = a.criticalStyle
		} else if alert.Severity == Warning {
			style = a.warningStyle
		} else {
			style = a.style
		}

		b.WriteString(style.Render(alert.Message))
	}

	return b.String()
}
