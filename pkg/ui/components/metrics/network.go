package metrics

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// NetworkMetrics renders network metrics
type NetworkMetrics struct {
	label lipgloss.Style
	value lipgloss.Style
	muted lipgloss.Style
	width int
}

// NewNetworkMetrics creates a new network metrics renderer
func NewNetworkMetrics() *NetworkMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")

	return &NetworkMetrics{
		label: lipgloss.NewStyle().Foreground(colorCyan),
		value: lipgloss.NewStyle().Foreground(colorForeground),
		muted: lipgloss.NewStyle().Foreground(colorComment),
	}
}

// SetWidth sets the render width
func (n *NetworkMetrics) SetWidth(w int) {
	n.width = w
}

// Render returns the rendered network metrics
func (n *NetworkMetrics) Render(systemData *data.SystemData) string {
	if systemData == nil || systemData.Network == nil {
		return n.muted.Render("Loading network data...")
	}

	net := systemData.Network
	var content string

	// Title
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Network Interfaces")
	content += "\n\n"

	// Network stats per interface
	for _, iface := range net.Interfaces {
		io, ok := net.IO[iface.Name]
		if !ok {
			continue
		}

		content += fmt.Sprintf("%s%s%s\n",
			n.label,
			iface.Name,
			n.value,
		)

		if len(iface.Addrs) > 0 {
			content += fmt.Sprintf("  %sAddr:%s %s\n",
				n.muted,
				n.value,
				iface.Addrs[0].Addr,
			)
		}

		content += fmt.Sprintf("  %sRX:%s %s  %sTX:%s %s\n\n",
			n.muted,
			n.value,
			n.formatBytes(io.BytesRecv),
			n.muted,
			n.value,
			n.formatBytes(io.BytesSent),
		)
	}

	return content
}

func (n *NetworkMetrics) formatBytes(b uint64) string {
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
