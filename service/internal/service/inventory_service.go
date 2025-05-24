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
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/valyala/fasthttp"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
)

type InventoryService struct {
	repo     *repository.InventoryRepository
	utilRepo *repository.UtilRepository
	rabbitmq *config.RabbitMQ
	tracer   opentracing.Tracer
}

func NewInventoryService(repo *repository.InventoryRepository, utilRepo *repository.UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *InventoryService {
	return &InventoryService{
		repo:     repo,
		utilRepo: utilRepo,
		rabbitmq: rabbitmq,
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

	customerID := req.CustomerID
	customerInvCreatedThisMonthNumber := 0
	var err error

	log.Println("CreateInventory-customerID", customerID)
	if req.CustomerID != nil {
		customerInvCreatedThisMonthNumber, err = s.repo.GetCustomerInventoryCreatedThisMonth(ctx, tx, *customerID, childSpan)
	}

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

		createdInventoryIDs := make([]uint, 0)
		createdInventoryIDs = append(createdInventoryIDs, inventory.ID)

		tx, err = s.updateRefQtyInOut(ctx, req, createdInventoryIDs, userID, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, nil, err
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

	oldInventory, err := s.GetInventoryByID(ctx, &dtos.GetInventoryParams{ID: *req.ID}, tx, childSpan)

	log.Println("UpdateInventory-oldInventory", oldInventory.TotalQty, *oldInventory.IngoingAt)
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

	tx, err = s.updateRefReverseQtyInOut(ctx, oldInvDts, createdInventoryIDs, userID, tx, childSpan)
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

	tx, err = s.updateRefQtyInOut(ctx, req, createdInventoryIDs, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	newInv, err := s.GetInventoryByID(ctx, &dtos.GetInventoryParams{ID: *req.ID}, tx, childSpan)

	if oldInventory.TotalQty != newInv.TotalQty || *oldInventory.IngoingAt != *newInv.IngoingAt {
		log.Println("UpdateInventory-PublishSyncCreateStockClosings", oldInventory.TotalQty, newInv.TotalQty, *oldInventory.IngoingAt, *newInv.IngoingAt)
		err = s.PublishSyncCreateStockClosings(ctx, tx, oldInventory, newInv, userID, branchID, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
		}
	}

	return &inventory, nil

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

func (s *InventoryService) UpdateResetStocksOnUpdateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, oldInvDts []dtos.InventoryInvDtListDTO, createdInventoryIDs []uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-UpdateResetStocksOnUpdateInventory", opentracing.ChildOf(span.Context()))

	if req.IoType == "INVENTORY_OUT" {
		err := s.repo.ResetCreateOrUpdateStockOut(ctx, tx, req, oldInvDts, childSpan)
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
	}

	return tx, nil
}

func (s *InventoryService) updateRefReverseQtyInOut(ctx *fiber.Ctx, oldInvDts []dtos.InventoryInvDtListDTO, createdInventoryIDs []uint, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-updateRefReverseQtyInOut", opentracing.ChildOf(span.Context()))

	refSoDtID, refSoDtBomID, refRoDtID, refPoDtID, refPoDtBomID, refInvDtID, err := utils.MapOldUpdateInvDts(ctx, oldInvDts, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	log.Println("updateRefReverseQtyInOut-refSoDtID", refSoDtID, "refSoDtBomID", refSoDtBomID, "refPoDtID", refPoDtID, "refPoDtBomID", refPoDtBomID, "refInvDtID", refInvDtID)

	var soDt []map[string]interface{}
	var soDtBom []map[string]interface{}
	var roDt []map[string]interface{}
	var poDt []map[string]interface{}
	var poDtBom []map[string]interface{}
	var invDt []map[string]interface{}

	if len(refSoDtID) > 0 {
		soDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "so_dts", "sales_order_id", refSoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refSoDtBomID) > 0 {
		soDtBom, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "so_dt_boms", "sales_order_id", refSoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtID) > 0 {
		roDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "request_order_dts", "request_order_id", refRoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtID) > 0 {
		poDt, err = s.repo.GetRefInDtByRefDtID(ctx, tx, "purchase_order_dts", "po_id", refPoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtBomID) > 0 {
		poDtBom, err = s.repo.GetRefInDtByRefDtID(ctx, tx, "po_dt_boms", "po_id", refPoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refInvDtID) > 0 {
		invDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "inv_dts", "inventory_id", refInvDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	newRefSoDt, newRefSoDtBom, newRefRoDt, newRefPoDt, newRefPoDtBom, newRefInvDt, err := utils.MapNewUpdatedReverseRefs(ctx, oldInvDts, soDt, soDtBom, roDt, poDt, poDtBom, invDt)

	log.Println("updateRefReverseQtyInOut-soDt", soDt, "soDtBom", soDtBom, "poDt", poDt, "poDtBom", poDtBom, "invDt", invDt)
	log.Println("updateRefReverseQtyInOut-newRef-soDt", newRefSoDt, "newRef-soDtBom", newRefSoDtBom, "newRef-poDt", newRefPoDt, "newRef-poDtBom", newRefPoDtBom, "newRef-invDt", newRefInvDt)

	if len(newRefSoDt) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefSoDt, "so_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefSoDtBom) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefSoDtBom, "so_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefRoDt) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefRoDt, "request_order_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefPoDt) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefPoDt, "purchase_order_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		poHead, err := s.repo.GetRefInHeadByRefDtID(ctx, tx, "purchase_order_dts", "purchase_orders", "po_id", refPoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		mappedPoHead := utils.MapInvRefHeadUpdateStatus(poHead, "po")

		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, mappedPoHead, "purchase_orders", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefPoDtBom) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefPoDtBom, "po_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefInvDt) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefInvDt, "inv_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *InventoryService) updateRefQtyInOut(ctx *fiber.Ctx, req dtos.FormInventoryRequest, createdInventoryIDs []uint, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-updateRefQtyInOut", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch InvDts
	refSoDtID, refSoDtBomID, refRoDtID, refPoDtID, refPoDtBomID, refInvDtID, err := utils.MapNewUpdateInvDts(ctx, req, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	log.Println("updateRefQtyInOut-refSoDtID", refSoDtID, "refSoDtBomID", refSoDtBomID, "refRoDtID", refRoDtID, "refPoDtID", refPoDtID, "refPoDtBomID", refPoDtBomID, "refInvDtID", refInvDtID)

	var soDt []map[string]interface{}
	var soDtBom []map[string]interface{}
	var roDt []map[string]interface{}
	var poDt []map[string]interface{}
	var poDtBom []map[string]interface{}
	var invDt []map[string]interface{}

	if len(refSoDtID) > 0 {
		soDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "so_dts", "sales_order_id", refSoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refSoDtBomID) > 0 {
		soDtBom, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "so_dt_boms", "sales_order_id", refSoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtID) > 0 {
		roDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "request_order_dts", "request_order_id", refRoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtID) > 0 {
		poDt, err = s.repo.GetRefInDtByRefDtID(ctx, tx, "purchase_order_dts", "po_id", refPoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refPoDtBomID) > 0 {
		poDtBom, err = s.repo.GetRefInDtByRefDtID(ctx, tx, "po_dt_boms", "po_id", refPoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refInvDtID) > 0 {
		invDt, err = s.repo.GetRefOutDtByRefDtID(ctx, tx, "inv_dts", "inventory_id", refInvDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	newRefSoDt, newRefSoDtBom, newRefRoDt, newRefPoDt, newRefPoDtBom, newRefInvDt, err := utils.MapNewUpdatedRefs(ctx, req, soDt, soDtBom, roDt, poDt, poDtBom, invDt)

	log.Println("updateRefQtyInOut-newRef-soDt", newRefSoDt, "newRef-soDtBom", newRefSoDtBom, "newRef-roDt", newRefRoDt, "newRef-poDt", newRefPoDt, "newRef-poDtBom", newRefPoDtBom, "newRef-invDt", newRefInvDt)

	if len(newRefSoDt) > 0 {
		log.Println("updateRefQtyInOut-newRefSoDt>0", newRefSoDt)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefSoDt, "so_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefSoDtBom) > 0 {
		log.Println("updateRefQtyInOut-newRefSoDtBom>0", newRefSoDtBom)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefSoDtBom, "so_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefRoDt) > 0 {
		log.Println("updateRefQtyInOut-newRefRoDt>0", newRefRoDt)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefRoDt, "request_order_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefPoDt) > 0 {
		log.Println("updateRefQtyInOut-newRefPoDt>0", newRefPoDt)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefPoDt, "purchase_order_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		poHead, err := s.repo.GetRefInHeadByRefDtID(ctx, tx, "purchase_order_dts", "purchase_orders", "po_id", refPoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}

		mappedPoHead := utils.MapInvRefHeadUpdateStatus(poHead, "po")

		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, mappedPoHead, "purchase_orders", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefPoDtBom) > 0 {
		log.Println("updateRefQtyInOut-newRefPoDtBom>0", newRefPoDtBom)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefPoDtBom, "po_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(newRefInvDt) > 0 {
		log.Println("updateRefQtyInOut-newRefInvDt>0", newRefInvDt)
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefInvDt, "inv_dts", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
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
func (s *InventoryService) BulkCreateUpdateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, updatedInventory *models.Inventory, userID uint, soDts []models.InvDt, inventoryID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
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
			log.Println("BulkCreateUpdateInvDts-invDt.ID == 0", invDt.ID)
			invDt.CreatedByID = &userID
			invDt.CreatedAt = time.Now()
			bulkCreateInvDts = append(bulkCreateInvDts, invDt)
		} else {
			log.Println("BulkCreateUpdateInvDts-invDt.ID != 0", invDt.ID)
			bulkUpdateInvDts = append(bulkUpdateInvDts, invDt)
			invDtIDs = append(invDtIDs, invDt.ID)
		}
	}

	// delete invDts that are not in the list
	if len(invDtIDs) > 0 {
		if tx, err := s.repo.DeleteInvDtsWhereNotIn(ctx, tx, inventoryID, invDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreateInvDts) > 0 {
		if tx, err := s.repo.CreateInvDts(tx, bulkCreateInvDts, inventoryID, childSpan); err != nil {
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

func (s *InventoryService) DeleteInvDtsByInventoryID(ctx *fiber.Ctx, params *dtos.GetInventoryParams, tx *gorm.DB, inventory *dtos.InventoryDetailDTO, userID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-DeleteInvDtsByInventoryID", opentracing.ChildOf(span.Context()))

	// Get all inv dts by inventory IDs
	inventoryIDs := make([]uint, 0)
	inventoryIDs = append(inventoryIDs, inventory.ID)

	oldInvDts, err := s.repo.GetInvDtsByInventoryIDs(ctx, tx, inventoryIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	tx, err = s.updateRefReverseQtyInOut(ctx, oldInvDts, inventoryIDs, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	var req dtos.FormInventoryRequest
	req.WarehouseID = *inventory.WarehouseID
	req.IoType = *inventory.IoType

	tx, err = s.UpdateResetStocksOnUpdateInventory(ctx, req, oldInvDts, inventoryIDs, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

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

func (s *InventoryService) GetRefIndexRoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexRoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetRefIndexRoDts", opentracing.ChildOf(span.Context()))

	soDts, total, err := s.repo.GetRefIndexRoDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return soDts, total, nil
}

func (s *InventoryService) GetRefIndexPoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexPoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetRefIndexPoDts", opentracing.ChildOf(span.Context()))

	soDts, total, err := s.repo.GetRefIndexPoDts(ctx, filters, childSpan)
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

func (s *InventoryService) GetRefIndexInvDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInvIndexInvDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetRefIndexInvDts", opentracing.ChildOf(span.Context()))

	soDts, total, err := s.repo.GetRefIndexInvDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	return soDts, total, nil
}

func (s *InventoryService) GetStockClosings(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.StockClosingListDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetStockClosings", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	inventories, total, err := s.repo.GetStockClosings(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return inventories, total, nil
}

func (s *InventoryService) CreateStockClosings(ctx *fiber.Ctx, tx *gorm.DB, date string, userID uint, branchID uint, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("InventoryService-CreateStockClosings", opentracing.ChildOf(span.Context()))

	// delete closing by date
	if err := s.repo.DeleteStockClosingByDate(ctx, tx, date, userID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	stockClosings, err := s.repo.CreateStockClosings(ctx, tx, date, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return stockClosings, nil
}

// func (s *InventoryService) BulkCreateStockClosingsByRange(ctx *fiber.Ctx, tx *gorm.DB, req dtos.FormClosingStockStoreRequest, span opentracing.Span) (*gorm.DB, error) {
// 	childSpan := opentracing.StartSpan("InventoryService-BulkCreateStockClosingsByRange", opentracing.ChildOf(span.Context()))

// 	stockClosings, err := s.repo.CreateStockClosings(ctx, tx, req, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, err
// 	}

// 	return stockClosings, nil
// }

// func (s *InventoryService) BackgroundCreateStockClosings(ctx *fiber.Ctx, tx *gorm.DB, date string, span opentracing.Span) ([]dtos.RefInvIndexInvDtListDTO, error) {
// 	childSpan := opentracing.StartSpan("InventoryService-BackgroundCreateStockClosings", opentracing.ChildOf(span.Context()))

// 	// stockClosings, err := s.repo.CreateStockClosings(ctx, date, childSpan)
// 	// if err != nil {
// 	// 	defer childSpan.Finish()
// 	// 	return nil, err
// 	// }

// 	// check last stock closing by date
// 	stockClosings, err := s.repo.GetStockClosingByDate(ctx, date, childSpan)
// 	if err != nil {
// 		defer childSpan.Finish()
// 		return nil, err
// 	}

// 	// backfillMissingClosings
// 	return stockClosings, nil
// }

func (s *InventoryService) GetInventoriesStatus(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InventoryStatusDtDTO, int, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetInventoriesStatus", opentracing.ChildOf(span.Context()))

	items, total, err := s.repo.GetInventoriesStatus(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	var params dtos.GetInventoriesStatusDtParams
	params.ItemIDs = utils.GetInventoryDtItemIDs(items)

	invDt, _, err := s.repo.GetInventoriesStatusDt(ctx, filters, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	mappedInvDt := utils.MapInventoryStatusList(items, invDt)

	return mappedInvDt, total, nil
}

func (s *InventoryService) PublishSyncCreateStockClosings(ctx *fiber.Ctx, tx *gorm.DB, oldInv *dtos.InventoryDetailDTO, newInv *dtos.InventoryDetailDTO, userID uint, branchID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-SyncCreateStockClosings", opentracing.ChildOf(span.Context()))

	var req dtos.SyncStockInventoryRequest
	req.StartClosingAt = oldInv.IngoingAt
	req.EndClosingAt = *newInv.IngoingAt
	req.Password = os.Getenv("RABBITMQ_PASSWORD")
	req.UserID = userID
	req.BranchID = branchID

	err := utils.PublishSyncCreateStockClosings(ctx, s.rabbitmq, req)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
	}

	return nil
}

func (s *InventoryService) BackgroundSyncCreateStockClosingsByRangeDate(ctx *fiber.Ctx, tx *gorm.DB, params dtos.FormClosingStockStoreRequest, userID uint, branchID uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InventoryService-BackgroundSyncCreateStockClosingsByRangeDate", opentracing.ChildOf(span.Context()))

	var req dtos.SyncStockInventoryRequest
	req.StartClosingAt = params.StartClosingAt
	req.EndClosingAt = params.EndClosingAt
	req.Password = params.Password
	req.UserID = userID
	req.BranchID = branchID

	err := utils.PublishSyncCreateStockClosings(ctx, s.rabbitmq, req)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
	}

	return nil
}

func (s *InventoryService) ConsumeSyncStock(req dtos.SyncStockInventoryRequest) error {
	parentSpan := opentracing.StartSpan("InventoryService-ConsumeSyncStock")

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	tx := s.repo.BeginTransaction()

	log.Printf("[INFO] Processing ConsumeSyncStock: %v", req)

	filters := make(map[string]string)

	latestInv, err := s.repo.GetLatestInventory(ctx, tx, filters, parentSpan)
	if err != nil {
		log.Printf("[ERROR] Failed to get latest inventory: %v", err)
		utils.LogErrors(parentSpan, err)
		tx.Rollback()
	}

	log.Println("UpdateInventory-latestInv", *latestInv.IngoingAt)
	dates, err := utils.GetDatesBetween(req, latestInv.IngoingAt)
	if err != nil {
		log.Printf("[ERROR] Failed to get dates between: %v", err)
		utils.LogErrors(parentSpan, err)
		tx.Rollback()
	}

	for _, date := range dates {
		log.Printf("[INFO] Processing date: %s", date)
		stockClosings, err := s.CreateStockClosings(ctx, tx, date, req.UserID, req.BranchID, parentSpan)
		if err != nil {
			log.Printf("[ERROR] Failed to create stock closings: %v", err)
			tx.Rollback()
		}
		log.Printf("[INFO] Created stock closings: %v", stockClosings)
	}

	tx.Commit()

	return nil
}

// func (s *InventoryService) ConsumeSyncStock(req dtos.SyncStockInventoryRequest) error {
// 	app := fiber.New()
// 	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
// 	defer app.ReleaseCtx(ctx)
// 	ctx.Locals("user_id", req.UserID)
// 	ctx.Locals("branch_id", req.BranchID)

// 	tx := s.repo.BeginTransaction()

// 	log.Printf("[INFO] Processing ConsumeSyncStock: %v", req)
// 	dates := utils.GetDatesBetween(req)

// 	for _, date := range dates {
// 		log.Printf("[INFO] Processing date: %s", date)
// 		stockClosings, err := s.CreateStockClosings(ctx, tx, date, nil)
// 		if err != nil {
// 			log.Printf("[ERROR] Failed to create stock closings: %v", err)
// 			tx.Rollback()
// 		}
// 		log.Printf("[INFO] Created stock closings: %v", stockClosings)
// 	}

// 	tx.Commit()

// 	return nil
// }

func (s *InventoryService) GetLatestInventory(ctx *fiber.Ctx, tx *gorm.DB, filters map[string]string, span opentracing.Span) (*dtos.InventoryListDTO, error) {
	childSpan := opentracing.StartSpan("InventoryService-GetLatestInventory", opentracing.ChildOf(span.Context()))

	inventories, err := s.repo.GetLatestInventory(ctx, tx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return inventories, nil
}

// PdfGetQuotations
func (s *InventoryService) Pdf(ctx *fiber.Ctx, req dtos.InventoryDetailDTO, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("InventoryService-Pdf", opentracing.ChildOf(span.Context()))

	form := req
	var inventory *dtos.InventoryDetailDTO
	var err error

	// req.IsIDOnly != nil
	var params dtos.GetInventoryParams
	if req.IsIDOnly != nil && *req.IsIDOnly == 1 {
		params.ID = req.ID

		inventory, err = s.repo.GetInventoryByID(ctx, &params, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			return nil, err
		}

		createdInventoryIDs := make([]uint, 0)
		createdInventoryIDs = append(createdInventoryIDs, inventory.ID)

		invDts, err := s.GetInvDtsByInventoryIDs(ctx, tx, createdInventoryIDs, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch invDts: %v", err)
		}

		companyParams := &dtos.GetCompanyProfileParams{ID: uint(*inventory.CompanyProfileID)}
		inventory.InvDts = invDts
		company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch company: %v", err)
		}

		inventory.Company = *company

		form = *inventory
		req.SuratJalanNo = inventory.SuratJalanNo
	}

	var num string
	if req.SuratJalanNo != nil {
		num = *req.SuratJalanNo
	} else {
		num = ""
	}

	data := dtos.InventoryPDFData{
		Num:  num,
		Form: form,
	}

	htmlFileName := "inventory-detail"
	log.Println("Pdf-htmlFileName-so", htmlFileName)

	// 2. Render HTML template with data
	// templateFile, err := templateFS.Open("templates/sales-order-detail.html")
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

	// htmlFile, err := os.CreateTemp("", "sales-order-detail-*.html")
	htmlFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error writing htmlFile:", err)
		return nil, err
	}
	defer os.Remove(htmlFile.Name())

	// Inside the Pdf function, before parsing the template
	funcMap := template.FuncMap{
		"formatNumber": func(n float64, args ...int) string {
			decimals := 2
			if len(args) > 0 {
				decimals = args[0]
			}

			format := fmt.Sprintf("%%.%df", decimals)
			p := message.NewPrinter(language.English)
			return p.Sprintf(format, n)
			// Handle different numeric types
			// switch v := n.(type) {
			// case float64:
			// 	return p.Sprintf(format, v)
			// case float32:
			// 	return p.Sprintf(format, v)
			// case int:
			// 	return p.Sprintf(format, float64(v))
			// case nil:
			// 	return "0"
			// default:
			// 	return "0"
			// }
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

	// Parse the template
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
	page.FooterSpacing.Set(10)                  // Space below content (mm)

	pdfg.AddPage(page)
	// pdfg.MarginBottom.Set(0)
	// pdfg.MarginTop.Set(0)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	// pdfg.Dpi.Set(300)

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
	fileName := fmt.Sprintf("so-%s.pdf", time.Now().Format("20060102150405"))
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
