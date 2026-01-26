package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/sensors"
)

// SensorMetrics holds sensor data (temperatures)
type SensorMetrics struct {
	Temperatures []sensors.TemperatureStat
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

	metrics := &SensorMetrics{
		Temperatures: temps,
		LastUpdate:   time.Now(),
	}

	c.mu.Lock()
	c.lastData = metrics
	c.mu.Unlock()

	return metrics, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *SensorsCollector) GetLastData() *SensorMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}
