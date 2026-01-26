# Hardware Monitoring TUI - Implementation Plan

## Project Overview

**Goal**: Build a real-time terminal UI that displays system hardware metrics (CPU, memory, disk, network, temperatures, fans, etc.)

**Tech Stack**:

- **Bubble Tea** - TUI framework (Elm Architecture)
- **gopsutil** - Hardware data collection
- **Cobra** - CLI interface
- **Viper** - Configuration management
- **Lipgloss** - Styling
- **Bubbles** - UI components (spinner, progress, viewport, list, table)

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    Bubble Tea Model                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │          Data Collection Layer             │   │
│  │  ┌────────┬────────┬────────┬────────┐ │   │
│  │  │ CPU    │ Memory │ Disk  │ Net   │ │   │
│  │  │ Collector│       │ Collector│ Collector│ │   │
│  │  └────────┴────────┴────────┴────────┘ │   │
│  │                     ↓                        │   │
│  │              Aggregated Data               │   │
│  └───────────────────────────────────────────┘   │
│                     ↓                            │
│              View Rendering                    │
│  ┌──────────────────────────────────────────┐   │
│  │ Header │ Sidebar │ Metrics │ Footer │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

### Key Components

1. **Collectors** - Separate packages for each metric type
2. **Aggregator** - Combines data from all collectors
3. **Model** - Bubble Tea state management
4. **View** - Lipgloss-styled rendering
5. **Config** - Viper-based configuration
6. **CLI** - Cobra command structure

---

## File Structure

```
monitor-tui/
├── cmd/
│   └── root.go                    # Cobra CLI setup
├── pkg/
│   ├── collectors/                  # Data collection layer
│   │   ├── cpu.go
│   │   ├── memory.go
│   │   ├── disk.go
│   │   ├── network.go
│   │   ├── sensors.go
│   │   ├── host.go
│   │   └── collector.go             # Interface definition
│   ├── ui/                         # UI components
│   │   ├── model.go                   # Bubble Tea model
│   │   ├── update.go                  # Update logic
│   │   ├── view.go                   # View rendering
│   │   ├── styles.go                 # Lipgloss styles
│   │   ├── components/                # Reusable UI parts
│   │   │   ├── header.go
│   │   │   ├── footer.go
│   │   │   ├── sidebar.go
│   │   │   ├── metrics/
│   │   │   │   ├── cpu.go
│   │   │   │   ├── memory.go
│   │   │   │   ├── disk.go
│   │   │   │   ├── network.go
│   │   │   │   ├── temperature.go
│   │   │   │   └── fan.go
│   │   │   ├── progressbar.go
│   │   │   └── sparkline.go          # Mini history charts
│   └── config/                     # Configuration
│       ├── config.go                  # Config struct
│       └── defaults.go
├── internal/
│   └── data/                      # Data models
│       ├── system.go
│       └── history.go                # Time-series data
├── config.yaml.example                  # Example config
├── go.mod
├── go.sum
├── README.md
└── main.go
```

---

## gopsutil API Analysis

### Available Functions by Package

**CPU Package** (`github.com/shirou/gopsutil/v4/cpu`):

```go
cpu.Percent(interval, percpu bool)           // CPU usage % (per-core or total)
cpu.Counts(logical bool)                    // Number of cores
cpu.Times(percpu bool)                     // CPU time breakdown
cpu.Info()                                  // Detailed CPU info (model, freq, etc.)
```

**Memory Package** (`github.com/shirou/gopsutil/v4/mem`):

```go
mem.VirtualMemory()              // Total, used, free, available, %
mem.SwapMemory()                // Swap usage
mem.SwapDevices()              // Per-device swap
mem.NewExLinux().VirtualMemory() // Linux-specific extended stats
```

**Disk Package** (`github.com/shirou/gopsutil/v4/disk`):

```go
disk.Usage(path)                 // Usage for specific mount
disk.Partitions(all)             // List all partitions
disk.IOCounters(names)           // Read/write stats per device
disk.Label(name)                 // Device label
```

**Network Package** (`github.com/shirou/gopsutil/v4/net`):

```go
net.IOCounters(pernic)           // Per-interface I/O stats
net.Interfaces()                    // Interface list
net.Connections(kind)             // Network connections
```

**Sensors Package** (`github.com/shirou/gopsutil/v4/sensors`):

```go
sensors.SensorsTemperatures()    // Temp sensors with thresholds
sensors.NewExLinux().Temperature()  // Extended: min/max/historical
// Note: No direct fan speed API - may need lm-sensors fallback
```

**Host Package** (`github.com/shirou/gopsutil/v4/host`):

```go
host.Info()                          // Hostname, uptime, OS, platform, kernel
host.Uptime()                         // System uptime
host.KernelVersion()                 // Kernel version
host.Virtualization()                 // VM detection
```

**Load Package** (`github.com/shirou/gopsutil/v4/load`):

```go
load.Avg()                          // Load averages (1, 5, 15 min)
```

### Performance Considerations

- **Context support**: All functions have `*WithContext()` variants
- **Caching**: `host.EnableBootTimeCache(true)` available
- **Error handling**: Collectors should handle errors gracefully (fallback to "N/A")
- **Update intervals**:
  - CPU/Mem/Disk/Net: 1-5 seconds
  - Sensors: 2-10 seconds
  - Load avg: 5-15 seconds

---

## Configuration (Viper)

### Config Structure

```yaml
# config.yaml
refresh:
  interval: 2s # Default update interval
  cpu: 1s # CPU specific
  memory: 2s
  disk: 5s
  network: 2s
  sensors: 5s

display:
  theme: auto # auto, dark, light
  showGraphs: true # Mini sparklines
  showPercentages: true
  precision: 1 # Decimal places
  units: auto # auto, binary, decimal

thresholds:
  cpu-warning: 70
  cpu-critical: 90
  memory-warning: 80
  memory-critical: 95
  temp-warning: 70
  temp-critical: 85

ui:
  pageSize: 50 # History length for sparklines
  showLoadAverage: true
  showUptime: true
  showHostname: true
```

### CLI Flags (Cobra)

```bash
monitor-tui [flags]

Flags:
  --config string        # Config file path (default: ~/.config/monitor-tui/config.yaml)
  --refresh duration    # Override refresh interval
  --theme string        # Force theme (dark/light)
  --no-graphs           # Disable sparklines
  --interval duration    # Set specific interval
  --list-disks         # Show available disks and exit
  --debug               # Enable debug logging
```

---

## Data Collection Strategy

### Collector Interface

```go
type Collector interface {
    Name() string
    Collect(ctx context.Context) (interface{}, error)
    Interval() time.Duration
}
```

### Concurrent Collection Pattern

```go
type Aggregator struct {
    collectors map[string]Collector
    data      map[string]interface{}
    mu        sync.RWMutex
}

func (a *Aggregator) Start(ctx context.Context) {
    for name, collector := range a.collectors {
        go func(c Collector) {
            ticker := time.NewTicker(c.Interval())
            for {
                select {
                case <-ticker.C:
                    if data, err := c.Collect(ctx); err == nil {
                        a.mu.Lock()
                        a.data[c.Name()] = data
                        a.mu.Unlock()
                    }
                case <-ctx.Done():
                    return
                }
            }
        }(collector)
    }
}
```

### Metric Data Models

```go
type CPUMetrics struct {
    Usage    []float64          // Per-core usage
    Total    float64           // Combined usage
    Times    []cpu.TimesStat   // Time breakdown
}

type MemoryMetrics struct {
    Total       uint64
    Used        uint64
    Available   uint64
    UsedPercent float64
    Swap        SwapMemoryStat
}

type DiskMetrics struct {
    Partitions []disk.PartitionStat
    Usage      map[string]disk.UsageStat
    IO         map[string]disk.IOCountersStat
}

type NetworkMetrics struct {
    Interfaces  []net.InterfaceStat
    IO         map[string]net.IOCountersStat
}

type SensorMetrics struct {
    Temperatures []sensors.TemperatureStat
}

type SystemData struct {
    CPUMetrics    CPUMetrics
    MemoryMetrics  MemoryMetrics
    DiskMetrics    DiskMetrics
    NetworkMetrics NetworkMetrics
    SensorMetrics  SensorMetrics
    LoadAvg       load.AvgStat
    HostInfo      host.InfoStat
    Timestamp     time.Time
}
```

---

## UI Design

### Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ Hostname | Uptime: 2d 3h 45m | Load: 1.2 1.5 1.8  │ ← Header
├──────┬───────────────────────────────────────────────────────────────┤
│      │                                                       │
│ CPU   │  CPU Usage                                              │
│ MEM   │  ▓▓▓▓▓▓▓▓░░░░  75%                             │
│ DISK  │  ▒▒▓▒▓▒▓▒▒▒▓  82% (32.1/39.1 GB)              │
│ NET   │  ▓▓▓▓▓▓▓░░░░░  65% (6.5/10.0 GB)              │
│ TEMP  │                                                        │
│ FAN   │  Core Usage                                             │
│ LOAD  │  ▓▓░░░░░░░░░░░░  42%  ▓▓░░░░░░░░░░░  38%     │
│       │  ▓▓▓░░░░░░░░░░░  31%  ▓▓▓░░░░░░░░░░  28%     │
│      │  sparkline: ▂▃▄▅▆▇█▇▆▅▄▃▂▃▄▅▆▇█▇▆▅     │
│      │                                                        │
│      │  Memory Usage                                          │
│      │  Total: 16.0 GB  Used: 12.8 GB  Avail: 3.2 GB     │
│      │  ▓▓▓▓▓▓▓▓▓░░░░░  80%                      │
│      │  sparkline: ▆▇██████████████▇▆▅▄▂            │
│      │                                                        │
│      │  Disk I/O (10s)                                     │
│      │  sda: ↓ 15.2 MB/s  ↑ 2.1 MB/s                  │
│      │  sdb: ↓ 0.0 MB/s   ↑ 0.0 MB/s                   │
│      │                                                        │
│      │  Network I/O (10s)                                  │
│      │  eth0: ↓ 125 KB/s  ↑ 23 KB/s                   │
│      │  wlan0: ↓ 2.3 MB/s  ↑ 145 KB/s                │
│      │                                                        │
│      │  Temperatures                                         │
│      │  cpu_thermal:  65°C (critical: 85°C)            │
│      │  acpitz:  62°C                                       │
│      │  coretemp:  68°C / 67°C / 66°C                    │
│      │                                                        │
│      │  Load Average                                       │
│      │  1 min: 1.23   5 min: 1.45   15 min: 1.51     │
│      │                                                        │
└──────┴───────────────────────────────────────────────────────────────┘
│ [q] quit [h] help [1-8] select panel [↑↓] scroll     │ ← Footer
└─────────────────────────────────────────────────────────────────────┘
```

### Bubbles Components to Use

- **spinner** - For initial data loading
- **progress** - For metric bars (CPU, memory, etc.)
- **viewport** - For scrolling long lists
- **list** - For navigation between metric panels
- **table** - For structured data (disk partitions, network interfaces)

### Lipgloss Styling

```go
type Styles struct {
    base       lipgloss.Style
    header     lipgloss.Style
    activeTab  lipgloss.Style
    inactiveTab lipgloss.Style
    normal     lipgloss.Style
    warning    lipgloss.Style
    critical   lipgloss.Style
    success    lipgloss.Style
}

var Styles = Styles{
    base:       lipgloss.NewStyle(),
    header:     lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7")).Bold(true),
    activeTab:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#3d8fd8")),
    warning:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b")),
    critical:   lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Bold(true),
    success:    lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")),
}
```

---

## Bubble Tea Implementation

### Message Types

```go
type tickMsg time.Time
type dataMsg struct{ data SystemData }
type errMsg  struct{ err error }

type Model struct {
    data      SystemData
    config    Config
    styles    Styles
    activeTab int
    history   map[string][]float64    // For sparklines
    spinner   spinner.Model
    quitting  bool
}
```

### Update Loop

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.quitting = true
            return m, tea.Quit
        case "1", "2", "3", "4", "5", "6", "7", "8":
            m.activeTab = int(msg.String()[0]) - '1'
        case "h", "?":
            return m, nil  // Show help
        }

    case tickMsg:
        // Re-schedule tick
        return m, tickCmd()

    case dataMsg:
        m.data = msg.data
        // Update history for sparklines
        for _, name := range []string{"cpu", "memory", "disk", "network"} {
            m.history[name] = append(m.history[name][1:], getValue(name, msg.data))
        }

    case errMsg:
        // Log error but continue
        log.Printf("Collection error: %v", msg.err)
    }

    return m, nil
}

func tickCmd() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### View Rendering

```go
func (m Model) View() string {
    if m.quitting {
        return "\n  Goodbye!\n\n"
    }

    return lipgloss.JoinVertical(
        m.renderHeader(),
        m.renderSidebar(),
        m.renderMainContent(),
        m.renderFooter(),
    )
}
```

---

## Implementation Phases

### Phase 1: Project Setup (Day 1)

- [ ] Initialize Go module
- [ ] Set up directory structure
- [ ] Add dependencies (bubbletea, gopsutil, cobra, viper, lipgloss, bubbles)
- [ ] Create `main.go` with basic Bubble Tea boilerplate
- [ ] Set up Cobra CLI skeleton

### Phase 2: Data Collectors (Days 2-3)

- [ ] Create collector interface
- [ ] Implement CPU collector
- [ ] Implement memory collector
- [ ] Implement disk collector
- [ ] Implement network collector
- [ ] Implement sensors collector
- [ ] Implement host info collector
- [ ] Create aggregator with concurrent collection
- [ ] Add error handling and fallbacks

### Phase 3: Basic TUI (Days 4-5)

- [ ] Create basic Bubble Tea model
- [ ] Implement tab-based navigation
- [ ] Create header component
- [ ] Create footer component
- [ ] Implement simple metric displays (text-based)
- [ ] Add quit handling

### Phase 4: Visual Enhancements (Days 6-7)

- [ ] Create Lipgloss style system
- [ ] Implement progress bars (bubbles/progress)
- [ ] Add sparkline component for history visualization
- [ ] Add color thresholds (normal/warning/critical)
- [ ] Implement sidebar for metric selection
- [ ] Create metric detail views

### Phase 5: Configuration (Days 8-9)

- [ ] Integrate Viper for config loading
- [ ] Create config struct and validation
- [ ] Implement config file (YAML)
- [ ] Add CLI flags with Cobra
- [ ] Support multiple config sources (file, flags, env)
- [ ] Create example config file

### Phase 6: Advanced Features (Days 10-12)

- [ ] Add process list view (optional)
- [ ] Implement disk usage per partition
- [ ] Add network connection view
- [ ] Implement export/snapshot feature
- [ ] Add alerting (threshold breaches)
- [ ] Create help screen
- [ ] Add debug mode

### Phase 7: Polish & Testing (Days 13-14)

- [ ] Performance testing (ensure low CPU usage)
- [ ] Test on different terminals
- [ ] Test on Linux, macOS, Windows
- [ ] Add error recovery
- [ ] Write README with examples
- [ ] Create screenshots/demos

---

## Testing Strategy

### Unit Tests

```go
// pkg/collectors/cpu_test.go
func TestCPUCollector(t *testing.T) {
    collector := NewCPUCollector()
    data, err := collector.Collect(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, data)
    assert.Greater(t, data.Total, 0.0)
}
```

### Integration Tests

```go
// Test actual gopsutil calls
func TestCollectorsIntegration(t *testing.T) {
    // Test that all collectors work together
    agg := NewAggregator([]Collector{...})
    ctx := context.Background()
    data := agg.CollectAll(ctx)
    assert.NoError(t, data.Error)
}
```

### Manual Testing Checklist

- [ ] Update rate is reasonable (1-5 seconds)
- [ ] Memory usage stays stable (no leaks)
- [ ] Resize handling works
- [ ] Terminal size limits are handled
- [ ] Errors are displayed gracefully
- [ ] Quit works (q, ctrl+c)
- [ ] Config reloads correctly
- [ ] Works with common terminal emulators

---

## Dependencies

```go
// go.mod
module github.com/yourname/monitor-tui

go 1.21

require (
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/charmbracelet/bubbles v0.21.0
    github.com/shirou/gopsutil/v4 v4.25.12
    github.com/spf13/cobra v1.10.2
    github.com/spf13/viper v1.21.0
    github.com/muesli/reflow v1.0.0
    gopkg.in/yaml.v3 v3.0.1
)
```

---

## Potential Challenges & Solutions

### Challenge 1: gopsutil Performance

**Issue**: Frequent gopsutil calls may impact system performance
**Solution**:

- Use appropriate update intervals (don't update sensors every 100ms)
- Cache expensive operations (host info, static hardware info)
- Use context for cancellation

### Challenge 2: Terminal Size Handling

**Issue**: TUI breaks on resize
**Solution**:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    return m, nil
```

### Challenge 3: Fan Speeds

**Issue**: gopsutil doesn't provide fan speeds directly
**Solutions**:

1. Parse `/sys/class/hwmon/` on Linux
2. Use `lm-sensors` output (requires subprocess)
3. Document as "not available" on non-Linux

### Challenge 4: Cross-Platform Differences

**Issue**: Some sensors only available on Linux
**Solution**:

```go
func collectTemps() []sensors.TemperatureStat {
    temps, err := sensors.SensorsTemperatures()
    if err != nil {
        log.Printf("Temp sensors unavailable: %v", err)
        return []sensors.TemperatureStat{}  // Return empty slice
    }
    return temps
}
```

---

## Future Enhancements

### V2 Features

- [ ] Historical data export (CSV, JSON)
- [ ] Alert thresholds with sound notifications
- [ ] Remote monitoring (server mode)
- [ ] Process list with sorting/filtering
- [ ] Customizable dashboard layouts
- [ ] Plugin system for custom collectors
- [ ] GPU monitoring (NVIDIA/AMD)
- [ ] Battery status for laptops
- [ ] Dark/light theme auto-detection
- [ ] Internationalization (i18n)

---

## Example Usage

```bash
# Install
go install github.com/yourname/monitor-tui@latest

# Run with defaults
monitor-tui

# Custom refresh interval
monitor-tui --refresh 5s

# Use specific config
monitor-tui --config /path/to/config.yaml

# List available disks and exit
monitor-tui --list-disks

# Debug mode
monitor-tui --debug
```

---

## Summary

This plan provides:

- ✅ **Clear architecture** with separation of concerns
- ✅ **Comprehensive gopsutil integration** for all major metrics
- ✅ **Real-time updates** using Bubble Tea's tick mechanism
- ✅ **Beautiful UI** with Lipgloss styling and sparkline visualizations
- ✅ **Flexible configuration** via Viper and Cobra
- ✅ **Production-ready** with error handling and testing strategy

**Estimated Timeline**: 2 weeks (14 days) to MVP
**Lines of Code**: ~2500-3000 lines (excluding tests)
