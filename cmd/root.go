package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ctcac00/monitor-tui/internal/data"
	"github.com/ctcac00/monitor-tui/pkg/collectors"
	"github.com/ctcac00/monitor-tui/pkg/config"
	"github.com/ctcac00/monitor-tui/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var appConfig *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "monitor-tui",
	Short: "A real-time terminal UI for hardware monitoring",
	Long: `monitor-tui displays real-time system metrics including CPU, memory,
disk, network, temperatures, and more in a terminal-based dashboard.

Built with Bubble Tea for a beautiful, responsive TUI experience.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		var err error
		appConfig, err = config.Load()
		if err != nil {
			cmd.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		debug := viper.GetBool("debug")
		listDisks := viper.GetBool("list-disks")

		if listDisks {
			listAvailableDisks(cmd)
			return
		}

		if debug {
			testCollectors(cmd)
			return
		}

		// Launch the TUI
		model := ui.NewModel()
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			cmd.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag: config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/monitor-tui/config.yaml)")

	// Flag: refresh interval
	rootCmd.PersistentFlags().StringP("refresh", "r", "2s", "Override refresh interval")

	// Flag: theme
	rootCmd.PersistentFlags().String("theme", "auto", "Color theme (auto|dark|light)")

	// Flag: no-graphs
	rootCmd.PersistentFlags().Bool("no-graphs", false, "Disable sparklines")

	// Flag: list disks
	rootCmd.PersistentFlags().Bool("list-disks", false, "Show available disks and exit")

	// Flag: debug
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")

	// Flag: precision
	rootCmd.PersistentFlags().IntP("precision", "p", 1, "Decimal places for values (0-3)")

	// Bind flags to viper
	viper.BindPFlag("refresh", rootCmd.PersistentFlags().Lookup("refresh"))
	viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	viper.BindPFlag("display.no_graphs", rootCmd.PersistentFlags().Lookup("no-graphs"))
	viper.BindPFlag("list-disks", rootCmd.PersistentFlags().Lookup("list-disks"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("display.precision", rootCmd.PersistentFlags().Lookup("precision"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory
		viper.AddConfigPath(home+"/.config/monitor-tui")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		// Config file found and successfully parsed
	}
}

// listAvailableDisks lists available disk partitions
func listAvailableDisks(cmd *cobra.Command) {
	ctx := context.Background()
	diskCollector := collectors.NewDiskCollector(1, nil, true)

	data, err := diskCollector.Collect(ctx)
	if err != nil {
		cmd.Printf("Error collecting disk info: %v\n", err)
		return
	}

	metrics, ok := data.(*collectors.DiskMetrics)
	if !ok {
		cmd.Println("Error: unexpected data type")
		return
	}

	cmd.Println("Available disk partitions:")
	cmd.Println()
	for _, p := range metrics.Partitions {
		cmd.Printf("  Device: %-20s Mount: %-15s Type: %s\n", p.Device, p.Mountpoint, p.Fstype)
	}
}

// testCollectors tests all collectors and prints their data
func testCollectors(cmd *cobra.Command) {
	ctx := context.Background()
	cmd.Println("\n=== Testing Collectors ===\n")

	// Test CPU collector
	cmd.Println("CPU Collector:")
	cpuCollector := collectors.NewCPUCollector(1)
	if data, err := cpuCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.CPUMetrics); ok {
			cmd.Printf("  Cores: %d\n", metrics.CoreCount)
			cmd.Printf("  Total Usage: %.1f%%\n", metrics.Total)
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	// Test Memory collector
	cmd.Println("\nMemory Collector:")
	memCollector := collectors.NewMemoryCollector(1)
	if data, err := memCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.MemoryMetrics); ok {
			cmd.Printf("  Total: %s\n", formatBytes(metrics.Total))
			cmd.Printf("  Used: %s (%.1f%%)\n", formatBytes(metrics.Used), metrics.UsedPercent)
			cmd.Printf("  Available: %s\n", formatBytes(metrics.Available))
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	// Test Disk collector
	cmd.Println("\nDisk Collector:")
	diskCollector := collectors.NewDiskCollector(1, nil, true)
	if data, err := diskCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.DiskMetrics); ok {
			cmd.Printf("  Partitions: %d\n", len(metrics.Partitions))
			for mount, usage := range metrics.Usage {
				cmd.Printf("    %s: %s used (%.1f%%)\n", mount, formatBytes(usage.Used), usage.UsedPercent)
			}
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	// Test Network collector
	cmd.Println("\nNetwork Collector:")
	netCollector := collectors.NewNetworkCollector(1, nil, true)
	if data, err := netCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.NetworkMetrics); ok {
			cmd.Printf("  Interfaces: %d\n", len(metrics.Interfaces))
			for name, io := range metrics.IO {
				cmd.Printf("    %s: RX %s, TX %s\n", name, formatBytes(io.BytesRecv), formatBytes(io.BytesSent))
			}
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	// Test Sensors collector
	cmd.Println("\nSensors Collector:")
	sensorCollector := collectors.NewSensorsCollector(1)
	if data, err := sensorCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.SensorMetrics); ok {
			cmd.Printf("  Temperatures: %d\n", len(metrics.Temperatures))
			for _, temp := range metrics.Temperatures {
				cmd.Printf("    %s: %.1f°C\n", temp.SensorKey, temp.Temperature)
			}
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	// Test Host collector
	cmd.Println("\nHost Collector:")
	hostCollector := collectors.NewHostCollector(1)
	if data, err := hostCollector.Collect(ctx); err == nil {
		if metrics, ok := data.(*collectors.HostMetrics); ok {
			cmd.Printf("  Hostname: %s\n", metrics.Info.Hostname)
			cmd.Printf("  Uptime: %s\n", formatDuration(time.Duration(metrics.Info.Uptime)*time.Second))
			if metrics.LoadAvg != nil {
				cmd.Printf("  Load Average: %.2f %.2f %.2f\n", metrics.LoadAvg.Load1, metrics.LoadAvg.Load5, metrics.LoadAvg.Load15)
			}
		}
	} else {
		cmd.Printf("  Error: %v\n", err)
	}

	cmd.Println("\n=== Testing Aggregator ===\n")

	// Test aggregator
	aggConfig := &collectors.AggregatorConfig{
		CPUInterval:          1,
		MemoryInterval:       1,
		DiskInterval:         1,
		NetworkInterval:      1,
		SensorsInterval:      1,
		HostInterval:         1,
		DiskIncludeAll:       true,
		NetworkExcludeVirtual: true,
	}
	aggregator := collectors.NewAggregator(aggConfig)

	// Set up a callback to receive data
	done := make(chan bool)
	aggregator.SetOnDataUpdate(func(d *data.SystemData) {
		cmd.Println("Aggregator received data:")
		if d.CPU != nil {
			cmd.Printf("  CPU: %.1f%%\n", d.CPU.Total)
		}
		if d.Memory != nil {
			cmd.Printf("  Memory: %.1f%%\n", d.Memory.UsedPercent)
		}
		done <- true
	})

	aggregator.Start()

	// Wait for first data update
	select {
	case <-done:
		cmd.Println("\nAggregator test successful!")
	case <-time.After(5 * time.Second):
		cmd.Println("\nAggregator test timed out")
	}

	aggregator.Stop()
}

// formatBytes formats a byte count as human-readable
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats a duration as human-readable
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
