package models

import "time"

type Certificate struct {
	Path         string    `json:"path" yaml:"path"`
	Domain       string    `json:"domain" yaml:"domain"`
	Issuer       string    `json:"issuer" yaml:"issuer"`
	Subject      string    `json:"subject" yaml:"subject"`
	NotBefore    time.Time `json:"not_before" yaml:"not_before"`
	NotAfter     time.Time `json:"not_after" yaml:"not_after"`
	SerialNumber string    `json:"serial_number" yaml:"serial_number"`
	Fingerprint  string    `json:"fingerprint" yaml:"fingerprint"`
	IsValid      bool      `json:"is_valid" yaml:"is_valid"`
	IsExpired    bool      `json:"is_expired" yaml:"is_expired"`
	DaysToExpire int       `json:"days_to_expire" yaml:"days_to_expire"`
}
