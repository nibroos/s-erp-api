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

type Pph23Service struct {
	repo     *repository.Pph23Repository
	utilRepo *repository.UtilRepository
}

func NewPph23Service(repo *repository.Pph23Repository, utilRepo *repository.UtilRepository) *Pph23Service {
	return &Pph23Service{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *Pph23Service) GetPph23s(ctx context.Context, filters map[string]string) ([]dtos.Pph23ListDTO, int, error) {
	pph23s, total, err := s.repo.GetPph23s(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return pph23s, total, nil
}

func (s *Pph23Service) CreatePph23(ctx context.Context, pph23 *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreatePph23(tx, pph23); err != nil {
		tx.Rollback()
		return nil, err
	}

	return pph23, nil
}

func (s *Pph23Service) GetPph23ByID(ctx context.Context, params *dtos.GetPph23Params) (*dtos.Pph23DetailDTO, error) {
	pph23, err := s.repo.GetPph23ByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return pph23, nil
}

func (s *Pph23Service) UpdatePph23(ctx context.Context, pph23 *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdatePph23(tx, pph23); err != nil {
		tx.Rollback()
		return nil, err
	}

	return pph23, nil
}

func (s *Pph23Service) DeletePph23(ctx context.Context, params *dtos.GetPph23Params, tx *gorm.DB) error {
	if err := s.repo.DeletePph23(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *Pph23Service) RestorePph23(ctx context.Context, params *dtos.GetPph23Params, tx *gorm.DB) error {
	if err := s.repo.RestorePph23(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *Pph23Service) ExcelGetPph23s(ctx context.Context, filters map[string]string) ([]byte, error) {
	pph23s, _, err := s.GetPph23s(ctx, filters)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "pph23s"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Pph23s", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, pph23 := range pph23s {
		row := []interface{}{
			pph23.ID,
			pph23.Name,
			pph23.Description,
			pph23.CreatedAt,
			pph23.UpdatedAt,
		}
		file.SetSheetRow("Pph23s", fmt.Sprintf("A%d", i+2), &row)
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
func (s *Pph23Service) CsvGetPph23s(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	pph23s, _, err := s.GetPph23s(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetPph23s error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, pph23 := range pph23s {
		csv += fmt.Sprintf("%d,%s,%f,%s,%s,%s,%s\n",
			pph23.ID,
			pph23.Name,
			pph23.Num,
			utils.GetPtrVal(pph23.Description),
			utils.GetPtrVal(pph23.Remark),
			utils.GetPtrVal(pph23.CreatedAt),
			utils.GetPtrVal(pph23.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
