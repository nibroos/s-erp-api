package service

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/opentracing/opentracing-go"
)

type CompanyProfileService struct {
	repo   *repository.CompanyProfileRepository
	tracer opentracing.Tracer
}

func NewCompanyProfileService(repo *repository.CompanyProfileRepository, tracer opentracing.Tracer) *CompanyProfileService {
	return &CompanyProfileService{repo: repo, tracer: tracer}
}

func (s *CompanyProfileService) GetCompanyProfiles(ctx context.Context, filters map[string]string) ([]dtos.CompanyProfileListDTO, int, error) {
	companyProfiles, total, err := s.repo.GetCompanyProfiles(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return companyProfiles, total, nil
}

func (s *CompanyProfileService) CreateCompanyProfile(ctx context.Context, companyProfile *models.CompanyProfile) (*models.CompanyProfile, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create companyProfile
	if err := s.repo.CreateCompanyProfile(tx, companyProfile); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) GetCompanyProfileByID(ctx context.Context, params *dtos.GetCompanyProfileParams) (*dtos.CompanyProfileDetailDTO, error) {
	companyProfile, err := s.repo.GetCompanyProfileByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return companyProfile, nil
}

func (s *CompanyProfileService) UpdateCompanyProfile(ctx context.Context, companyProfile *models.CompanyProfile) (*models.CompanyProfile, error) {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Update companyProfile
	if err := s.repo.UpdateCompanyProfile(tx, companyProfile); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) DeleteCompanyProfile(ctx context.Context, params *dtos.GetCompanyProfileParams) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete companyProfile
	if err := s.repo.DeleteCompanyProfile(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *CompanyProfileService) RestoreCompanyProfile(ctx context.Context, params *dtos.GetCompanyProfileParams) error {
	// Transaction handling
	tx := s.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	// Restore companyProfile
	if err := s.repo.RestoreCompanyProfile(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
