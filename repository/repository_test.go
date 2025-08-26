package repository

import (
	"network-scanner/model"
	"reflect"
	"testing"
	"time"
)

func TestSaveAndGetAll_NewModel(t *testing.T) {
	repo := NewInMemoryRepository()

	dev := model.Device{
		ID:           "device-1",
		IPAddress:    "192.168.1.10",
		MACAddress:   "AA:BB:CC:DD:EE:FF",
		Hostname:     "test-device",
		Status:       "online",
		Manufacturer: "VendorX",
		Tags:         []string{"lab", "server"},
		LastSeen:     time.Now().UTC(),
		FirstSeen:    time.Now().UTC(),
	}
	repo.Save(dev)

	all := repo.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 device, got %d", len(all))
	}

	got := all[0]
	if got.IPAddress != dev.IPAddress {
		t.Errorf("expected IP %s, got %s", dev.IPAddress, got.IPAddress)
	}
	if got.Status != dev.Status {
		t.Errorf("expected status %q, got %q", dev.Status, got.Status)
	}
	if got.ID != dev.ID {
		t.Errorf("expected ID %s, got %s", dev.ID, got.ID)
	}
	if !reflect.DeepEqual(got.Tags, dev.Tags) {
		t.Errorf("expected tags %v, got %v", dev.Tags, got.Tags)
	}
	if got.Manufacturer != dev.Manufacturer {
		t.Errorf("expected manufacturer %s, got %s", dev.Manufacturer, got.Manufacturer)
	}
}

func TestClear(t *testing.T) {
	repo := NewInMemoryRepository()

	repo.Save(model.Device{ID: "d1", IPAddress: "192.168.1.5"})
	repo.Save(model.Device{ID: "d2", IPAddress: "192.168.1.6"})

	repo.Clear()

	all := repo.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 devices after Clear, got %d", len(all))
	}
}

func TestFindByIP(t *testing.T) {
	repo := NewInMemoryRepository()

	dev := model.Device{ID: "n1", IPAddress: "192.168.1.20", Hostname: "node-1"}
	repo.Save(dev)

	got := repo.FindByIP("192.168.1.20")
	if got == nil {
		t.Fatalf("expected to find device by IP")
	}
	if got.ID != "n1" {
		t.Errorf("expected ID 'n1', got %q", got.ID)
	}
	if got.Hostname != "node-1" {
		t.Errorf("expected hostname 'node-1', got %q", got.Hostname)
	}
}

func TestUpdateTags_Normalization(t *testing.T) {
	repo := NewInMemoryRepository()

	d := model.Device{ID: "n2", IPAddress: "192.168.1.21"}
	repo.Save(d)

	err := repo.UpdateTags("n2", []string{"  Web ", "db", "web", "DB  ", " ", "", "  ", "cache", "web", "db", "api"})
	if err != nil {
		t.Fatalf("UpdateTags error: %v", err)
	}

	got, _ := repo.FindByID("n2")
	if got == nil {
		t.Fatalf("expected device after UpdateTags")
	}

	expected := []string{"web", "db", "cache", "api"}
	actualSet := make(map[string]bool)
	for _, tag := range got.Tags {
		actualSet[tag] = true
	}

	for _, tag := range expected {
		if !actualSet[tag] {
			t.Errorf("expected tag %q to be in result, got %v", tag, got.Tags)
		}
	}
	if len(got.Tags) > 10 {
		t.Errorf("expected at most 10 tags, got %d", len(got.Tags))
	}
}

func TestSearch(t *testing.T) {
	repo := NewInMemoryRepository()

	repo.Save(model.Device{ID: "a", IPAddress: "192.168.1.30", Hostname: "alpha", Tags: []string{"lab"}})
	repo.Save(model.Device{ID: "b", IPAddress: "192.168.1.31", Hostname: "beta", Tags: []string{"prod"}})
	repo.Save(model.Device{ID: "c", IPAddress: "10.0.0.5", Hostname: "gamma", Tags: []string{"Lab", "db"}})

	res, err := repo.Search("192.168.1.3")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("expected 2 results by IP prefix, got %d", len(res))
	}

	res, _ = repo.Search("beta")
	if len(res) != 1 || res[0].ID != "b" {
		t.Errorf("expected 1 result with ID 'b', got %v", ids(res))
	}

	res, _ = repo.Search("lab")
	if len(res) != 2 {
		t.Errorf("expected 2 results with tag 'lab', got %d", len(res))
	}

	res, _ = repo.Search("")
	if len(res) != 3 {
		t.Errorf("expected 3 results for empty query, got %d", len(res))
	}
}

func ids(devs []model.Device) []string {
	var out []string
	for _, d := range devs {
		out = append(out, d.ID)
	}
	return out
}
