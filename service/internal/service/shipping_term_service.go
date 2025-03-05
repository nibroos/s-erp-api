package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ShippingTermService struct {
	repo     *repository.ShippingTermRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewShippingTermService(repo *repository.ShippingTermRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *ShippingTermService {
	return &ShippingTermService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *ShippingTermService) GetShippingTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ShippingTermListDTO, int, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-GetShippingTerms", opentracing.ChildOf(span.Context()))

	shippingTerms, total, err := s.repo.GetShippingTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return shippingTerms, total, nil
}

func (s *ShippingTermService) GetShippingTermByID(ctx *fiber.Ctx, params *dtos.GetShippingTermParams, span opentracing.Span) (*dtos.ShippingTermDetailDTO, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-GetShippingTermByID", opentracing.ChildOf(span.Context()))

	term, err := s.repo.GetShippingTermByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return term, nil
}

func (s *ShippingTermService) CreateShippingTerm(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-CreateShippingTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateShippingTerm(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *ShippingTermService) UpdateShippingTerm(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-UpdateShippingTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateShippingTerm(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *ShippingTermService) DeleteShippingTerm(ctx *fiber.Ctx, params *dtos.GetShippingTermParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ShippingTermService-DeleteShippingTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteShippingTerm(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ShippingTermService) RestoreShippingTerm(ctx *fiber.Ctx, params *dtos.GetShippingTermParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ShippingTermService-RestoreShippingTerm", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreShippingTerm(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ShippingTermService) ExcelGetShippingTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-ExcelGetShippingTerms", opentracing.ChildOf(span.Context()))

	shippingTerms, _, err := s.GetShippingTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "ShippingTerms"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("ShippingTerms", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, term := range shippingTerms {
		row := []interface{}{
			term.ID,
			term.Name,
			term.Description,
			term.CreatedAt,
			term.UpdatedAt,
		}
		file.SetSheetRow("ShippingTerms", fmt.Sprintf("A%d", i+2), &row)
	}

	file.SetActiveSheet(index)

	if err := file.SaveAs("output.xlsx"); err != nil {
		fmt.Println("Error saving file:", err)
		return nil, err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *ShippingTermService) CsvGetShippingTerms(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ShippingTermService-CsvGetShippingTerms", opentracing.ChildOf(span.Context()))

	filters["is_csv"] = "1"
	shippingTerms, _, err := s.GetShippingTerms(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		defer childSpan.Finish()
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Shipping Terms\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	for _, term := range shippingTerms {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			term.ID,
			term.Name,
			utils.GetPtrVal(term.Description),
			utils.GetPtrVal(term.Remark),
			utils.GetPtrVal(term.CreatedAt),
			utils.GetPtrVal(term.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
