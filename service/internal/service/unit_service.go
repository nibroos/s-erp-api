package service

import (
	"context"
	"fmt"
	"log"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type UnitService struct {
	repo     *repository.UnitRepository
	utilRepo *repository.UtilRepository
}

func NewUnitService(repo *repository.UnitRepository, utilRepo *repository.UtilRepository) *UnitService {
	return &UnitService{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *UnitService) GetUnits(ctx context.Context, filters map[string]string) ([]dtos.UnitListDTO, int, error) {
	units, total, err := s.repo.GetUnits(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return units, total, nil
}

func (s *UnitService) CreateUnit(ctx context.Context, unit *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateUnit(tx, unit); err != nil {
		tx.Rollback()
		return nil, err
	}

	return unit, nil
}

func (s *UnitService) GetUnitByID(ctx context.Context, params *dtos.GetUnitParams) (*dtos.UnitDetailDTO, error) {
	unit, err := s.repo.GetUnitByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (s *UnitService) UpdateUnit(ctx context.Context, unit *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateUnit(tx, unit); err != nil {
		tx.Rollback()
		return nil, err
	}

	return unit, nil
}

func (s *UnitService) DeleteUnit(ctx context.Context, params *dtos.GetUnitParams, tx *gorm.DB) error {
	if err := s.repo.DeleteUnit(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *UnitService) RestoreUnit(ctx context.Context, params *dtos.GetUnitParams, tx *gorm.DB) error {
	if err := s.repo.RestoreUnit(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *UnitService) ExcelGetUnits(ctx context.Context, filters map[string]string) ([]byte, error) {
	units, _, err := s.GetUnits(ctx, filters)
	if err != nil {
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
func (s *UnitService) CsvGetUnits(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	units, _, err := s.GetUnits(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetUnits error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
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
