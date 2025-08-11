package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"network-scanner/repository"
	"network-scanner/service"
	"testing"
	"time"
)

func TestStartScanAndGetDevices(t *testing.T) {

	repo := repository.NewInMemoryRepository()
	scanner := service.NewScannerService(repo)
	handler := NewHandler(scanner)

	body := []byte(`{"ip_range": "127.0.0.1/32"}`)
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StartScan(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d", resp.StatusCode)
	}

	time.Sleep(2 * time.Second)

	req = httptest.NewRequest(http.MethodGet, "/devices", nil)
	w = httptest.NewRecorder()
	handler.GetDevices(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	var devices []map[string]interface{}
	err := json.Unmarshal(data, &devices)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	if devices[0]["ip_address"] != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %v", devices[0]["ip_address"])
	}
}
