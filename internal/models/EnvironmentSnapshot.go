// Environment snapshot will represen the complete environment state
package models

import "time"



type EnvironmentSnapshot struct{
	ID string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Hostname string `json:"hostname" yaml:"hostname"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	OS OSInfo `json:"os" yaml:"os"`
	Files map[string]FileInfo `json:"files" yaml:"files"`
	EnvVars map[string]EnvVar `json:"env_vars" yaml:"env_vars"`
	Packages map[string]PackageInfo `json:"packages" yaml:"packages"`
	Services map[string]ServiceInfo `json:"services" yaml:"services"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
