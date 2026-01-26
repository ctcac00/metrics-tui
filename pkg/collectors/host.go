package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
)

// HostMetrics holds host information
type HostMetrics struct {
	Info       host.InfoStat
	LoadAvg    *load.AvgStat
	LastUpdate time.Time
}

// HostCollector collects host information
type HostCollector struct {
	interval uint
	mu       sync.RWMutex
	lastData *HostMetrics
}

// NewHostCollector creates a new host collector
func NewHostCollector(interval uint) *HostCollector {
	return &HostCollector{
		interval: interval,
	}
}

// Name returns the collector name
func (c *HostCollector) Name() string {
	return "host"
}

// Interval returns the update interval in seconds
func (c *HostCollector) Interval() uint {
	return c.interval
}

// Collect gathers host metrics
func (c *HostCollector) Collect(ctx context.Context) (interface{}, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	loadAvg, err := load.Avg()
	if err != nil {
		// Load average might not be available on all systems
		loadAvg = &load.AvgStat{}
	}

	metrics := &HostMetrics{
		Info:       *info,
		LoadAvg:    loadAvg,
		LastUpdate: time.Now(),
	}

	c.mu.Lock()
	c.lastData = metrics
	c.mu.Unlock()

	return metrics, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *HostCollector) GetLastData() *HostMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}
