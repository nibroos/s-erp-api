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

type CurrencyService struct {
	repo     *repository.CurrencyRepository
	utilRepo *repository.UtilRepository
}

func NewCurrencyService(repo *repository.CurrencyRepository, utilRepo *repository.UtilRepository) *CurrencyService {
	return &CurrencyService{
		repo:     repo,
		utilRepo: utilRepo,
	}
}

func (s *CurrencyService) GetCurrencies(ctx context.Context, filters map[string]string) ([]dtos.CurrencyListDTO, int, error) {
	currencies, total, err := s.repo.GetCurrencies(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	return currencies, total, nil
}

func (s *CurrencyService) CreateCurrency(ctx context.Context, currency *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.CreateCurrency(tx, currency); err != nil {
		tx.Rollback()
		return nil, err
	}

	return currency, nil
}

func (s *CurrencyService) GetCurrencyByID(ctx context.Context, params *dtos.GetCurrencyParams) (*dtos.CurrencyDetailDTO, error) {
	currency, err := s.repo.GetCurrencyByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return currency, nil
}

func (s *CurrencyService) UpdateCurrency(ctx context.Context, currency *models.MixValue, tx *gorm.DB) (*models.MixValue, error) {
	if err := s.repo.UpdateCurrency(tx, currency); err != nil {
		tx.Rollback()
		return nil, err
	}

	return currency, nil
}

func (s *CurrencyService) DeleteCurrency(ctx context.Context, params *dtos.GetCurrencyParams, tx *gorm.DB) error {
	if err := s.repo.DeleteCurrency(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CurrencyService) RestoreCurrency(ctx context.Context, params *dtos.GetCurrencyParams, tx *gorm.DB) error {
	if err := s.repo.RestoreCurrency(tx, params); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CurrencyService) ExcelGetCurrencies(ctx context.Context, filters map[string]string) ([]byte, error) {
	currencies, _, err := s.GetCurrencies(ctx, filters)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "currencies"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Currencies", "A1", &[]string{"ID", "Name", "Description", "Created At", "Updated At"})

	for i, currency := range currencies {
		row := []interface{}{
			currency.ID,
			currency.Name,
			currency.Description,
			currency.CreatedAt,
			currency.UpdatedAt,
		}
		file.SetSheetRow("Currencies", fmt.Sprintf("A%d", i+2), &row)
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
func (s *CurrencyService) CsvGetCurrencies(ctx context.Context, filters map[string]string) ([]byte, error) {
	// filters is_csv
	filters["is_csv"] = "1"
	currencies, _, err := s.GetCurrencies(ctx, filters)
	if err != nil {
		return nil, err
	}

	// get company profile
	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err != nil {
		log.Println("CsvGetCurrencies error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Item Groups\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, currency := range currencies {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			currency.ID,
			currency.Name,
			utils.GetPtrVal(currency.Description),
			utils.GetPtrVal(currency.Remark),
			utils.GetPtrVal(currency.CreatedAt),
			utils.GetPtrVal(currency.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
