package models

type Container struct {
	ID      string            `json:"id" yaml:"id"`
	Name    string            `json:"name" yaml:"name"`
	Image   string            `json:"image" yaml:"image"`
	Status  string            `json:"status" yaml:"status"`
	State   string            `json:"state" yaml:"state"`
	Created string            `json:"created" yaml:"created"`
	Ports   []string          `json:"ports,omitempty" yaml:"ports,omitempty"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type Image struct {
	ID      string            `json:"id" yaml:"id"`
	Name    string            `json:"name" yaml:"name"`
	Tag     string            `json:"tag" yaml:"tag"`
	Size    int64             `json:"size" yaml:"size"`
	Created string            `json:"created" yaml:"created"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type Volume struct {
	Name       string `json:"name" yaml:"name"`
	Driver     string `json:"driver" yaml:"driver"`
	Mountpoint string `json:"mountpoint" yaml:"mountpoint"`
	Size       int64  `json:"size,omitempty" yaml:"size,omitempty"`
}

type Network struct {
	ID     string            `json:"id" yaml:"id"`
	Name   string            `json:"name" yaml:"name"`
	Driver string            `json:"driver" yaml:"driver"`
	Scope  string            `json:"scope" yaml:"scope"`
	Subnet string            `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type DockerConfig struct {
	Containers map[string]Container `json:"containers" yaml:"containers"`
	Images     map[string]Image     `json:"images" yaml:"images"`
	Volumes    map[string]Volume    `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Networks   map[string]Network   `json:"networks,omitempty" yaml:"networks,omitempty"`
}
