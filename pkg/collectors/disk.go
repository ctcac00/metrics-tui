package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
)

// DiskMetrics holds disk usage data
type DiskMetrics struct {
	Partitions []disk.PartitionStat
	Usage      map[string]disk.UsageStat
	IO         map[string]disk.IOCountersStat
	LastUpdate time.Time
}

// DiskCollector collects disk metrics
type DiskCollector struct {
	interval     uint
	partitions   []string // Specific partitions to monitor
	includeAll   bool
	mu           sync.RWMutex
	lastData     *DiskMetrics
	lastIO       map[string]disk.IOCountersStat
	lastIOTime   time.Time
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector(interval uint, partitions []string, includeAll bool) *DiskCollector {
	return &DiskCollector{
		interval:   interval,
		partitions: partitions,
		includeAll: includeAll,
		lastIO:     make(map[string]disk.IOCountersStat),
	}
}

// Name returns the collector name
func (c *DiskCollector) Name() string {
	return "disk"
}

// Interval returns the update interval in seconds
func (c *DiskCollector) Interval() uint {
	return c.interval
}

// Collect gathers disk metrics
func (c *DiskCollector) Collect(ctx context.Context) (interface{}, error) {
	// Get all partitions
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	// Filter partitions based on configuration
	var filteredPartitions []disk.PartitionStat
	var devicesToMonitor []string

	for _, p := range partitions {
		// Skip non-physical filesystems
		if p.Fstype == "squashfs" || p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" ||
			p.Fstype == "proc" || p.Fstype == "sysfs" || p.Fstype == "cgroup" ||
			p.Fstype == "securityfs" || p.Fstype == "debugfs" {
			continue
		}

		if c.includeAll {
			filteredPartitions = append(filteredPartitions, p)
			devicesToMonitor = append(devicesToMonitor, p.Mountpoint)
		} else {
			// Check if this partition is in our list
			for _, target := range c.partitions {
				if p.Mountpoint == target || p.Device == target {
					filteredPartitions = append(filteredPartitions, p)
					devicesToMonitor = append(devicesToMonitor, p.Mountpoint)
					break
				}
			}
		}
	}

	// Get usage for each partition
	usageMap := make(map[string]disk.UsageStat)
	for _, p := range filteredPartitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			// Skip partitions we can't read
			continue
		}
		usageMap[p.Mountpoint] = *usage
	}

	// Get IO counters
	ioCounters, err := disk.IOCounters()
	if err != nil {
		// IO counters might not be available on all systems
		ioCounters = make(map[string]disk.IOCountersStat)
	}

	// Store IO data
	ioMap := make(map[string]disk.IOCountersStat)
	for device, stats := range ioCounters {
		ioMap[device] = stats
	}

	metrics := &DiskMetrics{
		Partitions: filteredPartitions,
		Usage:      usageMap,
		IO:         ioMap,
		LastUpdate: time.Now(),
	}

	c.mu.Lock()
	c.lastData = metrics
	c.lastIO = ioMap
	c.lastIOTime = time.Now()
	c.mu.Unlock()

	return metrics, nil
}

// GetLastData returns the last collected data (thread-safe)
func (c *DiskCollector) GetLastData() *DiskMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}

// GetIORate calculates IO rate since last collection (thread-safe)
func (c *DiskCollector) GetIORate() map[string]IORate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.lastIO) == 0 {
		return nil
	}

	elapsed := time.Since(c.lastIOTime).Seconds()
	if elapsed == 0 {
		return nil
	}

	rates := make(map[string]IORate)
	for device, currentIO := range c.lastIO {
		if lastIO, ok := c.lastIO[device]; ok {
			rates[device] = IORate{
				ReadBytesPerSec:  float64(currentIO.ReadBytes-lastIO.ReadBytes) / elapsed,
				WriteBytesPerSec: float64(currentIO.WriteBytes-lastIO.WriteBytes) / elapsed,
				ReadCountPerSec:  float64(currentIO.ReadCount-lastIO.ReadCount) / elapsed,
				WriteCountPerSec: float64(currentIO.WriteCount-lastIO.WriteCount) / elapsed,
			}
		}
	}

	return rates
}

// IORate represents IO rates between two samples
type IORate struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	ReadCountPerSec  float64
	WriteCountPerSec float64
}
