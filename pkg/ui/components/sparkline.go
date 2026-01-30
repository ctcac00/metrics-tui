package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SparkLine renders a sparkline chart from historical data
type SparkLine struct {
	width  int
	height int
	data   []float64
	style  lipgloss.Style
}

// SparklineChars defines the characters used for sparkline rendering
var SparklineChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// NewSparkLine creates a new sparkline component
func NewSparkLine() *SparkLine {
	var colorCyan = lipgloss.Color("#8be9fd")

	return &SparkLine{
		width:  40,
		height: 1,
		style:  lipgloss.NewStyle().Foreground(colorCyan),
	}
}

// SetWidth sets the width (number of data points to display)
func (s *SparkLine) SetWidth(w int) {
	s.width = w
}

// SetHeight sets the height (number of rows)
func (s *SparkLine) SetHeight(h int) {
	s.height = h
}

// SetData sets the data points to display
func (s *SparkLine) SetData(data []float64) {
	s.data = data
	if len(s.data) > s.width {
		s.data = s.data[len(s.data)-s.width:]
	}
}

// SetStyle sets the rendering style
func (s *SparkLine) SetStyle(style lipgloss.Style) {
	s.style = style
}

// AddValue adds a new value to the data
func (s *SparkLine) AddValue(value float64) {
	s.data = append(s.data, value)
	// Keep only the last width elements if we exceed it
	if len(s.data) > s.width {
		s.data = s.data[len(s.data)-s.width:]
	}
}

// Render returns the rendered sparkline
func (s *SparkLine) Render() string {
	if len(s.data) == 0 {
		return strings.Repeat(" ", s.width)
	}

	// Get the last width elements
	data := s.data
	if len(data) > s.width {
		data = data[len(data)-s.width:]
	}

	// Find min and max for normalization
	min, max := s.getMinMax(data)
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}

	var result strings.Builder

	// Left-pad with spaces to maintain fixed width when data is short
	padding := s.width - len(data)
	if padding > 0 {
		result.WriteString(strings.Repeat(" ", padding))
	}

	for _, value := range data {
		// Normalize to 0-1
		normalized := (value - min) / rangeVal

		// Map to character index
		charIndex := int(normalized * float64(len(SparklineChars)-1))
		if charIndex < 0 {
			charIndex = 0
		}
		if charIndex >= len(SparklineChars) {
			charIndex = len(SparklineChars) - 1
		}

		result.WriteRune(SparklineChars[charIndex])
	}

	return s.style.Render(result.String())
}

// RenderMultiLine renders a multi-row sparkline
func (s *SparkLine) RenderMultiLine() string {
	if s.height <= 1 {
		return s.Render()
	}

	if len(s.data) == 0 {
		return strings.Repeat("\n"+strings.Repeat(" ", s.width), s.height)
	}

	// Get the last width elements
	data := s.data
	if len(data) > s.width {
		data = data[len(data)-s.width:]
	}

	// Find min and max for normalization
	min, max := s.getMinMax(data)
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}

	// Build each line from top to bottom
	var lines []string
	padding := s.width - len(data)
	for row := s.height - 1; row >= 0; row-- {
		var line strings.Builder

		// Left-pad with spaces to maintain fixed width when data is short
		if padding > 0 {
			line.WriteString(strings.Repeat(" ", padding))
		}

		for _, value := range data {
			// Normalize to 0-1
			normalized := (value - min) / rangeVal

			// Calculate which row this value should appear on
			valueRow := int(normalized * float64(s.height-1))

			if valueRow >= row {
				line.WriteRune(SparklineChars[len(SparklineChars)-1])
			} else {
				line.WriteRune(' ')
			}
		}
		lines = append(lines, s.style.Render(line.String()))
	}

	return strings.Join(lines, "\n")
}

// RenderWithColor returns sparkline with color based on latest value
func (s *SparkLine) RenderWithColor(warning, critical float64) string {
	if len(s.data) == 0 {
		return strings.Repeat(" ", s.width)
	}

	// Update color based on latest value
	latest := s.data[len(s.data)-1]
	if latest >= critical {
		s.style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true)
	} else if latest >= warning {
		s.style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c"))
	} else {
		s.style = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b"))
	}

	return s.Render()
}

// getMinMax finds the minimum and maximum values in data
func (s *SparkLine) getMinMax(data []float64) (min, max float64) {
	if len(data) == 0 {
		return 0, 1
	}

	min = data[0]
	max = data[0]

	for _, v := range data[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	return min, max
}

// GetLastValue returns the most recent value
func (s *SparkLine) GetLastValue() float64 {
	if len(s.data) == 0 {
		return 0
	}
	return s.data[len(s.data)-1]
}

// GetAverage returns the average of all values
func (s *SparkLine) GetAverage() float64 {
	if len(s.data) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range s.data {
		sum += v
	}
	return sum / float64(len(s.data))
}
