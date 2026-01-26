package config

// This file contains default values documentation

// Default configuration file location:
// Linux/macOS: ~/.config/metrics-tui/config.yaml
// Windows: %APPDATA%\metrics-tui\config.yaml

/*
Example config.yaml:

# Refresh intervals for each collector
refresh:
  interval: 2s      # Global default (overridden by specific settings)
  cpu: 1s           # CPU metrics update interval
  memory: 2s        # Memory metrics update interval
  disk: 5s          # Disk metrics update interval
  network: 2s       # Network metrics update interval
  sensors: 5s       # Temperature sensors update interval
  host: 5s          # Host info update interval

# Display settings
display:
  theme: auto              # Theme: auto, dark, light
  show_graphs: true         # Enable sparkline graphs
  show_percentages: true    # Show percentage values
  precision: 1              # Decimal places (0-3)
  units: auto               # Unit system: auto, binary, decimal

# Alert thresholds (percentage or temperature)
thresholds:
  cpu_warning: 70           # CPU usage warning level (%)
  cpu_critical: 90          # CPU usage critical level (%)
  memory_warning: 80        # Memory usage warning level (%)
  memory_critical: 95       # Memory usage critical level (%)
  temp_warning: 70          # Temperature warning level (°C)
  temp_critical: 85         # Temperature critical level (°C)

# UI-specific settings
ui:
  page_size: 50             # History size for sparklines
  show_load_average: true   # Show load average in header
  show_uptime: true         # Show system uptime in header
  show_hostname: true       # Show hostname in header

# Debug mode
debug: false

# Environment variables:
# All config values can be overridden via environment variables with MONITOR_ prefix
# Examples:
#   MONITOR_REFRESH_INTERVAL=5s
#   MONITOR_DISPLAY_THEME=dark
#   MONITOR_THRESHOLDS_CPU_WARNING=80
#   MONITOR_DEBUG=true
*/
