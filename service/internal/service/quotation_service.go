package service

import (
	"fmt"
	"log"
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
		if tx, err := s.repo.DeleteQuoDtsWhereNotIn(tx, quotationID, quoDtIDs, childSpan); err != nil {
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

func (s *QuotationService) GetQuoDtsByQuotationID(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.QuotationQuoDtListDTO, error) {
	childSpan := opentracing.StartSpan("QuotationService-GetQuoDtsByQuotationID", opentracing.ChildOf(span.Context()))

	boms, err := s.repo.GetQuoDtsByQuotationIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return boms, nil
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

// func (s *QuotationService) MapCreateQuoDtBoms(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuoDts []models.QuoDt, userID uint, span opentracing.Span) ([]models.QuoDtBom, error) {
// 	// childSpan := opentracing.StartSpan("QuotationService-MapQuoDtsToQuotation", opentracing.ChildOf(span.Context()))

// 	quoDtBomsModel := []models.QuoDtBom{}

// 	log.Println("MapCreateQuoDtBoms-createdQuoDts", createdQuoDts)

// 	for _, reqQuoDt := range req.QuoDts {
// 		log.Println("reqQuoDt", reqQuoDt)
// 		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
// 			log.Println("reqQuoDt.QuoDtsBoms", reqQuoDtBom)
// 			for _, createdQuoDt := range createdQuoDts {
// 				log.Println("reqQuoDtBom.ProductID", reqQuoDtBom.ProductID)
// 				log.Println("createdQuoDt.ItemID", createdQuoDt.ItemID)
// 				log.Println("createdQuoDt", createdQuoDt)
// 				if reqQuoDtBom.ProductUuid != nil && createdQuoDt.ProductUuid != nil {
// 					log.Println("reqQuoDtBom.ProductUuid-createdQuoDt.ProductUuid", *reqQuoDtBom.ProductUuid, *createdQuoDt.ProductUuid)
// 					if *reqQuoDtBom.ProductUuid == *createdQuoDt.ProductUuid {
// 						log.Println("reqQuoDtBom.ProductUuid == createdQuoDt.ProductUuid")
// 						quoDtModel := models.QuoDtBom{
// 							QuotationID:  *createdQuoDt.QuotationID,
// 							ProductUuid:  reqQuoDtBom.ProductUuid,
// 							QuoDtID:      createdQuoDt.ID,
// 							ProductID:    reqQuoDtBom.ProductID,
// 							ItemID:       reqQuoDtBom.ProductItemID,
// 							ItemUnitID:   reqQuoDtBom.ItemUnitID,
// 							Remark:       reqQuoDtBom.Remark,
// 							Qty:          reqQuoDtBom.Qty,
// 							PriceSell:    reqQuoDtBom.PriceSell,
// 							PriceBuy:     reqQuoDtBom.PriceBuy,
// 							SubtotalSell: reqQuoDtBom.SubtotalSell,
// 							SubtotalBuy:  reqQuoDtBom.SubtotalBuy,
// 							CreatedByID:  &userID,
// 						}
// 						quoDtBomsModel = append(quoDtBomsModel, quoDtModel)
// 					}
// 				}
// 			}
// 		}
// 	}

// 	log.Println("MapCreateQuoDtBoms-quoDtBomsModel", quoDtBomsModel)

//		return quoDtBomsModel, nil
//	}
func (s *QuotationService) MapCreateQuoDtBoms(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuoDts []models.QuoDt, userID uint, span opentracing.Span) ([]models.QuoDtBom, error) {
	quoDtBomsModel := []models.QuoDtBom{}

	log.Println("MapCreateQuoDtBoms-createdQuoDts", createdQuoDts)

	for _, reqQuoDt := range req.QuoDts {
		log.Println("reqQuoDt", reqQuoDt)
		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
			log.Println("reqQuoDt.QuoDtsBoms", reqQuoDtBom)
			for _, createdQuoDt := range createdQuoDts {
				log.Println("reqQuoDtBom.ProductID", reqQuoDtBom.ProductID)
				log.Println("createdQuoDt.ItemID", createdQuoDt.ItemID)
				log.Println("createdQuoDt", createdQuoDt)
				log.Println("createdQuoDt.ID", createdQuoDt.ID)

				// Ensure that ProductUuid is not nil before dereferencing
				if reqQuoDtBom.ProductUuid != nil && createdQuoDt.ProductUuid != nil {
					log.Println("reqQuoDtBom.ProductUuid-createdQuoDt.ProductUuid", *reqQuoDtBom.ProductUuid, *createdQuoDt.ProductUuid)
					if *reqQuoDtBom.ProductUuid == *createdQuoDt.ProductUuid {
						log.Println("reqQuoDtBom.ProductUuid == createdQuoDt.ProductUuid")
						quoDtModel := models.QuoDtBom{
							ItemJSON:     "{}",
							QuotationID:  createdQuoDt.QuotationID,
							ProductUuid:  reqQuoDtBom.ProductUuid,
							QuoDtID:      createdQuoDt.ID,
							ProductID:    reqQuoDtBom.ProductID,
							ItemID:       reqQuoDtBom.ProductItemID,
							ItemUnitID:   reqQuoDtBom.ItemUnitID,
							Remark:       reqQuoDtBom.Remark,
							Qty:          reqQuoDtBom.Qty,
							PriceSell:    reqQuoDtBom.PriceSell,
							PriceBuy:     reqQuoDtBom.PriceBuy,
							SubtotalSell: reqQuoDtBom.SubtotalSell,
							SubtotalBuy:  reqQuoDtBom.SubtotalBuy,
							CreatedByID:  &userID,
						}
						quoDtBomsModel = append(quoDtBomsModel, quoDtModel)
						log.Println("abc", quoDtBomsModel)
					}
				}
			}
		}
	}

	log.Println("MapCreateQuoDtBoms-quoDtBomsModel", quoDtBomsModel)

	return quoDtBomsModel, nil
}

func (s *QuotationService) CreateQuoDtBoms(ctx *fiber.Ctx, quoDtBoms []models.QuoDtBom, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("QuotationService-CreateQuoDtBoms", opentracing.ChildOf(span.Context()))

	if err := s.repo.CreateQuoDtBoms(tx, quoDtBoms, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *QuotationService) BulkCreateUpdateQuoDtBoms(ctx *fiber.Ctx, quoDts []dtos.QuotationQuoDtListDTO, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("QuotationService-BulkCreateUpdateQuoDtBoms", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateQuoDtBoms := []models.QuoDtBom{}
	// filter with ID to bulk update
	bulkUpdateQuoDtBoms := []models.QuoDtBom{}
	// get all ids
	quoDtBomIDs := []uint{}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, quoDt := range quoDts {
		for _, quoDtBom := range quoDt.QuoDtsBoms {
			newQuoDtBom := models.QuoDtBom{
				QuoDtID:      *quoDt.ID,
				ProductID:    quoDtBom.ProductID,
				ItemID:       quoDtBom.ItemID,
				ItemUnitID:   quoDtBom.ItemUnitID,
				Remark:       quoDtBom.Remark,
				Qty:          quoDtBom.Qty,
				PriceSell:    quoDtBom.PriceSell,
				PriceBuy:     quoDtBom.PriceBuy,
				SubtotalSell: quoDtBom.SubtotalSell,
				SubtotalBuy:  quoDtBom.SubtotalBuy,
			}

			if quoDtBom.ID == nil {
				newQuoDtBom.CreatedByID = &userID
				newQuoDtBom.CreatedAt = time.Now()
				bulkCreateQuoDtBoms = append(bulkCreateQuoDtBoms, newQuoDtBom)
			} else {
				newQuoDtBom.UpdatedByID = &userID
				newQuoDtBom.UpdatedAt = time.Now()
				bulkUpdateQuoDtBoms = append(bulkUpdateQuoDtBoms, newQuoDtBom)
				quoDtBomIDs = append(quoDtBomIDs, *quoDtBom.ID)
			}
		}
	}

	// delete quoDts that are not in the list
	if len(quoDtBomIDs) > 0 {
		if err := s.repo.DeleteQuoDtBomsWhereNotIn(tx, quotationID, quoDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateQuoDtBoms) > 0 {
		if err := s.repo.CreateQuoDtBoms(tx, bulkCreateQuoDtBoms, childSpan); err != nil {
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
