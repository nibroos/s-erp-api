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

type CurrencyService struct {
	repo     *repository.CurrencyRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewCurrencyService(repo *repository.CurrencyRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *CurrencyService {
	return &CurrencyService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *CurrencyService) GetCurrencies(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.CurrencyListDTO, int, error) {
	childSpan := opentracing.StartSpan("CurrencyService-GetCurrencies", opentracing.ChildOf(span.Context()))

	currencies, total, err := s.repo.GetCurrencies(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return currencies, total, nil
}

func (s *CurrencyService) CreateCurrency(ctx *fiber.Ctx, currency *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CurrencyService-CreateCurrency", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateCurrency(tx, currency, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return currency, nil
}

func (s *CurrencyService) GetCurrencyByID(ctx *fiber.Ctx, params *dtos.GetCurrencyParams, span opentracing.Span) (*dtos.CurrencyDetailDTO, error) {
	childSpan := opentracing.StartSpan("CurrencyService-GetCurrencyByID", opentracing.ChildOf(span.Context()))

	currency, err := s.repo.GetCurrencyByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return currency, nil
}

func (s *CurrencyService) UpdateCurrency(ctx *fiber.Ctx, currency *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("CurrencyService-UpdateCurrency", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateCurrency(tx, currency, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return currency, nil
}

func (s *CurrencyService) DeleteCurrency(ctx *fiber.Ctx, params *dtos.GetCurrencyParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CurrencyService-DeleteCurrency", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteCurrency(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *CurrencyService) RestoreCurrency(ctx *fiber.Ctx, params *dtos.GetCurrencyParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("CurrencyService-RestoreCurrency", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreCurrency(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *CurrencyService) ExcelGetCurrencies(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CurrencyService-ExcelGetCurrencies", opentracing.ChildOf(span.Context()))

	currencies, _, err := s.GetCurrencies(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
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
func (s *CurrencyService) CsvGetCurrencies(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("CurrencyService-CsvGetCurrencies", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	currencies, _, err := s.GetCurrencies(ctx, filters, childSpan)
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
		log.Println("CsvGetCurrencies error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Currencies\n"
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
