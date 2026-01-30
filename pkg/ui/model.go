package ui

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ctcac00/metrics-tui/internal/data"
	"github.com/ctcac00/metrics-tui/pkg/collectors"
	"github.com/ctcac00/metrics-tui/pkg/ui/components"
)

// Model is the main Bubble Tea model for the TUI
type Model struct {
	width      int
	height     int
	quitting   bool
	showHelp   bool
	systemData *data.SystemData
	history    *data.HistoryData

	// Components
	header       *components.Header
	footer       *components.Footer
	help         *components.Help
	dashboard    *Dashboard
	alertBar     *components.AlertBar
	alertManager *components.AlertManager

	// Aggregator
	aggregator *collectors.Aggregator
}

// NewModel creates a new TUI model
func NewModel() *Model {
	m := &Model{
		showHelp:   false,
		systemData: &data.SystemData{},
		history:    data.NewHistoryData(50), // 50 data points for sparklines
	}

	// Initialize components
	m.header = components.NewHeader()
	m.footer = components.NewFooter()
	m.help = components.NewHelp()
	m.dashboard = NewDashboard()
	m.alertManager = components.NewAlertManager()
	m.alertBar = components.NewAlertBar(m.alertManager)

	// Set up alert thresholds
	m.alertManager.SetThreshold("cpu", 70, 90)
	m.alertManager.SetThreshold("memory", 80, 95)
	m.alertManager.SetThreshold("temperature", 70, 85)

	// Initialize aggregator
	config := collectors.DefaultAggregatorConfig()
	m.aggregator = collectors.NewAggregator(config)
	m.aggregator.SetOnDataUpdate(m.onDataUpdate)

	return m
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	m.aggregator.Start()
	return m.tickCmd()
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			m.aggregator.Stop()
			return m, tea.Quit

		case "h", "?":
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.help.Show()
			} else {
				m.help.Hide()
			}
			return m, nil

		case "esc", "escape":
			// Close help on escape
			if m.showHelp {
				m.showHelp = false
				m.help.Hide()
			}
			return m, nil

		case "s":
			// Take snapshot
			snapshotMgr := components.NewSnapshotManagerWithDefaults()
			snapshot, err := snapshotMgr.TakeSnapshot(m.systemData)
			if err == nil {
				snapshotMgr.SaveToFile(snapshot, "")
			}
			return m, nil

		case "up", "k":
			// Scroll CPU cores up
			m.dashboard.ScrollUpCPU()
			return m, nil

		case "down", "j":
			// Scroll CPU cores down
			m.dashboard.ScrollDownCPU()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.header.SetWidth(msg.Width)
		m.footer.SetWidth(msg.Width)
		m.help.SetSize(msg.Width, msg.Height)
		m.dashboard.SetWidth(msg.Width - 4)   // Leave padding
		m.dashboard.SetHeight(msg.Height - 4) // Leave room for header and footer
		m.alertBar.SetWidth(msg.Width)

	case tickMsg:
		// Update history with latest data
		m.updateHistory()
		return m, m.tickCmd()

	case dataMsg:
		m.systemData = msg.data
	}

	return m, nil
}

// View implements tea.Model
func (m *Model) View() string {
	if m.quitting {
		return "\n  Goodbye!\n\n"
	}

	// If no size yet, show loading
	if m.width == 0 {
		return "Loading..."
	}

	// If help is visible, show help overlay
	if m.showHelp {
		return m.help.Render()
	}

	// Update history data for dashboard
	if m.history != nil {
		m.dashboard.SetHistory(m.history.CPU, m.history.Memory)
	}

	// Render header with alert bar
	header := m.header.Render(m.systemData)
	alertBar := m.alertBar.Render()
	if alertBar != "" {
		header = lipgloss.JoinVertical(lipgloss.Left, header, alertBar)
	}

	// Render dashboard
	dashboard := m.dashboard.Render(m.systemData)

	// Render footer
	footer := m.footer.Render()

	// Add padding around dashboard
	dashboardStyle := lipgloss.NewStyle().Padding(1, 2)
	dashboardPadded := dashboardStyle.Render(dashboard)

	// Join all parts vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		dashboardPadded,
		footer,
	)
}

// onDataUpdate is called when new data is available from the aggregator
func (m *Model) onDataUpdate(d *data.SystemData) {
	m.systemData = d
}

// updateHistory updates the history data with current values
func (m *Model) updateHistory() {
	if m.systemData.CPU != nil {
		m.history.AddCPU(m.systemData.CPU.Total)
		// Check CPU alerts
		m.alertManager.CheckValue("cpu", m.systemData.CPU.Total)
	}
	if m.systemData.Memory != nil {
		m.history.AddMemory(m.systemData.Memory.UsedPercent)
		// Check memory alerts
		m.alertManager.CheckValue("memory", m.systemData.Memory.UsedPercent)
	}
	// Check temperature alerts
	if m.systemData.Sensors != nil && len(m.systemData.Sensors.Temperatures) > 0 {
		// Get the highest temperature
		maxTemp := 0.0
		for _, temp := range m.systemData.Sensors.Temperatures {
			if temp.Temperature > maxTemp {
				maxTemp = temp.Temperature
			}
		}
		m.alertManager.CheckValue("temperature", maxTemp)
	}

	// Update alert bar visibility
	hasAlerts := len(m.alertManager.GetActiveAlerts()) > 0
	if hasAlerts {
		m.alertBar.Show()
	} else {
		m.alertBar.Hide()
	}
}

// tickMsg is sent on each tick
type tickMsg time.Time

// tickCmd returns a command that sends tick messages
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// dataMsg wraps new system data
type dataMsg struct {
	data *data.SystemData
}
