package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Refresh   RefreshConfig
	Display   DisplayConfig
	Threshold ThresholdConfig
	UI        UIConfig
	Debug     bool
}

// RefreshConfig holds refresh interval settings
type RefreshConfig struct {
	Interval time.Duration
	CPU      time.Duration
	Memory   time.Duration
	Disk     time.Duration
	Network  time.Duration
	Sensors  time.Duration
	Host     time.Duration
}

// DisplayConfig holds display settings
type DisplayConfig struct {
	Theme          string
	ShowGraphs     bool
	ShowPercentages bool
	Precision      int
	Units          string
}

// ThresholdConfig holds alert threshold settings
type ThresholdConfig struct {
	CPUWarning  float64
	CPUCritical float64
	MemWarning  float64
	MemCritical float64
	TempWarning float64
	TempCritical float64
}

// UIConfig holds UI-specific settings
type UIConfig struct {
	PageSize      int
	ShowLoadAverage bool
	ShowUptime      bool
	ShowHostname    bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Refresh: RefreshConfig{
			Interval: 2 * time.Second,
			CPU:      1 * time.Second,
			Memory:   2 * time.Second,
			Disk:     5 * time.Second,
			Network:  2 * time.Second,
			Sensors:  5 * time.Second,
			Host:     5 * time.Second,
		},
		Display: DisplayConfig{
			Theme:           "auto",
			ShowGraphs:      true,
			ShowPercentages: true,
			Precision:       1,
			Units:           "auto",
		},
		Threshold: ThresholdConfig{
			CPUWarning:    70.0,
			CPUCritical:  90.0,
			MemWarning:    80.0,
			MemCritical:   95.0,
			TempWarning:   70.0,
			TempCritical:  85.0,
		},
		UI: UIConfig{
			PageSize:        50,
			ShowLoadAverage: true,
			ShowUptime:      true,
			ShowHostname:    true,
		},
		Debug: false,
	}
}

// Load loads configuration from file, flags, and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Set up Viper
	viper.SetDefault("refresh.interval", cfg.Refresh.Interval)
	viper.SetDefault("refresh.cpu", cfg.Refresh.CPU)
	viper.SetDefault("refresh.memory", cfg.Refresh.Memory)
	viper.SetDefault("refresh.disk", cfg.Refresh.Disk)
	viper.SetDefault("refresh.network", cfg.Refresh.Network)
	viper.SetDefault("refresh.sensors", cfg.Refresh.Sensors)
	viper.SetDefault("refresh.host", cfg.Refresh.Host)

	viper.SetDefault("display.theme", cfg.Display.Theme)
	viper.SetDefault("display.show_graphs", cfg.Display.ShowGraphs)
	viper.SetDefault("display.show_percentages", cfg.Display.ShowPercentages)
	viper.SetDefault("display.precision", cfg.Display.Precision)
	viper.SetDefault("display.units", cfg.Display.Units)

	viper.SetDefault("thresholds.cpu_warning", cfg.Threshold.CPUWarning)
	viper.SetDefault("thresholds.cpu_critical", cfg.Threshold.CPUCritical)
	viper.SetDefault("thresholds.memory_warning", cfg.Threshold.MemWarning)
	viper.SetDefault("thresholds.memory_critical", cfg.Threshold.MemCritical)
	viper.SetDefault("thresholds.temp_warning", cfg.Threshold.TempWarning)
	viper.SetDefault("thresholds.temp_critical", cfg.Threshold.TempCritical)

	viper.SetDefault("ui.page_size", cfg.UI.PageSize)
	viper.SetDefault("ui.show_load_average", cfg.UI.ShowLoadAverage)
	viper.SetDefault("ui.show_uptime", cfg.UI.ShowUptime)
	viper.SetDefault("ui.show_hostname", cfg.UI.ShowHostname)

	viper.SetDefault("debug", cfg.Debug)

	// Read config file if it exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/monitor-tui")
	viper.AddConfigPath(".")

	// Allow environment variables with prefix
	viper.SetEnvPrefix("MONITOR")
	viper.AutomaticEnv()

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate refresh intervals (minimum 100ms)
	minInterval := 100 * time.Millisecond
	if c.Refresh.Interval < minInterval {
		c.Refresh.Interval = minInterval
	}
	if c.Refresh.CPU < minInterval {
		c.Refresh.CPU = minInterval
	}
	if c.Refresh.Memory < minInterval {
		c.Refresh.Memory = minInterval
	}
	if c.Refresh.Disk < minInterval {
		c.Refresh.Disk = minInterval
	}
	if c.Refresh.Network < minInterval {
		c.Refresh.Network = minInterval
	}
	if c.Refresh.Sensors < minInterval {
		c.Refresh.Sensors = minInterval
	}
	if c.Refresh.Host < minInterval {
		c.Refresh.Host = minInterval
	}

	// Validate display precision (0-3 decimal places)
	if c.Display.Precision < 0 {
		c.Display.Precision = 0
	}
	if c.Display.Precision > 3 {
		c.Display.Precision = 3
	}

	// Validate theme
	if c.Display.Theme != "auto" && c.Display.Theme != "dark" && c.Display.Theme != "light" {
		c.Display.Theme = "auto"
	}

	// Validate thresholds (0-100 range)
	validateThreshold(&c.Threshold.CPUWarning, &c.Threshold.CPUCritical)
	validateThreshold(&c.Threshold.MemWarning, &c.Threshold.MemCritical)
	validateThreshold(&c.Threshold.TempWarning, &c.Threshold.TempCritical)

	// Validate page size (10-200)
	if c.UI.PageSize < 10 {
		c.UI.PageSize = 10
	}
	if c.UI.PageSize > 200 {
		c.UI.PageSize = 200
	}

	return nil
}

// validateThreshold ensures warning < critical and both are in range 0-100
func validateThreshold(warning, critical *float64) {
	if *warning < 0 {
		*warning = 0
	}
	if *warning > 100 {
		*warning = 100
	}
	if *critical < 0 {
		*critical = 0
	}
	if *critical > 100 {
		*critical = 100
	}
	if *warning >= *critical {
		*warning = *critical - 10
		if *warning < 0 {
			*warning = 0
		}
	}
}

// GetIntervalMap returns a map of collector intervals
func (c *Config) GetIntervalMap() map[string]uint {
	return map[string]uint{
		"cpu":     uint(c.Refresh.CPU.Seconds()),
		"memory":  uint(c.Refresh.Memory.Seconds()),
		"disk":    uint(c.Refresh.Disk.Seconds()),
		"network": uint(c.Refresh.Network.Seconds()),
		"sensors": uint(c.Refresh.Sensors.Seconds()),
		"host":    uint(c.Refresh.Host.Seconds()),
	}
}
