package collector

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectProcessEnvVars(ctx context.Context) (map[int]models.ProcessEnvVar, error) {
	processEnvVars := make(map[int]models.ProcessEnvVar)

	if runtime.GOOS == "windows" {
		return processEnvVars, nil
	}

	compileExcludePatterns := func() []*regexp.Regexp {
		var patterns []*regexp.Regexp
		for _, pattern := range c.config.ProcessEnvVars.Exclude {
			if re, err := regexp.Compile(pattern); err == nil {
				patterns = append(patterns, re)
			}
		}
		return patterns
	}
	excludePatterns := compileExcludePatterns()

	maxProcs := c.config.ProcessEnvVars.MaxProcesses
	if maxProcs <= 0 {
		maxProcs = 10
	}

	procNames := c.config.ProcessEnvVars.Processes
	if len(procNames) == 0 {
		procNames = []string{"node", "php", "python", "python3", "ruby", "java", "go"}
	}

	procMap := make(map[string]bool)
	for _, name := range procNames {
		procMap[name] = true
	}

	if runtime.GOOS == "darwin" {
		return c.collectProcessEnvVarsDarwin(ctx, procMap, maxProcs, excludePatterns)
	}

	return c.collectProcessEnvVarsLinux(ctx, procMap, maxProcs, excludePatterns)
}

func (c *Collector) collectProcessEnvVarsLinux(ctx context.Context, procMap map[string]bool, maxProcs int, excludePatterns []*regexp.Regexp) (map[int]models.ProcessEnvVar, error) {
	processEnvVars := make(map[int]models.ProcessEnvVar)

	procsDir := "/proc"
	if _, err := os.Stat(procsDir); os.IsNotExist(err) {
		return processEnvVars, nil
	}

	entries, err := os.ReadDir(procsDir)
	if err != nil {
		return nil, err
	}

	count := 0
	for _, entry := range entries {
		if count >= maxProcs {
			break
		}

		select {
		case <-ctx.Done():
			return processEnvVars, ctx.Err()
		default:
		}

		if !entry.IsDir() {
			continue
		}

		pidStr := entry.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		cmdlinePath := filepath.Join(procsDir, pidStr, "cmdline")
		cmdlineBytes, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		cmdline := strings.TrimSpace(string(cmdlineBytes))
		cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
		if cmdline == "" {
			continue
		}

		argv := strings.Fields(cmdline)
		if len(argv) == 0 {
			continue
		}

		procName := filepath.Base(argv[0])
		if !procMap[procName] {
			continue
		}

		envPath := filepath.Join(procsDir, pidStr, "environ")
		envBytes, err := os.ReadFile(envPath)
		if err != nil {
			continue
		}

		envStr := string(envBytes)
		envPairs := strings.Split(envStr, "\x00")

		envVars := make(map[string]models.EnvVar)
		for _, pair := range envPairs {
			if pair == "" {
				continue
			}

			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				continue
			}

			name := parts[0]
			value := parts[1]

			excluded := false
			for _, re := range excludePatterns {
				if re.MatchString(name) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			if c.config.ProcessEnvVars.MaskSecrets && isSecretVar(name) {
				value = maskValue(value)
			}

			envVars[name] = models.EnvVar{
				Name:   name,
				Value:  value,
				Exists: true,
			}
		}

		if len(envVars) == 0 {
			continue
		}

		processEnvVars[pid] = models.ProcessEnvVar{
			PID:     pid,
			Cmdline: cmdline,
			EnvVars: envVars,
		}

		count++
	}

	return processEnvVars, nil
}

func (c *Collector) collectProcessEnvVarsDarwin(ctx context.Context, procMap map[string]bool, maxProcs int, excludePatterns []*regexp.Regexp) (map[int]models.ProcessEnvVar, error) {
	processEnvVars := make(map[int]models.ProcessEnvVar)

	for procName := range procMap {
		select {
		case <-ctx.Done():
			return processEnvVars, ctx.Err()
		default:
		}

		var stdout bytes.Buffer
		cmd := exec.CommandContext(ctx, "pgrep", "-d", "\n", procName)
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			continue
		}

		pidLines := strings.Split(stdout.String(), "\n")
		count := 0

		for _, pidLine := range pidLines {
			if count >= maxProcs {
				break
			}

			pidStr := strings.TrimSpace(pidLine)
			if pidStr == "" {
				continue
			}

			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			var envStdout bytes.Buffer
			envCmd := exec.CommandContext(ctx, "ps", "-E", "-p", pidStr, "-o", "command")
			envCmd.Stdout = &envStdout

			if err := envCmd.Run(); err != nil {
				continue
			}

			envLines := strings.Split(envStdout.String(), "\n")
			if len(envLines) < 2 {
				continue
			}

			fullLine := strings.Join(envLines[1:], " ")
			fields := strings.Fields(fullLine)

			if len(fields) == 0 {
				continue
			}

			cmdline := strings.Join(fields, " ")
			envVars := make(map[string]models.EnvVar)
			cmdlineParts := []string{}
			isEnvSection := false

			for _, field := range fields {
				if strings.Contains(field, "=") {
					isEnvSection = true
					parts := strings.SplitN(field, "=", 2)
					if len(parts) != 2 {
						continue
					}

					name := parts[0]
					value := parts[1]

					excluded := false
					for _, re := range excludePatterns {
						if re.MatchString(name) {
							excluded = true
							break
						}
					}
					if excluded {
						continue
					}

					if c.config.ProcessEnvVars.MaskSecrets && isSecretVar(name) {
						value = maskValue(value)
					}

					envVars[name] = models.EnvVar{
						Name:   name,
						Value:  value,
						Exists: true,
					}
				} else if !isEnvSection {
					cmdlineParts = append(cmdlineParts, field)
				}
			}

			cmdline = strings.Join(cmdlineParts, " ")

			if len(envVars) > 0 || cmdline != "" {
				processEnvVars[pid] = models.ProcessEnvVar{
					PID:     pid,
					Cmdline: cmdline,
					EnvVars: envVars,
				}
				count++
			}
		}

		if len(processEnvVars) >= maxProcs {
			break
		}
	}

	return processEnvVars, nil
}
