package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AshitomW/Drifty/internal/collector"
	"github.com/AshitomW/Drifty/internal/comparator"
	"github.com/AshitomW/Drifty/internal/models"
	"github.com/AshitomW/Drifty/internal/reporter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configFile   string
	outputFormat string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "drift",
		Short: "Environment Drift Detector",
		Long:  `Detect and report differences between environments including files, env vars, packages, and services.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (json, yaml, table, text)")

	// Commands
	rootCmd.AddCommand(snapshotCmd())
	rootCmd.AddCommand(compareCmd())
	rootCmd.AddCommand(diffCmd())
	rootCmd.AddCommand(daemonCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func snapshotCmd() *cobra.Command {
	var name string
	var outputPath string
	var format string

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create environment snapshot",
		Long:  `Collect and save current environment state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := loadConfig()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			c := collector.New(config.Collector)
			snapshot, err := c.Collect(ctx, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}

			// Output snapshot
			output := os.Stdout
			if outputPath != "" {
				f, err := os.Create(outputPath)
				if err != nil {
					return err
				}
				defer f.Close()
				output = f
			}

			switch format {
			case "yaml", "yml":
				encoder := yaml.NewEncoder(output)
				encoder.SetIndent(2)
				return encoder.Encode(snapshot)
			case "table":
				return generateSnapshotTable(snapshot, output)
			default:
				encoder := json.NewEncoder(output)
				encoder.SetIndent("", "  ")
				return encoder.Encode(snapshot)
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "default", "snapshot name")
	cmd.Flags().StringVarP(&outputPath, "file", "f", "", "output file path")
	cmd.Flags().StringVarP(&format, "format", "F", "json", "output format (json, yaml, table)")

	return cmd
}

func compareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare <source-snapshot> <target-snapshot>",
		Short: "Compare two snapshot files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceFile := args[0]
			targetFile := args[1]

			source, err := loadSnapshot(sourceFile)
			if err != nil {
				return fmt.Errorf("loading source snapshot: %w", err)
			}

			target, err := loadSnapshot(targetFile)
			if err != nil {
				return fmt.Errorf("loading target snapshot: %w", err)
			}

			config := loadConfig()
			comp := comparator.New(comparator.SeverityRules{
				CriticalPackages: config.SeverityRules.CriticalPackages,
				CriticalServices: config.SeverityRules.CriticalServices,
				CriticalFiles:    config.SeverityRules.CriticalFiles,
				CriticalEnvVars:  config.SeverityRules.CriticalEnvVars,
			})

			report := comp.Compare(source, target)

			// Output report
			format := reporter.Format(outputFormat)
			rep := reporter.New(format, os.Stdout)
			return rep.Generate(report)
		},
	}

	return cmd
}

func diffCmd() *cobra.Command {
	var snapshotFile string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare current environment against a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load baseline snapshot
			baseline, err := loadSnapshot(snapshotFile)
			if err != nil {
				return fmt.Errorf("loading baseline snapshot: %w", err)
			}

			// Collect current state
			config := loadConfig()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			c := collector.New(config.Collector)
			current, err := c.Collect(ctx, "current")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}

			// Compare
			comp := comparator.New(comparator.SeverityRules{
				CriticalPackages: config.SeverityRules.CriticalPackages,
				CriticalServices: config.SeverityRules.CriticalServices,
				CriticalFiles:    config.SeverityRules.CriticalFiles,
				CriticalEnvVars:  config.SeverityRules.CriticalEnvVars,
			})

			report := comp.Compare(baseline, current)

			// Output report
			format := reporter.Format(outputFormat)
			rep := reporter.New(format, os.Stdout)

			if err := rep.Generate(report); err != nil {
				return err
			}

			// Exit with error code if drift detected
			if report.Summary.CriticalCount > 0 {
				os.Exit(2)
			}
			if report.HasDrift {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&snapshotFile, "baseline", "b", "", "baseline snapshot file")
	cmd.MarkFlagRequired("baseline")

	return cmd
}

type Config struct {
	Collector     models.CollectorConfig `yaml:"collector"`
	SeverityRules struct {
		CriticalPackages []string `yaml:"critical_packages"`
		CriticalServices []string `yaml:"critical_services"`
		CriticalFiles    []string `yaml:"critical_files"`
		CriticalEnvVars  []string `yaml:"critical_env_vars"`
	} `yaml:"severity_rules"`
}

func loadConfig() *Config {
	config := &Config{
		Collector: models.CollectorConfig{
			Files: models.FileCollectorConfig{
				Enabled:  true,
				Paths:    []string{"/etc"},
				HashAlgo: "sha256",
				MaxDepth: 10,
			},
			EnvVars: models.EnvVarCollectorConfig{
				Enabled:     true,
				MaskSecrets: true,
			},
			ProcessEnvVars: models.ProcessEnvVarCollectorConfig{
				Enabled:      false,
				MaxProcesses: 10,
				MaskSecrets:  true,
			},
			Network: models.NetworkCollectorConfig{
				Enabled:       true,
				Interfaces:    true,
				Routes:        true,
				DNS:           true,
				FirewallRules: false,
			},
			Docker: models.DockerCollectorConfig{
				Enabled:    false,
				Containers: true,
				Images:     true,
				Volumes:    false,
				Networks:   false,
			},
			SystemResources: models.SystemResourcesCollectorConfig{
				Enabled: true,
				CPU:     true,
				Memory:  true,
				Disks:   true,
				Load:    true,
			},
			ScheduledTasks: models.ScheduledTasksCollectorConfig{
				Enabled:       true,
				CronJobs:      true,
				SystemdTimers: true,
				LaunchdJobs:   true,
			},
			Certificates: models.CertificateCollectorConfig{
				Enabled:       false,
				DaysThreshold: 30,
			},
			UsersGroups: models.UserGroupCollectorConfig{
				Enabled:   true,
				Users:     true,
				Groups:    true,
				SudoRules: true,
			},
			Packages: models.PackageCollectorConfig{
				Enabled:  true,
				Managers: []string{"dpkg", "pip"},
			},
			Services: models.ServiceCollectorConfig{
				Enabled:  true,
				InitType: "systemd",
			},
		},
	}

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err == nil {
			yaml.Unmarshal(data, config)
		}
	}

	return config
}

func loadSnapshot(path string) (*models.EnvironmentSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var snapshot models.EnvironmentSnapshot

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if err := yaml.Unmarshal(data, &snapshot); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return nil, err
		}
	}

	return &snapshot, nil
}

func generateSnapshotTable(snapshot *models.EnvironmentSnapshot, output *os.File) error {
	fmt.Fprintf(output, "\nEnvironment Snapshot\n")
	fmt.Fprintf(output, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(output, "ID:        %s\n", snapshot.ID)
	fmt.Fprintf(output, "Name:      %s\n", snapshot.Name)
	fmt.Fprintf(output, "Hostname:  %s\n", snapshot.Hostname)
	fmt.Fprintf(output, "Timestamp: %s\n", snapshot.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(output, "OS:        %s %s (%s)\n", snapshot.OS.Name, snapshot.OS.Version, snapshot.OS.Arch)
	fmt.Fprintf(output, "Kernel:    %s\n\n", snapshot.OS.Kernel)

	if len(snapshot.EnvVars) > 0 {
		fmt.Fprintf(output, "Environment Variables (%d)\n", len(snapshot.EnvVars))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for _, name := range sortedKeys(snapshot.EnvVars) {
			env := snapshot.EnvVars[name]
			value := env.Value
			if len(value) > 60 {
				value = value[:57] + "..."
			}
			fmt.Fprintf(output, "  %-30s : %s\n", name, value)
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.ProcessEnvVars) > 0 {
		fmt.Fprintf(output, "Process Environment Variables (%d processes)\n", len(snapshot.ProcessEnvVars))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for pid, procEnv := range snapshot.ProcessEnvVars {
			cmdline := procEnv.Cmdline
			if len(cmdline) > 50 {
				cmdline = cmdline[:47] + "..."
			}
			fmt.Fprintf(output, "  PID: %-8d [%s]\n", pid, cmdline)
			for _, name := range sortedKeys(procEnv.EnvVars) {
				env := procEnv.EnvVars[name]
				value := env.Value
				if len(value) > 50 {
					value = value[:47] + "..."
				}
				fmt.Fprintf(output, "    %-28s : %s\n", name, value)
			}
			fmt.Fprintln(output)
		}
	}

	if len(snapshot.Packages) > 0 {
		fmt.Fprintf(output, "Packages (%d)\n", len(snapshot.Packages))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for _, pkgName := range sortedKeys(snapshot.Packages) {
			pkg := snapshot.Packages[pkgName]
			fmt.Fprintf(output, "  %-30s : %-15s (%s)\n", pkg.Name, pkg.Version, pkg.Manager)
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.Services) > 0 {
		fmt.Fprintf(output, "Services (%d)\n", len(snapshot.Services))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for _, svcName := range sortedKeys(snapshot.Services) {
			svc := snapshot.Services[svcName]
			fmt.Fprintf(output, "  %-30s : %-10s (enabled: %v)\n", svc.Name, svc.Status, svc.Enabled)
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.Files) > 0 {
		fmt.Fprintf(output, "Files (%d)\n", len(snapshot.Files))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for _, path := range sortedKeys(snapshot.Files) {
			file := snapshot.Files[path]
			hash := file.Hash
			if len(hash) > 8 {
				hash = hash[:8]
			}
			fmt.Fprintf(output, "  %-40s : %s\n", file.Path, hash)
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.NetworkConfig.Interfaces) > 0 {
		fmt.Fprintf(output, "Network Interfaces (%d)\n", len(snapshot.NetworkConfig.Interfaces))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for _, name := range sortedKeys(snapshot.NetworkConfig.Interfaces) {
			iface := snapshot.NetworkConfig.Interfaces[name]
			fmt.Fprintf(output, "  %-20s : IP: %v, MAC: %s\n", name, iface.IPAddresses, iface.MACAddress)
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.DockerConfig.Containers) > 0 {
		fmt.Fprintf(output, "Docker Containers (%d)\n", len(snapshot.DockerConfig.Containers))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for id, cont := range snapshot.DockerConfig.Containers {
			fmt.Fprintf(output, "  %-30s : %s (%s)\n", cont.Name, cont.Image, cont.Status)
			_ = id
		}
		fmt.Fprintln(output)
	}

	if snapshot.SystemResources.CPU.Cores > 0 {
		fmt.Fprintf(output, "System Resources\n")
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		fmt.Fprintf(output, "  CPU Cores      : %d\n", snapshot.SystemResources.CPU.Cores)
		fmt.Fprintf(output, "  CPU Usage     : %.1f%%\n", snapshot.SystemResources.CPU.Usage)
		fmt.Fprintf(output, "  Memory Total   : %d MB\n", snapshot.SystemResources.Memory.Total/1024/1024)
		fmt.Fprintf(output, "  Memory Usage  : %.1f%%\n", snapshot.SystemResources.Memory.Usage)
		fmt.Fprintf(output, "  Process Count : %d\n", snapshot.SystemResources.ProcessCount)
		fmt.Fprintln(output)
	}

	if len(snapshot.ScheduledTasks.CronJobs) > 0 || len(snapshot.ScheduledTasks.SystemdTimers) > 0 || len(snapshot.ScheduledTasks.LaunchdJobs) > 0 {
		fmt.Fprintf(output, "Scheduled Tasks\n")
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		if len(snapshot.ScheduledTasks.CronJobs) > 0 {
			for name, cron := range snapshot.ScheduledTasks.CronJobs {
				fmt.Fprintf(output, "  %-20s (cron) : %s\n", name, cron.Schedule)
				_ = name
			}
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.Certificates) > 0 {
		fmt.Fprintf(output, "Certificates (%d)\n", len(snapshot.Certificates))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for path, cert := range snapshot.Certificates {
			expiryStatus := "VALID"
			if cert.IsExpired {
				expiryStatus = "EXPIRED"
			} else if cert.DaysToExpire < 30 {
				expiryStatus = "EXPIRING SOON"
			}
			fmt.Fprintf(output, "  %-40s : %s (%d days)\n", cert.Domain, expiryStatus, cert.DaysToExpire)
			_ = path
		}
		fmt.Fprintln(output)
	}

	if len(snapshot.UserGroupConfig.Users) > 0 {
		fmt.Fprintf(output, "Users (%d)\n", len(snapshot.UserGroupConfig.Users))
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 60))
		for name, user := range snapshot.UserGroupConfig.Users {
			fmt.Fprintf(output, "  %-20s : UID: %d, Shell: %s\n", name, user.UID, user.Shell)
			_ = name
		}
		fmt.Fprintln(output)
	}

	return nil
}

func sortedKeys(m interface{}) []string {
	var keys []string
	switch v := m.(type) {
	case map[string]models.EnvVar:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]models.PackageInfo:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]models.ServiceInfo:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]models.FileInfo:
		for k := range v {
			keys = append(keys, k)
		}
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func daemonCmd() *cobra.Command {
	var baselineFile string
	var intervalStr string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run drift detection daemon",
		Long:  `Continuously monitor for drift and generate reports`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if baselineFile == "" {
				return fmt.Errorf("baseline file is required")
			}

			config := loadConfig()
			ctx := context.Background()

			interval, err := time.ParseDuration(intervalStr)
			if err != nil {
				return fmt.Errorf("invalid interval format: %w", err)
			}

			if interval < time.Minute {
				return fmt.Errorf("interval must be at least 1 minute")
			}

			fmt.Printf("Starting drift daemon...\n")
			fmt.Printf("Baseline: %s\n", baselineFile)
			fmt.Printf("Interval: %s\n", interval)
			fmt.Printf("Output directory: %s\n", outputDir)
			fmt.Printf("Press Ctrl+C to stop\n\n")

			baseline, err := loadSnapshot(baselineFile)
			if err != nil {
				return fmt.Errorf("loading baseline: %w", err)
			}

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			runCount := 0

			for {
				select {
				case <-ctx.Done():
					fmt.Printf("\nDrift daemon stopped\n")
					return nil
				case <-ticker.C:
					runCount++
					fmt.Printf("\n[%s] Run #%d\n", time.Now().Format("2006-01-02 15:04:05"), runCount)

					c := collector.New(config.Collector)
					current, err := c.Collect(ctx, "current")
					if err != nil {
						fmt.Printf("Warning: %v\n", err)
					}

					comp := comparator.New(comparator.SeverityRules{
						CriticalPackages: config.SeverityRules.CriticalPackages,
						CriticalServices: config.SeverityRules.CriticalServices,
						CriticalFiles:    config.SeverityRules.CriticalFiles,
						CriticalEnvVars:  config.SeverityRules.CriticalEnvVars,
					})

					report := comp.Compare(baseline, current)

					fmt.Printf("Total drifts: %d (Critical: %d, Warning: %d, Info: %d)\n",
						report.Summary.TotalDrifts, report.Summary.CriticalCount,
						report.Summary.WarningCount, report.Summary.InfoCount)

					if report.HasDrift {
						fmt.Printf("Drift detected!\n")

						if outputDir != "" {
							reportFile := fmt.Sprintf("%s/drift-%s.json", outputDir, time.Now().Format("20060102-150405"))
							reportData, err := json.MarshalIndent(report, "", "  ")
							if err == nil {
								if err := os.WriteFile(reportFile, reportData, 0644); err != nil {
									fmt.Printf("Error writing report: %v\n", err)
								} else {
									fmt.Printf("Report saved: %s\n", reportFile)
								}
							}
						}

						if report.Summary.CriticalCount > 0 {
							fmt.Printf("CRITICAL DRIFT DETECTED!\n")
						}
					} else {
						fmt.Printf("No drift detected\n")
					}
				}
			}
		},
	}

	cmd.Flags().StringVarP(&baselineFile, "baseline", "b", "", "baseline snapshot file to compare against")
	cmd.MarkFlagRequired("baseline")
	cmd.Flags().StringVarP(&intervalStr, "interval", "i", "5m", "interval between checks (e.g., 5m, 1h)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory for reports")

	return cmd
}
