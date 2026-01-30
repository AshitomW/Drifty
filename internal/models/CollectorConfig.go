package models

// Collector config defines what to collect

type CollectorConfig struct {
	Files          FileCollectorConfig          `yaml:"files"`
	EnvVars        EnvVarCollectorConfig        `yaml:"env_vars"`
	ProcessEnvVars ProcessEnvVarCollectorConfig `yaml:"process_env_vars"`
	Packages       PackageCollectorConfig       `yaml:"packages"`
	Services       ServiceCollectorConfig       `yaml:"services"`
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
