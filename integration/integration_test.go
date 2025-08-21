package api_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"network-scanner/api"
// 	"network-scanner/logger"
// 	"network-scanner/model"
// 	"network-scanner/repository"
// 	"network-scanner/service"
// 	"testing"
// 	"time"
// )

// type dummyLogger struct{}

// func (l *dummyLogger) Info(args ...interface{})  {}
// func (l *dummyLogger) Error(args ...interface{}) {}
// func (l *dummyLogger) Debug(args ...interface{}) {}
// func (l *dummyLogger) Warn(args ...interface{})  {}

// var _ logger.Logger = (*dummyLogger)(nil)

// type fakeHistoryRepo struct {
// 	DeletedIDs []int64
// 	Cleared    bool
// }

// func (f *fakeHistoryRepo) Save(h model.ScanHistory) (int64, error) { return 1, nil }
// func (f *fakeHistoryRepo) GetAll() ([]model.ScanHistory, error)    { return nil, nil }
// func (f *fakeHistoryRepo) GetByID(id int64) (*model.ScanHistory, error) {
// 	return &model.ScanHistory{ID: id, IPRange: "127.0.0.1/32", StartedAt: time.Now(), DeviceCount: 1}, nil
// }
// func (f *fakeHistoryRepo) Delete(id int64) error { f.DeletedIDs = append(f.DeletedIDs, id); return nil }
// func (f *fakeHistoryRepo) Clear() error          { f.Cleared = true; return nil }

// var _ repository.ScanHistoryRepository = (*fakeHistoryRepo)(nil)

// func TestScanAndGetDevicesIntegration(t *testing.T) {
// 	devRepo := repository.NewInMemoryRepository()
// 	histRepo := &fakeHistoryRepo{}
// 	log := &dummyLogger{}

// 	scanner := service.NewScannerService(devRepo, histRepo, log)
// 	scanHandler := api.NewScanHandler(scanner, log, histRepo)
// 	deviceHandler := api.NewDeviceHandler(scanner, log)
// 	histHandler := api.NewScanHistoryHandler(histRepo, log)

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/scan", scanHandler.StartScan)
// 	mux.HandleFunc("/devices", deviceHandler.GetDevices)
// 	mux.HandleFunc("/scan-history", func(w http.ResponseWriter, r *http.Request) {
// 		switch r.Method {
// 		case http.MethodGet:
// 			histHandler.GetScanHistory(w, r)
// 		case http.MethodDelete:
// 			histHandler.ClearScanHistory(w, r)
// 		default:
// 			http.NotFound(w, r)
// 		}
// 	})
// 	mux.HandleFunc("/scan-history/", histHandler.DeleteScanHistory)

// 	server := httptest.NewServer(mux)
// 	defer server.Close()

// 	payload := map[string]string{"ip_range": "127.0.0.1/32"}
// 	body, _ := json.Marshal(payload)

// 	resp, err := http.Post(server.URL+"/scan", "application/json", bytes.NewReader(body))
// 	if err != nil {
// 		t.Fatalf("scan POST failed: %v", err)
// 	}
// 	if resp.StatusCode != http.StatusAccepted {
// 		t.Fatalf("expected status 202 but got %d", resp.StatusCode)
// 	}
// 	_ = resp.Body.Close()

// 	deadline := time.Now().Add(3 * time.Second)
// 	for {
// 		devicesResp, err := http.Get(server.URL + "/devices")
// 		if err == nil && devicesResp.StatusCode == http.StatusOK {
// 			var devices []map[string]interface{}
// 			_ = json.NewDecoder(devicesResp.Body).Decode(&devices)
// 			_ = devicesResp.Body.Close()
// 			if len(devices) == 1 && devices[0]["ip_address"] == "127.0.0.1" {
// 				break
// 			}
// 		}
// 		if time.Now().After(deadline) {
// 			t.Fatal("timeout waiting for device record")
// 		}
// 		time.Sleep(50 * time.Millisecond)
// 	}

// 	resp, err = http.Get(server.URL + "/devices")
// 	if err != nil {
// 		t.Fatalf("failed to get devices: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	var devices []map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
// 		t.Fatalf("failed to decode response: %v", err)
// 	}
// 	if len(devices) != 1 {
// 		t.Fatalf("expected 1 device, got %d", len(devices))
// 	}
// 	if devices[0]["ip_address"] != "127.0.0.1" {
// 		t.Errorf("expected IP 127.0.0.1, got %v", devices[0]["ip_address"])
// 	}
// }

// func TestDeleteScanHistory_OK(t *testing.T) {
// 	histRepo := &fakeHistoryRepo{}
// 	log := &dummyLogger{}
// 	histHandler := api.NewScanHistoryHandler(histRepo, log)

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/scan-history/", histHandler.DeleteScanHistory)

// 	server := httptest.NewServer(mux)
// 	defer server.Close()

// 	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/scan-history/42", nil)
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		t.Fatalf("delete request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
// 	}
// 	if len(histRepo.DeletedIDs) != 1 || histRepo.DeletedIDs[0] != 42 {
// 		t.Fatalf("expected DeletedIDs=[42], got %v", histRepo.DeletedIDs)
// 	}
// }

// func TestClearScanHistory_OK(t *testing.T) {
// 	histRepo := &fakeHistoryRepo{}
// 	log := &dummyLogger{}
// 	histHandler := api.NewScanHistoryHandler(histRepo, log)

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/scan-history", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodDelete {
// 			histHandler.ClearScanHistory(w, r)
// 			return
// 		}
// 		http.NotFound(w, r)
// 	})

// 	server := httptest.NewServer(mux)
// 	defer server.Close()

// 	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/scan-history", nil)
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		t.Fatalf("clear request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
// 	}
// 	if !histRepo.Cleared {
// 		t.Fatalf("expected Cleared=true, got false")
// 	}
// }
