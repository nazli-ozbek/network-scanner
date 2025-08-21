package repository

import "network-scanner/model"

type DeviceRepository interface {
	Save(device model.Device)
	GetAll() []model.Device
	Clear()
	FindByIP(ip string) *model.Device

	FindByID(id string) (*model.Device, error)
	UpdateTags(id string, tags []string) error
	Search(query string) ([]model.Device, error)
}
