package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

// SwapMemoryStat holds swap memory information
type SwapMemoryStat struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// MemoryMetrics holds memory usage data
type MemoryMetrics struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Free        uint64
	Buffers     uint64 // Linux-specific
	Cached      uint64 // Linux-specific
	Swap        SwapMemoryStat
	LastUpdate  time.Time
}

// MemoryCollector collects memory metrics
type MemoryCollector struct {
	interval uint
	mu       sync.RWMutex
	lastData *MemoryMetrics
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector(interval uint) *MemoryCollector {
	return &MemoryCollector{
		interval: interval,
	}
}

// Name returns the collector name
func (c *MemoryCollector) Name() string {
	return "memory"
}

// Interval returns the update interval in seconds
func (c *MemoryCollector) Interval() uint {
	return c.interval
}

// Collect gathers memory metrics
func (c *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	swapMem, err := mem.SwapMemory()
	if err != nil {
		// Swap stats are optional, continue without them
		swapMem = &mem.SwapMemoryStat{}
	}

	metrics := &MemoryMetrics{
		Total:       vmem.Total,
		Available:   vmem.Available,
		Used:        vmem.Used,
		UsedPercent: vmem.UsedPercent,
		Free:        vmem.Free,
		Swap: SwapMemoryStat{
			Total:       swapMem.Total,
			Used:        swapMem.Used,
			Free:        swapMem.Free,
			UsedPercent: swapMem.UsedPercent,
		},
		LastUpdate: time.Now(),
	}

	// Try to get extended stats (buffers/cached) on Linux
	if vmem.SwapCached > 0 {
		metrics.Cached = vmem.SwapCached
	}

	c.mu.Lock()
	c.lastData = metrics
	c.mu.Unlock()

	return metrics, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *MemoryCollector) GetLastData() *MemoryMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}
