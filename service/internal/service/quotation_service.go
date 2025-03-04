package service

import (
	"fmt"
	"time"

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

func (s *QuotationService) CreateQuotation(ctx *fiber.Ctx, quotation *models.Quotation, tx *gorm.DB, span opentracing.Span) (*models.Quotation, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuotation", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateQuotation(tx, quotation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return quotation, nil
}

func (s *QuotationService) GetQuotationByID(ctx *fiber.Ctx, params *dtos.GetQuotationParams, span opentracing.Span) (*dtos.QuotationDetailDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuotationByID", opentracing.ChildOf(span.Context()))

	quotation, err := s.repo.GetQuotationByID(ctx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return quotation, nil
}

func (s *QuotationService) UpdateQuotation(ctx *fiber.Ctx, quotation *models.Quotation, tx *gorm.DB, span opentracing.Span) (*models.Quotation, error) {
	childSpan := opentracing.StartSpan("QuotationService-UpdateQuotation", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateQuotation(tx, quotation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return quotation, nil
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
func (s *QuotationService) CreateQuoDts(ctx *fiber.Ctx, boms []*models.QuoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDts", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateQuoDts(tx, boms, quotationID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// bulk create/update boms for a quotation
func (s *QuotationService) BulkCreateUpdateQuoDts(ctx *fiber.Ctx, boms []*models.QuoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-BulkCreateQuoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateQuoDts := []*models.QuoDt{}
	// filter with ID to bulk update
	bulkUpdateQuoDts := []*models.QuoDt{}
	// get all ids
	bomIDs := []uint{}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, bom := range boms {
		if bom.ID == nil {
			bom.CreatedByID = &userID
			bom.CreatedAt = time.Now()
			bulkCreateQuoDts = append(bulkCreateQuoDts, bom)
		} else {
			bulkUpdateQuoDts = append(bulkUpdateQuoDts, bom)
			bomIDs = append(bomIDs, *bom.ID)
		}
	}

	// delete boms that are not in the list
	if len(bomIDs) > 0 {
		if err := s.repo.DeleteQuoDtsWhereNotIn(tx, quotationID, bomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateQuoDts) > 0 {
		if err := s.repo.CreateQuoDts(tx, bulkCreateQuoDts, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateQuoDts) > 0 {
		if err := s.repo.UpdateQuoDts(tx, bulkUpdateQuoDts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *QuotationService) GetQuoDtsByQuotationID(ctx *fiber.Ctx, quotationID uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	boms, err := s.repo.GetQuoDtsByQuotationID(ctx, quotationID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return boms, nil
}

func (s *QuotationService) MapCreateQuoDts(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuotation *models.Quotation, userID uint, span opentracing.Span) ([]*models.QuoDt, error) {
	// childSpan := opentracing.StartSpan("QuotationService-MapQuoDtsToQuotation", opentracing.ChildOf(span.Context()))

	quoDts := []*models.QuoDt{}
	for _, quoDt := range req.QuoDts {
		if *quoDt.RefType == "item" {
			quoDt := &models.QuoDt{
				QuotationID:       &createdQuotation.ID,
				RefID:             quoDt.RefID,
				ItemUnitID:        quoDt.ItemUnitID,
				VatID:             quoDt.VatID,
				ProductID:         quoDt.ProductID,
				ProductItemUnitID: quoDt.ProductItemUnitID,
				// QuoDtRefID: quoDt.QuoDtRefID,
				RefType: quoDt.RefType,
				// RefJSON: 	quoDt.RefJSON,
				Remark:      quoDt.Remark,
				VatPerc:     quoDt.VatPerc,
				QtySO:       quoDt.QtySO,
				Qty:         quoDt.Qty,
				PriceSell:   quoDt.PriceSell,
				Subtotal:    quoDt.Subtotal,
				DiscAm:      quoDt.DiscAm,
				DiscPerc:    quoDt.DiscPerc,
				TotalAm:     quoDt.TotalAm,
				PQtySO:      quoDt.PQtySO,
				PQty:        quoDt.PQty,
				PPriceSell:  quoDt.PPriceSell,
				PSubtotal:   quoDt.PSubtotal,
				PDiscAm:     quoDt.PDiscAm,
				PDiscPerc:   quoDt.PDiscPerc,
				PTotalAm:    quoDt.PTotalAm,
				CreatedByID: &userID,
			}
			quoDts = append(quoDts, quoDt)
		}

		if *quoDt.RefType == "bom" {

			for _, bom := range quoDt.Boms {
				quoDt := &models.QuoDt{
					QuotationID:       &createdQuotation.ID,
					RefID:             bom.RefID,
					ItemUnitID:        bom.ItemUnitID,
					VatID:             quoDt.VatID,
					ProductID:         quoDt.ProductID,
					ProductItemUnitID: bom.ProductItemUnitID,
					// QuoDtRefID: quoDt.QuoDtRefID,
					RefType: bom.RefType,
					// RefJSON: 	quoDt.RefJSON,
					Remark:      quoDt.Remark,
					VatPerc:     quoDt.VatPerc,
					QtySO:       quoDt.QtySO,
					Qty:         quoDt.Qty,
					PriceSell:   quoDt.PriceSell,
					Subtotal:    quoDt.Subtotal,
					DiscAm:      quoDt.DiscAm,
					DiscPerc:    quoDt.DiscPerc,
					TotalAm:     quoDt.TotalAm,
					PQtySO:      quoDt.PQtySO,
					PQty:        quoDt.PQty,
					PPriceSell:  quoDt.PPriceSell,
					PSubtotal:   quoDt.PSubtotal,
					PDiscAm:     quoDt.PDiscAm,
					PDiscPerc:   quoDt.PDiscPerc,
					PTotalAm:    quoDt.PTotalAm,
					CreatedByID: &userID,
				}
				quoDts = append(quoDts, quoDt)
			}
		}
	}

	return quoDts, nil
}
