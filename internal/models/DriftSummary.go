package models

// Drift Summary provides the drift statistics

type DriftSummary struct {
	TotalDrifts   int            `json:"total_drifts" yaml:"total_drifts"`
	CriticalCount int            `json:"critical_count" yaml:"critical_count"`
	WarningCount  int            `json:"warning_count" yaml:"warning_count"`
	InfoCount     int            `json:"info_count" yaml:"info_count"`
	ByCategory    map[string]int `json:"by_category" yaml:"by_category"`
	ByType        map[string]int `json:"by_type" yaml:"by_type"`
}
