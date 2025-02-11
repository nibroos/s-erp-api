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

type UnitService struct {
	repo     *repository.UnitRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewUnitService(repo *repository.UnitRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *UnitService {
	return &UnitService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *UnitService) GetUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.UnitListDTO, int, error) {
	childSpan := opentracing.StartSpan("UnitService-GetUnits", opentracing.ChildOf(span.Context()))

	units, total, err := s.repo.GetUnits(ctx.Context(), filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return units, total, nil
}

func (s *UnitService) CreateUnit(ctx *fiber.Ctx, unit *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("UnitService-CreateUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateUnit(tx, unit, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return unit, nil
}

func (s *UnitService) GetUnitByID(ctx *fiber.Ctx, params *dtos.GetUnitParams, span opentracing.Span) (*dtos.UnitDetailDTO, error) {
	childSpan := opentracing.StartSpan("UnitService-GetUnitByID", opentracing.ChildOf(span.Context()))

	unit, err := s.repo.GetUnitByID(ctx.Context(), params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return unit, nil
}

func (s *UnitService) UpdateUnit(ctx *fiber.Ctx, unit *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("UnitService-UpdateUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateUnit(tx, unit, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return unit, nil
}

func (s *UnitService) DeleteUnit(ctx *fiber.Ctx, params *dtos.GetUnitParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UnitService-DeleteUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteUnit(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *UnitService) RestoreUnit(ctx *fiber.Ctx, params *dtos.GetUnitParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("UnitService-RestoreUnit", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreUnit(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *UnitService) ExcelGetUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("UnitService-ExcelGetUnits", opentracing.ChildOf(span.Context()))

	units, _, err := s.GetUnits(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "units"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Units", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, unit := range units {
		row := []interface{}{
			unit.ID,
			unit.Name,
			unit.Description,
			unit.CreatedAt,
			unit.UpdatedAt,
		}
		file.SetSheetRow("Units", fmt.Sprintf("A%d", i+2), &row)
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
func (s *UnitService) CsvGetUnits(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("UnitService-CsvGetUnits", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	units, _, err := s.GetUnits(ctx, filters, childSpan)
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
		log.Println("CsvGetUnits error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Units\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, unit := range units {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			unit.ID,
			unit.Name,
			utils.GetPtrVal(unit.Description),
			utils.GetPtrVal(unit.Remark),
			utils.GetPtrVal(unit.CreatedAt),
			utils.GetPtrVal(unit.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
