package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// TemperatureMetrics renders temperature metrics
type TemperatureMetrics struct {
	label    lipgloss.Style
	value    lipgloss.Style
	muted    lipgloss.Style
	normal   lipgloss.Style
	warning  lipgloss.Style
	critical lipgloss.Style
	width    int
}

// NewTemperatureMetrics creates a new temperature metrics renderer
func NewTemperatureMetrics() *TemperatureMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &TemperatureMetrics{
		label:    lipgloss.NewStyle().Foreground(colorCyan),
		value:    lipgloss.NewStyle().Foreground(colorForeground),
		muted:    lipgloss.NewStyle().Foreground(colorComment),
		normal:   lipgloss.NewStyle().Foreground(colorGreen),
		warning:  lipgloss.NewStyle().Foreground(colorOrange),
		critical: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	}
}

// SetWidth sets the render width
func (t *TemperatureMetrics) SetWidth(w int) {
	t.width = w
}

// Render returns the rendered temperature metrics
func (t *TemperatureMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Sensors == nil {
		return t.muted.Render("Loading temperature data...")
	}

	sensors := systemData.Sensors
	var content string

	// Title
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Temperatures")
	content += "\n\n"

	if len(sensors.Temperatures) == 0 {
		return t.muted.Render("No temperature sensors found")
	}

	// Group temperatures by sensor type
	tempGroups := make(map[string][]TempEntry)
	for _, temp := range sensors.Temperatures {
		sensorType := extractSensorType(temp.SensorKey)
		tempGroups[sensorType] = append(tempGroups[sensorType], TempEntry{
			Key:      temp.SensorKey,
			Temp:     temp.Temperature,
			Critical: temp.Critical,
		})
	}

	// Display temperatures grouped by sensor type (limit to first few)
	count := 0
	maxTemps := 20 // Limit display
	for sensorType, temps := range tempGroups {
		if count >= maxTemps {
			break
		}

		content += t.label.Render(sensorType)
		content += "\n"

		for _, temp := range temps {
			if count >= maxTemps {
				break
			}
			count++

			tempStyle := t.getMetricStyle(temp.Temp, 70, 85)
			content += fmt.Sprintf("  %s: %s%.1f°C%s",
				temp.Key,
				tempStyle,
				temp.Temp,
				t.value,
			)

			if temp.Critical != 0 {
				content += t.muted.Render(fmt.Sprintf(" (critical: %.0f°C)", temp.Critical))
			}
			content += "\n"
		}
		content += "\n"
	}

	if len(sensors.Temperatures) > maxTemps {
		content += t.muted.Render(fmt.Sprintf("... and %d more sensors\n", len(sensors.Temperatures)-maxTemps))
	}

	return content
}

// TempEntry holds temperature info for display
type TempEntry struct {
	Key      string
	Temp     float64
	Critical float64
}

// extractSensorType extracts the base sensor type from the sensor key
func extractSensorType(key string) string {
	for i, c := range key {
		if c == '_' || c == '-' {
			return key[:i]
		}
	}
	return key
}

func (t *TemperatureMetrics) getMetricStyle(value float64, warning, critical float64) lipgloss.Style {
	if value >= critical {
		return t.critical
	}
	if value >= warning {
		return t.warning
	}
	return t.normal
}
