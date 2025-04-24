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

func (s *InventoryService) CreateInventory(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Inventory, *gorm.DB, error) {
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
		// bulk create item soDts ref ms items / product->boms
		tx, err = s.CreateInvDts(ctx, req, userID, &inventory, tx, childSpan)

		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		quoDtIDs := utils.GetLockInventoryQuoIDs(req)
		quoDtIDsString := utils.JoinUintPtrsToString(quoDtIDs, ",")
		filters := map[string]string{"ids": quoDtIDsString}
		getQuoDtsQtyUpdate, err := s.repo.GetQuoDtQtyUpdate(ctx, tx, filters, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, tx, err
		}

		mapUpdateInvSoDtsQty := utils.MapUpdateInvSoDtsQty(getQuoDtsQtyUpdate, req)

		// bulk update so dts qty_so = qty_so - qty
		if len(mapUpdateInvSoDtsQty) > 0 {
			if err := s.repo.BulkUpdateInvSoDtsQty(ctx, tx, mapUpdateInvSoDtsQty, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}

			paramSalesOrder := map[string]string{"sales_order_id": fmt.Sprintf("%d", inventory.ID)}
			// get quotation_id
			soID, err := s.repo.GetSoIDByInventoryID(ctx, tx, paramSalesOrder, childSpan)
			if err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
			}

			// params dtos.UpdateSalesOrderStatusRequest
			updateSoParams := dtos.UpdateInvSalesOrderStatusRequest{
				ID:     soID,
				Status: "APPROVED",
			}

			// update status header quotations
			if err := s.repo.UpdateSoStatus(ctx, tx, updateSoParams, childSpan); err != nil {
				defer childSpan.Finish()
				tx.Rollback()
				return nil, tx, err
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

func (s *InventoryService) UpdateInventory(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.Inventory, error) {
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
	soDts, err := s.MapUpdateInvDts(ctx, req, &inventory, userID, childSpan)

	tx, err = s.BulkCreateUpdateInvDts(ctx, req, &inventory, userID, soDts, inventory.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return &inventory, nil
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
func (s *InventoryService) CreateInvDts(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, userID uint, createdInventory *models.Inventory, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-CreateInvDts", opentracing.ChildOf(span.Context()))

	soDts, err := s.MapCreateInvDts(ctx, req, createdInventory, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.CreateInvDts(tx, soDts, createdInventory.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// bulk create/update boms for a quotation
func (s *InventoryService) BulkCreateUpdateInvDts(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, updatedInventory *models.Inventory, userID uint, soDts []models.InvDt, quotationID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-BulkCreateUpdateInvDts", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch InvDts
	soDts, err := s.MapUpdateInvDts(ctx, req, updatedInventory, userID, childSpan)
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
	soDtIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.ID == 0 {
			soDt.CreatedByID = &userID
			soDt.CreatedAt = time.Now()
			bulkCreateInvDts = append(bulkCreateInvDts, soDt)
		} else {
			bulkUpdateInvDts = append(bulkUpdateInvDts, soDt)
			soDtIDs = append(soDtIDs, soDt.ID)
		}
	}

	// delete soDts that are not in the list
	if len(soDtIDs) > 0 {
		if tx, err := s.repo.DeleteInvDtsWhereNotIn(ctx, tx, quotationID, soDtIDs, childSpan); err != nil {
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

func (s *InventoryService) GetInvDtsByInventoryIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.InventoryInvDtListDTO, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInvDtsByInventoryIDs", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetInvDtsByInventoryIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDts, nil
}

func (s *InventoryService) GetUpdatedInvDtsByInventoryIDs(ctx *fiber.Ctx, tx *gorm.DB, quotationIDs []uint, span opentracing.Span) ([]dtos.InventoryInvDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInvDtsByInventoryID", opentracing.ChildOf(span.Context()))

	soDts, err := s.repo.GetUpdatedInvDtsByInventoryIDs(ctx, tx, quotationIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return soDts, nil
}

func (s *InventoryService) MapCreateInvDts(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, createdInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {
	childSpan := opentracing.StartSpan("InventoryService-MapCreateInvDts", opentracing.ChildOf(span.Context()))

	soDtsModel, err := utils.MapCreateInvDts(ctx, req, createdInventory, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
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

func (s *InventoryService) MapUpdateInvDts(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, updatedInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {
	childSpan := opentracing.StartSpan("InventoryService-MapUpdateInvDts", opentracing.ChildOf(span.Context()))

	soDtsModel, err := utils.MapUpdateInvDts(ctx, req, updatedInventory, userID, span)

	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return soDtsModel, nil
}

// Lock all quotation table update
func (s *InventoryService) LockInventoryTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdateInventoryRequest, span opentracing.Span) (*gorm.DB, error) {
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

func (s *InventoryService) LockCreateInventoryTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.CreateInventoryRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-LockCreateInventoryTable", opentracing.ChildOf(span.Context()))

	soDtIDs := utils.GetLockInventoryQuoIDs(req)

	if len(soDtIDs) > 0 {
		var err error
		if tx, err = s.utilRepo.LockRowTable(ctx, tx, soDtIDs, "so_dts", childSpan); err != nil {
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

	soIDs := utils.GetInvSoDtIDs(soDts)

	var soDtBoms []dtos.InvSalesOrderQuoDtBomListDTO
	if len(soIDs) > 0 {
		soDtBoms, err = s.repo.GetRefSoDtsBomByQuoDtIDs(ctx, filters, soIDs, childSpan)

		if err != nil {
			defer childSpan.Finish()
			return nil, 0, err
		}
	}

	if len(soDtBoms) > 0 {
		soDts = utils.MapInvRefSoDtBomsToQuoDts(soDtBoms, soDts)
	}

	return soDts, total, nil
}
