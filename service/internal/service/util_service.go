package service

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type UtilService struct {
	repo *repository.UtilRepository
}

func NewUtilService(repo *repository.UtilRepository) *UtilService {
	return &UtilService{repo: repo}
}

func (s *UtilService) GetCompanyProfileByID(ctx context.Context, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
	companyProfile, err := s.repo.GetCompanyProfileByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return companyProfile, nil
}
