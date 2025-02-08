package service

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

type CompanyProfileService struct {
	repo *repository.CompanyProfileRepository
}

func NewCompanyProfileService(repo *repository.CompanyProfileRepository) *CompanyProfileService {
	return &CompanyProfileService{repo: repo}
}

func (s *CompanyProfileService) GetCompanyProfiles(ctx context.Context, filters map[string]string) ([]dtos.CompanyProfileListDTO, int, error) {

	resultChan := make(chan dtos.GetCompanyProfilesResult, 1)

	go func() {
		companyProfiles, total, err := s.repo.GetCompanyProfiles(ctx, filters)
		resultChan <- dtos.GetCompanyProfilesResult{CompanyProfiles: companyProfiles, Total: total, Err: err}
	}()

	select {
	case res := <-resultChan:
		return res.CompanyProfiles, res.Total, res.Err
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
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
	companyProfileChan := make(chan *dtos.CompanyProfileDetailDTO, 1)
	errChan := make(chan error, 1)

	go func() {
		companyProfile, err := s.repo.GetCompanyProfileByID(ctx, params)
		if err != nil {
			errChan <- err
			return
		}
		companyProfileChan <- companyProfile
	}()

	select {
	case companyProfile := <-companyProfileChan:
		return companyProfile, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
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
