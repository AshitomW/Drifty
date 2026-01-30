// Environment snapshot will represen the complete environment state
package models

import "time"

type EnvironmentSnapshot struct {
	ID              string                 `json:"id" yaml:"id"`
	Name            string                 `json:"name" yaml:"name"`
	Hostname        string                 `json:"hostname" yaml:"hostname"`
	Timestamp       time.Time              `json:"timestamp" yaml:"timestamp"`
	OS              OSInfo                 `json:"os" yaml:"os"`
	Files           map[string]FileInfo    `json:"files" yaml:"files"`
	EnvVars         map[string]EnvVar      `json:"env_vars" yaml:"env_vars"`
	ProcessEnvVars  map[int]ProcessEnvVar  `json:"process_env_vars,omitempty" yaml:"process_env_vars,omitempty"`
	Packages        map[string]PackageInfo `json:"packages" yaml:"packages"`
	Services        map[string]ServiceInfo `json:"services" yaml:"services"`
	NetworkConfig   NetworkConfig          `json:"network_config,omitempty" yaml:"network_config,omitempty"`
	DockerConfig    DockerConfig           `json:"docker_config,omitempty" yaml:"docker_config,omitempty"`
	SystemResources SystemResources        `json:"system_resources,omitempty" yaml:"system_resources,omitempty"`
	ScheduledTasks  ScheduledTasks         `json:"scheduled_tasks,omitempty" yaml:"scheduled_tasks,omitempty"`
	Certificates    map[string]Certificate `json:"certificates,omitempty" yaml:"certificates,omitempty"`
	UserGroupConfig UserGroupConfig        `json:"user_group_config,omitempty" yaml:"user_group_config,omitempty"`
	Metadata        map[string]string      `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
