package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// LoadMetrics renders load average metrics
type LoadMetrics struct {
	label    lipgloss.Style
	value    lipgloss.Style
	muted    lipgloss.Style
	normal   lipgloss.Style
	warning  lipgloss.Style
	critical lipgloss.Style
	width    int
}

// NewLoadMetrics creates a new load metrics renderer
func NewLoadMetrics() *LoadMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &LoadMetrics{
		label:    lipgloss.NewStyle().Foreground(colorCyan),
		value:    lipgloss.NewStyle().Foreground(colorForeground),
		muted:    lipgloss.NewStyle().Foreground(colorComment),
		normal:   lipgloss.NewStyle().Foreground(colorGreen),
		warning:  lipgloss.NewStyle().Foreground(colorOrange),
		critical: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	}
}

// SetWidth sets the render width
func (l *LoadMetrics) SetWidth(w int) {
	l.width = w
}

// Render returns the rendered load metrics
func (l *LoadMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Host == nil {
		return l.muted.Render("Loading load average data...")
	}

	if systemData.Host.LoadAvg == nil {
		return l.muted.Render("Load average not available")
	}

	load := systemData.Host.LoadAvg
	var content string

	// Title
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Load Average")
	content += "\n\n"

	// Get CPU count for context
	cpuCount := 1.0
	if systemData.CPU != nil && systemData.CPU.CoreCount > 0 {
		cpuCount = float64(systemData.CPU.CoreCount)
	}

	// 1 minute average
	load1Style := l.getMetricStyle(load.Load1/cpuCount*100, 70, 90)
	content += fmt.Sprintf("%s1 min:%s  %s%.2f%s",
		l.label,
		l.value,
		load1Style,
		load.Load1,
		l.value,
	)

	// Show percentage of CPU capacity
	content += l.muted.Render(fmt.Sprintf(" (%.0f%% of %d core%s)\n",
		load.Load1/cpuCount*100,
		int(cpuCount),
		map[bool]string{true: "s", false: ""}[cpuCount > 1],
	))

	// 5 minute average
	load5Style := l.getMetricStyle(load.Load5/cpuCount*100, 70, 90)
	content += fmt.Sprintf("%s5 min:%s  %s%.2f%s",
		l.label,
		l.value,
		load5Style,
		load.Load5,
		l.value,
	)

	content += l.muted.Render(fmt.Sprintf(" (%.0f%%)\n", load.Load5/cpuCount*100))

	// 15 minute average
	load15Style := l.getMetricStyle(load.Load15/cpuCount*100, 70, 90)
	content += fmt.Sprintf("%s15 min:%s %s%.2f%s",
		l.label,
		l.value,
		load15Style,
		load.Load15,
		l.value,
	)

	content += l.muted.Render(fmt.Sprintf(" (%.0f%%)\n\n", load.Load15/cpuCount*100))

	// System info
	if systemData.Host.Info.Uptime > 0 {
		content += l.label.Render("System Uptime:")
		content += "\n"
		content += fmt.Sprintf("  %s\n", formatUptime(systemData.Host.Info.Uptime))
	}

	if systemData.Host.Info.OS != "" {
		content += l.label.Render("Operating System:")
		content += "\n"
		content += fmt.Sprintf("  %s %s\n",
			systemData.Host.Info.Platform,
			systemData.Host.Info.PlatformVersion,
		)
	}

	if systemData.Host.Info.KernelVersion != "" {
		content += l.label.Render("Kernel:")
		content += "\n"
		content += fmt.Sprintf("  %s\n", systemData.Host.Info.KernelVersion)
	}

	return content
}

func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func (l *LoadMetrics) getMetricStyle(value float64, warning, critical float64) lipgloss.Style {
	if value >= critical {
		return l.critical
	}
	if value >= warning {
		return l.warning
	}
	return l.normal
}
