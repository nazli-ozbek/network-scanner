package service

import (
	"net"
	"network-scanner/model"
	"network-scanner/repository"
)

type RangeService struct {
	repo repository.IPRangeRepository
}

func NewRangeService(repo repository.IPRangeRepository) *RangeService {
	return &RangeService{repo: repo}
}

func (s *RangeService) Save(r model.IPRange) error {
	if _, _, err := net.ParseCIDR(r.Range); err != nil {
		return err
	}
	return s.repo.Save(r)
}

func (s *RangeService) List() ([]model.IPRange, error) {
	return s.repo.GetAll()
}

func (s *RangeService) Delete(id string) error {
	return s.repo.Delete(id)
}
