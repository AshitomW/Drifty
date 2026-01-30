package models

type UserInfo struct {
	Name    string `json:"name" yaml:"name"`
	UID     int    `json:"uid" yaml:"uid"`
	GID     int    `json:"gid" yaml:"gid"`
	HomeDir string `json:"home_dir" yaml:"home_dir"`
	Shell   string `json:"shell" yaml:"shell"`
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
}

type GroupInfo struct {
	Name    string   `json:"name" yaml:"name"`
	GID     int      `json:"gid" yaml:"gid"`
	Members []string `json:"members" yaml:"members"`
}

type SudoRule struct {
	Alias    string `json:"alias,omitempty" yaml:"alias,omitempty"`
	User     string `json:"user" yaml:"user"`
	Host     string `json:"host" yaml:"host"`
	RunAs    string `json:"runas,omitempty" yaml:"runas,omitempty"`
	Commands string `json:"commands" yaml:"commands"`
}

type UserGroupConfig struct {
	Users     map[string]UserInfo  `json:"users" yaml:"users"`
	Groups    map[string]GroupInfo `json:"groups" yaml:"groups"`
	SudoRules []SudoRule           `json:"sudo_rules,omitempty" yaml:"sudo_rules,omitempty"`
}
