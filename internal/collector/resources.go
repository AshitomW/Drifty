package collector

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectSystemResources(ctx context.Context) (models.SystemResources, error) {
	resources := models.SystemResources{}

	if c.config.SystemResources.CPU {
		cpu, err := c.collectCPUInfo(ctx)
		if err == nil {
			resources.CPU = cpu
		}
	}

	if c.config.SystemResources.Memory {
		memory, err := c.collectMemoryInfo(ctx)
		if err == nil {
			resources.Memory = memory
		}
	}

	if c.config.SystemResources.Disks {
		disks, err := c.collectDiskInfo(ctx)
		if err == nil {
			resources.Disks = disks
		}
	}

	if c.config.SystemResources.Load {
		load, err := c.collectLoadAverage(ctx)
		if err == nil {
			resources.LoadAverage = load
		}
	}

	procCount, err := c.getProcessCount(ctx)
	if err == nil {
		resources.ProcessCount = procCount
	}

	return resources, nil
}

func (c *Collector) collectCPUInfo(ctx context.Context) (models.CPUInfo, error) {
	cpu := models.CPUInfo{}

	if runtime.GOOS == "darwin" {
		return c.collectCPUInfoDarwin(ctx)
	} else if runtime.GOOS == "linux" {
		return c.collectCPUInfoLinux(ctx)
	}

	cpu.Cores = runtime.NumCPU()
	return cpu, nil
}

func (c *Collector) collectCPUInfoDarwin(ctx context.Context) (models.CPUInfo, error) {
	cpu := models.CPUInfo{}

	cmd := exec.CommandContext(ctx, "sysctl", "-n", "hw.ncpu")
	output, err := cmd.Output()
	if err == nil {
		if cores, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			cpu.Cores = cores
		}
	}

	cmd = exec.CommandContext(ctx, "sysctl", "-n", "machdep.cpu.brand_string")
	output, err = cmd.Output()
	if err == nil {
		cpu.Model = strings.TrimSpace(string(output))
	}

	cmd = exec.CommandContext(ctx, "top", "-l", "1", "-n", "0")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "CPU usage:") {
				parts := strings.Fields(line)
				for i := 0; i < len(parts); i++ {
					word := strings.TrimSuffix(parts[i], ",")
					if word == "user" {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.User = val
						}
					} else if word == "sys" {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.System = val
						}
					} else if word == "idle" {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.Idle = val
						}
					}
				}
				cpu.Usage = cpu.User + cpu.System
				break
			}
		}
	}

	return cpu, nil
}

func (c *Collector) collectCPUInfoLinux(ctx context.Context) (models.CPUInfo, error) {
	cpu := models.CPUInfo{}

	data, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					cpu.Model = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "processor") {
				cpu.Cores++
			}
		}
	}

	if cpu.Cores == 0 {
		cpu.Cores = runtime.NumCPU()
	}

	cmd := exec.CommandContext(ctx, "top", "-bn1", "|", "grep", "Cpu(s)")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Cpu(s):") {
				parts := strings.Fields(line)
				for i := 0; i < len(parts); i++ {
					if parts[i] == "us," {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.User = val
						}
					} else if parts[i] == "sy," {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.System = val
						}
					} else if parts[i] == "id," {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i-1], "%"), 64); err == nil {
							cpu.Idle = val
						}
					}
				}
				cpu.Usage = cpu.User + cpu.System
				break
			}
		}
	}

	return cpu, nil
}

func (c *Collector) collectMemoryInfo(ctx context.Context) (models.MemoryInfo, error) {
	memory := models.MemoryInfo{}

	if runtime.GOOS == "darwin" {
		return c.collectMemoryInfoDarwin(ctx)
	} else if runtime.GOOS == "linux" {
		return c.collectMemoryInfoLinux(ctx)
	}

	return memory, nil
}

func (c *Collector) collectMemoryInfoDarwin(ctx context.Context) (models.MemoryInfo, error) {
	memory := models.MemoryInfo{}

	cmd := exec.CommandContext(ctx, "vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return memory, err
	}

	var pageSize int64 = 4096
	var free, active, inactive, speculative, wired int64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				if val, err := strconv.ParseInt(strings.TrimSuffix(parts[3], "."), 10, 64); err == nil {
					pageSize = val
				}
			}
		} else if strings.HasPrefix(line, "Pages free:") {
			if val, err := strconv.ParseInt(strings.TrimSuffix(strings.Fields(line)[2], "."), 10, 64); err == nil {
				free = val
			}
		} else if strings.HasPrefix(line, "Pages active:") {
			if val, err := strconv.ParseInt(strings.TrimSuffix(strings.Fields(line)[2], "."), 10, 64); err == nil {
				active = val
			}
		} else if strings.HasPrefix(line, "Pages inactive:") {
			if val, err := strconv.ParseInt(strings.TrimSuffix(strings.Fields(line)[2], "."), 10, 64); err == nil {
				inactive = val
			}
		} else if strings.HasPrefix(line, "Pages speculative:") {
			if val, err := strconv.ParseInt(strings.TrimSuffix(strings.Fields(line)[2], "."), 10, 64); err == nil {
				speculative = val
			}
		} else if strings.HasPrefix(line, "Pages wired down:") {
			if val, err := strconv.ParseInt(strings.TrimSuffix(strings.Fields(line)[3], "."), 10, 64); err == nil {
				wired = val
			}
		}
	}

	memory.Free = free * pageSize
	memory.Used = (active + inactive + speculative + wired) * pageSize
	memory.Cached = (inactive + speculative) * pageSize

	cmd = exec.CommandContext(ctx, "sysctl", "-n", "hw.memsize")
	output, err = cmd.Output()
	if err == nil {
		if total, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
			memory.Total = total
		}
	}

	if memory.Total > 0 {
		memory.Available = memory.Total - memory.Used
		memory.Usage = float64(memory.Used) / float64(memory.Total) * 100
	}

	return memory, nil
}

func (c *Collector) collectMemoryInfoLinux(ctx context.Context) (models.MemoryInfo, error) {
	memory := models.MemoryInfo{}

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return memory, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		value, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			memory.Total = value
		case "MemFree":
			memory.Free = value
		case "Cached":
			memory.Cached = value
		case "MemAvailable":
			memory.Available = value
		}
	}

	memory.Used = memory.Total - memory.Available
	if memory.Total > 0 {
		memory.Usage = float64(memory.Used) / float64(memory.Total) * 100
	}

	return memory, nil
}

func (c *Collector) collectDiskInfo(ctx context.Context) (map[string]models.DiskInfo, error) {
	disks := make(map[string]models.DiskInfo)

	cmd := exec.CommandContext(ctx, "df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return disks, err
	}

	lines := strings.Split(string(output), "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		mountpoint := fields[8]
		if mountpoint == "/tmp" || strings.HasPrefix(mountpoint, "/tmp") {
			continue
		}

		total := parseDiskSize(fields[1])
		used := parseDiskSize(fields[2])
		free := parseDiskSize(fields[3])
		usage := float64(0)
		if total > 0 {
			usage = float64(used) / float64(total) * 100
		}

		disks[mountpoint] = models.DiskInfo{
			MountPoint: mountpoint,
			FileSystem: fields[0],
			Total:      total,
			Used:       used,
			Free:       free,
			Usage:      usage,
		}
	}

	return disks, nil
}

func parseDiskSize(s string) int64 {
	multiplier := int64(1)
	if strings.HasSuffix(s, "Ti") || strings.HasSuffix(s, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "Ti")
		s = strings.TrimSuffix(s, "T")
	} else if strings.HasSuffix(s, "Gi") || strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "Gi")
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "Mi") || strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "Mi")
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "Ki") || strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "Ki")
		s = strings.TrimSuffix(s, "K")
	}

	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val * multiplier
	}
	return 0
}

func (c *Collector) collectLoadAverage(ctx context.Context) (models.LoadAverage, error) {
	load := models.LoadAverage{}

	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "uptime")
		output, err := cmd.Output()
		if err != nil {
			return load, err
		}

		line := strings.TrimSpace(string(output))
		fields := strings.Fields(line)
		for i := 0; i < len(fields); i++ {
			if strings.Contains(fields[i], "load") && i+3 < len(fields) {
				if val, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(fields[i+1], "load"), ","), 64); err == nil {
					load.OneMin = val
				}
				if val, err := strconv.ParseFloat(strings.TrimSuffix(fields[i+2], ","), 64); err == nil {
					load.FiveMin = val
				}
				if val, err := strconv.ParseFloat(fields[i+3], 64); err == nil {
					load.FifteenMin = val
				}
				break
			}
		}
	} else if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/loadavg")
		if err != nil {
			return load, err
		}

		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			if val, err := strconv.ParseFloat(fields[0], 64); err == nil {
				load.OneMin = val
			}
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				load.FiveMin = val
			}
			if val, err := strconv.ParseFloat(fields[2], 64); err == nil {
				load.FifteenMin = val
			}
		}
	}

	return load, nil
}

func (c *Collector) getProcessCount(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, "ps", "-axo", "pid")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count - 1, nil
}
