package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
)

// ProcessList displays process information
type ProcessList struct {
	titleStyle    lipgloss.Style
	headerStyle   lipgloss.Style
	pidStyle      lipgloss.Style
	nameStyle     lipgloss.Style
	cpuStyle      lipgloss.Style
	memStyle      lipgloss.Style
	normalStyle   lipgloss.Style
	warningStyle  lipgloss.Style
	criticalStyle lipgloss.Style
	mutedStyle    lipgloss.Style
	width         int
	height        int
	processes      []ProcessInfo
}

// ProcessInfo holds information about a single process
type ProcessInfo struct {
	PID     int
	Name    string
	CPU     float64
	Memory  float64
	Command string
}

// NewProcessList creates a new process list component
func NewProcessList() *ProcessList {
	var colorPurple = lipgloss.Color("#bd93f9")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")
	var colorComment = lipgloss.Color("#6272a4")
	var colorForeground = lipgloss.Color("#f8f8f2")

	return &ProcessList{
		titleStyle:    lipgloss.NewStyle().Foreground(colorPurple).Bold(true),
		headerStyle:   lipgloss.NewStyle().Foreground(colorCyan).Bold(true),
		pidStyle:      lipgloss.NewStyle().Foreground(colorComment),
		nameStyle:     lipgloss.NewStyle().Foreground(colorForeground),
		cpuStyle:      lipgloss.NewStyle().Foreground(colorGreen),
		memStyle:      lipgloss.NewStyle().Foreground(colorGreen),
		normalStyle:   lipgloss.NewStyle().Foreground(colorGreen),
		warningStyle:  lipgloss.NewStyle().Foreground(colorOrange),
		criticalStyle: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
		mutedStyle:    lipgloss.NewStyle().Foreground(colorComment),
		processes:      make([]ProcessInfo, 0, 10),
	}
}

// SetWidth sets the render width
func (p *ProcessList) SetWidth(w int) {
	p.width = w
}

// SetHeight sets the render height
func (p *ProcessList) SetHeight(h int) {
	p.height = h
}

// SetProcesses sets the process list
func (p *ProcessList) SetProcesses(procs []ProcessInfo) {
	p.processes = procs
}

// AddProcess adds a process to the list
func (p *ProcessList) AddProcess(proc ProcessInfo) {
	p.processes = append(p.processes, proc)
}

// Clear clears the process list
func (p *ProcessList) Clear() {
	p.processes = make([]ProcessInfo, 0, 10)
}

// Render returns the rendered process list
func (p *ProcessList) Render(systemData *data.SystemData) string {
	var b strings.Builder

	// Title
	b.WriteString(p.titleStyle.Render("Top Processes"))
	b.WriteString("\n\n")

	if len(p.processes) == 0 {
		b.WriteString(p.mutedStyle.Render("No process data available"))
		b.WriteString("\n\n")
		b.WriteString(p.mutedStyle.Render("(Process listing requires additional permissions)"))
		return b.String()
	}

	// Header
	b.WriteString(fmt.Sprintf("%-7s %-20s %-8s %-8s\n",
		p.headerStyle.Render("PID"),
		p.headerStyle.Render("NAME"),
		p.headerStyle.Render("CPU%"),
		p.headerStyle.Render("MEM%"),
	))
	b.WriteString(p.mutedStyle.Render(strings.Repeat("-", p.width-4)))
	b.WriteString("\n")

	// Process rows
	for _, proc := range p.processes {
		cpuStyle := p.getCPUStyle(proc.CPU)
		memStyle := p.getMemStyle(proc.Memory)

		// Truncate name if too long
		name := proc.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		b.WriteString(fmt.Sprintf("%-7d %-20s %-8s %-8s\n",
			p.pidStyle.Render(fmt.Sprintf("%d", proc.PID)),
			p.nameStyle.Render(name),
			cpuStyle.Render(fmt.Sprintf("%.1f", proc.CPU)),
			memStyle.Render(fmt.Sprintf("%.1f", proc.Memory)),
		))
	}

	b.WriteString("\n")
	b.WriteString(p.mutedStyle.Render(fmt.Sprintf("Showing %d processes", len(p.processes))))

	return b.String()
}

// getCPUStyle returns style based on CPU usage
func (p *ProcessList) getCPUStyle(cpu float64) lipgloss.Style {
	if cpu >= 50 {
		return p.criticalStyle
	}
	if cpu >= 20 {
		return p.warningStyle
	}
	return p.cpuStyle
}

// getMemStyle returns style based on memory usage
func (p *ProcessList) getMemStyle(mem float64) lipgloss.Style {
	if mem >= 50 {
		return p.criticalStyle
	}
	if mem >= 20 {
		return p.warningStyle
	}
	return p.memStyle
}
