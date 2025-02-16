package service

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ItemGroupService struct {
	repo     *repository.ItemGroupRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewItemGroupService(repo *repository.ItemGroupRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *ItemGroupService {
	return &ItemGroupService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *ItemGroupService) GetItemGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ItemGroupListDTO, int, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-GetItemGroups", opentracing.ChildOf(span.Context()))

	itemGroups, total, err := s.repo.GetItemGroups(ctx.Context(), filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return itemGroups, total, nil
}

func (s *ItemGroupService) CreateItemGroup(ctx *fiber.Ctx, itemGroup *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-CreateItemGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateItemGroup(tx, itemGroup, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) GetItemGroupByID(ctx *fiber.Ctx, params *dtos.GetItemGroupParams, span opentracing.Span) (*dtos.ItemGroupDetailDTO, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-GetItemGroupByID", opentracing.ChildOf(span.Context()))

	itemGroup, err := s.repo.GetItemGroupByID(ctx.Context(), params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return itemGroup, nil
}

func (s *ItemGroupService) UpdateItemGroup(ctx *fiber.Ctx, itemGroup *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-UpdateItemGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateItemGroup(tx, itemGroup, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemGroup, nil
}

func (s *ItemGroupService) DeleteItemGroup(ctx *fiber.Ctx, params *dtos.GetItemGroupParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemGroupService-DeleteItemGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteItemGroup(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ItemGroupService) RestoreItemGroup(ctx *fiber.Ctx, params *dtos.GetItemGroupParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemGroupService-RestoreItemGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreItemGroup(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemGroupService) ExcelGetItemGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-ExcelGetItemGroups", opentracing.ChildOf(span.Context()))

	itemGroups, _, err := s.GetItemGroups(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "itemGroups"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("ItemGroups", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, itemGroup := range itemGroups {
		row := []interface{}{
			itemGroup.ID,
			itemGroup.Name,
			itemGroup.Description,
			itemGroup.CreatedAt,
			itemGroup.UpdatedAt,
		}
		file.SetSheetRow("ItemGroups", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ItemGroupService) CsvGetItemGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemGroupService-CsvGetItemGroups", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	itemGroups, _, err := s.GetItemGroups(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx.Context(), &companyProfileParams)
	appName := "App"
	if err != nil {
		defer childSpan.Finish()
		log.Println("CsvGetItemGroups error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, itemGroup := range itemGroups {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			itemGroup.ID,
			itemGroup.Name,
			utils.GetPtrVal(itemGroup.Description),
			utils.GetPtrVal(itemGroup.Remark),
			utils.GetPtrVal(itemGroup.CreatedAt),
			utils.GetPtrVal(itemGroup.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
