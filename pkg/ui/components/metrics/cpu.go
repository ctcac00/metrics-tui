package metrics

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
	"github.com/ctcac00/metrics-tui/pkg/ui/components"
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
	progressBar  *components.ProgressBar
	sparkline    *components.SparkLine
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
		progressBar:  components.NewProgressBar(),
		sparkline:    components.NewSparkLine(),
	}
}

// SetWidth sets the render width
func (c *CPUMetrics) SetWidth(w int) {
	c.width = w
	c.progressBar.SetWidth(30)
	sparkWidth := w - 24
	if sparkWidth < 10 {
		sparkWidth = 10
	}
	c.sparkline.SetWidth(sparkWidth)
}

// SetHistory sets the historical data for sparklines
func (c *CPUMetrics) SetHistory(data []float64) {
	c.sparkline.SetData(data)
}

// Render returns the rendered CPU metrics
func (c *CPUMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.CPU == nil {
		return c.muted.Render("Loading CPU data...")
	}

	cpu := systemData.CPU
	var b strings.Builder

	// Title
	b.WriteString(c.sectionTitle.Render("CPU Usage"))
	b.WriteString("\n\n")

	// Total usage with progress bar
	totalStyle := c.getMetricStyle(cpu.Total, 70, 90)
	b.WriteString(fmt.Sprintf("Total: %s%.1f%%%s\n",
		totalStyle,
		cpu.Total,
		c.value,
	))

	// Progress bar for total usage
	c.progressBar.SetWidth(30)
	b.WriteString(c.progressBar.RenderDynamic(cpu.Total, 70, 90))
	b.WriteString("\n\n")

	// Sparkline for CPU history
	if c.sparkline.GetLastValue() > 0 {
		b.WriteString(c.label.Render("History:"))
		b.WriteString(" ")
		b.WriteString(fmt.Sprintf("%.1f%% ", c.sparkline.GetLastValue()))
		b.WriteString(c.sparkline.RenderWithColor(70, 90))
		b.WriteString("\n\n")
	}

	// Core count
	b.WriteString(c.muted.Render(fmt.Sprintf("Cores: %d", cpu.CoreCount)))
	b.WriteString("\n\n")

	// Per-core usage with progress bars
	if len(cpu.Usage) > 0 {
		b.WriteString(c.label.Render("Per-Core Usage:"))
		b.WriteString("\n")

		coresPerRow := 2
		for i, usage := range cpu.Usage {
			if i > 0 && i%coresPerRow == 0 {
				b.WriteString("\n")
			}

			coreStyle := c.getMetricStyle(usage, 70, 90)
			c.progressBar.SetWidth(15)
			bar := c.progressBar.RenderDynamic(usage, 70, 90)

			b.WriteString(fmt.Sprintf("%sCore %2d:%s %5.1f%% %s\n",
				c.muted,
				i,
				coreStyle,
				usage,
				bar,
			))
		}
	}

	return b.String()
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
