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

type Pph23Service struct {
	repo     *repository.Pph23Repository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPph23Service(repo *repository.Pph23Repository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *Pph23Service {
	return &Pph23Service{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *Pph23Service) GetPph23s(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.Pph23ListDTO, int, error) {
	childSpan := opentracing.StartSpan("Pph23Service-GetPph23s", opentracing.ChildOf(span.Context()))

	pph23s, total, err := s.repo.GetPph23s(ctx.Context(), filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return pph23s, total, nil
}

func (s *Pph23Service) CreatePph23(ctx *fiber.Ctx, pph23 *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("Pph23Service-CreatePph23", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreatePph23(tx, pph23, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return pph23, nil
}

func (s *Pph23Service) GetPph23ByID(ctx *fiber.Ctx, params *dtos.GetPph23Params, span opentracing.Span) (*dtos.Pph23DetailDTO, error) {
	childSpan := opentracing.StartSpan("Pph23Service-GetPph23ByID", opentracing.ChildOf(span.Context()))

	pph23, err := s.repo.GetPph23ByID(ctx.Context(), params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return pph23, nil
}

func (s *Pph23Service) UpdatePph23(ctx *fiber.Ctx, pph23 *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("Pph23Service-UpdatePph23", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdatePph23(tx, pph23, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return pph23, nil
}

func (s *Pph23Service) DeletePph23(ctx *fiber.Ctx, params *dtos.GetPph23Params, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("Pph23Service-DeletePph23", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePph23(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *Pph23Service) RestorePph23(ctx *fiber.Ctx, params *dtos.GetPph23Params, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("Pph23Service-RestorePph23", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestorePph23(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *Pph23Service) ExcelGetPph23s(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("Pph23Service-ExcelGetPph23s", opentracing.ChildOf(span.Context()))

	pph23s, _, err := s.GetPph23s(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
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
func (s *Pph23Service) CsvGetPph23s(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("Pph23Service-CsvGetPph23s", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	pph23s, _, err := s.GetPph23s(ctx, filters, childSpan)
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
		log.Println("CsvGetPph23s error:", err)
	} else {
		appName = companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Pph23s\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Created At,Updated At\n"
	// Build CSV rows
	for _, pph23 := range pph23s {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			pph23.ID,
			pph23.Name,
			utils.GetPtrVal(pph23.Description),
			utils.GetPtrVal(pph23.Remark),
			utils.GetPtrVal(pph23.CreatedAt),
			utils.GetPtrVal(pph23.UpdatedAt),
		)
	}

	return []byte(csv), nil
}
