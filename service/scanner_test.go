package service

import (
	"strings"
	"sync"
	"testing"
	"time"

	"network-scanner/logger"
	"network-scanner/model"
	"network-scanner/repository"
)

type dummyLogger struct{}

func (l *dummyLogger) Info(args ...interface{})  {}
func (l *dummyLogger) Error(args ...interface{}) {}
func (l *dummyLogger) Debug(args ...interface{}) {}
func (l *dummyLogger) Warn(args ...interface{})  {}

var _ logger.Logger = (*dummyLogger)(nil)

type fakeDeviceRepo struct {
	mu      sync.RWMutex
	byID    map[string]model.Device
	ipIndex map[string]string
}

func newFakeDeviceRepo() *fakeDeviceRepo {
	return &fakeDeviceRepo{
		byID:    make(map[string]model.Device),
		ipIndex: make(map[string]string),
	}
}

func (r *fakeDeviceRepo) Save(d model.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if d.ID == "" {
		d.ID = d.IPAddress
	}
	r.byID[d.ID] = d
	r.ipIndex[d.IPAddress] = d.ID
}

func (r *fakeDeviceRepo) GetAll() []model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []model.Device
	for _, d := range r.byID {
		out = append(out, d)
	}
	return out
}

func (r *fakeDeviceRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID = make(map[string]model.Device)
	r.ipIndex = make(map[string]string)
}

func (r *fakeDeviceRepo) FindByIP(ip string) *model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.ipIndex[ip]; ok {
		if d, ok := r.byID[id]; ok {
			copy := d
			return &copy
		}
	}
	return nil
}

func (r *fakeDeviceRepo) FindByID(id string) (*model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.byID[id]; ok {
		copy := d
		return &copy, nil
	}
	return nil, nil
}

func (r *fakeDeviceRepo) UpdateTags(id string, tags []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.byID[id]
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	norm := make([]string, 0, len(tags))
	for _, t := range tags {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt == "" || seen[tt] {
			continue
		}
		seen[tt] = true
		norm = append(norm, tt)
	}
	if len(norm) > 10 {
		norm = norm[:10]
	}
	d.Tags = norm
	r.byID[id] = d
	return nil
}

func (r *fakeDeviceRepo) Search(q string) ([]model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q = strings.ToLower(strings.TrimSpace(q))
	var out []model.Device
	for _, d := range r.byID {
		if strings.Contains(d.IPAddress, q) || strings.Contains(d.Hostname, q) {
			out = append(out, d)
			continue
		}
		for _, tag := range d.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				out = append(out, d)
				break
			}
		}
	}
	if q == "" {
		for _, d := range r.byID {
			out = append(out, d)
		}
	}
	return out, nil
}

var _ repository.DeviceRepository = (*fakeDeviceRepo)(nil)

func TestStartScan_SingleIP(t *testing.T) {
	repo := newFakeDeviceRepo()
	log := &dummyLogger{}
	svc := NewScannerService(repo, log)

	svc.StartScan("127.0.0.1/32")

	deadline := time.Now().Add(2 * time.Second)
	for {
		if len(svc.GetDevices()) == 1 || time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	devices := svc.GetDevices()
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	d := devices[0]
	if d.IPAddress != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %s", d.IPAddress)
	}
	if d.ID == "" {
		t.Errorf("device ID should not be empty")
	}
	if d.Status == "online" {
		if d.FirstSeen.IsZero() {
			t.Errorf("expected FirstSeen to be set for online device")
		}
		if d.LastSeen.IsZero() {
			t.Errorf("expected LastSeen to be set for online device")
		}
	} else if d.Status == "offline" {
		if !d.LastSeen.IsZero() || !d.FirstSeen.IsZero() {
			t.Logf("offline device still had seen timestamps: %v / %v", d.FirstSeen, d.LastSeen)
		}
	} else {
		t.Errorf("unexpected device status: %s", d.Status)
	}
}
