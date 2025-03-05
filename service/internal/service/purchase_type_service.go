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

type PurchaseTypeService struct {
	repo     *repository.PurchaseTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPurchaseTypeService(repo *repository.PurchaseTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *PurchaseTypeService {
	return &PurchaseTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *PurchaseTypeService) GetPurchaseTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-GetPurchaseTypes", opentracing.ChildOf(span.Context()))

	purchaseTypes, total, err := s.repo.GetPurchaseTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return purchaseTypes, total, nil
}

func (s *PurchaseTypeService) GetPurchaseTypeByID(ctx *fiber.Ctx, params *dtos.GetPurchaseTypeParams, span opentracing.Span) (*dtos.PurchaseTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-GetPurchaseTypeByID", opentracing.ChildOf(span.Context()))

	term, err := s.repo.GetPurchaseTypeByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return term, nil
}

func (s *PurchaseTypeService) CreatePurchaseType(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-CreatePurchaseType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreatePurchaseType(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *PurchaseTypeService) UpdatePurchaseType(ctx *fiber.Ctx, term *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-UpdatePurchaseType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdatePurchaseType(tx, term, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return term, nil
}

func (s *PurchaseTypeService) DeletePurchaseType(ctx *fiber.Ctx, params *dtos.GetPurchaseTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseTypeService-DeletePurchaseType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePurchaseType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseTypeService) RestorePurchaseType(ctx *fiber.Ctx, params *dtos.GetPurchaseTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseTypeService-RestorePurchaseType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestorePurchaseType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseTypeService) ExcelGetPurchaseTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-ExcelGetPurchaseTypes", opentracing.ChildOf(span.Context()))

	purchaseTypes, _, err := s.GetPurchaseTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "PurchaseTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("PurchaseTypes", "A1", &[]string{"ID", "Name", "Code", "Description", "Created At", "Updated At"})

	for i, term := range purchaseTypes {
		row := []interface{}{
			term.ID,
			term.Name,
			term.Code,
			term.Description,
			term.CreatedAt,
			term.UpdatedAt,
		}
		file.SetSheetRow("PurchaseTypes", fmt.Sprintf("A%d", i+2), &row)
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

func (s *PurchaseTypeService) CsvGetPurchaseTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("PurchaseTypeService-CsvGetPurchaseTypes", opentracing.ChildOf(span.Context()))

	filters["is_csv"] = "1"
	purchaseTypes, _, err := s.GetPurchaseTypes(ctx, filters, childSpan)
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
	csv += "Purchase Types\n"
	csv += "\n"

	csv += "ID,Name,Code,Description,Remark,Created At,Updated At\n"
	for _, term := range purchaseTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s\n",
			term.ID,
			term.Name,
			utils.GetPtrVal(term.Code),
			utils.GetPtrVal(term.Description),
			utils.GetPtrVal(term.Remark),
			utils.GetPtrVal(term.CreatedAt),
			utils.GetPtrVal(term.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
