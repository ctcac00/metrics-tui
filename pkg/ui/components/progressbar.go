package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a progress bar
type ProgressBar struct {
	width      int
	fillChar   string
	emptyChar  string
	fullStyle  lipgloss.Style
	emptyStyle lipgloss.Style
}

// NewProgressBar creates a new progress bar component
func NewProgressBar() *ProgressBar {
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorComment = lipgloss.Color("#44475a")

	return &ProgressBar{
		fillChar:   "█",
		emptyChar:  "░",
		fullStyle:  lipgloss.NewStyle().Foreground(colorGreen),
		emptyStyle: lipgloss.NewStyle().Foreground(colorComment),
	}
}

// SetWidth sets the total width of the progress bar
func (p *ProgressBar) SetWidth(w int) {
	p.width = w
}

// SetFillChar sets the character used for filled portion
func (p *ProgressBar) SetFillChar(c string) {
	p.fillChar = c
}

// SetEmptyChar sets the character used for empty portion
func (p *ProgressBar) SetEmptyChar(c string) {
	p.emptyChar = c
}

// SetFullStyle sets the style for filled portion
func (p *ProgressBar) SetFullStyle(s lipgloss.Style) {
	p.fullStyle = s
}

// SetEmptyStyle sets the style for empty portion
func (p *ProgressBar) SetEmptyStyle(s lipgloss.Style) {
	p.emptyStyle = s
}

// Render returns the rendered progress bar at the given percentage (0-100)
func (p *ProgressBar) Render(percent float64) string {
	if p.width <= 0 {
		return ""
	}

	// Clamp percent to 0-100
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Calculate filled width
	filledWidth := int(float64(p.width) * percent / 100.0)
	emptyWidth := p.width - filledWidth

	// Build the bar
	fullPart := strings.Repeat(p.fillChar, filledWidth)
	emptyPart := strings.Repeat(p.emptyChar, emptyWidth)

	return p.fullStyle.Render(fullPart) + p.emptyStyle.Render(emptyPart)
}

// RenderWithLabel returns the progress bar with a label showing percentage
func (p *ProgressBar) RenderWithLabel(percent float64, label string) string {
	bar := p.Render(percent)
	return lipgloss.JoinHorizontal(lipgloss.Left, bar, " "+label)
}

// RenderDynamic returns the progress bar with dynamic styling based on thresholds
func (p *ProgressBar) RenderDynamic(percent float64, warning, critical float64) string {
	// Update color based on thresholds
	if percent >= critical {
		p.fullStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true)
	} else if percent >= warning {
		p.fullStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c"))
	} else {
		p.fullStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b"))
	}

	return p.Render(percent)
}
