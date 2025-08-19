package service

import (
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
	data map[string]model.Device
}

func newFakeDeviceRepo() *fakeDeviceRepo      { return &fakeDeviceRepo{data: map[string]model.Device{}} }
func (r *fakeDeviceRepo) Save(d model.Device) { r.data[d.IPAddress] = d }
func (r *fakeDeviceRepo) GetAll() []model.Device {
	out := make([]model.Device, 0, len(r.data))
	for _, v := range r.data {
		out = append(out, v)
	}
	return out
}
func (r *fakeDeviceRepo) Clear() { r.data = map[string]model.Device{} }
func (r *fakeDeviceRepo) FindByIP(ip string) *model.Device {
	if d, ok := r.data[ip]; ok {
		return &d
	}
	return nil
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
