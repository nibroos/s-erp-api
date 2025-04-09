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

type SalesOrderService struct {
	repo     *repository.SalesOrderRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewSalesOrderService(repo *repository.SalesOrderRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *SalesOrderService {
	return &SalesOrderService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *SalesOrderService) GetSalesOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSalesOrders", opentracing.ChildOf(span.Context()))

	salesOrders, total, err := s.repo.GetSalesOrders(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return salesOrders, total, nil
}

func (s *SalesOrderService) CreateSalesOrder(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSalesOrder", opentracing.ChildOf(span.Context()))

	customerSoCreatedThisMonthNumber, err := s.repo.GetCustomerSalesOrderCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)

	salesOrder, err := utils.MapCreateSalesOrder(ctx, req, userID, branchID, customerSoCreatedThisMonthNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, tx, err
	}

	if tx, err := s.repo.CreateSalesOrder(tx, &salesOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	// go routine to create schedule entity related to sales order
	// go func() {
	// 	if tx, err := s.CreateSchedule(ctx, *req.Schedule, salesOrder.ID, userID, tx, childSpan); err != nil {
	// 		childSpan.Finish()
	// 		tx.Rollback()
	// 		return
	// 	}
	// }()

	// bulk create item soDts ref ms items / product->boms
	var soDts []models.SoDt
	tx, soDts, err = s.CreateSoDts(ctx, req, userID, &salesOrder, tx, childSpan)

	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	quoDtIDs := utils.GetLockSalesOrderQuoIDs(req)
	quoDtIDsString := utils.JoinUintPtrsToString(quoDtIDs, ",")
	filters := map[string]string{"ids": quoDtIDsString}
	getQuoDtsQtyUpdate, err := s.repo.GetQuoDtQtyUpdate(ctx, tx, filters, childSpan)
	mapUpdateQuoDtsQty := utils.MapUpdateQuoDtsQty(getQuoDtsQtyUpdate, req)

	// bulk update quo dts qty_so = qty_so - qty
	if len(mapUpdateQuoDtsQty) > 0 {
		if err := s.repo.BulkUpdateQuoDtsQty(ctx, tx, mapUpdateQuoDtsQty, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		paramQuotation := map[string]string{"sales_order_id": fmt.Sprintf("%d", salesOrder.ID)}
		// get quotation_id
		quotationID, err := s.repo.GetQuotationIDBySalesOrderID(ctx, tx, paramQuotation, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		// params dtos.UpdateQuotationStatusRequest
		updateQuoParams := dtos.UpdateQuotationStatusRequest{
			ID:     quotationID,
			Status: "APPROVED",
		}

		// update status header quotations
		if err := s.repo.UpdateQuoStatus(ctx, tx, updateQuoParams, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}
	}

	// tx, err = s.CreateSoDtBoms(ctx, soDtBoms, tx, childSpan)
	tx, err = s.CreateSoDtBoms(ctx, soDts, req, &salesOrder, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return &salesOrder, tx, nil
}

// CreateSchedule
func (s *SalesOrderService) CreateSchedule(ctx *fiber.Ctx, req dtos.CreateScheduleRequest, salesOrderID uint, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSchedule", opentracing.ChildOf(span.Context()))

	schedule, err := utils.MapCreateSchedule(ctx, req, salesOrderID, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	if tx, err := s.repo.CreateSchedule(ctx, tx, &schedule, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return tx, err
	}

	// create steps
	if len(req.Steps) > 0 {
		steps, err := utils.MapCreateScheduleSteps(ctx, req.Steps, schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err := s.repo.CreateScheduleSteps(ctx, tx, steps, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		// create schedule task
		tasks, err := utils.MapCreateScheduleTasks(ctx, req.Steps, steps, schedule.ID, userID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}

		if tx, err := s.repo.CreateScheduleTasks(ctx, tx, tasks, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return tx, err
		}
	}

	return tx, nil
}

func (s *SalesOrderService) GetSalesOrderByID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.SalesOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSalesOrderByID", opentracing.ChildOf(span.Context()))

	salesOrder, err := s.repo.GetSalesOrderByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return salesOrder, nil
}

func (s *SalesOrderService) UpdateSalesOrder(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesOrder, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-UpdateSalesOrder", opentracing.ChildOf(span.Context()))

	salesOrder, err := utils.MapUpdateSalesOrder(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.UpdateSalesOrder(tx, &salesOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Bulk/Create Update Batch SoDts
	soDts, err := s.MapUpdateSoDts(ctx, req, &salesOrder, userID, childSpan)

	tx, err = s.BulkCreateUpdateSoDts(ctx, req, &salesOrder, userID, soDts, salesOrder.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	updatedSalesOrderIDs := make([]uint, 0)
	updatedSalesOrderIDs = append(updatedSalesOrderIDs, salesOrder.ID)

	// get updated soDts
	updatedSoDts, err := s.GetUpdatedSoDtsBySalesOrderIDs(ctx, tx, updatedSalesOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Bulk/Create Update Batch SoDtBoms
	err = s.BulkCreateUpdateSoDtBoms(ctx, updatedSoDts, req, salesOrder.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return &salesOrder, nil
}

func (s *SalesOrderService) DeleteSalesOrder(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-DeleteSalesOrder", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteSalesOrder(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *SalesOrderService) RestoreSalesOrder(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-RestoreSalesOrder", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreSalesOrder(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *SalesOrderService) ExcelGetSalesOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-ExcelGetSalesOrders", opentracing.ChildOf(span.Context()))

	salesOrders, _, err := s.GetSalesOrders(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "salesOrders"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("SalesOrders", "A1", &[]string{"ID", "Branch", "Code", "Factory Code", "Name", "Sku", "Barcode", "Unit", "Specification", "Desc", "Remark", "Price Sell", "Price Buy"})

	for i, salesOrder := range salesOrders {
		row := []interface{}{
			salesOrder.ID,
			utils.GetPtrVal(salesOrder.Remark),
		}
		file.SetSheetRow("SalesOrders", fmt.Sprintf("A%d", i+2), &row)
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
func (s *SalesOrderService) CsvGetSalesOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CsvGetSalesOrders", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	salesOrders, _, err := s.GetSalesOrders(ctx, filters, childSpan)
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
	csv += "Master SalesOrder\n"
	csv += "\n"

	csv += "ID,Branch,Code,Factory Code,Name,Sku,Barcode,Unit,Specification,Desc,Remark,Price Sell,Price Buy\n"
	// Build CSV rows
	for _, salesOrder := range salesOrders {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			salesOrder.ID,
			utils.GetPtrVal(salesOrder.Remark),
		)
	}

	return []byte(csv), nil
}

// quotation *models.SalesOrder
func (s *SalesOrderService) CreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, userID uint, createdSalesOrder *models.SalesOrder, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSoDts", opentracing.ChildOf(span.Context()))

	soDts, err := s.MapCreateSoDts(ctx, req, createdSalesOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, soDts, err
	}

	soDtsModel := []models.SoDt{}

	tx, soDtsModel, err = s.repo.CreateSoDts(tx, soDts, createdSalesOrder.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, soDtsModel, err
	}

	return tx, soDtsModel, nil
}

// bulk create/update boms for a quotation
func (s *SalesOrderService) BulkCreateUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, soDts []models.SoDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-BulkCreateUpdateSoDts", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch SoDts
	soDts, err := s.MapUpdateSoDts(ctx, req, updatedSalesOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// filter without ID to bulk create
	bulkCreateSoDts := []models.SoDt{}
	// filter with ID to bulk update
	bulkUpdateSoDts := []models.SoDt{}
	// get all ids
	soDtIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.ID == 0 {
			soDt.CreatedByID = &userID
			soDt.CreatedAt = time.Now()
			bulkCreateSoDts = append(bulkCreateSoDts, soDt)
		} else {
			bulkUpdateSoDts = append(bulkUpdateSoDts, soDt)
			soDtIDs = append(soDtIDs, soDt.ID)
		}
	}

	// delete soDts that are not in the list
	if len(soDtIDs) > 0 {
		if tx, err := s.repo.DeleteSoDtsWhereNotIn(ctx, tx, quotationID, soDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateSoDts) > 0 {
		if tx, _, err := s.repo.CreateSoDts(tx, bulkCreateSoDts, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdateSoDts) > 0 {
		if tx, err := s.repo.UpdateSoDts(tx, bulkUpdateSoDts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *SalesOrderService) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetSoDtsBySalesOrderIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDts, nil
}

func (s *SalesOrderService) GetUpdatedSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBySalesOrderID", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetUpdatedSoDtsBySalesOrderIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return soDts, nil
}

func (s *SalesOrderService) MapCreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-MapCreateSoDts", opentracing.ChildOf(span.Context()))

	soDtsModel, err := utils.MapCreateSoDts(ctx, req, createdSalesOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
}

func (s *SalesOrderService) MapCreateSoDtBoms(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateSalesOrderRequest, createdSoDts []models.SoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	soDtBomsModel := utils.MapCreateSoDtBoms(ctx, req, createdSoDts, userID, span)

	return soDtBomsModel
}

func (s *SalesOrderService) CreateSoDtBoms(ctx *fiber.Ctx, soDts []models.SoDt, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSoDtBoms", opentracing.ChildOf(span.Context()))

	// bulk create boms
	soDtBoms := s.MapCreateSoDtBoms(ctx, tx, req, soDts, userID, childSpan)

	if tx, err := s.repo.CreateSoDtBoms(tx, soDtBoms, createdSalesOrder.ID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *SalesOrderService) BulkCreateUpdateSoDtBoms(ctx *fiber.Ctx, soDts []dtos.SalesOrderSoDtListUpdateDTO, req dtos.UpdateSalesOrderRequest, quotationID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-BulkCreateUpdateSoDtBoms", opentracing.ChildOf(span.Context()))

	bulkCreateSoDtBoms, bulkUpdateSoDtBoms, soDtBomIDs, err := utils.MapFilterUpdateSoDtBomsToSoDts(ctx, soDts, req, quotationID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// delete soDts that are not in the list
	if len(soDtBomIDs) > 0 {
		if err := s.repo.DeleteSoDtBomsWhereNotIn(ctx, tx, quotationID, soDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateSoDtBoms) > 0 {
		if tx, err := s.repo.CreateSoDtBoms(tx, bulkCreateSoDtBoms, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdateSoDtBoms) > 0 {
		if err := s.repo.UpdateSoDtBoms(tx, bulkUpdateSoDtBoms, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *SalesOrderService) DeleteSoDtBomsBySalesOrderID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-DeleteSoDtBomsBySalesOrderID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteSoDtBomsBySalesOrderID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *SalesOrderService) DeleteSoDtsBySalesOrderID(ctx *fiber.Ctx, params *dtos.GetSalesOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-DeleteSoDtsBySalesOrderID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteSoDtsBySalesOrderID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *SalesOrderService) GetSoDtsBomBySalesOrders(ctx *fiber.Ctx, filters map[string]string, quotationIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBomBySalesOrders", opentracing.ChildOf(span.Context()))

	soDtBoms, err := s.repo.GetSoDtsBomBySalesOrders(ctx, filters, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return soDtBoms, nil
}

func (s *SalesOrderService) MapFilterSoDtBomsToSoDts(ctx *fiber.Ctx, soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.SalesOrderSoDtListDTO, span opentracing.Span) []dtos.SalesOrderSoDtListDTO {
	return utils.MapFilterSoDtBomsToSoDts(soDtBoms, soDts)
}

func (s *SalesOrderService) MapUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-MapUpdateSoDts", opentracing.ChildOf(span.Context()))

	soDtsModel, err := utils.MapUpdateSoDts(ctx, req, updatedSalesOrder, userID, span)

	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
}

// Lock all quotation table update
func (s *SalesOrderService) LockSalesOrderTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateSalesOrderRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-LockSalesOrderTable", opentracing.ChildOf(span.Context()))

	if req.ID > 0 {
		if err := s.repo.LockSalesOrderHeader(ctx, tx, req, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	soDtIDs, soDtBomIDs, productIDs, itemUnitIDs := utils.GetSoIDs(req)

	if len(soDtIDs) > 0 {
		if err := s.repo.LockSoDts(ctx, tx, soDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(soDtBomIDs) > 0 {
		if err := s.repo.LockSoDtBoms(ctx, tx, soDtBomIDs, childSpan); err != nil {
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

func (s *SalesOrderService) LockCreateSalesOrderTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateSalesOrderRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-LockCreateSalesOrderTable", opentracing.ChildOf(span.Context()))

	quoDtIDs := utils.GetLockSalesOrderQuoIDs(req)

	if len(quoDtIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, quoDtIDs, "quo_dts", childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *SalesOrderService) GetRefIndexQuoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefIndexQuoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetRefIndexQuoDts", opentracing.ChildOf(span.Context()))

	quoDts, total, err := s.repo.GetRefIndexQuoDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	quotationIDs := utils.GetQuoDtIDs(quoDts)

	var quoDtBoms []dtos.QuotationQuoDtBomListDTO
	if len(quotationIDs) > 0 {
		quoDtBoms, err = s.repo.GetRefQuoDtsBomByQuoDtIDs(ctx, filters, quotationIDs, childSpan)

		if err != nil {
			defer childSpan.Finish()
			return nil, 0, err
		}
	}

	if len(quoDtBoms) > 0 {
		quoDts = utils.MapRefQuoDtBomsToQuoDts(quoDtBoms, quoDts)
	}

	return quoDts, total, nil
}
