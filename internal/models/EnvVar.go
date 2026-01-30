// EnvVar will represent an environment variable

package models

type EnvVar struct {
	Name   string `json:"name" yaml:"name"`
	Value  string `json:"value" yaml:"value"`
	Exists bool   `json:"exists" yaml:"exists"`
}

type ProcessEnvVar struct {
	PID     int               `json:"pid" yaml:"pid"`
	Cmdline string            `json:"cmdline" yaml:"cmdline"`
	EnvVars map[string]EnvVar `json:"env_vars" yaml:"env_vars"`
}
