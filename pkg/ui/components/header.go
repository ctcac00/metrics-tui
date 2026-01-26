package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
)

// Header displays the top bar with host info
type Header struct {
	headerStyle lipgloss.Style
	width       int
}

// NewHeader creates a new header component with default styles
func NewHeader() *Header {
	var colorCyan = lipgloss.Color("#8be9fd")

	return &Header{
		headerStyle: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Padding(0, 1),
	}
}

// SetWidth sets the header width
func (h *Header) SetWidth(w int) {
	h.width = w
}

// Render returns the rendered header
func (h *Header) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Host == nil {
		return h.headerStyle.Render("Loading...")
	}

	var parts []string

	// Hostname
	if systemData.Host.Info.Hostname != "" {
		parts = append(parts, systemData.Host.Info.Hostname)
	}

	// Uptime
	if systemData.Host.Info.Uptime > 0 {
		uptime := formatUptime(systemData.Host.Info.Uptime)
		parts = append(parts, fmt.Sprintf("Uptime: %s", uptime))
	}

	// Load Average
	if systemData.Host.LoadAvg != nil {
		loadAvg := fmt.Sprintf("Load: %.2f %.2f %.2f",
			systemData.Host.LoadAvg.Load1,
			systemData.Host.LoadAvg.Load5,
			systemData.Host.LoadAvg.Load15)
		parts = append(parts, loadAvg)
	}

	// Join parts with spacing
	var content string
	for i, part := range parts {
		if i > 0 {
			content += " | "
		}
		content += part
	}

	return h.headerStyle.Width(h.width).Render(content)
}
