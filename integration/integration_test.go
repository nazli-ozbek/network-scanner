package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"network-scanner/api"
	"network-scanner/repository"
	"network-scanner/service"
	"testing"
	"time"
)

type dummyLogger struct{}

func (l *dummyLogger) Info(args ...interface{})  {}
func (l *dummyLogger) Error(args ...interface{}) {}
func (l *dummyLogger) Debug(args ...interface{}) {}
func (l *dummyLogger) Warn(args ...interface{})  {}

func TestScanAndGetDevicesIntegration(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	logger := &dummyLogger{}
	scanner := service.NewScannerService(repo, logger)
	scanHandler := api.NewScanHandler(scanner, logger)
	deviceHandler := api.NewDeviceHandler(scanner, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", scanHandler.StartScan)
	mux.HandleFunc("/devices", deviceHandler.GetDevices)

	server := httptest.NewServer(mux)
	defer server.Close()

	payload := map[string]string{"ip_range": "8.8.8.8/32"}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(server.URL+"/scan", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202 but got %d", resp.StatusCode)
	}

	time.Sleep(3 * time.Second)

	resp, err = http.Get(server.URL + "/devices")
	if err != nil {
		t.Fatalf("failed to get devices: %v", err)
	}
	defer resp.Body.Close()

	var devices []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(devices) == 0 {
		t.Errorf("expected at least one device, got 0")
	}

	found := false
	for _, d := range devices {
		if d["ip_address"] == "8.8.8.8" && d["is_online"] == true {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 8.8.8.8 to be online, but was not found")
	}
}
