package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// CPUMetrics holds CPU usage data
type CPUMetrics struct {
	Usage      []float64 // Per-core usage percentage
	Total      float64   // Combined usage percentage
	CoreCount  int       // Number of logical cores
	Times      []cpu.TimesStat
	LastUpdate time.Time
}

// CPUCollector collects CPU metrics
type CPUCollector struct {
	interval uint
	mu       sync.RWMutex
	lastData *CPUMetrics
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector(interval uint) *CPUCollector {
	return &CPUCollector{
		interval: interval,
	}
}

// Name returns the collector name
func (c *CPUCollector) Name() string {
	return "cpu"
}

// Interval returns the update interval in seconds
func (c *CPUCollector) Interval() uint {
	return c.interval
}

// Collect gathers CPU metrics
func (c *CPUCollector) Collect(ctx context.Context) (interface{}, error) {
	// Get CPU counts (logical cores)
	cores, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU counts: %w", err)
	}

	// Get per-core and total usage
	percentages, err := cpu.Percent(time.Duration(c.interval)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percentages: %w", err)
	}

	// Calculate total usage from individual cores
	var total float64
	if len(percentages) > 0 {
		sum := 0.0
		for _, p := range percentages {
			sum += p
		}
		total = sum / float64(len(percentages))
	}

	// Get CPU times for more detailed info
	times, err := cpu.Times(true)
	if err != nil {
		// Times are optional, continue without them
		times = []cpu.TimesStat{}
	}

	metrics := &CPUMetrics{
		Usage:      percentages,
		Total:      total,
		CoreCount:  cores,
		Times:      times,
		LastUpdate: time.Now(),
	}

	c.mu.Lock()
	c.lastData = metrics
	c.mu.Unlock()

	return metrics, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *CPUCollector) GetLastData() *CPUMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}
