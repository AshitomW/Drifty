// Package Info will represent the installed packages information

package models

type PackageInfo struct {
	Name         string `json:"name" yaml:"name"`
	Version      string `json:"version" yaml:"version"`
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Manager      string `json:"manager" yaml:"manager"` // Represens the package manager like yum , apt etc.
	Exists       bool   `json:"exists" yaml:"exists"`
}
