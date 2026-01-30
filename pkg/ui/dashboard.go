package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
	"github.com/ctcac00/metrics-tui/pkg/ui/components/metrics"
)

// Dashboard renders a consolidated view of all metrics
type Dashboard struct {
	border lipgloss.Style
	width  int
	height int

	// Metric components (reuse existing components with all their graphics)
	cpuMetrics     *metrics.CPUMetrics
	memoryMetrics  *metrics.MemoryMetrics
	networkMetrics *metrics.NetworkMetrics
	tempMetrics    *metrics.TemperatureMetrics
}

// NewDashboard creates a new dashboard component
func NewDashboard() *Dashboard {
	var colorBorder = lipgloss.Color("#44475a")

	return &Dashboard{
		border:         lipgloss.NewStyle().Foreground(colorBorder),
		cpuMetrics:     metrics.NewCPUMetrics(),
		memoryMetrics:  metrics.NewMemoryMetrics(),
		networkMetrics: metrics.NewNetworkMetrics(),
		tempMetrics:    metrics.NewTemperatureMetrics(),
	}
}

// SetWidth sets the dashboard width
func (d *Dashboard) SetWidth(w int) {
	d.width = w
	// Distribute width among panels (3 columns with spacing)
	panelWidth := (w - 8) / 3
	d.cpuMetrics.SetWidth(panelWidth)
	d.memoryMetrics.SetWidth(panelWidth)
	d.networkMetrics.SetWidth(panelWidth)
	d.tempMetrics.SetWidth(panelWidth)
}

// SetHeight sets the dashboard height
func (d *Dashboard) SetHeight(h int) {
	d.height = h
}

// SetHistory sets the historical data for sparklines
func (d *Dashboard) SetHistory(cpuHistory, memHistory []float64) {
	d.cpuMetrics.SetHistory(cpuHistory)
	d.memoryMetrics.SetHistory(memHistory)
}

// ScrollUpCPU scrolls the CPU core list up
func (d *Dashboard) ScrollUpCPU() {
	d.cpuMetrics.ScrollUp()
}

// ScrollDownCPU scrolls the CPU core list down
func (d *Dashboard) ScrollDownCPU() {
	d.cpuMetrics.ScrollDown()
}

// CanScrollUpCPU returns true if CPU core list can scroll up
func (d *Dashboard) CanScrollUpCPU() bool {
	return d.cpuMetrics.CanScrollUp()
}

// CanScrollDownCPU returns true if CPU core list can scroll down
func (d *Dashboard) CanScrollDownCPU() bool {
	return d.cpuMetrics.CanScrollDown()
}

// Render returns the rendered dashboard
func (d *Dashboard) Render(systemData *data.SystemData) string {
	if systemData == nil {
		return "Loading system data..."
	}

	// First, render Memory and Network to determine their combined height
	// These don't need padding, so we render them first
	memContent := d.memoryMetrics.Render(systemData)
	netContent := d.networkMetrics.Render(systemData)

	// Calculate the combined height of column 3 (Memory + Network)
	memLines := len(strings.Split(memContent, "\n"))
	netLines := len(strings.Split(netContent, "\n"))
	col3ContentHeight := memLines + netLines + 2 // +2 for spacing between panels

	// Set target height for Temperature to match column 3
	d.tempMetrics.SetHeight(col3ContentHeight)

	// Now render Temperature with padding to match
	tempContent := d.tempMetrics.Render(systemData)

	// CPU content - render last as it scrolls independently
	cpuContent := d.cpuMetrics.Render(systemData)

	// Wrap each in a bordered panel
	cpuPanel := d.wrapInBox("CPU", cpuContent)
	memPanel := d.wrapInBox("Memory", memContent)
	netPanel := d.wrapInBox("Network", netContent)
	tempPanel := d.wrapInBox("Temperature", tempContent)

	// Layout: 3 columns
	// Column 1: CPU
	// Column 2: Temperature
	// Column 3: Memory on top of Network

	col3 := d.stackRows(memPanel, netPanel)

	return d.joinThreeColumns(cpuPanel, tempPanel, col3)
}

// wrapInBox wraps content in a nice bordered box
func (d *Dashboard) wrapInBox(title string, content string) string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.border.GetForeground()).
		Padding(0, 1)

	return borderStyle.Render(content)
}

// stackRows stacks two panels vertically
func (d *Dashboard) stackRows(top, bottom string) string {
	return top + "\n\n" + bottom
}

// joinThreeColumns joins three panels side by side
func (d *Dashboard) joinThreeColumns(col1, col2, col3 string) string {
	lines1 := strings.Split(col1, "\n")
	lines2 := strings.Split(col2, "\n")
	lines3 := strings.Split(col3, "\n")

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}
	if len(lines3) > maxLines {
		maxLines = len(lines3)
	}

	// Get visible width of each column's first line (ignores ANSI codes)
	col1Width := 0
	if len(lines1) > 0 {
		col1Width = lipgloss.Width(lines1[0])
	}
	col2Width := 0
	if len(lines2) > 0 {
		col2Width = lipgloss.Width(lines2[0])
	}

	var result strings.Builder
	for i := 0; i < maxLines; i++ {
		// Column 1
		if i < len(lines1) {
			result.WriteString(lines1[i])
		} else {
			result.WriteString(strings.Repeat(" ", col1Width))
		}

		result.WriteString("  ") // Spacing between columns

		// Column 2
		if i < len(lines2) {
			result.WriteString(lines2[i])
		} else {
			result.WriteString(strings.Repeat(" ", col2Width))
		}

		result.WriteString("  ") // Spacing between columns

		// Column 3
		if i < len(lines3) {
			result.WriteString(lines3[i])
		}
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
