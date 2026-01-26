package data

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

// CPUMetrics holds CPU usage data
type CPUMetrics struct {
	Usage      []float64
	Total      float64
	CoreCount  int
	Times      []cpu.TimesStat
	LastUpdate time.Time
}

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
	Buffers     uint64
	Cached      uint64
	Swap        SwapMemoryStat
	LastUpdate  time.Time
}

// IORate represents IO rates between two samples
type IORate struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
	ReadCountPerSec  float64
	WriteCountPerSec float64
}

// DiskMetrics holds disk usage data
type DiskMetrics struct {
	Partitions []disk.PartitionStat
	Usage      map[string]disk.UsageStat
	IO         map[string]disk.IOCountersStat
	LastUpdate time.Time
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

// NetworkMetrics holds network usage data
type NetworkMetrics struct {
	Interfaces []net.InterfaceStat
	IO         map[string]net.IOCountersStat
	LastUpdate time.Time
}

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

// HostMetrics holds host information
type HostMetrics struct {
	Info       host.InfoStat
	LoadAvg    *load.AvgStat
	LastUpdate time.Time
}

// SystemData aggregates all system metrics
type SystemData struct {
	CPU       *CPUMetrics
	Memory    *MemoryMetrics
	Disk      *DiskMetrics
	Network   *NetworkMetrics
	Sensors   *SensorMetrics
	Host      *HostMetrics
	Timestamp time.Time
	Error     error
}

// HistoryData holds historical data for sparklines
type HistoryData struct {
	CPU     []float64
	Memory  []float64
	Network RxTxHistory
	Disk    RWHistory
	maxSize int
}

// RxTxHistory tracks network receive/transmit history
type RxTxHistory struct {
	Rx []float64
	Tx []float64
}

// RWHistory tracks disk read/write history
type RWHistory struct {
	Read  []float64
	Write []float64
}

// NewHistoryData creates a new history tracker
func NewHistoryData(maxSize int) *HistoryData {
	return &HistoryData{
		CPU:     make([]float64, 0, maxSize),
		Memory:  make([]float64, 0, maxSize),
		Network: RxTxHistory{Rx: make([]float64, 0, maxSize), Tx: make([]float64, 0, maxSize)},
		Disk:    RWHistory{Read: make([]float64, 0, maxSize), Write: make([]float64, 0, maxSize)},
		maxSize: maxSize,
	}
}

// AddCPU adds a CPU usage value to history
func (h *HistoryData) AddCPU(value float64) {
	h.CPU = h.appendAndTrim(h.CPU, value)
}

// AddMemory adds a memory usage value to history
func (h *HistoryData) AddMemory(value float64) {
	h.Memory = h.appendAndTrim(h.Memory, value)
}

// AddNetworkRx adds a network receive value to history
func (h *HistoryData) AddNetworkRx(value float64) {
	h.Network.Rx = h.appendAndTrim(h.Network.Rx, value)
}

// AddNetworkTx adds a network transmit value to history
func (h *HistoryData) AddNetworkTx(value float64) {
	h.Network.Tx = h.appendAndTrim(h.Network.Tx, value)
}

// AddDiskRead adds a disk read value to history
func (h *HistoryData) AddDiskRead(value float64) {
	h.Disk.Read = h.appendAndTrim(h.Disk.Read, value)
}

// AddDiskWrite adds a disk write value to history
func (h *HistoryData) AddDiskWrite(value float64) {
	h.Disk.Write = h.appendAndTrim(h.Disk.Write, value)
}

// appendAndTrim adds a value to a slice and keeps it at maxSize
func (h *HistoryData) appendAndTrim(slice []float64, value float64) []float64 {
	slice = append(slice, value)
	if len(slice) > h.maxSize {
		slice = slice[1:]
	}
	return slice
}

// GetLatestCPU returns the most recent CPU usage
func (h *HistoryData) GetLatestCPU() float64 {
	if len(h.CPU) == 0 {
		return 0
	}
	return h.CPU[len(h.CPU)-1]
}

// GetLatestMemory returns the most recent memory usage
func (h *HistoryData) GetLatestMemory() float64 {
	if len(h.Memory) == 0 {
		return 0
	}
	return h.Memory[len(h.Memory)-1]
}

// GetLatestNetworkRx returns the most recent network receive rate
func (h *HistoryData) GetLatestNetworkRx() float64 {
	if len(h.Network.Rx) == 0 {
		return 0
	}
	return h.Network.Rx[len(h.Network.Rx)-1]
}

// GetLatestNetworkTx returns the most recent network transmit rate
func (h *HistoryData) GetLatestNetworkTx() float64 {
	if len(h.Network.Tx) == 0 {
		return 0
	}
	return h.Network.Tx[len(h.Network.Tx)-1]
}
