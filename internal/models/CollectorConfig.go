package models

// Collector config defines what to collect

type CollectorConfig struct {
	Files           FileCollectorConfig            `yaml:"files"`
	EnvVars         EnvVarCollectorConfig          `yaml:"env_vars"`
	ProcessEnvVars  ProcessEnvVarCollectorConfig   `yaml:"process_env_vars"`
	Packages        PackageCollectorConfig         `yaml:"packages"`
	Services        ServiceCollectorConfig         `yaml:"services"`
	Network         NetworkCollectorConfig         `yaml:"network"`
	Docker          DockerCollectorConfig          `yaml:"docker"`
	SystemResources SystemResourcesCollectorConfig `yaml:"system_resources"`
	ScheduledTasks  ScheduledTasksCollectorConfig  `yaml:"scheduled_tasks"`
	Certificates    CertificateCollectorConfig     `yaml:"certificates"`
	UsersGroups     UserGroupCollectorConfig       `yaml:"users_groups"`
}

type FileCollectorConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Paths        []string `yaml:"paths"`
	ExcludePaths []string `yaml:"exclude_paths"`
	FollowLinks  bool     `yaml:"follow_links"`
	MaxDepth     int      `yaml:"max_depth"`
	HashAlgo     string   `yaml:"hash_algo"` // md5 sha256
}

type EnvVarCollectorConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Include     []string `yaml:"include"`      // regex patterns
	Exclude     []string `yaml:"exclude"`      // regex patterns
	MaskSecrets bool     `yaml:"mask_secrets"` // mask sensitive values
}

type PackageCollectorConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Managers []string `yaml:"managers"` // apt ,yum, go...
}

type ServiceCollectorConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Include  []string `yaml:"include"`
	Exclude  []string `yaml:"exclude"`
	InitType string   `yaml:"init_type"` // systemd, sysvinit, openrc
}

type ProcessEnvVarCollectorConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Processes    []string `yaml:"processes"`     // process names to collect env vars for (e.g., "node", "php", "python")
	MaxProcesses int      `yaml:"max_processes"` // max number of processes to collect env vars from
	MaskSecrets  bool     `yaml:"mask_secrets"`  // mask sensitive values
	Exclude      []string `yaml:"exclude"`       // regex patterns to exclude specific env vars
}

type NetworkCollectorConfig struct {
	Enabled       bool `yaml:"enabled"`
	Interfaces    bool `yaml:"interfaces"`
	Routes        bool `yaml:"routes"`
	DNS           bool `yaml:"dns"`
	FirewallRules bool `yaml:"firewall_rules"`
}

type DockerCollectorConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Containers bool   `yaml:"containers"`
	Images     bool   `yaml:"images"`
	Volumes    bool   `yaml:"volumes"`
	Networks   bool   `yaml:"networks"`
	SocketPath string `yaml:"socket_path"` // e.g., /var/run/docker.sock
}

type SystemResourcesCollectorConfig struct {
	Enabled bool `yaml:"enabled"`
	CPU     bool `yaml:"cpu"`
	Memory  bool `yaml:"memory"`
	Disks   bool `yaml:"disks"`
	Load    bool `yaml:"load"`
}

type ScheduledTasksCollectorConfig struct {
	Enabled       bool `yaml:"enabled"`
	CronJobs      bool `yaml:"cron_jobs"`
	SystemdTimers bool `yaml:"systemd_timers"`
	LaunchdJobs   bool `yaml:"launchd_jobs"`
}

type CertificateCollectorConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Paths         []string `yaml:"paths"`          // paths to scan for certificates
	Extensions    []string `yaml:"extensions"`     // .pem, .crt, .cer, .key
	DaysThreshold int      `yaml:"days_threshold"` // alert if expiring within X days
}

type UserGroupCollectorConfig struct {
	Enabled   bool `yaml:"enabled"`
	Users     bool `yaml:"users"`
	Groups    bool `yaml:"groups"`
	SudoRules bool `yaml:"sudo_rules"`
}
