package model

import "time"

type Device struct {
	ID           string    `json:"id"`
	IPAddress    string    `json:"ip_address"`
	MACAddress   string    `json:"mac_address"`
	Hostname     string    `json:"hostname"`
	Status       string    `json:"status"`
	Manufacturer string    `json:"manufacturer"`
	Tags         []string  `json:"tags"`
	LastSeen     time.Time `json:"last_seen"`
	FirstSeen    time.Time `json:"first_seen"`
}
