package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	outputFile   string
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func snapshotCmd() *cobra.Command {
	var name string
	var outputPath string

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

			encoder := json.NewEncoder(output)
			encoder.SetIndent("", "  ")
			return encoder.Encode(snapshot)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "default", "snapshot name")
	cmd.Flags().StringVarP(&outputPath, "file", "f", "", "output file path")

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
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}