package collectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ctcac00/monitor-tui/internal/data"
)

// Aggregator manages multiple collectors and aggregates their data
type Aggregator struct {
	collectors      map[string]Collector
	data            map[string]any
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	updateInterval  time.Duration
	onDataUpdate    func(*data.SystemData)
}

// AggregatorConfig holds configuration for the aggregator
type AggregatorConfig struct {
	CPUInterval          uint
	MemoryInterval       uint
	DiskInterval         uint
	NetworkInterval      uint
	SensorsInterval      uint
	HostInterval         uint
	DiskPartitions       []string
	DiskIncludeAll       bool
	NetworkInterfaces    []string
	NetworkExcludeVirtual bool
}

// DefaultAggregatorConfig returns default configuration
func DefaultAggregatorConfig() *AggregatorConfig {
	return &AggregatorConfig{
		CPUInterval:          1,
		MemoryInterval:       2,
		DiskInterval:         5,
		NetworkInterval:      2,
		SensorsInterval:      5,
		HostInterval:         5,
		DiskIncludeAll:       true,
		NetworkExcludeVirtual: true,
	}
}

// NewAggregator creates a new aggregator with the given configuration
func NewAggregator(config *AggregatorConfig) *Aggregator {
	if config == nil {
		config = DefaultAggregatorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	agg := &Aggregator{
		collectors:     make(map[string]Collector),
		data:           make(map[string]any),
		ctx:            ctx,
		cancel:         cancel,
		updateInterval: 500 * time.Millisecond, // Check for updates twice per second
	}

	// Initialize collectors
	agg.collectors["cpu"] = NewCPUCollector(config.CPUInterval)
	agg.collectors["memory"] = NewMemoryCollector(config.MemoryInterval)
	agg.collectors["disk"] = NewDiskCollector(config.DiskInterval, config.DiskPartitions, config.DiskIncludeAll)
	agg.collectors["network"] = NewNetworkCollector(config.NetworkInterval, config.NetworkInterfaces, config.NetworkExcludeVirtual)
	agg.collectors["sensors"] = NewSensorsCollector(config.SensorsInterval)
	agg.collectors["host"] = NewHostCollector(config.HostInterval)

	return agg
}

// SetOnDataUpdate sets a callback function to be called when data is updated
func (a *Aggregator) SetOnDataUpdate(fn func(*data.SystemData)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onDataUpdate = fn
}

// Start begins concurrent collection from all collectors
func (a *Aggregator) Start() {
	for _, collector := range a.collectors {
		a.wg.Add(1)
		go a.startCollector(collector)
	}

	// Start update checker goroutine
	a.wg.Add(1)
	go a.updateChecker()
}

// Stop gracefully stops all collectors
func (a *Aggregator) Stop() {
	a.cancel()
	a.wg.Wait()
}

// startCollector runs a single collector in a loop
func (a *Aggregator) startCollector(collector Collector) {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(collector.Interval()) * time.Second)
	defer ticker.Stop()

	// Do initial collection
	a.collectFrom(collector)

	for {
		select {
		case <-ticker.C:
			a.collectFrom(collector)
		case <-a.ctx.Done():
			return
		}
	}
}

// collectFrom performs a single collection from a collector
func (a *Aggregator) collectFrom(collector Collector) {
	result, err := collector.Collect(a.ctx)
	if err != nil {
		log.Printf("[%s] Collection error: %v", collector.Name(), err)
		return
	}

	a.mu.Lock()
	a.data[collector.Name()] = result
	a.mu.Unlock()
}

// updateChecker periodically checks for data updates and triggers callbacks
func (a *Aggregator) updateChecker() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.notifyUpdate()
		case <-a.ctx.Done():
			return
		}
	}
}

// notifyUpdate triggers the data update callback with current data
func (a *Aggregator) notifyUpdate() {
	a.mu.RLock()
	onDataUpdate := a.onDataUpdate
	a.mu.RUnlock()

	if onDataUpdate != nil {
		systemData := a.GetSystemData()
		onDataUpdate(systemData)
	}
}

// convertCPUMetrics converts from collectors.CPUMetrics to data.CPUMetrics
func convertCPUMetrics(m *CPUMetrics) *data.CPUMetrics {
	if m == nil {
		return nil
	}
	return &data.CPUMetrics{
		Usage:      m.Usage,
		Total:      m.Total,
		CoreCount:  m.CoreCount,
		Times:      m.Times,
		LastUpdate: m.LastUpdate,
	}
}

// convertMemoryMetrics converts from collectors.MemoryMetrics to data.MemoryMetrics
func convertMemoryMetrics(m *MemoryMetrics) *data.MemoryMetrics {
	if m == nil {
		return nil
	}
	return &data.MemoryMetrics{
		Total:       m.Total,
		Available:   m.Available,
		Used:        m.Used,
		UsedPercent: m.UsedPercent,
		Free:        m.Free,
		Buffers:     m.Buffers,
		Cached:      m.Cached,
		Swap:        data.SwapMemoryStat(m.Swap),
		LastUpdate:  m.LastUpdate,
	}
}

// convertDiskMetrics converts from collectors.DiskMetrics to data.DiskMetrics
func convertDiskMetrics(m *DiskMetrics) *data.DiskMetrics {
	if m == nil {
		return nil
	}
	return &data.DiskMetrics{
		Partitions: m.Partitions,
		Usage:      m.Usage,
		IO:         m.IO,
		LastUpdate: m.LastUpdate,
	}
}

// convertNetworkMetrics converts from collectors.NetworkMetrics to data.NetworkMetrics
func convertNetworkMetrics(m *NetworkMetrics) *data.NetworkMetrics {
	if m == nil {
		return nil
	}
	return &data.NetworkMetrics{
		Interfaces: m.Interfaces,
		IO:         m.IO,
		LastUpdate: m.LastUpdate,
	}
}

// convertSensorMetrics converts from collectors.SensorMetrics to data.SensorMetrics
func convertSensorMetrics(m *SensorMetrics) *data.SensorMetrics {
	if m == nil {
		return nil
	}
	return &data.SensorMetrics{
		Temperatures: m.Temperatures,
		LastUpdate:   m.LastUpdate,
	}
}

// convertHostMetrics converts from collectors.HostMetrics to data.HostMetrics
func convertHostMetrics(m *HostMetrics) *data.HostMetrics {
	if m == nil {
		return nil
	}
	return &data.HostMetrics{
		Info:       m.Info,
		LoadAvg:    m.LoadAvg,
		LastUpdate: m.LastUpdate,
	}
}

// GetSystemData returns the current system data from all collectors
func (a *Aggregator) GetSystemData() *data.SystemData {
	a.mu.RLock()
	defer a.mu.RUnlock()

	systemData := &data.SystemData{
		Timestamp: time.Now(),
	}

	if cpuData, ok := a.data["cpu"].(*CPUMetrics); ok {
		systemData.CPU = convertCPUMetrics(cpuData)
	}
	if memData, ok := a.data["memory"].(*MemoryMetrics); ok {
		systemData.Memory = convertMemoryMetrics(memData)
	}
	if diskData, ok := a.data["disk"].(*DiskMetrics); ok {
		systemData.Disk = convertDiskMetrics(diskData)
	}
	if netData, ok := a.data["network"].(*NetworkMetrics); ok {
		systemData.Network = convertNetworkMetrics(netData)
	}
	if sensorData, ok := a.data["sensors"].(*SensorMetrics); ok {
		systemData.Sensors = convertSensorMetrics(sensorData)
	}
	if hostData, ok := a.data["host"].(*HostMetrics); ok {
		systemData.Host = convertHostMetrics(hostData)
	}

	return systemData
}

// GetCollector returns a collector by name
func (a *Aggregator) GetCollector(name string) (Collector, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	collector, ok := a.collectors[name]
	if !ok {
		return nil, fmt.Errorf("collector not found: %s", name)
	}

	return collector, nil
}

// ListCollectors returns all collector names
func (a *Aggregator) ListCollectors() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	names := make([]string, 0, len(a.collectors))
	for name := range a.collectors {
		names = append(names, name)
	}

	return names
}

// GetCPUCollector returns the CPU collector
func (a *Aggregator) GetCPUCollector() (*CPUCollector, error) {
	c, err := a.GetCollector("cpu")
	if err != nil {
		return nil, err
	}
	return c.(*CPUCollector), nil
}

// GetMemoryCollector returns the memory collector
func (a *Aggregator) GetMemoryCollector() (*MemoryCollector, error) {
	c, err := a.GetCollector("memory")
	if err != nil {
		return nil, err
	}
	return c.(*MemoryCollector), nil
}

// GetDiskCollector returns the disk collector
func (a *Aggregator) GetDiskCollector() (*DiskCollector, error) {
	c, err := a.GetCollector("disk")
	if err != nil {
		return nil, err
	}
	return c.(*DiskCollector), nil
}

// GetNetworkCollector returns the network collector
func (a *Aggregator) GetNetworkCollector() (*NetworkCollector, error) {
	c, err := a.GetCollector("network")
	if err != nil {
		return nil, err
	}
	return c.(*NetworkCollector), nil
}
