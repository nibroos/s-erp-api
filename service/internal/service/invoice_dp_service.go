package service

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
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

	tx, _, err = s.CreateInvoiceDpDts(ctx, req, userID, &invoiceDp, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	soDtIDs := utils.GetLockInvoiceDpSalesOrderIDs(req)
	if len(soDtIDs) > 0 {
		soDtIDsUint := make([]uint, 0)
		for _, id := range soDtIDs {
			if id != nil {
				soDtIDsUint = append(soDtIDsUint, *id)
			}
		}

		if len(soDtIDsUint) > 0 {
			soDtsQtyUpdate, err := s.repo.GetSoDtQtyUpdateForInvoice(ctx, soDtIDsUint, childSpan)
			if err != nil {
				tx.Rollback()
				return nil, tx, err
			}

			mapUpdateSoDtsQty := utils.MapUpdateSoDtsQtyForInvoice(soDtsQtyUpdate, req)

			if len(mapUpdateSoDtsQty) > 0 {
				tx, err = s.repo.BulkUpdateSoDtsQty(tx, mapUpdateSoDtsQty, childSpan)
				if err != nil {
					tx.Rollback()
					return nil, tx, err
				}
			}

			salesOrderIDs := make(map[uint]bool)
			for _, soDt := range soDtsQtyUpdate {
				if soDt.SalesOrderID != nil {
					salesOrderIDs[*soDt.SalesOrderID] = true
				}
			}

			for salesOrderID := range salesOrderIDs {
				tx, err = s.repo.UpdateSalesOrderStatus(tx, salesOrderID, "invoiced", childSpan)
				if err != nil {
					tx.Rollback()
					return nil, tx, err
				}
			}
		}
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

	return invoiceDp, nil
}

func (s *InvoiceDpService) UpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceDp, error) {
	childSpan := opentracing.StartSpan("InvoiceDpService-UpdateInvoiceDp", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceDp, err := utils.MapUpdateInvoiceDp(ctx, req, userID, branchID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.UpdateInvoiceDp(tx, &invoiceDp, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	params := dtos.GetInvoiceDpParams{ID: req.ID}
	existingInvoiceDp, err := s.GetInvoiceDpByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
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
	}

	if len(updateInvoiceDpDts) > 0 {
		tx, err = s.repo.BulkUpdateInvoiceDpDts(tx, updateInvoiceDpDts, childSpan)
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

	invoiceDpDts, err := utils.MapCreateInvoiceDpDts(ctx, req, createdInvoiceDp, userID, childSpan)
	if err != nil {
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

// func (s *InvoiceDpService) ExcelGetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
// 	childSpan := opentracing.StartSpan("InvoiceDpService-ExcelGetInvoiceDps", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	invoiceDps, _, err := s.GetInvoiceDps(ctx, filters, childSpan)
// 	if err != nil {
// 		return nil, err
// 	}

// 	file := excelize.NewFile()

// 	sheetName := "InvoiceDps"
// 	index, err := file.NewSheet(sheetName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	file.SetSheetRow(sheetName, "A1", &[]string{
// 		"ID", "Invoice No", "Invoice Date", "Customer", "Currency", "Payment Term",
// 		"DP Percentage", "Subtotal", "Total Discount", "Total PPh23", "Total VAT", "Grand Total",
// 		"Remark", "Created By", "Created At",
// 	})

// 	for i, invoiceDp := range invoiceDps {
// 		row := []interface{}{
// 			invoiceDp.ID,
// 			utils.GetPtrVal(invoiceDp.InvoiceNo),
// 			utils.GetPtrVal(invoiceDp.InvoiceDate),
// 			utils.GetPtrVal(invoiceDp.CustomerName),
// 			utils.GetPtrVal(invoiceDp.CurrencyName),
// 			utils.GetPtrVal(invoiceDp.PaymentTermName),
// 			utils.GetPtrVal(invoiceDp.DpPercentage),
// 			utils.GetPtrVal(invoiceDp.Subtotal),
// 			utils.GetPtrVal(invoiceDp.TotalDiscount),
// 			utils.GetPtrVal(invoiceDp.TotalPph23),
// 			utils.GetPtrVal(invoiceDp.TotalVat),
// 			utils.GetPtrVal(invoiceDp.GrandTotal),
// 			utils.GetPtrVal(invoiceDp.Remark),
// 			utils.GetPtrVal(invoiceDp.CreatedByName),
// 			utils.GetPtrVal(invoiceDp.CreatedAt),
// 		}
// 		file.SetSheetRow(sheetName, fmt.Sprintf("A%d", i+2), &row)
// 	}

// 	file.SetActiveSheet(index)

// 	buffer, err := file.WriteToBuffer()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return buffer.Bytes(), nil
// }

// func (s *InvoiceDpService) CsvGetInvoiceDps(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
// 	childSpan := opentracing.StartSpan("InvoiceDpService-CsvGetInvoiceDps", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	filters["is_csv"] = "1"

// 	invoiceDps, _, err := s.GetInvoiceDps(ctx, filters, childSpan)
// 	if err != nil {
// 		return nil, err
// 	}

// 	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
// 	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
// 	appName := "App"
// 	if err != nil {
// 	} else {
// 		appName = *companyProfile.CompanyName
// 	}

// 	csv := fmt.Sprintf("%s\n", appName)
// 	csv += "\n"
// 	csv += "Invoice Down Payment\n"
// 	csv += "\n"

// 	csv += "ID,Invoice No,Invoice Date,Customer,Currency,Payment Term,DP Percentage,Subtotal,Total Discount,Total PPh23,Total VAT,Grand Total,Remark,Created By,Created At\n"

// 	for _, invoiceDp := range invoiceDps {
// 		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%v,%v,%v,%v,%v,%v,%s,%s,%s\n",
// 			invoiceDp.ID,
// 			utils.GetPtrValString(invoiceDp.InvoiceNo),
// 			utils.GetPtrValString(invoiceDp.InvoiceDate),
// 			utils.GetPtrValString(invoiceDp.CustomerName),
// 			utils.GetPtrValString(invoiceDp.CurrencyName),
// 			utils.GetPtrValString(invoiceDp.PaymentTermName),
// 			utils.GetPtrValFloat64(invoiceDp.DpPercentage),
// 			utils.GetPtrValFloat64(invoiceDp.Subtotal),
// 			utils.GetPtrValFloat64(invoiceDp.TotalDiscount),
// 			utils.GetPtrValFloat64(invoiceDp.TotalPph23),
// 			utils.GetPtrValFloat64(invoiceDp.TotalVat),
// 			utils.GetPtrValFloat64(invoiceDp.GrandTotal),
// 			utils.GetPtrValString(invoiceDp.Remark),
// 			utils.GetPtrValString(invoiceDp.CreatedByName),
// 			utils.GetPtrValString(invoiceDp.CreatedAt),
// 		)
// 	}

// 	return []byte(csv), nil
// }

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
