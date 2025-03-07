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

type ItemSubGroupService struct {
	repo     *repository.ItemSubGroupRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewItemSubGroupService(repo *repository.ItemSubGroupRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *ItemSubGroupService {
	return &ItemSubGroupService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *ItemSubGroupService) GetItemSubGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ItemSubGroupListDTO, int, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-GetItemSubGroups", opentracing.ChildOf(span.Context()))

	itemSubGroups, total, err := s.repo.GetItemSubGroups(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return itemSubGroups, total, nil
}

func (s *ItemSubGroupService) CreateItemSubGroup(ctx *fiber.Ctx, itemSubGroup *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-CreateItemSubGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateItemSubGroup(tx, itemSubGroup, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) GetItemSubGroupByID(ctx *fiber.Ctx, params *dtos.GetItemSubGroupParams, span opentracing.Span) (*dtos.ItemSubGroupDetailDTO, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-GetItemSubGroupByID", opentracing.ChildOf(span.Context()))

	itemSubGroup, err := s.repo.GetItemSubGroupByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return itemSubGroup, nil
}

func (s *ItemSubGroupService) UpdateItemSubGroup(ctx *fiber.Ctx, itemSubGroup *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-UpdateItemSubGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateItemSubGroup(tx, itemSubGroup, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemSubGroup, nil
}

func (s *ItemSubGroupService) DeleteItemSubGroup(ctx *fiber.Ctx, params *dtos.GetItemSubGroupParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemSubGroupService-DeleteItemSubGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteItemSubGroup(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ItemSubGroupService) RestoreItemSubGroup(ctx *fiber.Ctx, params *dtos.GetItemSubGroupParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemSubGroupService-RestoreItemSubGroup", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreItemSubGroup(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemSubGroupService) ExcelGetItemSubGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-ExcelGetItemSubGroups", opentracing.ChildOf(span.Context()))

	itemSubGroups, _, err := s.GetItemSubGroups(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "itemSubGroups"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("ItemSubGroups", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, itemSubGroup := range itemSubGroups {
		row := []interface{}{
			itemSubGroup.ID,
			itemSubGroup.Name,
			itemSubGroup.Description,
			itemSubGroup.CreatedAt,
			itemSubGroup.UpdatedAt,
		}
		file.SetSheetRow("ItemSubGroups", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ItemSubGroupService) CsvGetItemSubGroups(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemSubGroupService-CsvGetItemSubGroups", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	itemSubGroups, _, err := s.GetItemSubGroups(ctx, filters, childSpan)
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
	csv += "Item Subgroups\n"
	csv += "\n"

	csv += "ID,Name,Sub Group,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, itemSubGroup := range itemSubGroups {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s\n",
			itemSubGroup.ID,
			itemSubGroup.Name,
			itemSubGroup.GroupName,
			utils.GetPtrVal(itemSubGroup.Description),
			utils.GetPtrVal(itemSubGroup.Remark),
			utils.GetPtrVal(itemSubGroup.CreatedAt),
			utils.GetPtrVal(itemSubGroup.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
