package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type QuotationService struct {
	repo     *repository.QuotationRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewQuotationService(repo *repository.QuotationRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *QuotationService {
	return &QuotationService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *QuotationService) GetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.QuotationListDTO, int, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuotations", opentracing.ChildOf(span.Context()))

	quotations, total, err := s.repo.GetQuotations(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return quotations, total, nil
}

func (s *QuotationService) GetWidgetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.QuotationStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetWidgetQuotations", opentracing.ChildOf(span.Context()))

	quotations, total, err := s.repo.GetWidgetQuotations(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return quotations, total, nil
}

func (s *QuotationService) CreateQuotation(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Quotation, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuotation", opentracing.ChildOf(span.Context()))

	customerQuoCreatedThisMonthNumber, err := s.repo.GetCustomerQuotationCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	quoGlobalCreatedThisMonthNumber, err := s.repo.GetGlobalQuotationCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)

	quotation, err := utils.MapCreateQuotation(ctx, req, userID, branchID, customerQuoCreatedThisMonthNumber, quoGlobalCreatedThisMonthNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, tx, err
	}

	if tx, err := s.repo.CreateQuotation(tx, &quotation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	// bulk create item quoDts ref ms items / product->boms
	var quoDts []models.QuoDt
	tx, quoDts, err = s.CreateQuoDts(ctx, req, userID, &quotation, tx, childSpan)

	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	// tx, err = s.CreateQuoDtBoms(ctx, quoDtBoms, tx, childSpan)
	tx, err = s.CreateQuoDtBoms(ctx, quoDts, req, &quotation, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return &quotation, tx, nil
}

func (s *QuotationService) GetQuotationByID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) (*dtos.QuotationDetailDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuotationByID", opentracing.ChildOf(span.Context()))

	quotation, err := s.repo.GetQuotationByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return quotation, nil
}

func (s *QuotationService) UpdateQuotation(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Quotation, error) {
	childSpan := opentracing.StartSpan("QuotationService-UpdateQuotation", opentracing.ChildOf(span.Context()))

	quotation, err := utils.MapUpdateQuotation(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	if err := s.repo.UpdateQuotation(tx, &quotation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Bulk/Create Update Batch QuoDts
	quoDts, err := s.MapCreateUpdateQuoDts(ctx, req, &quotation, userID, childSpan)

	tx, err = s.BulkCreateUpdateQuoDts(ctx, req, &quotation, userID, quoDts, quotation.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	updatedQuotationIDs := make([]uint, 0)
	updatedQuotationIDs = append(updatedQuotationIDs, quotation.ID)

	// get updated quoDts
	updatedQuoDts, err := s.GetUpdatedQuoDtsByQuotationIDs(ctx, tx, updatedQuotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Bulk/Create Update Batch QuoDtBoms
	err = s.BulkCreateUpdateQuoDtBoms(ctx, updatedQuoDts, req, quotation.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return &quotation, nil
}

func (s *QuotationService) DeleteQuotation(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-DeleteQuotation", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteQuotation(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *QuotationService) RestoreQuotation(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-RestoreQuotation", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreQuotation(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *QuotationService) ExcelGetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("QuotationService-ExcelGetQuotations", opentracing.ChildOf(span.Context()))

	quotations, _, err := s.GetQuotations(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "quotations"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Quotations", "A1", &[]string{"ID", "Branch", "Code", "Factory Code", "Name", "Sku", "Barcode", "Unit", "Specification", "Desc", "Remark", "Price Sell", "Price Buy"})

	for i, quotation := range quotations {
		row := []interface{}{
			quotation.ID,
			utils.GetPtrVal(quotation.Remark),
		}
		file.SetSheetRow("Quotations", fmt.Sprintf("A%d", i+2), &row)
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
func (s *QuotationService) CsvGetQuotations(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("QuotationService-CsvGetQuotations", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	quotations, _, err := s.GetQuotations(ctx, filters, childSpan)
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
	csv += "Master Quotation\n"
	csv += "\n"

	csv += "ID,Branch,Code,Factory Code,Name,Sku,Barcode,Unit,Specification,Desc,Remark,Price Sell,Price Buy\n"
	// Build CSV rows
	for _, quotation := range quotations {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			quotation.ID,
			utils.GetPtrVal(quotation.Remark),
		)
	}

	return []byte(csv), nil
}

// quotation *models.Quotation
func (s *QuotationService) CreateQuoDts(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, userID uint, createdQuotation *models.Quotation, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.QuoDt, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDts", opentracing.ChildOf(span.Context()))

	quoDts, err := s.MapCreateQuoDts(ctx, req, createdQuotation, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, quoDts, err
	}

	quoDtsModel := []models.QuoDt{}

	tx, quoDtsModel, err = s.repo.CreateQuoDts(tx, quoDts, createdQuotation.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, quoDtsModel, err
	}

	return tx, quoDtsModel, nil
}

// bulk create/update boms for a quotation
func (s *QuotationService) BulkCreateUpdateQuoDts(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, updatedQuotation *models.Quotation, userID uint, quoDts []models.QuoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-BulkCreateUpdateQuoDts", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch QuoDts
	quoDts, err := s.MapCreateUpdateQuoDts(ctx, req, updatedQuotation, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// filter without ID to bulk create
	bulkCreateQuoDts := []models.QuoDt{}
	// filter with ID to bulk update
	bulkUpdateQuoDts := []models.QuoDt{}
	// get all ids
	quoDtIDs := []uint{}

	for _, quoDt := range quoDts {
		if quoDt.ID == 0 {
			quoDt.CreatedByID = &userID
			quoDt.CreatedAt = time.Now()
			bulkCreateQuoDts = append(bulkCreateQuoDts, quoDt)
		} else {
			bulkUpdateQuoDts = append(bulkUpdateQuoDts, quoDt)
			quoDtIDs = append(quoDtIDs, quoDt.ID)
		}
	}

	// delete quoDts that are not in the list
	if len(quoDtIDs) > 0 {
		if tx, err := s.repo.DeleteQuoDtsWhereNotIn(ctx, tx, quotationID, quoDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateQuoDts) > 0 {
		if tx, _, err := s.repo.CreateQuoDts(tx, bulkCreateQuoDts, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdateQuoDts) > 0 {
		if tx, err := s.repo.UpdateQuoDts(tx, bulkUpdateQuoDts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *QuotationService) GetQuoDtsByQuotationIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuoDtsByQuotationIDs", opentracing.ChildOf(span.Context()))

	quoDts, err := s.repo.GetQuoDtsByQuotationIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return quoDts, nil
}

func (s *QuotationService) GetUpdatedQuoDtsByQuotationIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	quoDts, err := s.repo.GetUpdatedQuoDtsByQuotationIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return quoDts, nil
}

func (s *QuotationService) MapCreateQuoDts(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuotation *models.Quotation, userID uint, span opentracing.Span) ([]models.QuoDt, error) {
	childSpan := opentracing.StartSpan("QuotationService-MapCreateQuoDts", opentracing.ChildOf(span.Context()))

	quoDtsModel, err := utils.MapCreateQuoDts(ctx, req, createdQuotation, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return quoDtsModel, nil
}

func (s *QuotationService) MapCreateQuoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateQuotationRequest, createdQuoDts []models.QuoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	quoDtBomsModel := utils.MapCreateQuoDtBoms(ctx, req, createdQuoDts, userID, span)

	return quoDtBomsModel
}

func (s *QuotationService) CreateQuoDtBoms(ctx *fiber.Ctx, quoDts []models.QuoDt, req dtos.CreateQuotationRequest, createdQuotation *models.Quotation, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	// bulk create boms
	quoDtBoms := s.MapCreateQuoDtBoms(ctx, tx, req, quoDts, userID, childSpan)

	if tx, err := s.repo.CreateQuoDtBoms(tx, quoDtBoms, createdQuotation.ID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a quotation
// func (s *QuotationService) BulkCreateUpdateQuoDtBoms(ctx *fiber.Ctx, quoDts []dtos.QuotationQuoDtListDTO, req dtos.UpdateQuotationRequest, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
// func (s *QuotationService) BulkCreateUpdateQuoDtBoms(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
func (s *QuotationService) BulkCreateUpdateQuoDtBoms(ctx *fiber.Ctx, quoDts []dtos.QuotationQuoDtListUpdateDTO, req dtos.UpdateQuotationRequest, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-BulkCreateUpdateQuoDtBoms", opentracing.ChildOf(span.Context()))

	bulkCreateQuoDtBoms, bulkUpdateQuoDtBoms, quoDtBomIDs, err := utils.MapFilterUpdateQuoDtBomsToQuoDts(ctx, quoDts, req, quotationID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// delete quoDts that are not in the list
	if len(quoDtBomIDs) > 0 {
		if err := s.repo.DeleteQuoDtBomsWhereNotIn(ctx, tx, quotationID, quoDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateQuoDtBoms) > 0 {
		if tx, err := s.repo.CreateQuoDtBoms(tx, bulkCreateQuoDtBoms, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateQuoDtBoms) > 0 {
		if err := s.repo.UpdateQuoDtBoms(tx, bulkUpdateQuoDtBoms, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *QuotationService) DeleteQuoDtBomsByQuotationID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-DeleteQuoDtBomsByQuotationID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteQuoDtBomsByQuotationID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *QuotationService) DeleteQuoDtsByQuotationID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-DeleteQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteQuoDtsByQuotationID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *QuotationService) GetQuoDtsBomByQuotations(ctx *fiber.Ctx, filters map[string]string, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuoDtsBomByQuotations", opentracing.ChildOf(span.Context()))

	quoDtBoms, err := s.repo.GetQuoDtsBomByQuotations(ctx, filters, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return quoDtBoms, nil
}

func (s *QuotationService) MapFilterQuoDtBomsToQuoDts(ctx *fiber.Ctx, quoDtBoms []dtos.QuotationQuoDtBomListDTO, quoDts []dtos.QuotationQuoDtListDTO, span opentracing.Span) []dtos.QuotationQuoDtListDTO {
	return utils.MapFilterQuoDtBomsToQuoDts(quoDtBoms, quoDts)
}

func (s *QuotationService) MapCreateUpdateQuoDts(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, updatedQuotation *models.Quotation, userID uint, span opentracing.Span) ([]models.QuoDt, error) {
	childSpan := opentracing.StartSpan("QuotationService-MapCreateUpdateQuoDts", opentracing.ChildOf(span.Context()))

	quoDtsModel, err := utils.MapCreateUpdateQuoDts(ctx, req, updatedQuotation, userID, span)

	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return quoDtsModel, nil
}

// Lock all quotation table update
func (s *QuotationService) LockQuotationTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateQuotationRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-LockQuotationTable", opentracing.ChildOf(span.Context()))

	if req.ID > 0 {
		if err := s.repo.LockQuotationHeader(ctx, tx, req, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	quoDtIDs, quoDtBomIDs, productIDs, itemUnitIDs := utils.GetQuoIDs(req)

	if len(quoDtIDs) > 0 {
		if err := s.repo.LockQuoDts(ctx, tx, quoDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(quoDtBomIDs) > 0 {
		if err := s.repo.LockQuoDtBoms(ctx, tx, quoDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(productIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, productIDs, "products", childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(itemUnitIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, itemUnitIDs, "item_units", childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

// PdfGetQuotations
func (s *QuotationService) PdfGetQuotations(ctx *fiber.Ctx, req dtos.FormQuotationRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("QuotationService-PdfGetQuotations", opentracing.ChildOf(span.Context()))

	// quotation, err := s.repo.GetQuotationByID(ctx, &req.GetQuotationParams, tx, childSpan)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	return nil, err
	// }

	// if err := s.repo.UpdateQuotation(tx, quotation, childSpan); err != nil {
	// 	defer childSpan.Finish()
	// 	tx.Rollback()
	// 	return nil, err
	// }
	type InvoiceItem struct {
		Name     string
		Quantity int
		Price    float64
	}

	type InvoiceData struct {
		InvoiceNumber string
		Items         []InvoiceItem
	}

	data := InvoiceData{
		InvoiceNumber: "INV-2024-001",
		Items: []InvoiceItem{
			{Name: "Laptop", Quantity: 1, Price: 999.99},
			{Name: "Mouse", Quantity: 2, Price: 25.50},
		},
	}

	htmlFileName := "quotation-detail"

	// 2. Render HTML template with data
	// templateFile, err := templateFS.Open("templates/quotation-detail.html")
	templateFile, err := templateFS.Open(fmt.Sprintf("templates/%s.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to open embedded template: %v", err)
		return nil, err
	}

	// Read the template content
	templateContent, err := io.ReadAll(templateFile)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to read template content: %v", err)
		return nil, err
	}

	// htmlFile, err := os.CreateTemp("", "quotation-detail-*.html")
	htmlFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error writing htmlFile:", err)
		return nil, err
	}
	defer os.Remove(htmlFile.Name())

	// Parse the template
	tmpl, err := template.New(fmt.Sprintf("%s.html", htmlFileName)).Parse(string(templateContent))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to parse template: %v", err)
		return nil, err
	}

	if err := tmpl.Execute(htmlFile, data); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error Execute:", err)
		return nil, err
	}

	// 3. Generate PDF using wkhtmltopdf (Docker or local)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg:", err)
		return nil, err

	}

	// Read embedded templates
	headerContent, err := templateFS.ReadFile("templates/header.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	footerContent, err := templateFS.ReadFile("templates/footer.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}

	// Write to temp files
	headerPath, err := createTempFileFromEmbed(string(headerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create header temp file: %w", err)
	}
	defer os.Remove(headerPath) // Clean up

	footerPath, err := createTempFileFromEmbed(string(footerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create footer temp file: %w", err)
	}
	defer os.Remove(footerPath) // Clean up

	page := wkhtmltopdf.NewPage(htmlFile.Name())
	page.EnableLocalFileAccess.Set(true)
	page.HeaderHTML.Set("file://" + headerPath) // Set header
	page.FooterHTML.Set("file://" + footerPath) // Set footer
	// page.FooterSpacing.Set(10)                  // Space below content (mm)

	pdfg.AddPage(page)

	if err := pdfg.Create(); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg create:", err)
		return nil, err
	}

	uploadDir := "./public/generated_pdfs"

	// Ensure the directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	// 4. Save PDF to the "public" folder
	fileName := fmt.Sprintf("invoice-%s-%s.pdf", data.InvoiceNumber, time.Now().Format("20060102150405"))
	// pdfPath := filepath.Join("public", pdfName)
	pdfPath := filepath.Join(uploadDir, fileName)
	if err := pdfg.WriteFile(pdfPath); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdf path:", err)
		return nil, err
	}

	return &pdfPath, nil
}

func createTempFileFromEmbed(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "wkhtml-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
