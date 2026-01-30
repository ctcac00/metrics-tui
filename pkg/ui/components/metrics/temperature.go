package metrics

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
)

// TemperatureMetrics renders temperature metrics
type TemperatureMetrics struct {
	label        lipgloss.Style
	value        lipgloss.Style
	muted        lipgloss.Style
	normal       lipgloss.Style
	warning      lipgloss.Style
	critical     lipgloss.Style
	width        int
	targetHeight int
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
		label:        lipgloss.NewStyle().Foreground(colorCyan),
		value:        lipgloss.NewStyle().Foreground(colorForeground),
		muted:        lipgloss.NewStyle().Foreground(colorComment),
		normal:       lipgloss.NewStyle().Foreground(colorGreen),
		warning:      lipgloss.NewStyle().Foreground(colorOrange),
		critical:     lipgloss.NewStyle().Foreground(colorRed).Bold(true),
		targetHeight: 0,
	}
}

// SetWidth sets the render width
func (t *TemperatureMetrics) SetWidth(w int) {
	t.width = w
}

// SetHeight sets the target height for padding
func (t *TemperatureMetrics) SetHeight(h int) {
	t.targetHeight = h
}

// Render returns the rendered temperature metrics
func (t *TemperatureMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Sensors == nil {
		result := t.muted.Render("Loading temperature data...")
		return t.padToHeight(result)
	}

	sensors := systemData.Sensors
	var content strings.Builder

	// Title
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Temperatures"))
	content.WriteString("\n\n")

	// Display fan speeds first with visual gauge (always visible if available)
	if len(sensors.Fans) > 0 {
		content.WriteString(t.label.Render("Fan Speeds"))
		content.WriteString("\n")
		for _, fan := range sensors.Fans {
			// Estimate max RPM for gauge (typically ~2000-3000 for case fans, GPU can be higher)
			maxRPM := estimateMaxFanRPM(fan.Name, fan.RPM)
			gauge := renderGauge(float64(fan.RPM), maxRPM, 20, t.normal, t.warning)
			content.WriteString(fmt.Sprintf("  %s\n    %s%d RPM\n",
				fan.Name,
				gauge,
				fan.RPM,
			))
		}
		content.WriteString("\n")
	}

	if len(sensors.Temperatures) == 0 {
		result := t.muted.Render("No temperature sensors found")
		return t.padToHeight(result)
	}

	// Group temperatures by sensor type and select representative temps
	tempGroups := make(map[string][]TempEntry)
	for _, temp := range sensors.Temperatures {
		sensorType := extractSensorType(temp.SensorKey)
		tempGroups[sensorType] = append(tempGroups[sensorType], TempEntry{
			Key:      temp.SensorKey,
			Temp:     temp.Temperature,
			Critical: temp.Critical,
		})
	}

	// Display temperatures with visual gauges
	for sensorType, temps := range tempGroups {
		// For coretemp and amdgpu, only show the highest (package) temp
		if sensorType == "coretemp" || sensorType == "amdgpu" {
			content.WriteString(t.renderSummaryTemp(sensorType, temps))
		} else {
			// For other sensors, show all individually
			content.WriteString(t.label.Render(sensorType))
			content.WriteString("\n")
			for _, temp := range temps {
				content.WriteString(t.renderTempGauge(temp))
			}
			content.WriteString("\n")
		}
	}

	return t.padToHeight(content.String())
}

// padToHeight pads the content with blank lines to reach target height
func (t *TemperatureMetrics) padToHeight(content string) string {
	if t.targetHeight <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	currentHeight := len(lines)

	// Pad with empty lines to reach target height
	for currentHeight < t.targetHeight {
		content += "\n"
		currentHeight++
	}

	return content
}

// renderSummaryTemp shows only the max temperature for a sensor type
func (t *TemperatureMetrics) renderSummaryTemp(sensorType string, temps []TempEntry) string {
	if len(temps) == 0 {
		return ""
	}

	// Find the highest temperature (usually the package temp)
	maxTemp := temps[0]
	for _, temp := range temps[1:] {
		if temp.Temp > maxTemp.Temp {
			maxTemp = temp
		}
	}

	var sb strings.Builder
	sb.WriteString(t.label.Render(sensorType))
	sb.WriteString("\n")
	sb.WriteString(t.renderTempGauge(maxTemp))
	sb.WriteString("\n")
	return sb.String()
}

// renderTempGauge renders a temperature with visual gauge
func (t *TemperatureMetrics) renderTempGauge(temp TempEntry) string {
	tempStyle := t.getMetricStyle(temp.Temp, 70, 85)

	// Temperature gauge: 0-100°C range
	gauge := renderGauge(temp.Temp, 100, 20, t.normal, tempStyle)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %s\n    %s%.1f°C",
		temp.Key,
		gauge,
		temp.Temp,
	))

	if temp.Critical != 0 {
		sb.WriteString(t.muted.Render(fmt.Sprintf(" (crit: %.0f°C)", temp.Critical)))
	}
	sb.WriteString("\n")
	return sb.String()
}

// renderGauge creates a horizontal bar gauge
func renderGauge(value, max float64, width int, normalStyle, fillStyle lipgloss.Style) string {
	if max == 0 {
		max = 1
	}

	percent := value / max
	if percent > 1 {
		percent = 1
	}

	filledWidth := int(math.Round(float64(width) * percent))
	if filledWidth > width {
		filledWidth = width
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", width-filledWidth)

	return fillStyle.Render(filled) + normalStyle.Render(empty)
}

// estimateMaxFanRPM estimates the maximum RPM for a fan based on its name
func estimateMaxFanRPM(name string, currentRPM uint64) float64 {
	name = strings.ToLower(name)

	// GPU fans typically spin faster
	if strings.Contains(name, "gpu") || strings.Contains(name, "amdgpu") || strings.Contains(name, "nvidia") {
		return 3500
	}

	// CPU fans
	if strings.Contains(name, "cpu") || strings.Contains(name, "coretemp") {
		return 2500
	}

	// Default for case fans
	if currentRPM > 2000 {
		return float64(currentRPM) * 1.2
	}
	return 2000
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
