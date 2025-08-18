package service

import (
	"network-scanner/repository"
	"testing"
	"time"
)

type dummyLogger struct{}

func (l *dummyLogger) Info(args ...interface{})  {}
func (l *dummyLogger) Error(args ...interface{}) {}
func (l *dummyLogger) Debug(args ...interface{}) {}
func (l *dummyLogger) Warn(args ...interface{})  {}

func TestStartScan_SingleIP(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	testLogger := &dummyLogger{}
	service := NewScannerService(repo, testLogger)

	service.StartScan("127.0.0.1/32")

	time.Sleep(2 * time.Second)

	devices := service.GetDevices()

	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	device := devices[0]
	if device.IPAddress != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %s", device.IPAddress)
	}
}
