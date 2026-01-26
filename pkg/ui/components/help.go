package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Help displays the help screen
type Help struct {
	titleStyle   lipgloss.Style
	headerStyle  lipgloss.Style
	keyStyle     lipgloss.Style
	descStyle    lipgloss.Style
	footerStyle  lipgloss.Style
	visible      bool
	width        int
	height       int
}

// NewHelp creates a new help component
func NewHelp() *Help {
	var colorPurple = lipgloss.Color("#bd93f9")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorComment = lipgloss.Color("#6272a4")

	return &Help{
		titleStyle:  lipgloss.NewStyle().Foreground(colorPurple).Bold(true),
		headerStyle: lipgloss.NewStyle().Foreground(colorCyan).Bold(true),
		keyStyle:    lipgloss.NewStyle().Foreground(colorGreen),
		descStyle:   lipgloss.NewStyle().Foreground(colorComment),
		footerStyle: lipgloss.NewStyle().Foreground(colorComment).Italic(true),
		visible:     false,
	}
}

// Show displays the help screen
func (h *Help) Show() {
	h.visible = true
}

// Hide hides the help screen
func (h *Help) Hide() {
	h.visible = false
}

// IsVisible returns whether help is currently visible
func (h *Help) IsVisible() bool {
	return h.visible
}

// SetSize sets the dimensions
func (h *Help) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// Render returns the rendered help screen
func (h *Help) Render() string {
	if !h.visible {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(h.titleStyle.Render("Monitor TUI - Help"))
	b.WriteString("\n\n")

	// Navigation
	b.WriteString(h.headerStyle.Render("Navigation"))
	b.WriteString("\n")
	helpItems := [][]string{
		{"q, Ctrl+C", "Quit the application"},
		{"h, ?", "Show/hide this help screen"},
		{"1-6", "Switch between metric panels"},
		{"↑, k", "Scroll up"},
		{"↓, j", "Scroll down"},
	}

	for _, item := range helpItems {
		b.WriteString(h.keyStyle.Render(item[0]))
		b.WriteString("   ")
		b.WriteString(h.descStyle.Render(item[1]))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Panels
	b.WriteString(h.headerStyle.Render("Panels"))
	b.WriteString("\n")
	panelItems := [][]string{
		{"1", "CPU - Processor usage and load"},
		{"2", "Memory - RAM and swap usage"},
		{"3", "Disk - Storage usage and I/O stats"},
		{"4", "Network - Interface traffic statistics"},
		{"5", "Temperature - Sensor readings"},
		{"6", "Load - System load average"},
	}

	for _, item := range panelItems {
		b.WriteString(h.keyStyle.Render(item[0]))
		b.WriteString("   ")
		b.WriteString(h.descStyle.Render(item[1]))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Indicators
	b.WriteString(h.headerStyle.Render("Color Indicators"))
	b.WriteString("\n")
	indicatorItems := [][]string{
		{"Green", "Normal usage"},
		{"Orange", "Warning threshold exceeded"},
		{"Red/Bold", "Critical threshold exceeded"},
	}

	for _, item := range indicatorItems {
		b.WriteString(h.keyStyle.Render(item[0]))
		b.WriteString("  ")
		b.WriteString(h.descStyle.Render(item[1]))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Footer
	b.WriteString(h.footerStyle.Render("Press any key to close"))

	// Center the help content if we have space
	content := b.String()
	lines := strings.Split(content, "\n")

	// Calculate padding to center
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	padding := (h.width - maxWidth) / 2
	if padding < 0 {
		padding = 0
	}

	padStyle := lipgloss.NewStyle().Padding(0, padding)

	// Add padding to each line and center vertically
	var result strings.Builder
	verticalPadding := (h.height - len(lines)) / 2
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	for i := 0; i < verticalPadding; i++ {
		result.WriteString("\n")
	}

	for _, line := range lines {
		result.WriteString(padStyle.Render(line))
		result.WriteString("\n")
	}

	return result.String()
}
