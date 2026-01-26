package models

// Represents a single drift detection



type DriftItem struct{
	Type string `json:"type" yaml:"type"` // added , removed modified
	Category string `json:"category" yaml:"category"` // file, environment variables (envvar), packages , services
	Name string `json:"name" yaml:"name"`
	SourceVal interface{} `json:"source_value,omitempty" yaml:"source_value,omitempty"`
	TargetVal interface{} `json:"target_value,omitempty" yaml:"target_value,omitempty"`
	Severity string `json:"severity" yaml:"severity"` // critical , warning , infromation
	Message string `json:"message" yaml:"message"`
}



