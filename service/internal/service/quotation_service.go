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

func (s *QuotationService) CreateQuotation(ctx *fiber.Ctx, quotation *models.Quotation, tx *gorm.DB, span opentracing.Span) (*models.Quotation, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuotation", opentracing.ChildOf(span.Context()))

	if tx, err := s.repo.CreateQuotation(tx, quotation, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return quotation, tx, nil
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
func (s *QuotationService) CreateQuoDts(ctx *fiber.Ctx, boms []models.QuoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.QuoDt, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDts", opentracing.ChildOf(span.Context()))

	quoDtsModel := []models.QuoDt{}
	var err error

	tx, quoDtsModel, err = s.repo.CreateQuoDts(tx, boms, quotationID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, quoDtsModel, err
	}

	tx.SavePoint("create_quo_dts")

	return tx, quoDtsModel, nil
}

// bulk create/update boms for a quotation
func (s *QuotationService) BulkCreateUpdateQuoDts(ctx *fiber.Ctx, quoDts []models.QuoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-BulkCreateUpdateQuoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateQuoDts := []models.QuoDt{}
	// filter with ID to bulk update
	bulkUpdateQuoDts := []models.QuoDt{}
	// get all ids
	quoDtIDs := []uint{}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

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

	tx.SavePoint("bulk_create_update_quo_dts")

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
	// childSpan := opentracing.StartSpan("QuotationService-MapQuoDtsToQuotation", opentracing.ChildOf(span.Context()))

	quoDtsModel := []models.QuoDt{}

	for _, quoDt := range req.QuoDts {
		quoDtModel := models.QuoDt{
			ProductUuid: quoDt.ProductUuid,
			QuotationID: &createdQuotation.ID,
			ItemUnitID:  quoDt.ItemUnitID,
			VatID:       quoDt.VatID,
			RefID:       quoDt.RefID,
			ItemID:      quoDt.ItemID,
			RefType:     quoDt.RefType,
			ItemType:    quoDt.ItemType,
			// RefJSON: 	quoDt.RefJSON,
			Remark:       quoDt.Remark,
			VatPerc:      quoDt.VatPerc,
			VatPercAm:    quoDt.VatPercAm,
			QtySO:        quoDt.QtySO,
			Qty:          quoDt.Qty,
			PriceSell:    quoDt.PriceSell,
			PriceBuy:     quoDt.PriceBuy,
			SubtotalSell: quoDt.SubtotalSell,
			SubtotalBuy:  quoDt.SubtotalBuy,
			DiscAm:       quoDt.DiscAm,
			DiscPerc:     quoDt.DiscPerc,
			DiscPercNum:  quoDt.DiscPercNum,
			DiscPercAm:   quoDt.DiscPercAm,
			DiscFinal:    quoDt.DiscFinal,
			DiscType:     quoDt.DiscType,
			TotalAm:      quoDt.TotalAm,
			CreatedByID:  &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}

func (s *QuotationService) MapCreateQuoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateQuotationRequest, createdQuoDts []models.QuoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	// var quoDtBomsModel []models.QuoDtBom
	quoDtBomsModel := make([]map[string]interface{}, 0)

	for _, reqQuoDt := range req.QuoDts {
		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
			for _, createdQuoDt := range createdQuoDts {
				if reqQuoDtBom.ProductUuid == createdQuoDt.ProductUuid {
					genCode := "-"
					itemJson := "{}"
					quoDtBomsModel = append(quoDtBomsModel, map[string]interface{}{
						"id":            0,
						"quotation_id":  createdQuoDt.QuotationID,
						"product_uuid":  reqQuoDtBom.ProductUuid,
						"quo_dt_id":     createdQuoDt.ID,
						"product_id":    reqQuoDtBom.ProductID,
						"item_id":       reqQuoDtBom.ItemID,
						"item_unit_id":  reqQuoDtBom.ItemUnitID,
						"remark":        reqQuoDtBom.Remark,
						"qty":           reqQuoDtBom.Qty,
						"price_sell":    reqQuoDtBom.PriceSell,
						"price_buy":     reqQuoDtBom.PriceBuy,
						"subtotal_sell": reqQuoDtBom.SubtotalSell,
						"subtotal_buy":  reqQuoDtBom.SubtotalBuy,
						"gen_code":      &genCode,
						"item_json":     &itemJson,
						"created_by_id": userID,
					})
				}

			}
		}
	}

	return quoDtBomsModel
}

func (s *QuotationService) CreateQuoDtBoms(ctx *fiber.Ctx, quoDtBoms []map[string]interface{}, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	if tx, err := s.repo.CreateQuoDtBoms(tx, quoDtBoms, quotationID, childSpan); err != nil {
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
	// childSpan := opentracing.StartSpan("QuotationService-MapQuoDtsToQuotation", opentracing.ChildOf(span.Context()))

	quoDtsModel := []models.QuoDt{}

	// refJson := "{}"
	// itemJson := "{}"

	for _, reqQuoDt := range req.QuoDts {
		quoDtID := uint(0)
		if reqQuoDt.QuoDtID != nil {
			quoDtID = *reqQuoDt.QuoDtID
		}

		quoDtModel := models.QuoDt{
			ID:          quoDtID,
			ProductUuid: reqQuoDt.ProductUuid,
			QuotationID: &updatedQuotation.ID,
			ItemUnitID:  reqQuoDt.ItemUnitID,
			VatID:       reqQuoDt.VatID,
			RefID:       reqQuoDt.RefID,
			ItemID:      reqQuoDt.ItemID,
			RefType:     reqQuoDt.RefType,
			ItemType:    reqQuoDt.ItemType,
			// RefJSON: 	refJson,
			// ItemJSON: 	 itemJson,
			GenCode:      reqQuoDt.GenCode,
			Remark:       reqQuoDt.Remark,
			VatPerc:      reqQuoDt.VatPerc,
			VatPercAm:    reqQuoDt.VatPercAm,
			QtySO:        reqQuoDt.QtySO,
			Qty:          reqQuoDt.Qty,
			PriceSell:    reqQuoDt.PriceSell,
			PriceBuy:     reqQuoDt.PriceBuy,
			SubtotalSell: reqQuoDt.SubtotalSell,
			SubtotalBuy:  reqQuoDt.SubtotalBuy,
			DiscAm:       reqQuoDt.DiscAm,
			DiscPerc:     reqQuoDt.DiscPerc,
			DiscPercNum:  reqQuoDt.DiscPercNum,
			DiscPercAm:   reqQuoDt.DiscPercAm,
			DiscFinal:    reqQuoDt.DiscFinal,
			DiscType:     reqQuoDt.DiscType,
			TotalAm:      reqQuoDt.TotalAm,
			CreatedByID:  &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}
