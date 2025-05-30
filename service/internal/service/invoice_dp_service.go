package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"text/template"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
)

type InvoiceDpService struct {
	repo     *repository.InvoiceDpRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewInvoiceDpService(repo *repository.InvoiceDpRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *InvoiceDpService {
	return &InvoiceDpService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *InvoiceDpService) GetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceDpListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-GetInvoiceDps", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDps, total, err := s.repo.GetInvoiceDps(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return invoiceDps, total, nil
}

func (s *InvoiceDpService) CreateInvoiceDp(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceDp, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-CreateInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpCreatedThisMonthNumber, err := s.repo.GetInvoiceDpCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := invoiceDpCreatedThisMonthNumber + 1

	invoiceDp, err := utils.MapCreateInvoiceDp(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateInvoiceDp(tx, &invoiceDp, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, createdInvoiceDpDts, err := s.CreateInvoiceDpDts(ctx, req, userID, &invoiceDp, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.UpdateSoDtsTotalDp(tx, createdInvoiceDpDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	return &invoiceDp, tx, nil
}

func (s *InvoiceDpService) GetInvoiceDpByID(ctx *fiber.Ctx, params *dtos.GetInvoiceDpParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceDpDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-GetInvoiceDpByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDp, err := s.repo.GetInvoiceDpByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	invoiceDpDts, err := s.repo.GetInvoiceDpDts(ctx, invoiceDp.ID, childSpan)
	if err != nil {
		return nil, err
	}

	var soDtIDs []uint
	for _, dt := range invoiceDpDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
			soDtIDs = append(soDtIDs, *dt.RefDtID)
		}
	}

	if len(soDtIDs) > 0 {
		soDtBoms, err := s.repo.GetSoDtBoms(ctx, soDtIDs, childSpan)
		if err != nil {
			return nil, err
		}

		for i, dt := range invoiceDpDts {
			if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
				var dtBoms []dtos.SalesOrderSoDtBomListDTO
				for _, bom := range soDtBoms {
					if bom.SoDtID != nil && *bom.SoDtID == *dt.RefDtID {
						dtBoms = append(dtBoms, bom)
					}
				}
				invoiceDpDts[i].SoDtsBoms = dtBoms
			}
		}
	}

	invoiceDp.InvoiceDpDts = invoiceDpDts

	if invoiceDp.CompanyProfileID != nil {
		companyParams := &dtos.GetCompanyProfileParams{ID: uint(*invoiceDp.CompanyProfileID)}
		company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch company: %v", err)
		} else {
			invoiceDp.Company = *company
		}
	}

	return invoiceDp, nil
}

// func (s *InvoiceDpService) UpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceDp, error) {
// 	childSpan := opentracing.StartSpan("InvoiceDpService-UpdateInvoiceDp", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	params := dtos.GetInvoiceDpParams{ID: req.ID}
// 	existingInvoiceDp, err := s.GetInvoiceDpByID(ctx, &params, tx, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	invoiceDp, err := utils.MapUpdateInvoiceDp(ctx, req, userID, branchID, existingInvoiceDp.RevNo, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	tx, err = s.repo.UpdateInvoiceDp(tx, &invoiceDp, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	invoiceDpDts, err := utils.MapUpdateInvoiceDpDts(ctx, req, &invoiceDp, userID, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	var createInvoiceDpDts []models.InvoiceDpDt
// 	var updateInvoiceDpDts []models.InvoiceDpDt
// 	var deleteInvoiceDpDtIDs []uint

// 	existingDtMap := make(map[uint]bool)
// 	for _, dt := range existingInvoiceDp.InvoiceDpDts {
// 		if dt.ID != nil {
// 			existingDtMap[*dt.ID] = true
// 		}
// 	}

// 	for _, dt := range invoiceDpDts {
// 		if dt.ID == 0 {
// 			dt.CreatedByID = &userID
// 			dt.CreatedAt = time.Now()
// 			createInvoiceDpDts = append(createInvoiceDpDts, dt)
// 		} else {
// 			dt.UpdatedByID = &userID
// 			updateInvoiceDpDts = append(updateInvoiceDpDts, dt)
// 			delete(existingDtMap, dt.ID)
// 		}
// 	}

// 	for dtID := range existingDtMap {
// 		deleteInvoiceDpDtIDs = append(deleteInvoiceDpDtIDs, dtID)
// 	}

// 	if len(deleteInvoiceDpDtIDs) > 0 {
// 		tx, err = s.repo.ResetSoDtsTotalDp(tx, deleteInvoiceDpDtIDs, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}

// 		tx, err = s.repo.DeleteInvoiceDpDtsByIDs(tx, deleteInvoiceDpDtIDs, userID, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(createInvoiceDpDts) > 0 {
// 		tx, err = s.repo.BulkCreateInvoiceDpDts(tx, createInvoiceDpDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}

// 		tx, err = s.repo.UpdateSoDtsTotalDp(tx, createInvoiceDpDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(updateInvoiceDpDts) > 0 {
// 		tx, err = s.repo.BulkUpdateInvoiceDpDts(tx, updateInvoiceDpDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}

// 		tx, err = s.repo.UpdateSoDtsTotalDp(tx, updateInvoiceDpDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	return &invoiceDp, nil
// }

func (s *InvoiceDpService) UpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceDp, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-UpdateInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetInvoiceDpParams{ID: req.ID}
	existingInvoiceDp, err := s.GetInvoiceDpByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	isChangingToCancelled := req.Status != nil && *req.Status == "CANCELLED" &&
		(existingInvoiceDp.Status == nil || *existingInvoiceDp.Status != "CANCELLED")

	invoiceDp, err := utils.MapUpdateInvoiceDp(ctx, req, userID, branchID, existingInvoiceDp.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if isChangingToCancelled {
		tx, err = s.repo.ResetSoDtsTotalDpForCancelled(tx, invoiceDp.ID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx, err = s.repo.UpdateInvoiceDp(tx, &invoiceDp, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if isChangingToCancelled {
		return &invoiceDp, nil
	}

	invoiceDpDts, err := utils.MapUpdateInvoiceDpDts(ctx, req, &invoiceDp, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var createInvoiceDpDts []models.InvoiceDpDt
	var updateInvoiceDpDts []models.InvoiceDpDt
	var deleteInvoiceDpDtIDs []uint

	existingDtMap := make(map[uint]bool)
	for _, dt := range existingInvoiceDp.InvoiceDpDts {
		if dt.ID != nil {
			existingDtMap[*dt.ID] = true
		}
	}

	for _, dt := range invoiceDpDts {
		if dt.ID == 0 {
			dt.CreatedByID = &userID
			dt.CreatedAt = time.Now()
			createInvoiceDpDts = append(createInvoiceDpDts, dt)
		} else {
			dt.UpdatedByID = &userID
			updateInvoiceDpDts = append(updateInvoiceDpDts, dt)
			delete(existingDtMap, dt.ID)
		}
	}

	for dtID := range existingDtMap {
		deleteInvoiceDpDtIDs = append(deleteInvoiceDpDtIDs, dtID)
	}

	if len(deleteInvoiceDpDtIDs) > 0 {
		tx, err = s.repo.ResetSoDtsTotalDp(tx, deleteInvoiceDpDtIDs, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		tx, err = s.repo.DeleteInvoiceDpDtsByIDs(tx, deleteInvoiceDpDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createInvoiceDpDts) > 0 {
		tx, err = s.repo.BulkCreateInvoiceDpDts(tx, createInvoiceDpDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		tx, err = s.repo.UpdateSoDtsTotalDp(tx, createInvoiceDpDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(updateInvoiceDpDts) > 0 {
		tx, err = s.repo.BulkUpdateInvoiceDpDts(tx, updateInvoiceDpDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		tx, err = s.repo.UpdateSoDtsTotalDp(tx, updateInvoiceDpDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &invoiceDp, nil
}

func (s *InvoiceDpService) DeleteInvoiceDp(ctx *fiber.Ctx, invoiceDpID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceDpService-DeleteInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	var invoiceDpDtIDs []uint
	if err := tx.Model(&models.InvoiceDpDt{}).
		Where("invoice_dp_id = ?", invoiceDpID).
		Pluck("id", &invoiceDpDtIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(invoiceDpDtIDs) > 0 {
		tx, err := s.repo.ResetSoDtsTotalDp(tx, invoiceDpDtIDs, childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx, err := s.repo.DeleteInvoiceDp(tx, invoiceDpID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceDpService) RestoreInvoiceDp(ctx *fiber.Ctx, params *dtos.GetInvoiceDpParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceDpService-RestoreInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := s.repo.RestoreInvoiceDp(ctx, params, tx, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceDpService) CreateInvoiceDpDts(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, userID uint, createdInvoiceDp *models.InvoiceDp, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.InvoiceDpDt, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-CreateInvoiceDpDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDpDts, updateSalesOrderIDs, err := utils.MapCreateInvoiceDpDts(ctx, req, createdInvoiceDp, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// update status header quotations
	if err := s.repo.UpdateSoStatus(ctx, tx, updateSalesOrderIDs, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, nil, err
	}

	tx, createdInvoiceDpDts, err := s.repo.CreateInvoiceDpDts(tx, invoiceDpDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	return tx, createdInvoiceDpDts, nil
}

func (s *InvoiceDpService) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDts, total, err := s.repo.GetRefSalesOrderDts(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	soDtIDs := make([]uint, 0)
	for _, soDt := range soDts {
		if soDt.ID != nil {
			soDtIDs = append(soDtIDs, *soDt.ID)
		}
	}

	if len(soDtIDs) > 0 {
		soDtBoms, err := s.repo.GetSoDtBoms(ctx, soDtIDs, childSpan)
		if err != nil {
			return nil, 0, err
		}

		if len(soDtBoms) > 0 {
			soDts = utils.MapRefSoDtBomsToSoDts(soDtBoms, soDts)
		}
	}

	return soDts, total, nil
}

func (s *InvoiceDpService) LockInvoiceDpTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateInvoiceDpRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-LockInvoiceDpTable", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if req.ID > 0 {
		var err error
		idPtr := &req.ID
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, []*uint{idPtr}, "invoice_dps", childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	invoiceDpDtIDs, productIDs, itemUnitIDs := utils.GetInvoiceDpIDs(req)

	invoiceDpDtIDsUint := make([]uint, 0)
	for _, id := range invoiceDpDtIDs {
		if id != nil {
			invoiceDpDtIDsUint = append(invoiceDpDtIDsUint, *id)
		}
	}

	if len(invoiceDpDtIDsUint) > 0 {
		var err error
		invoiceDpDtIDsPtrs := make([]*uint, len(invoiceDpDtIDsUint))
		for i := range invoiceDpDtIDsUint {
			invoiceDpDtIDsPtrs[i] = &invoiceDpDtIDsUint[i]
		}
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, invoiceDpDtIDsPtrs, "invoice_dp_dts", childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	productIDsUint := make([]uint, 0)
	for _, id := range productIDs {
		if id != nil {
			productIDsUint = append(productIDsUint, *id)
		}
	}

	if len(productIDsUint) > 0 {
		var err error
		productIDsPtrs := make([]*uint, len(productIDsUint))
		for i := range productIDsUint {
			productIDsPtrs[i] = &productIDsUint[i]
		}
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, productIDsPtrs, "products", childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	itemUnitIDsUint := make([]uint, 0)
	for _, id := range itemUnitIDs {
		if id != nil {
			itemUnitIDsUint = append(itemUnitIDsUint, *id)
		}
	}

	if len(itemUnitIDsUint) > 0 {
		var err error
		itemUnitIDsPtrs := make([]*uint, len(itemUnitIDsUint))
		for i := range itemUnitIDsUint {
			itemUnitIDsPtrs[i] = &itemUnitIDsUint[i]
		}
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, itemUnitIDsPtrs, "item_units", childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *InvoiceDpService) LockCreateInvoiceDpTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateInvoiceDpRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-LockCreateInvoiceDpTable", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDtIDs := utils.GetLockInvoiceDpSalesOrderIDs(req)

	soDtIDsUint := make([]uint, 0)
	for _, id := range soDtIDs {
		if id != nil {
			soDtIDsUint = append(soDtIDsUint, *id)
		}
	}

	if len(soDtIDsUint) > 0 {
		var err error
		soDtIDsPtrs := make([]*uint, len(soDtIDsUint))
		for i := range soDtIDsUint {
			soDtIDsPtrs[i] = &soDtIDsUint[i]
		}
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, soDtIDsPtrs, "so_dts", childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *InvoiceDpService) GetWidgetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceDpStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-GetWidgetInvoiceDps", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDps, total, err := s.repo.GetWidgetInvoiceDps(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return invoiceDps, total, nil
}

func (s *InvoiceDpService) ExcelGetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-ExcelGetInvoiceDps", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceDps, _, err := s.GetInvoiceDps(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "InvoiceDps"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{
		"Customer", "Invoice No", "Title", "Invoice Date", "Due Date",
		"Bank", "Currency", "Exchange Rate", "VAT", "PPh23",
		"Qty", "Sub Amount", "Grand Total DP", "Status", "Created By",
	}
	file.SetSheetRow(sheetName, "A1", &headers)

	file.SetColWidth(sheetName, "A", "A", 25)
	file.SetColWidth(sheetName, "B", "B", 15)
	file.SetColWidth(sheetName, "C", "C", 30)
	file.SetColWidth(sheetName, "D", "E", 15)
	file.SetColWidth(sheetName, "F", "F", 20)
	file.SetColWidth(sheetName, "G", "G", 15)
	file.SetColWidth(sheetName, "H", "H", 15)
	file.SetColWidth(sheetName, "I", "J", 15)
	file.SetColWidth(sheetName, "K", "M", 15)
	file.SetColWidth(sheetName, "N", "N", 15)
	file.SetColWidth(sheetName, "O", "O", 20)

	headerStyle, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DCE6F1"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	file.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)

	currencyStyle, _ := file.NewStyle(&excelize.Style{
		NumFmt: 4,
	})

	for i, invoice := range invoiceDps {
		row := i + 2
		file.SetSheetRow(sheetName, fmt.Sprintf("A%d", row), &[]interface{}{
			invoice.CustomerName,
			invoice.InvoiceNo,
			invoice.Title,
			invoice.InvoiceDate,
			invoice.DueDate,
			invoice.BankName,
			invoice.CurrencyName,
			invoice.ExchangeRate,
			invoice.VatName,
			invoice.Pph23Name,
			invoice.TotalQty,
			invoice.Subtotal,
			invoice.GrandTotal,
			invoice.Status,
			invoice.CreatedByName,
		})

		file.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), currencyStyle)
		file.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("M%d", row), currencyStyle)
	}

	file.SetActiveSheet(index)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *InvoiceDpService) CsvGetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-CsvGetInvoiceDps", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceDps, _, err := s.GetInvoiceDps(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err == nil && companyProfile != nil && companyProfile.CompanyName != nil {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Invoice Down Payments\n"
	csv += "\n"

	csv += "Customer,Invoice No,Title,Invoice Date,Due Date,Bank,Currency,Exchange Rate,Total VAT,Total PPh23,Qty,Sub Amount,Grand Total DP,Status,Created By\n"

	for _, invoice := range invoiceDps {
		customerName := utils.GetPtrVal(invoice.CustomerName)
		invoiceNo := utils.GetPtrVal(invoice.InvoiceNo)
		title := utils.GetPtrVal(invoice.Title)
		invoiceDate := utils.GetPtrVal(invoice.InvoiceDate)
		dueDate := utils.GetPtrVal(invoice.DueDate)

		bankName := utils.GetPtrVal(invoice.BankName)
		accountNumber := utils.GetPtrVal(invoice.AccountNumber)
		accountName := utils.GetPtrVal(invoice.AccountName)

		bankInfo := bankName
		if accountNumber != "" {
			if bankInfo != "" {
				bankInfo += " - "
			}
			bankInfo += accountNumber
		}
		if accountName != "" {
			if bankInfo != "" {
				bankInfo += " - "
			}
			bankInfo += accountName
		}

		currencyName := utils.GetPtrVal(invoice.CurrencyName)
		status := utils.GetPtrVal(invoice.Status)
		createdByName := utils.GetPtrVal(invoice.CreatedByName)

		exchangeRate := utils.GetFloatPtrVal(invoice.ExchangeRate)
		totalQty := utils.GetFloatPtrVal(invoice.TotalQty)
		subtotal := utils.GetFloatPtrVal(invoice.Subtotal)
		grandTotal := utils.GetFloatPtrVal(invoice.GrandTotal)

		totalVat := utils.GetFloatPtrVal(invoice.TotalVat)
		totalPph23 := utils.GetFloatPtrVal(invoice.TotalPph23)

		customerName = utils.EscapeCsvField(customerName)
		invoiceNo = utils.EscapeCsvField(invoiceNo)
		title = utils.EscapeCsvField(title)
		bankInfo = utils.EscapeCsvField(bankInfo)
		currencyName = utils.EscapeCsvField(currencyName)
		status = utils.EscapeCsvField(status)
		createdByName = utils.EscapeCsvField(createdByName)

		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s,%s\n",
			customerName,
			invoiceNo,
			title,
			invoiceDate,
			dueDate,
			bankInfo,
			currencyName,
			exchangeRate,
			totalVat,
			totalPph23,
			totalQty,
			subtotal,
			grandTotal,
			status,
			createdByName,
		)
	}

	return []byte(csv), nil
}

func (s *InvoiceDpService) Pdf(ctx *fiber.Ctx, req dtos.InvoiceDpDetailDTO, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-Pdf", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	form := req
	var invoiceDp *dtos.InvoiceDpDetailDTO
	var err error

	var params dtos.GetInvoiceDpParams
	if req.IsIDOnly != nil && *req.IsIDOnly == 1 {
		params.ID = req.ID

		invoiceDp, err = s.repo.GetInvoiceDpByID(ctx, &params, tx, childSpan)
		if err != nil {
			return nil, err
		}

		invoiceDpDts, err := s.repo.GetInvoiceDpDts(ctx, invoiceDp.ID, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch invoiceDpDts: %v", err)
		}

		if invoiceDp.CompanyProfileID != nil {
			companyParams := &dtos.GetCompanyProfileParams{ID: uint(*invoiceDp.CompanyProfileID)}
			company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
			if err != nil {
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to fetch company: %v", err)
			} else if company != nil {
				invoiceDp.Company = *company
			}
		}

		invoiceDp.InvoiceDpDts = invoiceDpDts
		form = *invoiceDp
		req.InvoiceNo = invoiceDp.InvoiceNo
	}

	var num string
	if req.InvoiceNo != nil {
		num = *req.InvoiceNo
	} else {
		num = ""
	}

	data := dtos.InvoiceDpPDFData{
		Num:  num,
		Form: form,
	}

	htmlFileName := "invoice-dp-detail"
	log.Println("Pdf-htmlFileName-idp", htmlFileName)

	templateFile, err := templateFS.Open(fmt.Sprintf("templates/%s.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to open embedded template: %v", err)
		return nil, err
	}

	templateContent, err := io.ReadAll(templateFile)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to read template content: %v", err)
		return nil, err
	}

	htmlFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error writing htmlFile:", err)
		return nil, err
	}
	defer os.Remove(htmlFile.Name())

	funcMap := template.FuncMap{
		"formatNumber": func(n float64, args ...int) string {
			decimals := 2
			if len(args) > 0 {
				decimals = args[0]
			}

			format := fmt.Sprintf("%%.%df", decimals)
			p := message.NewPrinter(language.English)
			return p.Sprintf(format, n)
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New(fmt.Sprintf("%s.html", htmlFileName)).Funcs(funcMap).Parse(string(templateContent))
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

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg:", err)
		return nil, err
	}

	headerContent, err := templateFS.ReadFile("templates/header.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	footerContent, err := templateFS.ReadFile("templates/footer.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}

	headerPath, err := createTempFileFromEmbed(string(headerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create header temp file: %w", err)
	}
	defer os.Remove(headerPath)

	footerPath, err := createTempFileFromEmbed(string(footerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create footer temp file: %w", err)
	}
	defer os.Remove(footerPath)

	page := wkhtmltopdf.NewPage(htmlFile.Name())
	page.EnableLocalFileAccess.Set(true)
	page.HeaderHTML.Set("file://" + headerPath)
	page.FooterHTML.Set("file://" + footerPath)
	page.FooterSpacing.Set(10)

	pdfg.AddPage(page)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)

	if err := pdfg.Create(); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg create:", err)
		return nil, err
	}

	uploadDir := "./public/generated_pdfs"

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	fileName := fmt.Sprintf("invoice-dp-%s.pdf", time.Now().Format("20060102150405"))
	pdfPath := filepath.Join(uploadDir, fileName)
	if err := pdfg.WriteFile(pdfPath); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdf path:", err)
		return nil, err
	}

	return &pdfPath, nil
}
