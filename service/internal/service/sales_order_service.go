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

func (s *SalesOrderService) CreateSalesOrder(ctx *fiber.Ctx, salesOrder *models.SalesOrder, tx *gorm.DB, span opentracing.Span) (*models.SalesOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSalesOrder", opentracing.ChildOf(span.Context()))

	if tx, err := s.repo.CreateSalesOrder(tx, salesOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return salesOrder, tx, nil
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

func (s *SalesOrderService) UpdateSalesOrder(ctx *fiber.Ctx, salesOrder *models.SalesOrder, tx *gorm.DB, span opentracing.Span) (*models.SalesOrder, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-UpdateSalesOrder", opentracing.ChildOf(span.Context()))

	if err := s.repo.UpdateSalesOrder(tx, salesOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return salesOrder, nil
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

// salesOrder *models.SalesOrder
func (s *SalesOrderService) CreateSoDts(ctx *fiber.Ctx, boms []models.SoDt, salesOrderID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSoDts", opentracing.ChildOf(span.Context()))

	soDtsModel := []models.SoDt{}
	var err error

	tx, soDtsModel, err = s.repo.CreateSoDts(tx, boms, salesOrderID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, soDtsModel, err
	}

	return tx, soDtsModel, nil
}

// bulk create/update boms for a salesOrder
func (s *SalesOrderService) BulkCreateUpdateSoDts(ctx *fiber.Ctx, soDts []models.SoDt, salesOrderID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-BulkCreateUpdateSoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateSoDts := []models.SoDt{}
	// filter with ID to bulk update
	bulkUpdateSoDts := []models.SoDt{}
	// get all ids
	soDtIDs := []uint{}

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

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
		if tx, err := s.repo.DeleteSoDtsWhereNotIn(ctx, tx, salesOrderID, soDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateSoDts) > 0 {
		if tx, _, err := s.repo.CreateSoDts(tx, bulkCreateSoDts, salesOrderID, childSpan); err != nil {
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

func (s *SalesOrderService) GetSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBySalesOrderIDs", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetSoDtsBySalesOrderIDs(ctx, tx, salesOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDts, nil
}

func (s *SalesOrderService) GetUpdatedSoDtsBySalesOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBySalesOrderID", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetUpdatedSoDtsBySalesOrderIDs(ctx, tx, salesOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return soDts, nil
}

func (s *SalesOrderService) MapCreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-MapCreateSoDts", opentracing.ChildOf(span.Context()))
	soDtsModel, err := utils.MapCreateSoDts(ctx, req, createdSalesOrder, userID, span)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
}

func (s *SalesOrderService) MapCreateSoDtBoms(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSoDts []models.SoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	// var soDtBomsModel []models.SoDtBom
	soDtBomsModel := utils.MapCreateSoDtBoms(ctx, req, createdSoDts, userID, span)

	return soDtBomsModel
}

func (s *SalesOrderService) CreateSoDtBoms(ctx *fiber.Ctx, soDtBoms []map[string]interface{}, salesOrderID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-CreateSoDtBoms", opentracing.ChildOf(span.Context()))

	if tx, err := s.repo.CreateSoDtBoms(tx, soDtBoms, salesOrderID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a salesOrder
// func (s *SalesOrderService) BulkCreateUpdateSoDtBoms(ctx *fiber.Ctx, soDts []dtos.SalesOrderSoDtListDTO, req dtos.UpdateSalesOrderRequest, salesOrderID uint, tx *gorm.DB, span opentracing.Span) error {
// func (s *SalesOrderService) BulkCreateUpdateSoDtBoms(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, salesOrderID uint, tx *gorm.DB, span opentracing.Span) error {
func (s *SalesOrderService) BulkCreateUpdateSoDtBoms(ctx *fiber.Ctx, soDts []dtos.SalesOrderSoDtListUpdateDTO, req dtos.UpdateSalesOrderRequest, salesOrderID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesOrderService-BulkCreateUpdateSoDtBoms", opentracing.ChildOf(span.Context()))

	bulkCreateSoDtBoms, bulkUpdateSoDtBoms, soDtBomIDs, err := utils.MapFilterUpdateSoDtBomsToSoDts(ctx, soDts, req, salesOrderID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()

		return err
	}

	// delete soDts that are not in the list
	if len(soDtBomIDs) > 0 {
		if err := s.repo.DeleteSoDtBomsWhereNotIn(ctx, tx, salesOrderID, soDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreateSoDtBoms) > 0 {
		if tx, err := s.repo.CreateSoDtBoms(tx, bulkCreateSoDtBoms, salesOrderID, childSpan); err != nil {
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

func (s *SalesOrderService) GetSoDtsBomBySalesOrders(ctx *fiber.Ctx, filters map[string]string, salesOrderIDs []uint, span opentracing.Span) ([]dtos.SalesOrderSoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-GetSoDtsBomBySalesOrders", opentracing.ChildOf(span.Context()))

	soDtBoms, err := s.repo.GetSoDtsBomBySalesOrders(ctx, filters, salesOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return soDtBoms, nil
}

func (s *SalesOrderService) MapFilterSoDtBomsToSoDts(ctx *fiber.Ctx, soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.SalesOrderSoDtListDTO, span opentracing.Span) []dtos.SalesOrderSoDtListDTO {
	return utils.MapFilterSoDtBomsToSoDts(soDtBoms, soDts)
}

func (s *SalesOrderService) MapCreateUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-MapCreateUpdateSoDts", opentracing.ChildOf(span.Context()))

	soDtsModel, err := utils.MapCreateUpdateSoDts(ctx, req, updatedSalesOrder, userID, span)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
}

// Lock all salesOrder table update
func (s *SalesOrderService) LockSalesOrderTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateSalesOrderRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesOrderService-LockSalesOrderTable", opentracing.ChildOf(span.Context()))

	if err := s.repo.LockSalesOrderHeader(ctx, tx, req, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
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
