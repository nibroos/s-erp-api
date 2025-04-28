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

type InventoryService struct {
	repo     *repository.InventoryRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewInventoryService(repo *repository.InventoryRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *InventoryService {
	return &InventoryService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *InventoryService) GetInventories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InventoryListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInventories", opentracing.ChildOf(span.Context()))

	inventories, total, err := s.repo.GetInventories(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return inventories, total, nil
}

func (s *InventoryService) GetStocks(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.StockListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetStocks", opentracing.ChildOf(span.Context()))

	inventories, total, err := s.repo.GetStocks(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return inventories, total, nil
}

func (s *InventoryService) CreateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Inventory, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-CreateInventory", opentracing.ChildOf(span.Context()))

	customerInvCreatedThisMonthNumber, err := s.repo.GetCustomerInventoryCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)

	inventory, err := utils.MapCreateInventory(ctx, req, userID, branchID, customerInvCreatedThisMonthNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, tx, err
	}

	if tx, err = s.repo.CreateInventory(tx, &inventory, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	if len(req.InvDts) > 0 {
		tx, err = s.CreateInvDts(ctx, req, userID, &inventory, tx, childSpan)

		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		if req.IoType == "INVENTORY_OUT" {
			err = s.repo.CreateOrUpdateStockOut(ctx, tx, req, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}
		} else if req.IoType == "INVENTORY_IN" {
			err = s.repo.CreateOrUpdateStockIn(ctx, tx, req, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}
		}

		soDtIDs, soDtBomIDs := utils.GetLockInventorySoIDs(req)
		soDtIDsString := utils.JoinUintPtrsToString(soDtIDs, ",")
		soDtBomIDsString := utils.JoinUintPtrsToString(soDtBomIDs, ",")
		filters := map[string]string{"ids": soDtIDsString}

		if len(soDtIDs) > 0 {
			getSoDtsQtyUpdate, err := s.repo.GetSoDtQtyUpdate(ctx, tx, filters, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}

			mapUpdateInvSoDtsQty := utils.MapUpdateInvSoDtsQty(getSoDtsQtyUpdate, req)

			// bulk update so dts qty_out = qty_out - qty
			if len(mapUpdateInvSoDtsQty) > 0 {
				if err := s.repo.BulkUpdateInvSoDtsQty(ctx, tx, mapUpdateInvSoDtsQty, childSpan); err != nil {
					defer childSpan.Finish()
					tx.Rollback()
					return nil, tx, err
				}
			}
		}

		if len(soDtBomIDs) > 0 {
			filters = map[string]string{"ids": soDtBomIDsString}
			getSoDtBomsQtyUpdate, err := s.repo.GetSoDtBomQtyUpdate(ctx, tx, filters, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}
			mapUpdateInvSoDtBomsQty := utils.MapUpdateInvSoDtBomsQty(getSoDtBomsQtyUpdate, req)

			if len(mapUpdateInvSoDtBomsQty) > 0 {
				if err := s.repo.BulkUpdateInvSoDtBomsQty(ctx, tx, mapUpdateInvSoDtBomsQty, childSpan); err != nil {
					defer childSpan.Finish()
					tx.Rollback()
					return nil, tx, err
				}
			}
		}
	}

	return &inventory, tx, nil
}

func (s *InventoryService) GetInventoryByID(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, span opentracing.Span) (*dtos.InventoryDetailDTO, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInventoryByID", opentracing.ChildOf(span.Context()))

	inventory, err := s.repo.GetInventoryByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return inventory, nil
}

func (s *InventoryService) UpdateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Inventory, error) {
	childSpan := opentracing.StartSpan("InventoryService-UpdateInventory", opentracing.ChildOf(span.Context()))

	inventory, err := utils.MapUpdateInventory(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.UpdateInventory(tx, &inventory, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Bulk/Create Update Batch InvDts
	invDts, err := utils.MapUpdateInvDts(ctx, req, &inventory, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// Get all inv dts by inventory IDs
	createdInventoryIDs := make([]uint, 0)
	createdInventoryIDs = append(createdInventoryIDs, inventory.ID)

	oldInvDts, err := s.repo.GetInvDtsByInventoryIDs(ctx, tx, createdInventoryIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.updateRefReverseQtyInOut(ctx, req, oldInvDts, createdInventoryIDs, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.UpdateStocksOnUpdateInventory(ctx, req, oldInvDts, createdInventoryIDs, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.BulkCreateUpdateInvDts(ctx, req, &inventory, userID, invDts, inventory.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	utils.LogErrors(childSpan, fmt.Errorf("simulated error"))
	return &inventory, err

	// return &inventory, nil
}

// bulk create/update boms for a quotation
func (s *InventoryService) UpdateStocksOnUpdateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, oldInvDts []dtos.InventoryInvDtListDTO, createdInventoryIDs []uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-UpdateStocksOnUpdateInventory", opentracing.ChildOf(span.Context()))

	if req.IoType == "INVENTORY_OUT" {
		err := s.repo.ResetCreateOrUpdateStockOut(ctx, tx, req, oldInvDts, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		err = s.repo.CreateOrUpdateStockOut(ctx, tx, req, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	} else if req.IoType == "INVENTORY_IN" {
		err := s.repo.ResetCreateOrUpdateStockIn(ctx, tx, req, oldInvDts, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		err = s.repo.CreateOrUpdateStockIn(ctx, tx, req, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *InventoryService) updateRefReverseQtyInOut(ctx *fiber.Ctx, req dtos.FormInventoryRequest, oldInvDts []dtos.InventoryInvDtListDTO, createdInventoryIDs []uint, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-updateRefReverseQtyInOut", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch InvDts
	refSoDtID, refSoDtBomID, refPoDtID, refPoDtBomID, refInvDtID, err := utils.MapOldUpdateInvDts(ctx, req, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	var soDt []map[string]interface{}
	var soDtBom []map[string]interface{}
	var poDt []map[string]interface{}
	var poDtBom []map[string]interface{}
	var invDt []map[string]interface{}

	if len(refSoDtID) > 0 {
		soDt, err = s.repo.GetRefOutDtBySoDtID(ctx, tx, "so_dts", "sales_order_id", refSoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refSoDtBomID) > 0 {
		soDtBom, err = s.repo.GetRefOutDtBySoDtID(ctx, tx, "so_dt_boms", "sales_order_id", refSoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtID) > 0 {
		poDt, err = s.repo.GetRefOutDtBySoDtID(ctx, tx, "po_dts", "purchase_order_id", refPoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtBomID) > 0 {
		poDtBom, err = s.repo.GetRefOutDtBySoDtID(ctx, tx, "po_dt_boms", "purchase_order_id", refPoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refInvDtID) > 0 {
		invDt, err = s.repo.GetRefOutDtBySoDtID(ctx, tx, "inv_dts", "inventory_id", refInvDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(soDt) > 0 {

	}

	if len(soDtBom) > 0 {

	}

	if len(poDt) > 0 {

	}

	if len(poDtBom) > 0 {

	}

	if len(invDt) > 0 {

	}

	// err = s.repo.ResetCreateOrUpdateStockOut(ctx, tx, req, oldInvDts, childSpan)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	tx.Rollback()
	// 	return nil, err
	// }

	return tx, nil
}

func (s *InventoryService) DeleteInventory(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-DeleteInventory", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteInventory(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InventoryService) RestoreInventory(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-RestoreInventory", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestoreInventory(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// github.com/xuri/excelize/v2
func (s *InventoryService) ExcelGetInventories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InventoryService-ExcelGetInventories", opentracing.ChildOf(span.Context()))

	inventories, _, err := s.GetInventories(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	file := excelize.NewFile()

	// Create a new sheet
	sheetName := "inventories"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	file.SetSheetRow("Inventories", "A1", &[]string{"ID", "Branch", "Code", "Factory Code", "Name", "Sku", "Barcode", "Unit", "Specification", "Desc", "Remark", "Price Sell", "Price Buy"})

	for i, inventory := range inventories {
		row := []interface{}{
			inventory.ID,
			utils.GetPtrVal(inventory.Remark),
		}
		file.SetSheetRow("Inventories", fmt.Sprintf("A%d", i+2), &row)
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
func (s *InventoryService) CsvGetInventories(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InventoryService-CsvGetInventories", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"
	inventories, _, err := s.GetInventories(ctx, filters, childSpan)
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
	csv += "Master Inventory\n"
	csv += "\n"

	csv += "ID,Branch,Code,Factory Code,Name,Sku,Barcode,Unit,Specification,Desc,Remark,Price Sell,Price Buy\n"
	// Build CSV rows
	for _, inventory := range inventories {
		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			inventory.ID,
			utils.GetPtrVal(inventory.Remark),
		)
	}

	return []byte(csv), nil
}

// quotation *models.Inventory
func (s *InventoryService) CreateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, createdInventory *models.Inventory, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-CreateInvDts", opentracing.ChildOf(span.Context()))

	invDts, err := utils.MapCreateInvDts(ctx, req, createdInventory, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.CreateInvDts(tx, invDts, createdInventory.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *InventoryService) BulkCreateUpdateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, updatedInventory *models.Inventory, userID uint, soDts []models.InvDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-BulkCreateUpdateInvDts", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch InvDts
	invDts, err := utils.MapUpdateInvDts(ctx, req, updatedInventory, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	// filter without ID to bulk create
	bulkCreateInvDts := []models.InvDt{}
	// filter with ID to bulk update
	bulkUpdateInvDts := []models.InvDt{}
	// get all ids
	invDtIDs := []uint{}

	for _, invDt := range invDts {
		if invDt.ID == 0 {
			invDt.CreatedByID = &userID
			invDt.CreatedAt = time.Now()
			bulkCreateInvDts = append(bulkCreateInvDts, invDt)
		} else {
			bulkUpdateInvDts = append(bulkUpdateInvDts, invDt)
			invDtIDs = append(invDtIDs, invDt.ID)
		}
	}

	// delete invDts that are not in the list
	if len(invDtIDs) > 0 {
		if tx, err := s.repo.DeleteInvDtsWhereNotIn(ctx, tx, quotationID, invDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateInvDts) > 0 {
		if tx, err := s.repo.CreateInvDts(tx, bulkCreateInvDts, quotationID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdateInvDts) > 0 {
		if tx, err := s.repo.UpdateInvDts(tx, bulkUpdateInvDts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *InventoryService) GetInvDtsByInventoryIDs(ctx *fiber.Ctx, tx *gorm.DB, inventoryIDs []uint, span opentracing.Span) ([]dtos.InventoryInvDtListDTO, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInvDtsByInventoryIDs", opentracing.ChildOf(span.Context()))

	invDts, err := s.repo.GetInvDtsByInventoryIDs(ctx, tx, inventoryIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return invDts, nil
}

func (s *InventoryService) DeleteInvDtsByInventoryID(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-DeleteInvDtsByInventoryID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeleteInvDtsByInventoryID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

// Lock all quotation table update
func (s *InventoryService) LockInventoryTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-LockInventoryTable", opentracing.ChildOf(span.Context()))

	if err := s.repo.LockInventoryHeader(ctx, tx, req, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	soDtIDs, productIDs, itemUnitIDs := utils.GetInvIDs(req)

	if len(soDtIDs) > 0 {
		if err := s.repo.LockInvDts(ctx, tx, soDtIDs, childSpan); err != nil {
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

func (s *InventoryService) LockCreateInventoryTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormInventoryRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-LockCreateInventoryTable", opentracing.ChildOf(span.Context()))

	soDtIDs, soDtBomIDs := utils.GetLockInventorySoIDs(req)

	if len(soDtIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, soDtIDs, "so_dts", childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(soDtBomIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, soDtBomIDs, "so_dt_boms", childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *InventoryService) GetRefIndexSoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexSoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetRefIndexSoDts", opentracing.ChildOf(span.Context()))

	soDts, total, err := s.repo.GetRefIndexSoDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	// soIDs := utils.GetInvSoDtIDs(soDts)

	// var soDtBoms []dtos.InvSalesOrderQuoDtBomListDTO
	// if len(soIDs) > 0 {
	// 	soDtBoms, err = s.repo.GetRefSoDtsBomByQuoDtIDs(ctx, filters, soIDs, childSpan)

	// 	if err != nil {
	// 		defer childSpan.Finish()
	// 		return nil, 0, err
	// 	}
	// }

	// if len(soDtBoms) > 0 {
	// 	soDts = utils.MapInvRefSoDtBomsToQuoDts(soDtBoms, soDts)
	// }

	return soDts, total, nil
}
