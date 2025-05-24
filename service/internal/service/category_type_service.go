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

type CategoryTypeService struct {
	repo     *repository.CategoryTypeRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewCategoryTypeService(repo *repository.CategoryTypeRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *CategoryTypeService {
	return &CategoryTypeService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *CategoryTypeService) GetCategoryTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CategoryTypeListDTO, int, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-GetCategoryTypes", opentracing.ChildOf(span.Context()))

	categoryTypes, total, err := s.repo.GetCategoryTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return categoryTypes, total, nil
}

func (s *CategoryTypeService) CreateCategoryType(ctx *fiber.Ctx, categoryType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-CreateCategoryType", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCategoryType(tx, categoryType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return categoryType, nil
}

func (s *CategoryTypeService) GetCategoryTypeByID(ctx *fiber.Ctx, params *dtos.GetCategoryTypeParams, span opentracing.Span) (*dtos.CategoryTypeDetailDTO, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-GetCategoryTypeByID", opentracing.ChildOf(span.Context()))

	categoryType, err := s.repo.GetCategoryTypeByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return categoryType, nil
}

func (s *CategoryTypeService) UpdateCategoryType(ctx *fiber.Ctx, categoryType *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-UpdateCategoryType", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCategoryType(tx, categoryType, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return categoryType, nil
}

func (s *CategoryTypeService) DeleteCategoryType(ctx *fiber.Ctx, params *dtos.GetCategoryTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CategoryTypeService-DeleteCategoryType", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteCategoryType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CategoryTypeService) RestoreCategoryType(ctx *fiber.Ctx, params *dtos.GetCategoryTypeParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CategoryTypeService-RestoreCategoryType", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreCategoryType(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CategoryTypeService) ExcelGetCategoryTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-ExcelGetCategoryTypes", opentracing.ChildOf(span.Context()))

	categoryTypes, _, err := s.GetCategoryTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "categoryTypes"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("CategoryTypes", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, categoryType := range categoryTypes {
		row := []interface{}{
			categoryType.ID,
			categoryType.Name,
			categoryType.Description,
			categoryType.CreatedAt,
			categoryType.UpdatedAt,
		}
		file.SetSheetRow("CategoryTypes", fmt.Sprintf("A%d", i+2), &row)
	}

	// Set active sheet of the workbook
	file.SetActiveSheet(index)

	// Save the file
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

// github.com/xuri/excelize/v2
func (s *CategoryTypeService) CsvGetCategoryTypes(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CategoryTypeService-CsvGetCategoryTypes", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	categoryTypes, _, err := s.GetCategoryTypes(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// get company profile
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
	csv += "CategoryTypes\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, categoryType := range categoryTypes {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			categoryType.ID,
			categoryType.Name,
			utils.GetPtrVal(categoryType.Description),
			utils.GetPtrVal(categoryType.Remark),
			utils.GetPtrVal(categoryType.CreatedAt),
			utils.GetPtrVal(categoryType.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
