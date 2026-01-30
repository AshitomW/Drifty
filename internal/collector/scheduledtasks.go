package collector

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectScheduledTasks(ctx context.Context) (models.ScheduledTasks, error) {
	tasks := models.ScheduledTasks{
		CronJobs:      make(map[string]models.CronJob),
		SystemdTimers: make(map[string]models.SystemdTimer),
		LaunchdJobs:   make(map[string]models.LaunchdJob),
	}

	if runtime.GOOS == "darwin" {
		if c.config.ScheduledTasks.LaunchdJobs {
			jobs, err := c.collectLaunchdJobs(ctx)
			if err == nil {
				tasks.LaunchdJobs = jobs
			}
		}
	} else if runtime.GOOS == "linux" {
		if c.config.ScheduledTasks.CronJobs {
			crons, err := c.collectCronJobs(ctx)
			if err == nil {
				tasks.CronJobs = crons
			}
		}

		if c.config.ScheduledTasks.SystemdTimers {
			timers, err := c.collectSystemdTimers(ctx)
			if err == nil {
				tasks.SystemdTimers = timers
			}
		}
	}

	return tasks, nil
}

func (c *Collector) collectCronJobs(ctx context.Context) (map[string]models.CronJob, error) {
	cronJobs := make(map[string]models.CronJob)

	cronPaths := []string{
		"/etc/crontab",
		"/etc/cron.d/",
	}

	for _, path := range cronPaths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				fullPath := path + "/" + entry.Name()
				jobs, err := c.parseCronFile(ctx, fullPath)
				if err == nil {
					for name, job := range jobs {
						cronJobs[name+"_"+entry.Name()] = job
					}
				}
			}
		} else {
			jobs, err := c.parseCronFile(ctx, path)
			if err == nil {
				for name, job := range jobs {
					cronJobs[name] = job
				}
			}
		}
	}

	return cronJobs, nil
}

func (c *Collector) parseCronFile(ctx context.Context, path string) (map[string]models.CronJob, error) {
	jobs := make(map[string]models.CronJob)

	data, err := os.ReadFile(path)
	if err != nil {
		return jobs, err
	}

	lines := strings.Split(string(data), "\n")
	lineNum := 0
	for _, line := range lines {
		lineNum++
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		schedule := strings.Join(fields[0:5], " ")
		command := strings.Join(fields[5:], " ")

		user := "root"
		if strings.Contains(path, "/etc/cron.d/") || !strings.HasSuffix(path, "/etc/crontab") {
			if len(fields) >= 6 {
				user = fields[5]
				command = strings.Join(fields[6:], " ")
			}
		}

		jobs[path+":"+string(lineNum)] = models.CronJob{
			User:     user,
			Schedule: schedule,
			Command:  command,
			Enabled:  true,
		}
	}

	return jobs, nil
}

func (c *Collector) collectSystemdTimers(ctx context.Context) (map[string]models.SystemdTimer, error) {
	timers := make(map[string]models.SystemdTimer)

	cmd := exec.CommandContext(ctx, "systemctl", "list-timers", "--all", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return timers, err
	}

	lines := strings.Split(string(output), "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "UNIT") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		name := strings.TrimSuffix(fields[0], ".timer")
		nextRun := parseSystemdTime(fields[1])
		lastRun := parseSystemdTime(fields[2])

		timers[name] = models.SystemdTimer{
			Name:        name,
			NextTrigger: nextRun,
			LastTrigger: lastRun,
			Enabled:     true,
			Active:      true,
		}
	}

	return timers, nil
}

func parseSystemdTime(s string) time.Time {
	if s == "-" || s == "" {
		return time.Time{}
	}

	t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", s)
	if err == nil {
		return t
	}

	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return t
	}

	return time.Time{}
}

func (c *Collector) collectLaunchdJobs(ctx context.Context) (map[string]models.LaunchdJob, error) {
	jobs := make(map[string]models.LaunchdJob)

	paths := []string{
		"/Library/LaunchDaemons",
		"/Library/LaunchAgents",
		os.Getenv("HOME") + "/Library/LaunchAgents",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
				continue
			}

			fullPath := path + "/" + entry.Name()
			job, err := c.parseLaunchdPlist(ctx, fullPath)
			if err == nil {
				jobs[job.Label] = job
			}
		}
	}

	return jobs, nil
}

func (c *Collector) parseLaunchdPlist(ctx context.Context, path string) (models.LaunchdJob, error) {
	job := models.LaunchdJob{
		Label:   path,
		Path:    path,
		Enabled: true,
	}

	cmd := exec.CommandContext(ctx, "launchctl", "list")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, path) {
				job.Running = true
				break
			}
		}
	}

	return job, nil
}
