package model

import "time"

type ScanHistory struct {
	ID          int64     `json:"id"`
	IPRange     string    `json:"ip_range"`
	StartedAt   time.Time `json:"started_at"`
	DeviceCount int       `json:"device_count"`
}
