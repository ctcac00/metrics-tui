package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Footer displays the bottom bar with keybindings
type Footer struct {
	footerStyle lipgloss.Style
	width       int
}

// NewFooter creates a new footer component
func NewFooter() *Footer {
	var colorComment = lipgloss.Color("#6272a4")

	return &Footer{
		footerStyle: lipgloss.NewStyle().
			Foreground(colorComment).
			Padding(0, 1),
	}
}

// SetWidth sets the footer width
func (f *Footer) SetWidth(w int) {
	f.width = w
}

// Render returns the rendered footer
func (f *Footer) Render() string {
	help := "[q] quit [h] help [1-6] select panel [↑↓] scroll"
	return f.footerStyle.Width(f.width).Render(help)
}
