package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type CompanyProfileService struct {
	repo   *repository.CompanyProfileRepository
	tracer opentracing.Tracer
}

func NewCompanyProfileService(repo *repository.CompanyProfileRepository, tracer opentracing.Tracer) *CompanyProfileService {
	return &CompanyProfileService{repo: repo, tracer: tracer}
}

func (s *CompanyProfileService) GetCompanyProfiles(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CompanyProfileListDTO, int, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetCompanyProfiles")

	companyProfiles, total, err := s.repo.GetCompanyProfiles(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return companyProfiles, total, nil
}

func (s *CompanyProfileService) CreateCompanyProfile(ctx *fiber.Ctx, companyProfile *models.CompanyProfile, tx *gorm.DB, span opentracing.Span) (*models.CompanyProfile, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-CreateCompanyProfile", opentracing.ChildOf(span.Context()))

	// Create companyProfile
	if err := s.repo.CreateCompanyProfile(tx, companyProfile, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) GetCompanyProfileByID(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, span opentracing.Span) (*dtos.CompanyProfileDetailDTO, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetCompanyProfileByID", opentracing.ChildOf(span.Context()))

	companyProfile, err := s.repo.GetCompanyProfileByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return companyProfile, nil
}

func (s *CompanyProfileService) UpdateCompanyProfile(ctx *fiber.Ctx, companyProfile *models.CompanyProfile, tx *gorm.DB, span opentracing.Span) (*models.CompanyProfile, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-UpdateCompanyProfile", opentracing.ChildOf(span.Context()))

	// Update companyProfile
	if err := s.repo.UpdateCompanyProfile(tx, companyProfile, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) DeleteCompanyProfile(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-DeleteCompanyProfile", opentracing.ChildOf(span.Context()))
	// Delete companyProfile
	if err := s.repo.DeleteCompanyProfile(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CompanyProfileService) RestoreCompanyProfile(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-RestoreCompanyProfile", opentracing.ChildOf(span.Context()))
	// Restore companyProfile
	if err := s.repo.RestoreCompanyProfile(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
