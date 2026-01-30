package models

type CPUInfo struct {
	Cores  int     `json:"cores" yaml:"cores"`
	Model  string  `json:"model" yaml:"model"`
	Usage  float64 `json:"usage" yaml:"usage"`
	User   float64 `json:"user" yaml:"user"`
	System float64 `json:"system" yaml:"system"`
	Idle   float64 `json:"idle" yaml:"idle"`
}

type MemoryInfo struct {
	Total     int64   `json:"total" yaml:"total"`
	Used      int64   `json:"used" yaml:"used"`
	Available int64   `json:"available" yaml:"available"`
	Free      int64   `json:"free" yaml:"free"`
	Cached    int64   `json:"cached" yaml:"cached"`
	Usage     float64 `json:"usage" yaml:"usage"`
}

type DiskInfo struct {
	Device      string  `json:"device" yaml:"device"`
	MountPoint  string  `json:"mountpoint" yaml:"mountpoint"`
	FileSystem  string  `json:"filesystem" yaml:"filesystem"`
	Total       int64   `json:"total" yaml:"total"`
	Used        int64   `json:"used" yaml:"used"`
	Free        int64   `json:"free" yaml:"free"`
	Usage       float64 `json:"usage" yaml:"usage"`
	InodesTotal int64   `json:"inodes_total" yaml:"inodes_total"`
	InodesUsed  int64   `json:"inodes_used" yaml:"inodes_used"`
	InodesFree  int64   `json:"inodes_free" yaml:"inodes_free"`
}

type LoadAverage struct {
	OneMin     float64 `json:"one_min" yaml:"one_min"`
	FiveMin    float64 `json:"five_min" yaml:"five_min"`
	FifteenMin float64 `json:"fifteen_min" yaml:"fifteen_min"`
}

type SystemResources struct {
	CPU          CPUInfo             `json:"cpu" yaml:"cpu"`
	Memory       MemoryInfo          `json:"memory" yaml:"memory"`
	Disks        map[string]DiskInfo `json:"disks" yaml:"disks"`
	LoadAverage  LoadAverage         `json:"load_average" yaml:"load_average"`
	ProcessCount int                 `json:"process_count" yaml:"process_count"`
}
