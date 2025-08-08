package repository

import (
	"network-scanner/model"
	"testing"
)

func TestSaveAndGetAll(t *testing.T) {
	repo := NewInMemoryRepository()

	device := model.Device{
		IPAddress:  "192.168.1.10",
		MACAddress: "AA:BB:CC:DD:EE:FF",
		Hostname:   "test-device",
		IsOnline:   true,
	}

	repo.Save(device)

	devices := repo.GetAll()
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	got := devices[0]
	if got.IPAddress != device.IPAddress {
		t.Errorf("expected IP %s, got %s", device.IPAddress, got.IPAddress)
	}
	if !got.IsOnline {
		t.Errorf("expected device to be online")
	}
}

func TestClear(t *testing.T) {
	repo := NewInMemoryRepository()

	device := model.Device{IPAddress: "192.168.1.5"}
	repo.Save(device)

	repo.Clear()

	devices := repo.GetAll()
	if len(devices) != 0 {
		t.Errorf("expected 0 devices after Clear, got %d", len(devices))
	}
}
