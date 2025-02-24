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

type VatService struct {
	repo     *repository.VatRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewVatService(repo *repository.VatRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *VatService {
	return &VatService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *VatService) GetVats(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.VatListDTO, int, error) {
	childSpan := opentracing.StartSpan("VatService-GetVats", opentracing.ChildOf(span.Context()))

	vats, total, err := s.repo.GetVats(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return vats, total, nil
}

func (s *VatService) CreateVat(ctx *fiber.Ctx, vat *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("VatService-CreateVat", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateVat(tx, vat, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return vat, nil
}

func (s *VatService) GetVatByID(ctx *fiber.Ctx, params *dtos.GetVatParams, span opentracing.Span) (*dtos.VatDetailDTO, error) {
	childSpan := opentracing.StartSpan("VatService-GetVatByID", opentracing.ChildOf(span.Context()))

	vat, err := s.repo.GetVatByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return vat, nil
}

func (s *VatService) UpdateVat(ctx *fiber.Ctx, vat *models.MixValue, tx *gorm.DB, span opentracing.Span) (*models.MixValue, error) {
	childSpan := opentracing.StartSpan("VatService-UpdateVat", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateVat(tx, vat, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return vat, nil
}

func (s *VatService) DeleteVat(ctx *fiber.Ctx, params *dtos.GetVatParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatService-DeleteVat", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteVat(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *VatService) RestoreVat(ctx *fiber.Ctx, params *dtos.GetVatParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatService-RestoreVat", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreVat(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *VatService) RestoreVatHistory(ctx *fiber.Ctx, params *dtos.GetVatHistoryParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatService-RestoreVatHistory", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreVatHistory(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *VatService) ExcelGetVats(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("VatService-ExcelGetVats", opentracing.ChildOf(span.Context()))

	vats, _, err := s.GetVats(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
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
func (s *VatService) CsvGetVats(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("VatService-CsvGetVats", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	vats, _, err := s.GetVats(ctx, filters, childSpan)
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
		log.Println("CsvGetVats error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Vats\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Multiplier,Divider,Created At,Updated At\n"
	// Build CSV rows
	for _, vat := range vats {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s\n",
			vat.ID,
			vat.Name,
			utils.GetPtrVal(vat.Description),
			utils.GetPtrVal(vat.Remark),
			utils.GetPtrVal(vat.Multiplier),
			utils.GetPtrVal(vat.Divider),
			utils.GetPtrVal(vat.CreatedAt),
			utils.GetPtrVal(vat.UpdatedAt),
		)
	}

	return []byte(csv), nil
}

// github.com/xuri/excelize/v2
func (s *VatService) ExcelGetVatsHistory(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("VatService-ExcelGetVatsHistory", opentracing.ChildOf(span.Context()))

	vats, _, err := s.GetVatHistories(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "vats"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Vats", "A1", &[]string{"ID", "Name", "Percent", "Multiplier", "Divider", "Remark", "Created At", "Updated At"})

	for i, vat := range vats {
		row := []interface{}{
			vat.ID,
			vat.Name,
			vat.Num,
			vat.Multiplier,
			vat.Divider,
			vat.Remark,
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
func (s *VatService) CsvGetVatsHistory(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("VatService-CsvGetVatsHistory", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	vats, _, err := s.GetVatHistories(ctx, filters, childSpan)
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
		log.Println("CsvGetVats error:", err)
	} else {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Vat Histories\n"
	csv += "\n"

	csv += "ID,Name,Description,Remark,Multiplier,Divider,Changed At,Created At,Updated At\n"
	// Build CSV rows
	for _, vat := range vats {
		csv += fmt.Sprintf("%d,%s,%s,%f,%f,%s,%s,%s\n",
			vat.ID,
			vat.Name,
			utils.GetPtrVal(vat.Remark),
			vat.Multiplier,
			vat.Divider,
			utils.GetPtrVal(vat.ChangedAt),
			utils.GetPtrVal(vat.CreatedAt),
			utils.GetPtrVal(vat.UpdatedAt),
		)
	}

	return []byte(csv), nil
}

// create a new vat history
func (s *VatService) CreateVatHistory(ctx *fiber.Ctx, vatHistory *models.VatHistory, tx *gorm.DB, span opentracing.Span) (*models.VatHistory, error) {
	childSpan := opentracing.StartSpan("VatService-CreateVatHistory", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateVatHistory(tx, vatHistory, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return vatHistory, nil
}

// Get latest vat history by vat id
func (s *VatService) GetVatHistoryByID(ctx *fiber.Ctx, params *dtos.GetVatHistoryParams, span opentracing.Span) (*dtos.VatHistoryDetailDTO, error) {
	childSpan := opentracing.StartSpan("VatService-GetLatestVatHistoryByID", opentracing.ChildOf(span.Context()))

	vatHistory, err := s.repo.GetVatHistoryByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return vatHistory, nil
}

func (s *VatService) GetVatHistories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.VatHistoryListDTO, int, error) {
	childSpan := opentracing.StartSpan("VatService-GetVatHistories", opentracing.ChildOf(span.Context()))

	vats, total, err := s.repo.GetVatHistories(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return vats, total, nil
}

func (s *VatService) UpdateVatHistory(ctx *fiber.Ctx, vat *models.VatHistory, tx *gorm.DB, span opentracing.Span) (*models.VatHistory, error) {
	childSpan := opentracing.StartSpan("VatService-UpdateVatHistory", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateVatHistory(tx, vat, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return vat, nil
}

func (s *VatService) DeleteVatHistory(ctx *fiber.Ctx, params *dtos.GetVatHistoryParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("VatService-DeleteVatHistory", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteVatHistory(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}
