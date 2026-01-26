package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Tab represents a single tab in the sidebar
type Tab struct {
	Name   string
	Number int
}

// Sidebar displays the navigation tabs
type Sidebar struct {
	activeTabStyle lipgloss.Style
	inactiveTabStyle lipgloss.Style
	width     int
	height    int
	activeTab int
	tabs      []Tab
}

// NewSidebar creates a new sidebar component
func NewSidebar() *Sidebar {
	var colorComment = lipgloss.Color("#6272a4")
	var colorPink = lipgloss.Color("#ff79c6")

	return &Sidebar{
		activeTabStyle: lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true).
			Padding(0, 1),
		inactiveTabStyle: lipgloss.NewStyle().
			Foreground(colorComment).
			Padding(0, 1),
		tabs: []Tab{
			{Name: "CPU", Number: 1},
			{Name: "MEM", Number: 2},
			{Name: "DISK", Number: 3},
			{Name: "NET", Number: 4},
			{Name: "TEMP", Number: 5},
			{Name: "LOAD", Number: 6},
		},
		activeTab: 0,
	}
}

// SetWidth sets the sidebar width
func (s *Sidebar) SetWidth(w int) {
	s.width = w
}

// SetHeight sets the sidebar height
func (s *Sidebar) SetHeight(h int) {
	s.height = h
}

// SetActiveTab sets the active tab index
func (s *Sidebar) SetActiveTab(index int) {
	if index >= 0 && index < len(s.tabs) {
		s.activeTab = index
	}
}

// GetActiveTab returns the active tab index
func (s *Sidebar) GetActiveTab() int {
	return s.activeTab
}

// Render returns the rendered sidebar
func (s *Sidebar) Render() string {
	var tabs []string
	for i, tab := range s.tabs {
		if i == s.activeTab {
			tabs = append(tabs, s.activeTabStyle.Render(tab.Name))
		} else {
			tabs = append(tabs, s.inactiveTabStyle.Render(tab.Name))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabs...)
}
