package metrics

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/monitor-tui/internal/data"
)

// NetworkMetrics renders network metrics
type NetworkMetrics struct {
	label   lipgloss.Style
	value   lipgloss.Style
	muted   lipgloss.Style
	normal  lipgloss.Style
	warning lipgloss.Style
	width   int
}

// NewNetworkMetrics creates a new network metrics renderer
func NewNetworkMetrics() *NetworkMetrics {
	var colorForeground = lipgloss.Color("#f8f8f2")
	var colorComment = lipgloss.Color("#6272a4")
	var colorCyan = lipgloss.Color("#8be9fd")
	var colorGreen = lipgloss.Color("#50fa7b")
	var colorOrange = lipgloss.Color("#ffb86c")

	return &NetworkMetrics{
		label:   lipgloss.NewStyle().Foreground(colorCyan),
		value:   lipgloss.NewStyle().Foreground(colorForeground),
		muted:   lipgloss.NewStyle().Foreground(colorComment),
		normal:  lipgloss.NewStyle().Foreground(colorGreen),
		warning: lipgloss.NewStyle().Foreground(colorOrange),
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
	var content strings.Builder

	// Title
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true).Render("Network Interfaces"))
	content.WriteString("\n\n")

	// Network stats per interface
	for _, iface := range net.Interfaces {
		io, ok := net.IO[iface.Name]
		if !ok {
			continue
		}

		content.WriteString(fmt.Sprintf("%s%s%s\n",
			n.label,
			iface.Name,
			n.value,
		))

		if len(iface.Addrs) > 0 {
			content.WriteString(fmt.Sprintf("  %sAddr:%s %s\n",
				n.muted,
				n.value,
				iface.Addrs[0].Addr,
			))
		}

		// RX with gauge (scale to 1 GB max for visualization)
		maxBytes := uint64(1024 * 1024 * 1024) // 1 GB
		rxGauge := n.renderByteGauge(io.BytesRecv, maxBytes)
		txGauge := n.renderByteGauge(io.BytesSent, maxBytes)

		content.WriteString(fmt.Sprintf("  %sRX:%s %s %s\n",
			n.muted,
			n.value,
			n.formatBytes(io.BytesRecv),
			rxGauge,
		))

		content.WriteString(fmt.Sprintf("  %sTX:%s %s %s\n\n",
			n.muted,
			n.value,
			n.formatBytes(io.BytesSent),
			txGauge,
		))
	}

	return content.String()
}

// renderByteGauge creates a visual gauge for bytes transferred
func (n *NetworkMetrics) renderByteGauge(bytes, maxBytes uint64) string {
	width := 15

	if bytes == 0 {
		return strings.Repeat("░", width)
	}

	// Calculate fill percentage
	var percent float64
	if bytes >= maxBytes {
		percent = 1.0
	} else {
		percent = float64(bytes) / float64(maxBytes)
	}

	filledWidth := int(float64(width) * percent)
	if filledWidth > width {
		filledWidth = width
	}

	// Choose color based on usage
	style := n.normal
	if percent > 0.7 {
		style = n.warning
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", width-filledWidth)

	return style.Render(filled) + n.normal.Render(empty)
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
