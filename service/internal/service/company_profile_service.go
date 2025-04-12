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

func (s *CompanyProfileService) CreateCompanyProfile(ctx *fiber.Ctx, companyProfile *models.CompanyProfile, bankInformations []*models.BankInformation, tx *gorm.DB, span opentracing.Span) (*models.CompanyProfile, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-CreateCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCompanyProfile(tx, companyProfile, bankInformations, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) GetCompanyProfileByID(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, span opentracing.Span) (*dtos.CompanyProfileWithBanksDTO, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetCompanyProfileByID", opentracing.ChildOf(span.Context()))

	companyProfile, err := s.repo.GetCompanyProfileByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return companyProfile, nil
}

func (s *CompanyProfileService) UpdateCompanyProfile(ctx *fiber.Ctx, companyProfile *models.CompanyProfile, bankInformations []*models.BankInformation, tx *gorm.DB, span opentracing.Span) (*models.CompanyProfile, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-UpdateCompanyProfile", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCompanyProfile(tx, companyProfile, bankInformations, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return companyProfile, nil
}

func (s *CompanyProfileService) DeleteCompanyProfile(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-DeleteCompanyProfile", opentracing.ChildOf(span.Context()))
	if err := s.repo.DeleteCompanyProfile(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CompanyProfileService) RestoreCompanyProfile(ctx *fiber.Ctx, params *dtos.GetCompanyProfileParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-RestoreCompanyProfile", opentracing.ChildOf(span.Context()))
	if err := s.repo.RestoreCompanyProfile(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CompanyProfileService) GetBankInformations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BankInformationListDTO, int, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetBankInformations")

	bankInformations, total, err := s.repo.GetBankInformations(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return bankInformations, total, nil
}

func (s *CompanyProfileService) GetBankInformationByID(ctx *fiber.Ctx, params *dtos.GetBankInformationParams, span opentracing.Span) (*dtos.BankInformationDetailDTO, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetBankInformationByID", opentracing.ChildOf(span.Context()))

	bankInformation, err := s.repo.GetBankInformationByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return bankInformation, nil
}

func (s *CompanyProfileService) CreateBankInformation(ctx *fiber.Ctx, bankInformation *models.BankInformation, tx *gorm.DB, span opentracing.Span) (*models.BankInformation, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-CreateBankInformation", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateBankInformation(tx, bankInformation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return bankInformation, nil
}

func (s *CompanyProfileService) UpdateBankInformation(ctx *fiber.Ctx, bankInformation *models.BankInformation, tx *gorm.DB, span opentracing.Span) (*models.BankInformation, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-UpdateBankInformation", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateBankInformation(tx, bankInformation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return bankInformation, nil
}

func (s *CompanyProfileService) DeleteBankInformation(ctx *fiber.Ctx, id uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-DeleteBankInformation", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteBankInformation(tx, id, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CompanyProfileService) RestoreBankInformation(ctx *fiber.Ctx, id uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CompanyProfileService-RestoreBankInformation", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreBankInformation(tx, id, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CompanyProfileService) GetBankInformationsWithCompany(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.BankInformationWithCompanyDTO, int, error) {
	childSpan := opentracing.StartSpan("CompanyProfileService-GetBankInformationsWithCompany", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	bankInformations, total, err := s.repo.GetBankInformationsWithCompany(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return bankInformations, total, nil
}
