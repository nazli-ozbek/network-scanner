package repository

import "network-scanner/model"

type IPRangeRepository interface {
	Save(r model.IPRange) error
	GetAll() ([]model.IPRange, error)
	Delete(id string) error
}
