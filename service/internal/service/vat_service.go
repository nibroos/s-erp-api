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

type VatService struct {
	repo     *repository.VatRepository
	utilRepo *repository.UtilRepository
}

func NewVatService(repo *repository.VatRepository, utilRepo *repository.UtilRepository) *VatService {
	return &VatService{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *VatService) GetVats(ctx context.Context, filters map[string]string) ([]dtos.VatListDTO, int, error) {
	vats, total, err := s.repo.GetVats(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return vats, total, nil
}

func (s *VatService) CreateVat(ctx context.Context, vat *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateVat(tx, vat); err != nil {
		tx.Rollback()
		return nil, err
	}

	return vat, nil
}

func (s *VatService) GetVatByID(ctx context.Context, params *dtos.GetVatParams) (*dtos.VatDetailDTO, error) {
	vat, err := s.repo.GetVatByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return vat, nil
}

func (s *VatService) UpdateVat(ctx context.Context, vat *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateVat(tx, vat); err != nil {
		tx.Rollback()
		return nil, err
	}

	return vat, nil
}

func (s *VatService) DeleteVat(ctx context.Context, params *dtos.GetVatParams, tx *gorm.DB) error {
	if err := s.repo.DeleteVat(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *VatService) RestoreVat(ctx context.Context, params *dtos.GetVatParams, tx *gorm.DB) error {
	if err := s.repo.RestoreVat(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *VatService) ExcelGetVats(ctx context.Context, filters map[string]string) ([]byte, error) {
	vats, _, err := s.GetVats(ctx, filters)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "vats"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Vats", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, vat := range vats {
		row := []interface{}{
			vat.ID,
			vat.Name,
			vat.Description,
			vat.CreatedAt,
			vat.UpdatedAt,
		}
		file.SetSheetRow("Vats", fmt.Sprintf("A%d", i+2), &row)
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
func (s *VatService) CsvGetVats(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	vats, _, err := s.GetVats(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetVats error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, vat := range vats {
		csv += fmt.Sprintf("%d,%s,%f,%s,%s,%s,%s\n",
			vat.ID,
			vat.Name,
			vat.Num,
			utils.GetPtrVal(vat.Description),
			utils.GetPtrVal(vat.Remark),
			utils.GetPtrVal(vat.CreatedAt),
			utils.GetPtrVal(vat.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
