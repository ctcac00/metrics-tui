package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// CPUMetrics renders CPU metrics
type CPUMetrics struct {
	sectionTitle lipgloss.Style
	label        lipgloss.Style
	value        lipgloss.Style
	muted        lipgloss.Style
	normal       lipgloss.Style
	warning      lipgloss.Style
	critical     lipgloss.Style
	width        int
}

// NewCPUMetrics creates a new CPU metrics renderer
func NewCPUMetrics() *CPUMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorPurple = lipgloss.Color("#bd93f9")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &CPUMetrics{
		sectionTitle: lipgloss.NewStyle().Foreground(colorPurple).Bold(true),
		label:        lipgloss.NewStyle().Foreground(colorCyan),
		value:        lipgloss.NewStyle().Foreground(colorForeground),
		muted:        lipgloss.NewStyle().Foreground(colorComment),
		normal:       lipgloss.NewStyle().Foreground(colorGreen),
		warning:      lipgloss.NewStyle().Foreground(colorOrange),
		critical:     lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	}
}

// SetWidth sets the render width
func (c *CPUMetrics) SetWidth(w int) {
	c.width = w
}

// Render returns the rendered CPU metrics
func (c *CPUMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.CPU == nil {
		return c.muted.Render("Loading CPU data...")
	}

	cpu := systemData.CPU
	var content string

	// Title
	content += c.sectionTitle.Render("CPU Usage")
	content += "\n\n"

	// Total usage
	totalStyle := c.getMetricStyle(cpu.Total, 70, 90)
	content += fmt.Sprintf("Total: %s%.1f%%%s\n",
		totalStyle,
		cpu.Total,
		c.value,
	)

	// Core count
	content += c.muted.Render(fmt.Sprintf("Cores: %d", cpu.CoreCount))
	content += "\n\n"

	// Per-core usage
	if len(cpu.Usage) > 0 {
		content += c.label.Render("Per-Core Usage:")
		content += "\n"

		coresPerRow := 4
		for i, usage := range cpu.Usage {
			if i > 0 && i%coresPerRow == 0 {
				content += "\n"
			}

			coreStyle := c.getMetricStyle(usage, 70, 90)
			content += fmt.Sprintf("%sCore %2d:%s %5.1f%%  ",
				c.muted,
				i,
				coreStyle,
				usage,
			)
		}
		content += "\n"
	}

	return content
}

func (c *CPUMetrics) getMetricStyle(value float64, warning, critical float64) lipgloss.Style {
	if value >= critical {
		return c.critical
	}
	if value >= warning {
		return c.warning
	}
	return c.normal
}
