package repository

import (
	"network-scanner/model"
	"sync"
)

type InMemoryRepository struct {
	mu      sync.RWMutex
	devices map[string]model.Device
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		devices: make(map[string]model.Device),
	}
}

func (r *InMemoryRepository) Save(device model.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[device.IPAddress] = device
}

func (r *InMemoryRepository) GetAll() []model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Device, 0, len(r.devices))
	for _, device := range r.devices {
		result = append(result, device)
	}
	return result
}

func (r *InMemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices = make(map[string]model.Device)
}

func (r *InMemoryRepository) FindByIP(ip string) *model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.devices[ip]; ok {
		return &d
	}
	return nil
}

var _ DeviceRepository = (*InMemoryRepository)(nil)
