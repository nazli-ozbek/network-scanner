package service

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

type ManufacturerResolver interface {
	Resolve(mac string) string
}

type OfflineManufacturerResolver struct {
	lookup map[string]string
}

func NewOfflineManufacturerResolver(path string) *OfflineManufacturerResolver {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open MAC vendor CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read MAC vendor CSV: %v", err)
	}

	lookup := make(map[string]string)
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 2 {
			continue
		}
		macPrefix := strings.ToUpper(strings.TrimSpace(row[0]))
		vendorName := strings.TrimSpace(row[1])
		if macPrefix != "" && vendorName != "" {
			lookup[macPrefix] = vendorName
		}
	}

	return &OfflineManufacturerResolver{lookup: lookup}
}

func (r *OfflineManufacturerResolver) Resolve(mac string) string {
	parts := strings.Split(strings.ToUpper(mac), ":")
	if len(parts) < 3 {
		return ""
	}
	prefix := strings.Join(parts[:3], ":")
	return r.lookup[prefix]
}
