package components

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ctcac00/metrics-tui/internal/data"
)

// Snapshot represents a system state snapshot
type Snapshot struct {
	Timestamp   time.Time          `json:"timestamp"`
	CPU         *data.CPUMetrics  `json:"cpu"`
	Memory      *data.MemoryMetrics `json:"memory"`
	Disk        *data.DiskMetrics   `json:"disk"`
	Network     *data.NetworkMetrics `json:"network"`
	Sensors     *data.SensorMetrics `json:"sensors"`
	Host        *data.HostMetrics   `json:"host"`
}

// SnapshotManager handles snapshot operations
type SnapshotManager struct {
	outputDir string
	format    string // json, text
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(outputDir string, format string) *SnapshotManager {
	return &SnapshotManager{
		outputDir: outputDir,
		format:    format,
	}
}

// NewSnapshotManagerWithDefaults creates a snapshot manager with defaults
func NewSnapshotManagerWithDefaults() *SnapshotManager {
	homeDir, _ := os.UserHomeDir()
	return &SnapshotManager{
		outputDir: homeDir + "/snapshots",
		format:    "json",
	}
}

// TakeSnapshot captures the current system state
func (s *SnapshotManager) TakeSnapshot(systemData *data.SystemData) (*Snapshot, error) {
	snapshot := &Snapshot{
		Timestamp: time.Now(),
		CPU:       systemData.CPU,
		Memory:    systemData.Memory,
		Disk:      systemData.Disk,
		Network:   systemData.Network,
		Sensors:   systemData.Sensors,
		Host:      systemData.Host,
	}

	return snapshot, nil
}

// SaveToFile saves a snapshot to a file
func (s *SnapshotManager) SaveToFile(snapshot *Snapshot, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("monitor-snapshot-%s.%s",
			snapshot.Timestamp.Format("20060102-150405"),
			s.format,
		)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filepath := s.outputDir + "/" + filename

	var err error
	switch s.format {
	case "json":
		err = s.saveJSON(snapshot, filepath)
	case "text":
		err = s.saveText(snapshot, filepath)
	default:
		err = s.saveJSON(snapshot, filepath)
	}

	if err != nil {
		return err
	}

	fmt.Printf("Snapshot saved to: %s\n", filepath)
	return nil
}

// saveJSON saves snapshot as JSON
func (s *SnapshotManager) saveJSON(snapshot *Snapshot, filepath string) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// saveText saves snapshot as human-readable text
func (s *SnapshotManager) saveText(snapshot *Snapshot, filepath string) error {
	var content string

	content += fmt.Sprintf("Monitor TUI Snapshot\n")
	content += fmt.Sprintf("==================\n\n")
	content += fmt.Sprintf("Timestamp: %s\n\n", snapshot.Timestamp.Format("2006-01-02 15:04:05"))

	if snapshot.Host != nil {
		content += fmt.Sprintf("System: %s\n", snapshot.Host.Info.OS)
		content += fmt.Sprintf("Hostname: %s\n", snapshot.Host.Info.Hostname)
		content += fmt.Sprintf("Uptime: %s\n\n", formatUptime(snapshot.Host.Info.Uptime))
	}

	if snapshot.CPU != nil {
		content += "CPU Metrics\n"
		content += "------------\n"
		content += fmt.Sprintf("Total Usage: %.1f%%\n", snapshot.CPU.Total)
		content += fmt.Sprintf("Cores: %d\n\n", snapshot.CPU.CoreCount)
		for i, usage := range snapshot.CPU.Usage {
			content += fmt.Sprintf("  Core %d: %.1f%%\n", i, usage)
		}
	}

	if snapshot.Memory != nil {
		content += "\nMemory Metrics\n"
		content += "--------------\n"
		content += fmt.Sprintf("Total: %s\n", formatBytes(snapshot.Memory.Total))
		content += fmt.Sprintf("Used: %s (%.1f%%)\n", formatBytes(snapshot.Memory.Used), snapshot.Memory.UsedPercent)
		content += fmt.Sprintf("Available: %s\n\n", formatBytes(snapshot.Memory.Available))
	}

	if snapshot.Sensors != nil && len(snapshot.Sensors.Temperatures) > 0 {
		content += "\nTemperature Sensors\n"
		content += "------------------\n"
		for _, temp := range snapshot.Sensors.Temperatures {
			content += fmt.Sprintf("  %s: %.1f°C\n", temp.SensorKey, temp.Temperature)
		}
	}

	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// ExportCSV exports metrics history as CSV
func (s *SnapshotManager) ExportCSV(history map[string][]float64, filepath string) error {
	var content string

	// Header
	content += "timestamp"
	if cpuData, ok := history["cpu"]; ok {
		for i := range cpuData {
			content += fmt.Sprintf(",cpu_%d", i)
		}
	}
	if _, ok := history["memory"]; ok {
		content += ",memory"
	}
	content += "\n"

	// Data rows
	maxLen := 0
	for _, data := range history {
		if len(data) > maxLen {
			maxLen = len(data)
		}
	}

	for i := 0; i < maxLen; i++ {
		content += time.Now().Format(time.RFC3339)
		if cpuData, ok := history["cpu"]; ok && i < len(cpuData) {
			for _, val := range cpuData {
				content += fmt.Sprintf(",%.2f", val)
			}
		}
		if memData, ok := history["memory"]; ok && i < len(memData) {
			content += fmt.Sprintf(",%.2f", memData[i])
		}
		content += "\n"
	}

	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	fmt.Printf("CSV exported to: %s\n", filepath)
	return nil
}
