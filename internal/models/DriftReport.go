package models

import "time"

// Drift report will contain the complete drift analysis
type DriftReport struct{
	ID string `json:"id" yaml:"id"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	SourceEnv string `json:"source_env" yaml:"source_env"`
	TargetEnv string `json:"target_env" yaml:"target_env"`
	SourceSnapshot string `json:"source_snapshot" yaml:"source_snapshot"`
	HasDrift bool `json:"has_drift" yaml:"has_drift"`
	Summary DriftSummary `json:"summary" yaml:"summary"`
	Drifts []DriftItem `json:"drifts" yaml:"drifts"`
}
