package models

import "time"

// FileInfo will represent the file metadata

type FileInfo struct {
	Path        string    `json:"path" yaml:"path"`
	Hash        string    `json:"hash" yaml:"hash"`
	Size        int64     `json:"size" yaml:"size"`
	Mode        string    `json:"mode" yaml:"mode"`
	ModTime     time.Time `json:"mod_time" yaml:"mod_time"`
	Owner       string    `json:"owner" yaml:"owner"`
	Group       string    `json:"group" yaml:"group"`
	IsDirectory bool      `json:"is_directory" yaml:"is_directory"`
	Exists      bool      `json:"exists" yaml:"exists"`
}
