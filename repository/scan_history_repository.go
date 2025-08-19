package repository

import "network-scanner/model"

type ScanHistoryRepository interface {
	Save(h model.ScanHistory) (int64, error)
	GetAll() ([]model.ScanHistory, error)
	GetByID(id int64) (*model.ScanHistory, error)
	Delete(id int64) error
	Clear() error
}
