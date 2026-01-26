// Service infomation represnts the system service states

package models


type ServiceInfo struct {
	Name string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"` // Running , Stopped , Failed 
	Enabled bool `json:"enabled" yaml:"enabled"` // runs on boot or not?
	ActiveState string `json:"active_state" yaml:"active_state"`
	SubState string `json:"sub_state,omitempty" yaml:"sub_State,omitempty"`
	Exists bool `json:"exists" yaml:"exists"`
}