# Metrics TUI

A beautiful, real-time terminal user interface for monitoring system hardware metrics.

![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

> **Note**: This project was vibe coded with AI agents 🤖✨

## Features

- **Real-time Monitoring**: Live metrics updated at configurable intervals
- **Comprehensive Metrics**:
  - CPU usage (per-core and total)
  - Memory and swap usage
  - Disk usage and I/O statistics
  - Network interface statistics
  - Temperature sensors (CPU, GPU, thermal zones)
  - Fan speeds (Linux)
  - System load averages
  - Host information (hostname, uptime, OS)
- **Beautiful UI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Historical Data**: Sparkline visualizations showing metric trends
- **Smart Alerts**: Configurable threshold-based alerts with color coding
- **Snapshot Feature**: Capture and save system state
- **Highly Configurable**: YAML config files, CLI flags, and environment variables
- **Cross-platform**: Linux, macOS, and Windows support

## Installation

### From Source

```bash
go install github.com/ctcac00/metrics-tui@latest
```

### Build Locally

```bash
git clone https://github.com/ctcac00/metrics-tui.git
cd metrics-tui
go build -o metrics-tui
```

## Usage

### Basic Usage

```bash
# Run with default settings
metrics-tui

# Show help
metrics-tui --help

# List available disk partitions
metrics-tui --list-disks

# Run in debug mode (test collectors)
metrics-tui --debug
```

### Command-line Flags

```bash
# Use custom config file
metrics-tui --config /path/to/config.yaml

# Override refresh interval
metrics-tui --refresh 5s

# Set color theme
metrics-tui --theme dark

# Disable sparkline graphs
metrics-tui --no-graphs

# Set decimal precision
metrics-tui --precision 2
```

## Configuration

Metrics TUI supports configuration through:
1. Config file (`~/.config/metrics-tui/config.yaml`)
2. Command-line flags
3. Environment variables (with `MONITOR_` prefix)

### Example Configuration

Create `~/.config/metrics-tui/config.yaml`:

```yaml
# Refresh intervals for each collector
refresh:
  interval: 2s    # Global default
  cpu: 1s         # CPU metrics
  memory: 2s      # Memory metrics
  disk: 5s        # Disk metrics
  network: 2s     # Network metrics
  sensors: 5s     # Temperature sensors
  host: 5s        # Host info

# Display settings
display:
  theme: auto              # auto, dark, or light
  show_graphs: true        # Enable sparkline graphs
  show_percentages: true   # Show percentage values
  precision: 1             # Decimal places (0-3)
  units: auto              # auto, binary (KiB), or decimal (KB)

# Alert thresholds
thresholds:
  cpu_warning: 70          # CPU usage warning (%)
  cpu_critical: 90         # CPU usage critical (%)
  memory_warning: 80       # Memory usage warning (%)
  memory_critical: 95      # Memory usage critical (%)
  temp_warning: 70         # Temperature warning (°C)
  temp_critical: 85        # Temperature critical (°C)

# UI settings
ui:
  page_size: 50            # History size for sparklines
  show_load_average: true  # Show load averages
  show_uptime: true        # Show system uptime
  show_hostname: true      # Show hostname

# Debug mode
debug: false
```

### Environment Variables

Override any setting using environment variables:

```bash
export MONITOR_REFRESH_INTERVAL=5s
export MONITOR_DISPLAY_THEME=dark
export MONITOR_THRESHOLDS_CPU_WARNING=80
export MONITOR_DEBUG=true

metrics-tui
```

## Keyboard Shortcuts

- `q` or `Ctrl+C` - Quit
- `h` or `?` - Toggle help screen
- `Esc` - Close help overlay
- `s` - Take snapshot of current metrics

## Architecture

Metrics TUI uses a modular architecture:

- **Collectors**: Independent goroutines for each metric type (CPU, memory, disk, etc.)
- **Aggregator**: Centralizes data from all collectors and manages updates
- **UI Components**: Modular components for header, dashboard, metrics, alerts, etc.
- **Bubble Tea**: Elm Architecture pattern for state management and rendering

### Project Structure

```
metrics-tui/
├── cmd/                  # CLI commands
├── pkg/
│   ├── collectors/       # Data collection layer
│   ├── config/           # Configuration management
│   └── ui/               # UI components and rendering
├── internal/
│   └── data/             # Data models and history
├── config.yaml.example   # Example configuration
└── main.go
```

## Platform-Specific Features

### Linux
- Full sensor support (via `hwmon` and `gopsutil`)
- Fan speed monitoring (via `/sys/class/hwmon/`)
- Extended memory statistics

### macOS
- Temperature sensors (when available)
- Standard CPU, memory, disk, and network metrics

### Windows
- Core metrics (CPU, memory, disk, network)
- Limited sensor support

## Requirements

- Go 1.25.5 or higher
- Terminal with ANSI color support
- For best experience: 80x24 minimum terminal size

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [gopsutil](https://github.com/shirou/gopsutil) - System metrics collection
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with the excellent [Charm](https://charm.sh/) ecosystem and [gopsutil](https://github.com/shirou/gopsutil).

## Screenshots

> Add screenshots here showing the TUI in action

## Roadmap

- [ ] GPU monitoring (NVIDIA/AMD)
- [ ] Process list view with sorting/filtering
- [ ] Customizable dashboard layouts
- [ ] Historical data export (CSV, JSON)
- [ ] Remote monitoring mode
- [ ] Plugin system for custom collectors
