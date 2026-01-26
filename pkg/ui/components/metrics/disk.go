package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// DiskMetrics renders disk metrics
type DiskMetrics struct {
	label  lipgloss.Style
	value  lipgloss.Style
	muted  lipgloss.Style
	width  int
}

// NewDiskMetrics creates a new disk metrics renderer
func NewDiskMetrics() *DiskMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var _ = lipgloss.Color("#bd93f9") // unused

	return &DiskMetrics{
		label: lipgloss.NewStyle().Foreground(colorCyan),
		value: lipgloss.NewStyle().Foreground(colorForeground),
		muted: lipgloss.NewStyle().Foreground(colorComment),
	}
}

// SetWidth sets the render width
func (d *DiskMetrics) SetWidth(w int) {
	d.width = w
}

// Render returns the rendered disk metrics
func (d *DiskMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Disk == nil {
		return d.muted.Render("Loading disk data...")
	}

	disk := systemData.Disk
	var content string

	// Title
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Disk Usage")
	content += "\n\n"

	// Disk usage per partition
	for _, partition := range disk.Partitions {
		usage, ok := disk.Usage[partition.Mountpoint]
		if !ok {
			continue
		}

		content += fmt.Sprintf("%s%s%s\n",
			d.label,
			partition.Mountpoint,
			d.value,
		)

		content += fmt.Sprintf("  %s / %s (%.1f%%)\n",
			d.formatBytes(usage.Used),
			d.formatBytes(usage.Total),
			usage.UsedPercent,
		)

		content += "\n"
	}

	return content
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
