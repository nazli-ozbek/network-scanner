package repository

import (
	"network-scanner/model"
	"strings"
	"sync"
)

type InMemoryRepository struct {
	mu      sync.RWMutex
	byID    map[string]model.Device
	ipIndex map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		byID:    make(map[string]model.Device),
		ipIndex: make(map[string]string),
	}
}

func (r *InMemoryRepository) Save(device model.Device) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := device.ID
	if id == "" {
		id = device.IPAddress
		device.ID = id
	}

	if old, ok := r.byID[id]; ok && old.IPAddress != device.IPAddress && old.IPAddress != "" {
		delete(r.ipIndex, old.IPAddress)
	}

	r.byID[id] = device
	if device.IPAddress != "" {
		r.ipIndex[device.IPAddress] = id
	}
}

func (r *InMemoryRepository) GetAll() []model.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Device, 0, len(r.byID))
	for _, d := range r.byID {
		out = append(out, d)
	}
	return out
}

func (r *InMemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID = make(map[string]model.Device)
	r.ipIndex = make(map[string]string)
}

func (r *InMemoryRepository) FindByIP(ip string) *model.Device {
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

func (r *InMemoryRepository) FindByID(id string) (*model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.byID[id]; ok {
		dd := d
		return &dd, nil
	}
	return nil, nil
}

func (r *InMemoryRepository) UpdateTags(id string, tags []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.byID[id]
	if !ok {
		return nil
	}

	seen := make(map[string]struct{}, len(tags))
	norm := make([]string, 0, len(tags))
	for _, t := range tags {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt == "" {
			continue
		}
		if _, exists := seen[tt]; !exists {
			seen[tt] = struct{}{}
			norm = append(norm, tt)
		}
	}
	d.Tags = norm
	r.byID[id] = d
	return nil
}

func (r *InMemoryRepository) Search(q string) ([]model.Device, error) {
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

var _ DeviceRepository = (*InMemoryRepository)(nil)
