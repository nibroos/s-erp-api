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

type WarehouseService struct {
	repo     *repository.WarehouseRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewWarehouseService(repo *repository.WarehouseRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *WarehouseService {
	return &WarehouseService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *WarehouseService) GetWarehouses(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.WarehouseListDTO, int, error) {
	childSpan := opentracing.StartSpan("WarehouseService-GetWarehouses", opentracing.ChildOf(span.Context()))

	warehouses, total, err := s.repo.GetWarehouses(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return warehouses, total, nil
}

func (s *WarehouseService) GetWarehouseByID(ctx *fiber.Ctx, params *dtos.GetWarehouseParams, span opentracing.Span) (*dtos.WarehouseDetailDTO, error) {
	childSpan := opentracing.StartSpan("WarehouseService-GetWarehouseByID", opentracing.ChildOf(span.Context()))

	warehouse, err := s.repo.GetWarehouseByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return warehouse, nil
}

func (s *WarehouseService) CreateWarehouse(ctx *fiber.Ctx, warehouse *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("WarehouseService-CreateWarehouse", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateWarehouse(tx, warehouse, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return warehouse, nil
}

func (s *WarehouseService) UpdateWarehouse(ctx *fiber.Ctx, warehouse *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("WarehouseService-UpdateWarehouse", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateWarehouse(tx, warehouse, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return warehouse, nil
}

func (s *WarehouseService) DeleteWarehouse(ctx *fiber.Ctx, params *dtos.GetWarehouseParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("WarehouseService-DeleteWarehouse", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteWarehouse(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *WarehouseService) RestoreWarehouse(ctx *fiber.Ctx, params *dtos.GetWarehouseParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("WarehouseService-RestoreWarehouse", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreWarehouse(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *WarehouseService) ExcelGetWarehouses(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("WarehouseService-ExcelGetWarehouses", opentracing.ChildOf(span.Context()))

	warehouses, _, err := s.GetWarehouses(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "Warehouses"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Warehouses", "A1", &[]string{"ID", "Name", "Code", "Description", "Created At", "Updated At"})

	for i, warehouse := range warehouses {
		row := []interface{}{
			warehouse.ID,
			warehouse.Name,
			warehouse.Code,
			warehouse.Description,
			warehouse.CreatedAt,
			warehouse.UpdatedAt,
		}
		file.SetSheetRow("Warehouses", fmt.Sprintf("A%d", i+2), &row)
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

func (s *WarehouseService) CsvGetWarehouses(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("WarehouseService-CsvGetWarehouses", opentracing.ChildOf(span.Context()))

	filters["is_csv"] = "1"
	warehouses, _, err := s.GetWarehouses(ctx, filters, childSpan)
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
	csv += "Warehouses\n"
	csv += "\n"

	csv += "ID,Name,Code,Description,Remark,Created At,Updated At\n"
	for _, warehouse := range warehouses {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s\n",
			warehouse.ID,
			warehouse.Name,
			utils.GetPtrVal(warehouse.Code),
			utils.GetPtrVal(warehouse.Description),
			utils.GetPtrVal(warehouse.Remark),
			utils.GetPtrVal(warehouse.CreatedAt),
			utils.GetPtrVal(warehouse.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
