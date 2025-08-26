package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"network-scanner/model"
	"network-scanner/repository"
	"network-scanner/service"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type dummyLogger struct{}

func (l *dummyLogger) Info(args ...interface{})  {}
func (l *dummyLogger) Error(args ...interface{}) {}
func (l *dummyLogger) Debug(args ...interface{}) {}
func (l *dummyLogger) Warn(args ...interface{})  {}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbFile := "test_scan.db"
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}
	return db, cleanup
}

func TestStartScanAndGetDevices_SQLite(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := repository.NewSQLiteRepositoryWithDB(db, &dummyLogger{})
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	scanner := service.NewScannerService(repo, &dummyLogger{})
	scanHandler := NewScanHandler(scanner, &dummyLogger{})
	deviceHandler := NewDeviceHandler(scanner, &dummyLogger{})

	body := []byte(`{"ip_range": "127.0.0.1/32"}`)
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	scanHandler.StartScan(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	deadline := time.Now().Add(3 * time.Second)
	for {
		req = httptest.NewRequest(http.MethodGet, "/devices", nil)
		w = httptest.NewRecorder()
		deviceHandler.GetDevices(w, req)
		resp = w.Result()

		if resp.StatusCode == http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var devices []model.Device
			if err := json.Unmarshal(data, &devices); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			if len(devices) == 1 && devices[0].IPAddress == "127.0.0.1" {
				if devices[0].ID == "" {
					t.Errorf("device ID should not be empty")
				}
				if devices[0].Status == "online" {
					if devices[0].FirstSeen.IsZero() {
						t.Errorf("FirstSeen should not be zero for online device")
					}
					if devices[0].LastSeen.IsZero() {
						t.Errorf("LastSeen should not be zero for online device")
					}
				} else {
					t.Logf("Device is offline; skipping FirstSeen/LastSeen checks")
				}
				break
			}
		} else {
			resp.Body.Close()
		}

		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for device record")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
