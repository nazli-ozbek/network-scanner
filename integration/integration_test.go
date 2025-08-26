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
	devRepo := repository.NewInMemoryRepository()
	log := &dummyLogger{}

	scanner := service.NewScannerService(devRepo, log)
	scanHandler := api.NewScanHandler(scanner, log)
	deviceHandler := api.NewDeviceHandler(scanner, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", scanHandler.StartScan)
	mux.HandleFunc("/devices", deviceHandler.GetDevices)

	server := httptest.NewServer(mux)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"ip_range": "127.0.0.1/32"})
	resp, err := http.Post(server.URL+"/scan", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("scan POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202 but got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	deadline := time.Now().Add(3 * time.Second)
	for {
		devicesResp, err := http.Get(server.URL + "/devices")
		if err == nil && devicesResp.StatusCode == http.StatusOK {
			var devices []map[string]interface{}
			_ = json.NewDecoder(devicesResp.Body).Decode(&devices)
			_ = devicesResp.Body.Close()
			if len(devices) == 1 && devices[0]["ip_address"] == "127.0.0.1" {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for device record")
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err = http.Get(server.URL + "/devices")
	if err != nil {
		t.Fatalf("failed to get devices: %v", err)
	}
	defer resp.Body.Close()

	var devices []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0]["ip_address"] != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %v", devices[0]["ip_address"])
	}
}
