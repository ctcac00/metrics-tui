package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// MemoryMetrics renders memory metrics
type MemoryMetrics struct {
	label    lipgloss.Style
	value    lipgloss.Style
	muted    lipgloss.Style
	normal   lipgloss.Style
	warning  lipgloss.Style
	critical lipgloss.Style
	width    int
}

// NewMemoryMetrics creates a new memory metrics renderer
func NewMemoryMetrics() *MemoryMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var _ = lipgloss.Color("#bd93f9") // unused
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &MemoryMetrics{
		label:    lipgloss.NewStyle().Foreground(colorCyan),
		value:    lipgloss.NewStyle().Foreground(colorForeground),
		muted:    lipgloss.NewStyle().Foreground(colorComment),
		normal:   lipgloss.NewStyle().Foreground(colorGreen),
		warning:  lipgloss.NewStyle().Foreground(colorOrange),
		critical: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	}
}

// SetWidth sets the render width
func (m *MemoryMetrics) SetWidth(w int) {
	m.width = w
}

// Render returns the rendered memory metrics
func (m *MemoryMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Memory == nil {
		return m.muted.Render("Loading memory data...")
	}

	mem := systemData.Memory
	var content string

	// Title
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Memory Usage")
	content += "\n\n"

	// Memory stats
	content += fmt.Sprintf("%sTotal:%s     %s\n",
		m.label,
		m.value,
		m.formatBytes(mem.Total),
	)

	usedStyle := m.getMetricStyle(mem.UsedPercent, 80, 95)
	content += fmt.Sprintf("%sUsed:%s      %s (%s%.1f%%%s)\n",
		m.label,
		m.value,
		m.formatBytes(mem.Used),
		usedStyle,
		mem.UsedPercent,
		m.value,
	)

	content += fmt.Sprintf("%sAvailable:%s %s\n",
		m.label,
		m.value,
		m.formatBytes(mem.Available),
	)

	content += fmt.Sprintf("%sFree:%s      %s\n",
		m.label,
		m.value,
		m.formatBytes(mem.Free),
	)

	// Swap info
	if mem.Swap.Total > 0 {
		content += "\n"
		content += m.label.Render("Swap:")
		content += "\n"

		swapStyle := m.getMetricStyle(mem.Swap.UsedPercent, 50, 80)
		content += fmt.Sprintf("  %s / %s (%s%.1f%%%s)\n",
			m.formatBytes(mem.Swap.Used),
			m.formatBytes(mem.Swap.Total),
			swapStyle,
			mem.Swap.UsedPercent,
			m.value,
		)
	}

	return content
}

func (m *MemoryMetrics) getMetricStyle(value float64, warning, critical float64) lipgloss.Style {
	if value >= critical {
		return m.critical
	}
	if value >= warning {
		return m.warning
	}
	return m.normal
}

func (m *MemoryMetrics) formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
