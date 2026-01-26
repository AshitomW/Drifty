// EnvVar will represent an environment variable

package models

type EnvVar struct{
	Name string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
	Exists bool `json:"exists" yaml:"exists"`
}





