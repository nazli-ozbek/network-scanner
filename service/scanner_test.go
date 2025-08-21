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
		byID:    map[string]model.Device{},
		ipIndex: map[string]string{},
	}
}

func (r *fakeDeviceRepo) Save(d model.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := d.ID
	if id == "" {
		id = d.IPAddress
		d.ID = id
	}
	if old, ok := r.byID[id]; ok && old.IPAddress != d.IPAddress && old.IPAddress != "" {
		delete(r.ipIndex, old.IPAddress)
	}
	r.byID[id] = d
	if d.IPAddress != "" {
		r.ipIndex[d.IPAddress] = id
	}
}

func (r *fakeDeviceRepo) GetAll() []model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Device, 0, len(r.byID))
	for _, v := range r.byID {
		out = append(out, v)
	}
	return out
}

func (r *fakeDeviceRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID = map[string]model.Device{}
	r.ipIndex = map[string]string{}
}

func (r *fakeDeviceRepo) FindByIP(ip string) *model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.ipIndex[ip]; ok {
		if d, ok2 := r.byID[id]; ok2 {
			dd := d
			return &dd
		}
	}
	return nil
}

func (r *fakeDeviceRepo) FindByID(id string) (*model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.byID[id]; ok {
		dd := d
		return &dd, nil
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
	seen := map[string]struct{}{}
	norm := make([]string, 0, len(tags))
	for _, t := range tags {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt == "" {
			continue
		}
		if _, ex := seen[tt]; !ex {
			seen[tt] = struct{}{}
			norm = append(norm, tt)
		}
	}
	d.Tags = norm
	r.byID[id] = d
	return nil
}

func (r *fakeDeviceRepo) Search(q string) ([]model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		out := make([]model.Device, 0, len(r.byID))
		for _, d := range r.byID {
			out = append(out, d)
		}
		return out, nil
	}
	match := func(d model.Device) bool {
		if strings.Contains(strings.ToLower(d.IPAddress), q) {
			return true
		}
		if strings.Contains(strings.ToLower(d.Hostname), q) {
			return true
		}
		for _, tag := range d.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				return true
			}
		}
		return false
	}
	var out []model.Device
	for _, d := range r.byID {
		if match(d) {
			out = append(out, d)
		}
	}
	return out, nil
}

var _ repository.DeviceRepository = (*fakeDeviceRepo)(nil)

type fakeHistoryRepo struct{}

func (f *fakeHistoryRepo) Save(h model.ScanHistory) (int64, error)      { return 1, nil }
func (f *fakeHistoryRepo) GetAll() ([]model.ScanHistory, error)         { return nil, nil }
func (f *fakeHistoryRepo) GetByID(id int64) (*model.ScanHistory, error) { return nil, nil }
func (f *fakeHistoryRepo) Delete(id int64) error                        { return nil }
func (f *fakeHistoryRepo) Clear() error                                 { return nil }

var _ repository.ScanHistoryRepository = (*fakeHistoryRepo)(nil)

func TestStartScan_SingleIP(t *testing.T) {
	devRepo := newFakeDeviceRepo()
	histRepo := &fakeHistoryRepo{}
	log := &dummyLogger{}

	svc := NewScannerService(devRepo, histRepo, log)
	svc.StartScan("127.0.0.1/32")

	deadline := time.Now().Add(2 * time.Second)
	for {
		if len(svc.GetDevices()) == 1 || time.Now().After(deadline) {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	devices := svc.GetDevices()
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].IPAddress != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1, got %s", devices[0].IPAddress)
	}
}
