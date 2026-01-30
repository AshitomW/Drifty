package models

import "time"

type CronJob struct {
	User     string `json:"user" yaml:"user"`
	Schedule string `json:"schedule" yaml:"schedule"`
	Command  string `json:"command" yaml:"command"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

type SystemdTimer struct {
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	NextTrigger time.Time `json:"next_trigger" yaml:"next_trigger"`
	LastTrigger time.Time `json:"last_trigger,omitempty" yaml:"last_trigger,omitempty"`
	Enabled     bool      `json:"enabled" yaml:"enabled"`
	Active      bool      `json:"active" yaml:"active"`
}

type LaunchdJob struct {
	Label     string   `json:"label" yaml:"label"`
	Path      string   `json:"path" yaml:"path"`
	RunAtLoad bool     `json:"run_at_load" yaml:"run_at_load"`
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	Running   bool     `json:"running" yaml:"running"`
	Program   string   `json:"program,omitempty" yaml:"program,omitempty"`
	Arguments []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
}

type ScheduledTask struct {
	Type    string `json:"type" yaml:"type"` // cron, systemd, launchd
	Content string `json:"content" yaml:"content"`
	Exists  bool   `json:"exists" yaml:"exists"`
}

type ScheduledTasks struct {
	CronJobs      map[string]CronJob      `json:"cron_jobs,omitempty" yaml:"cron_jobs,omitempty"`
	SystemdTimers map[string]SystemdTimer `json:"systemd_timers,omitempty" yaml:"systemd_timers,omitempty"`
	LaunchdJobs   map[string]LaunchdJob   `json:"launchd_jobs,omitempty" yaml:"launchd_jobs,omitempty"`
}
