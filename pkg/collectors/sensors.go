package collectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/sensors"
)

// FanStat holds fan speed data
type FanStat struct {
	Name string
	RPM  uint64
}

// SensorMetrics holds sensor data (temperatures and fans)
type SensorMetrics struct {
	Temperatures []sensors.TemperatureStat
	Fans         []FanStat
	LastUpdate   time.Time
}

// SensorsCollector collects sensor metrics
type SensorsCollector struct {
	interval uint
	mu       sync.RWMutex
	lastData *SensorMetrics
}

// NewSensorsCollector creates a new sensors collector
func NewSensorsCollector(interval uint) *SensorsCollector {
	return &SensorsCollector{
		interval: interval,
	}
}

// Name returns the collector name
func (c *SensorsCollector) Name() string {
	return "sensors"
}

// Interval returns the update interval in seconds
func (c *SensorsCollector) Interval() uint {
	return c.interval
}

// Collect gathers sensor metrics
func (c *SensorsCollector) Collect(ctx context.Context) (interface{}, error) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature sensors: %w", err)
	}

	// Filter to only the most useful temperature sensors
	filteredTemps := filterUsefulTemperatures(temps)

	// Collect fan speeds from hwmon
	fans, err := collectFanSpeeds()
	if err != nil {
		// Don't fail entirely if fans can't be read, just log it
		fans = nil
	}

	metrics := &SensorMetrics{
		Temperatures: filteredTemps,
		Fans:         fans,
		LastUpdate:   time.Now(),
	}

	c.mu.Lock()
	c.lastData = metrics
	c.mu.Unlock()

	return metrics, nil
}

// filterUsefulTemperatures selects the most useful temperature sensors
func filterUsefulTemperatures(temps []sensors.TemperatureStat) []sensors.TemperatureStat {
	// Priority prefixes for sensors we want to show
	priorityPrefixes := []string{
		"coretemp",      // Intel CPU cores
		"k10temp",       // AMD CPU
		"cpu",           // Generic CPU
		"nvidia",        // NVIDIA GPU
		"amdgpu",        // AMD GPU
		"radeon",        // AMD GPU (older)
		"iwlwifi",       // Intel WiFi (can overheat)
		"BAT",           // Battery temps (laptops)
		"acpitz",        // ACPI thermal zone
		"soc_thermal",   // SoC temperature
		"gpu",           // Generic GPU
	}

	// Low priority prefixes (less useful)
	lowPriorityPrefixes := []string{
		"aux",
		"intrusion",
		" intrusion",
	}

	var result []sensors.TemperatureStat
	priorityCount := make(map[string]int)

	// First pass: add priority sensors (limited per type)
	for _, temp := range temps {
		key := strings.ToLower(temp.SensorKey)
		matched := false

		// Check priority prefixes
		for _, prefix := range priorityPrefixes {
			if strings.HasPrefix(key, prefix) {
				// Limit to 8 sensors per type to avoid clutter
				if priorityCount[prefix] < 8 {
					result = append(result, temp)
					priorityCount[prefix]++
					matched = true
				}
				break
			}
		}

		// Skip low priority sensors
		if !matched {
			for _, lowPrefix := range lowPriorityPrefixes {
				if strings.HasPrefix(key, lowPrefix) {
					matched = true
					break
				}
			}
		}
	}

	return result
}

// collectFanSpeeds reads fan speeds from hwmon sysfs
func collectFanSpeeds() ([]FanStat, error) {
	var fans []FanStat

	hwmonPath := "/sys/class/hwmon"
	entries, err := os.ReadDir(hwmonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read hwmon directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "hwmon") {
			continue
		}

		devicePath := filepath.Join(hwmonPath, entry.Name())
		name, err := readDeviceName(devicePath)
		if err != nil {
			continue
		}

		// Read fan inputs
		deviceFans, err := readFanInputs(devicePath, name)
		if err != nil {
			continue
		}
		fans = append(fans, deviceFans...)
	}

	return fans, nil
}

// readDeviceName reads the device name from hwmon
func readDeviceName(devicePath string) (string, error) {
	namePath := filepath.Join(devicePath, "name")
	nameData, err := os.ReadFile(namePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(nameData)), nil
}

// readFanInputs reads all fan*_input files for a device
func readFanInputs(devicePath string, deviceName string) ([]FanStat, error) {
	var fans []FanStat

	entries, err := os.ReadDir(devicePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "fan") || !strings.HasSuffix(entry.Name(), "_input") {
			continue
		}

		fanPath := filepath.Join(devicePath, entry.Name())
		rpmData, err := os.ReadFile(fanPath)
		if err != nil {
			continue
		}

		rpm, err := strconv.ParseUint(strings.TrimSpace(string(rpmData)), 10, 64)
		if err != nil {
			continue
		}

		// Skip fans that report 0 RPM (likely not connected or not spinning)
		if rpm == 0 {
			continue
		}

		// Extract fan number from filename (e.g., "fan1_input" -> "1")
		fanNum := strings.TrimPrefix(entry.Name(), "fan")
		fanNum = strings.TrimSuffix(fanNum, "_input")

		fanName := fmt.Sprintf("%s_fan%s", deviceName, fanNum)
		fans = append(fans, FanStat{
			Name: fanName,
			RPM:  rpm,
		})
	}

	return fans, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *SensorsCollector) GetLastData() *SensorMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}
