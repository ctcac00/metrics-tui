package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

// NetworkMetrics holds network usage data
type NetworkMetrics struct {
	Interfaces  []net.InterfaceStat
	IO          map[string]net.IOCountersStat
	LastUpdate  time.Time
}

// NetworkCollector collects network metrics
type NetworkCollector struct {
	interval      uint
	interfaces    []string // Specific interfaces to monitor (empty = all)
	excludeVirtual bool
	mu            sync.RWMutex
	lastData      *NetworkMetrics
	lastIO        map[string]net.IOCountersStat
	lastIOTime    time.Time
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector(interval uint, interfaces []string, excludeVirtual bool) *NetworkCollector {
	return &NetworkCollector{
		interval:       interval,
		interfaces:     interfaces,
		excludeVirtual: excludeVirtual,
		lastIO:         make(map[string]net.IOCountersStat),
	}
}

// Name returns the collector name
func (c *NetworkCollector) Name() string {
	return "network"
}

// Interval returns the update interval in seconds
func (c *NetworkCollector) Interval() uint {
	return c.interval
}

// Collect gathers network metrics
func (c *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	// Get all interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Filter interfaces
	var filteredInterfaces []net.InterfaceStat
	var interfacesToMonitor []string

	for _, iface := range interfaces {
		// Skip virtual interfaces if requested
		if c.excludeVirtual && isVirtualInterface(iface.Name) {
			continue
		}

		// Skip interfaces with no addresses (down)
		if iface.Addrs == nil || len(iface.Addrs) == 0 {
			continue
		}

		if len(c.interfaces) == 0 {
			// Monitor all non-virtual interfaces
			filteredInterfaces = append(filteredInterfaces, iface)
			interfacesToMonitor = append(interfacesToMonitor, iface.Name)
		} else {
			// Check if this interface is in our list
			for _, target := range c.interfaces {
				if iface.Name == target {
					filteredInterfaces = append(filteredInterfaces, iface)
					interfacesToMonitor = append(interfacesToMonitor, iface.Name)
					break
				}
			}
		}
	}

	// Get IO counters (per NIC)
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network IO counters: %w", err)
	}

	// Filter IO counters to monitored interfaces
	ioMap := make(map[string]net.IOCountersStat)
	for _, io := range ioCounters {
		for _, name := range interfacesToMonitor {
			if io.Name == name {
				ioMap[name] = io
				break
			}
		}
	}

	metrics := &NetworkMetrics{
		Interfaces: filteredInterfaces,
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
func (c *NetworkCollector) GetLastData() *NetworkMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastData
}

// GetIORate calculates network IO rate since last collection (thread-safe)
func (c *NetworkCollector) GetIORate() map[string]NetIORate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.lastIO) == 0 {
		return nil
	}

	elapsed := time.Since(c.lastIOTime).Seconds()
	if elapsed == 0 {
		return nil
	}

	rates := make(map[string]NetIORate)
	for iface, currentIO := range c.lastIO {
		rates[iface] = NetIORate{
			BytesSentPerSec:   float64(currentIO.BytesSent) / elapsed,
			BytesRecvPerSec:   float64(currentIO.BytesRecv) / elapsed,
			PacketsSentPerSec: float64(currentIO.PacketsSent) / elapsed,
			PacketsRecvPerSec: float64(currentIO.PacketsRecv) / elapsed,
			ErrInPerSec:       float64(currentIO.Errin) / elapsed,
			ErrOutPerSec:      float64(currentIO.Errout) / elapsed,
		}
	}

	return rates
}

// isVirtualInterface checks if an interface is virtual
func isVirtualInterface(name string) bool {
	virtualPrefixes := []string{
		"veth", "docker", "br-", "virbr", "tun", "tap",
		"vnet", "kube", "flannel", "cali", "cni",
	}

	for _, prefix := range virtualPrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// NetIORate represents network IO rate between two samples
type NetIORate struct {
	BytesSentPerSec   float64
	BytesRecvPerSec   float64
	PacketsSentPerSec float64
	PacketsRecvPerSec float64
	ErrInPerSec       float64
	ErrOutPerSec      float64
}
