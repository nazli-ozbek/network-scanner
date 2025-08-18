package repository

import "network-scanner/model"

type DeviceRepository interface {
	Save(device model.Device)
	GetAll() []model.Device
	Clear()
	FindByIP(ip string) *model.Device
}
