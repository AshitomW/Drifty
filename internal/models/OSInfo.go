// Represents / Contains the operating sysem details
package models

type OSInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Arch    string `json:"arch" yaml:"arch"`
	Kernel  string `json:"kernel" yaml:"kernel"`
}
