package metrics

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
	"github.com/ctcac00/monitor-tui/pkg/ui/components"
)

// DiskMetrics renders disk metrics
type DiskMetrics struct {
	label       lipgloss.Style
	value       lipgloss.Style
	muted       lipgloss.Style
	normal      lipgloss.Style
	warning     lipgloss.Style
	critical    lipgloss.Style
	width       int
	progressBar *components.ProgressBar
}

// NewDiskMetrics creates a new disk metrics renderer
func NewDiskMetrics() *DiskMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")
	var colorRed = lipgloss.Color("#ff5555")

	return &DiskMetrics{
		label:       lipgloss.NewStyle().Foreground(colorCyan),
		value:       lipgloss.NewStyle().Foreground(colorForeground),
		muted:       lipgloss.NewStyle().Foreground(colorComment),
		normal:      lipgloss.NewStyle().Foreground(colorGreen),
		warning:     lipgloss.NewStyle().Foreground(colorOrange),
		critical:    lipgloss.NewStyle().Foreground(colorRed).Bold(true),
		progressBar: components.NewProgressBar(),
	}
}

// SetWidth sets the render width
func (d *DiskMetrics) SetWidth(w int) {
	d.width = w
	d.progressBar.SetWidth(25)
}

// Render returns the rendered disk metrics
func (d *DiskMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Disk == nil {
		return d.muted.Render("Loading disk data...")
	}

	disk := systemData.Disk
	var b strings.Builder

	// Title
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Disk Usage"))
	b.WriteString("\n\n")

	// Disk usage per partition with progress bars
	for _, partition := range disk.Partitions {
		usage, ok := disk.Usage[partition.Mountpoint]
		if !ok {
			continue
		}

		b.WriteString(fmt.Sprintf("%s%s%s\n",
			d.label,
			partition.Mountpoint,
			d.value,
		))

		// Progress bar for disk usage
		d.progressBar.SetWidth(25)
		style := d.getMetricStyle(usage.UsedPercent, 80, 95)
		b.WriteString(style.Render(d.progressBar.RenderDynamic(usage.UsedPercent, 80, 95)))
		b.WriteString(fmt.Sprintf(" %s%.1f%%%s\n",
			style,
			usage.UsedPercent,
			d.value,
		))

		b.WriteString(fmt.Sprintf("  %s / %s\n\n",
			d.formatBytes(usage.Used),
			d.formatBytes(usage.Total),
		))
	}

	return b.String()
}

func (d *DiskMetrics) getMetricStyle(value float64, warning, critical float64) lipgloss.Style {
	if value >= critical {
		return d.critical
	}
	if value >= warning {
		return d.warning
	}
	return d.normal
}

func (d *DiskMetrics) formatBytes(b uint64) string {
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
