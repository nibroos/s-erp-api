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

type ItemUnitService struct {
	repo     *repository.ItemUnitRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewItemUnitService(repo *repository.ItemUnitRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *ItemUnitService {
	return &ItemUnitService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *ItemUnitService) GetItemUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ItemUnitListDTO, int, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-GetItemUnits", opentracing.ChildOf(span.Context()))

	itemUnits, total, err := s.repo.GetItemUnits(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return itemUnits, total, nil
}

func (s *ItemUnitService) CreateItemUnit(ctx *fiber.Ctx, itemUnit *models.ItemUnit, tx *gorm.DB, span opentracing.Span) (*models.ItemUnit, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-CreateItemUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateItemUnit(tx, itemUnit, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemUnit, nil
}

func (s *ItemUnitService) GetItemUnitByID(ctx *fiber.Ctx, params *dtos.GetItemUnitParams, span opentracing.Span) (*dtos.ItemUnitDetailDTO, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-GetItemUnitByID", opentracing.ChildOf(span.Context()))

	itemUnit, err := s.repo.GetItemUnitByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return itemUnit, nil
}

func (s *ItemUnitService) UpdateItemUnit(ctx *fiber.Ctx, itemUnit *models.ItemUnit, tx *gorm.DB, span opentracing.Span) (*models.ItemUnit, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-UpdateItemUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateItemUnit(tx, itemUnit, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return itemUnit, nil
}

func (s *ItemUnitService) DeleteItemUnit(ctx *fiber.Ctx, params *dtos.GetItemUnitParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitService-DeleteItemUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteItemUnit(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *ItemUnitService) RestoreItemUnit(ctx *fiber.Ctx, params *dtos.GetItemUnitParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("ItemUnitService-RestoreItemUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreItemUnit(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *ItemUnitService) ExcelGetItemUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-ExcelGetItemUnits", opentracing.ChildOf(span.Context()))

	itemUnits, _, err := s.GetItemUnits(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "itemUnits"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("ItemUnits", "A1", &[]string{"ID", "Item Name", "Unit Name", "Conversion", "Price Sell", "Price Buy", "Created At", "Updated At"})

	for i, itemUnit := range itemUnits {
		row := []interface{}{
			itemUnit.ID,
			utils.GetPtrVal(itemUnit.ProductName),
			utils.GetPtrVal(itemUnit.UnitName),
			utils.GetFloatPtrVal(itemUnit.Conversion),
			utils.GetFloatPtrVal(itemUnit.PriceSell),
			utils.GetFloatPtrVal(itemUnit.PriceBuy),
			itemUnit.CreatedAt,
			itemUnit.UpdatedAt,
		}
		file.SetSheetRow("ItemUnits", fmt.Sprintf("A%d", i+2), &row)
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
func (s *ItemUnitService) CsvGetItemUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("ItemUnitService-CsvGetItemUnits", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	itemUnits, _, err := s.GetItemUnits(ctx, filters, childSpan)
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
	csv += "Master Item\n"
	csv += "\n"

	csv += "ID,Item Name, Unit Name, Conversion, Price Sell, Price Buy, Created At, Updated At\n"
	// Build CSV rows
	for _, itemUnit := range itemUnits {
		csv += fmt.Sprintf("%d,%s,%s,%f,%f,%f,%s,%s\n",
			itemUnit.ID,
			utils.GetPtrVal(itemUnit.ProductName),
			utils.GetPtrVal(itemUnit.UnitName),
			utils.GetFloatPtrVal(itemUnit.Conversion),
			utils.GetFloatPtrVal(itemUnit.PriceSell),
			utils.GetFloatPtrVal(itemUnit.PriceBuy),
			utils.GetPtrVal(itemUnit.CreatedAt),
			utils.GetPtrVal(itemUnit.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
