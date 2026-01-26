# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Metrics TUI is a real-time terminal user interface for system hardware monitoring built in Go. It displays live metrics for CPU, memory, disk, network, temperature sensors, and more in a visually appealing dashboard using the Bubble Tea framework.

## Build and Run Commands

```bash
# Build the binary
go build -o metrics-tui

# Run directly (without building)
go run main.go

# Run with flags
./metrics-tui --debug          # Test all collectors
./metrics-tui --list-disks     # List available disk partitions
./metrics-tui --refresh 5s     # Set refresh interval
./metrics-tui --help           # Show all options

# Clean and rebuild
rm -f metrics-tui && go build -o metrics-tui

# Update dependencies
go mod tidy
```

## Architecture Overview

### Three-Layer Architecture

1. **Data Collection Layer** (`pkg/collectors/`)
   - Independent collectors for each metric type (CPU, memory, disk, network, sensors, host)
   - All collectors implement the `Collector` interface with `Name()`, `Collect(ctx)`, and `Interval()` methods
   - The `Aggregator` runs all collectors concurrently in separate goroutines with individual update intervals
   - Each collector has its own ticker and runs at its configured interval (e.g., CPU: 1s, Disk: 5s)

2. **Data Models** (`internal/data/`)
   - `SystemData` aggregates all metrics in a single structure
   - `HistoryData` maintains circular buffers for sparkline visualizations
   - Metrics structs are separate from gopsutil types to allow additional fields like `LastUpdate`

3. **UI Layer** (`pkg/ui/`)
   - Built with Bubble Tea (Elm Architecture): Model, Update, View pattern
   - `Model` in `pkg/ui/model.go` is the main Bubble Tea component that owns the aggregator
   - Dashboard uses a 3-column layout: CPU | Temperature | (Memory stacked on Network)
   - Individual metric components in `pkg/ui/components/metrics/` handle rendering of each metric type

### Data Flow

```
Collectors (goroutines) → Aggregator → SystemData → Model.onDataUpdate() → UI Components → Dashboard.Render()
                           ↓
                    History tracking in Model.updateHistory()
                           ↓
                    Alert checking via AlertManager
```

### Key Design Patterns

**Concurrent Collection with Aggregator Pattern:**
- Each collector runs independently in its own goroutine with a ticker
- The `Aggregator` manages all collectors and provides thread-safe access via RWMutex
- `updateChecker()` goroutine polls every 500ms and triggers `onDataUpdate` callbacks
- This decouples data collection frequency from UI update frequency

**Type Conversion Boundary:**
- Collectors use their own metric types (e.g., `collectors.CPUMetrics`)
- Aggregator converts to internal types (e.g., `data.CPUMetrics`) when assembling `SystemData`
- This allows collectors to be independently developed and tested

**Bubble Tea Message Flow:**
- `tickMsg` every 2 seconds triggers history updates and alert checking
- `dataMsg` carries new SystemData from aggregator callback
- `tea.WindowSizeMsg` triggers layout recalculation for responsive design

## Important Implementation Details

### Sensor Collection (Linux-specific)
- Temperature sensors are filtered by priority (CPU cores first, GPU second, then ACPI)
- Fan speeds are read directly from `/sys/class/hwmon/` filesystem on Linux
- `filterUsefulTemperatures()` reduces noise by selecting only the most important sensors

### Alert System
- `AlertManager` in `pkg/ui/components/alerts.go` tracks threshold breaches
- Thresholds: CPU (70%/90%), Memory (80%/95%), Temperature (70°C/85°C)
- Alerts use severity levels: Info, Warning, Critical
- Alert history is maintained with a circular buffer (max 100 entries)
- `AlertBar` component displays active alerts at the top of the dashboard

### Configuration Hierarchy
Configuration is loaded in this order (later overrides earlier):
1. Default values in `config.DefaultConfig()`
2. Config file at `~/.config/metrics-tui/config.yaml`
3. Environment variables with `MONITOR_` prefix (e.g., `MONITOR_REFRESH_INTERVAL=5s`)
4. Command-line flags

### Dashboard Layout Logic
- Dashboard uses `lipgloss.JoinVertical()` and custom column joining
- Each metric component is wrapped in a `RoundedBorder()` box
- Width is distributed: `(totalWidth - 8) / 3` per column
- Column 3 stacks two panels vertically using `stackRows()`

## Common Development Tasks

### Adding a New Collector

1. Create `pkg/collectors/newmetric.go` implementing the `Collector` interface
2. Define a metrics struct (e.g., `NewMetricMetrics`)
3. Add conversion function in `aggregator.go` (e.g., `convertNewMetricMetrics()`)
4. Register in `NewAggregator()`: `agg.collectors["newmetric"] = NewNewMetricCollector(config.NewMetricInterval)`
5. Add field to `data.SystemData` and update `Aggregator.GetSystemData()`

### Adding a New UI Component

1. Create component in `pkg/ui/components/metrics/newmetric.go`
2. Implement `Render(systemData *data.SystemData) string` method
3. Add to `Dashboard` struct and initialize in `NewDashboard()`
4. Update `Dashboard.Render()` to include the new component
5. Use existing components like `RenderProgressBar()` and `RenderSparkline()` from `pkg/ui/components/utils.go`

### Modifying Bubble Tea Model

- The `Model.Init()` starts the aggregator and returns the first tick command
- `Model.Update()` handles keyboard input (`q`, `h`, `s`, `esc`) and window resize
- `Model.onDataUpdate()` is the callback from aggregator - updates `systemData`
- `Model.updateHistory()` is called on `tickMsg` - adds to history and checks alerts
- Always call `aggregator.Stop()` before returning `tea.Quit` to clean up goroutines

### Testing Collectors

```bash
# Debug mode runs all collectors once and prints output
./metrics-tui --debug

# This is useful for testing new collectors without running the full TUI
```

## Module Path

The Go module path is `github.com/ctcac00/metrics-tui`. When adding imports:
- Internal packages: `github.com/ctcac00/metrics-tui/internal/data`
- Public packages: `github.com/ctcac00/metrics-tui/pkg/collectors`, `github.com/ctcac00/metrics-tui/pkg/ui`

## Key Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework (Elm Architecture)
- `github.com/charmbracelet/lipgloss` - Terminal styling and layout
- `github.com/shirou/gopsutil/v4` - Cross-platform system metrics
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## Styling Conventions

The project uses Dracula color scheme:
- Foreground: `#f8f8f2`
- Background: `#282a36`
- Borders: `#44475a`
- Green (success/normal): `#50fa7b`
- Orange (warning): `#ffb86c`
- Red (critical): `#ff5555`
- Cyan: `#8be9fd`
- Purple: `#bd93f9`
- Pink: `#ff79c6`

## Platform-Specific Code

- Fan speed collection only works on Linux (reads from `/sys/class/hwmon/`)
- Sensor filtering priorities favor Intel/AMD CPU sensors on Linux
- Extended memory stats (buffers, cached) are Linux-specific
- Always check for nil/empty data before rendering to handle cross-platform gracefully

## Performance Considerations

- Collectors run at different intervals to balance freshness vs. system load
- History buffers are limited to 50 data points to prevent unbounded memory growth
- RWMutex usage in aggregator allows concurrent reads while collectors update
- `updateChecker()` runs at 500ms but UI only redraws on user input or explicit tick (2s)
- Avoid calling gopsutil functions in tight loops - they can be expensive

## Configuration File Location

Default config path: `~/.config/metrics-tui/config.yaml`
See `config.yaml.example` for all available options.
