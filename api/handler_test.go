package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"network-scanner/logger"
	"network-scanner/model"
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

var _ logger.Logger = (*dummyLogger)(nil)

type fakeHistoryRepo struct {
	DeletedIDs []int64
	Cleared    bool
}

func (f *fakeHistoryRepo) Save(h model.ScanHistory) (int64, error) { return 1, nil }
func (f *fakeHistoryRepo) GetAll() ([]model.ScanHistory, error)    { return nil, nil }

func (f *fakeHistoryRepo) GetByID(id int64) (*model.ScanHistory, error) {
	now := time.Now()
	return &model.ScanHistory{
		ID:          id,
		IPRange:     "127.0.0.1/32",
		StartedAt:   now,
		DeviceCount: 1,
	}, nil
}

func (f *fakeHistoryRepo) Delete(id int64) error {
	f.DeletedIDs = append(f.DeletedIDs, id)
	return nil
}

func (f *fakeHistoryRepo) Clear() error {
	f.Cleared = true
	return nil
}

var _ repository.ScanHistoryRepository = (*fakeHistoryRepo)(nil)

func TestStartScanAndGetDevices(t *testing.T) {
	devRepo := repository.NewInMemoryRepository()
	histRepo := &fakeHistoryRepo{}
	log := &dummyLogger{}

	scanner := service.NewScannerService(devRepo, histRepo, log)
	scanHandler := NewScanHandler(scanner, log, histRepo)
	deviceHandler := NewDeviceHandler(scanner, log)

	body := []byte(`{"ip_range": "127.0.0.1/32"}`)
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	scanHandler.StartScan(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		_ = resp.Body.Close()
		t.Fatalf("expected 202 Accepted, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	deadline := time.Now().Add(3 * time.Second)
	for {
		req = httptest.NewRequest(http.MethodGet, "/devices", nil)
		w = httptest.NewRecorder()
		deviceHandler.GetDevices(w, req)
		resp = w.Result()

		if resp.StatusCode == http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			var devices []map[string]interface{}
			_ = json.Unmarshal(data, &devices)

			if len(devices) == 1 && devices[0]["ip_address"] == "127.0.0.1" {
				break
			}
		} else {
			_ = resp.Body.Close()
		}

		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for device record")
		}
		time.Sleep(50 * time.Millisecond)
	}

	req = httptest.NewRequest(http.MethodGet, "/devices", nil)
	w = httptest.NewRecorder()
	deviceHandler.GetDevices(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var devices []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0]["ip_address"] != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %v", devices[0]["ip_address"])
	}
}

func TestDeleteScanHistory_OK(t *testing.T) {

	histRepo := &fakeHistoryRepo{}
	log := &dummyLogger{}
	handler := NewScanHistoryHandler(histRepo, log)

	req := httptest.NewRequest(http.MethodDelete, "/scan-history/42", nil)
	w := httptest.NewRecorder()
	handler.DeleteScanHistory(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if len(histRepo.DeletedIDs) != 1 || histRepo.DeletedIDs[0] != 42 {
		t.Fatalf("expected DeletedIDs=[42], got %v", histRepo.DeletedIDs)
	}
}

func TestClearScanHistory_OK(t *testing.T) {
	histRepo := &fakeHistoryRepo{}
	log := &dummyLogger{}
	handler := NewScanHistoryHandler(histRepo, log)

	req := httptest.NewRequest(http.MethodDelete, "/scan-history", nil)
	w := httptest.NewRecorder()
	handler.ClearScanHistory(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !histRepo.Cleared {
		t.Fatalf("expected Cleared=true, got false")
	}
}
