package model

import "time"

type Device struct {
	IPAddress  string    `json:"ip_address"`
	MACAddress string    `json:"mac_address"`
	Hostname   string    `json:"hostname"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen,omitempty"`
}
