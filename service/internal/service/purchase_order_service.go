package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	repo     *repository.PurchaseOrderRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewPurchaseOrderService(repo *repository.PurchaseOrderRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *PurchaseOrderService {
	return &PurchaseOrderService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *PurchaseOrderService) GetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPurchaseOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	purchaseOrders, total, err := s.repo.GetPurchaseOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return purchaseOrders, total, nil
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	customerSoCreatedThisMonthNumber, err := s.repo.GetCustomerPurchaseOrderCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	purchaseOrder, err := utils.MapCreatePurchaseOrder(ctx, req, userID, branchID, customerSoCreatedThisMonthNumber, childSpan)
	if err != nil {
		return nil, err
	}

	createdPurchaseOrder, err := s.repo.CreatePurchaseOrder(tx, &purchaseOrder, childSpan)
	if err != nil {
		return nil, err
	}

	poDts, err := utils.MapCreatePoDts(ctx, req, createdPurchaseOrder, userID, childSpan)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.CreatePoDts(tx, poDts, childSpan)
	if err != nil {
		return nil, err
	}

	tx, err = s.updateRefQtyInOut(ctx, req, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return createdPurchaseOrder, nil
}

func (s *PurchaseOrderService) updateRefQtyInOut(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-updateRefQtyInOut", opentracing.ChildOf(span.Context()))

	// Bulk/Create Update Batch InvDts
	refSoDtID, refSoDtBomID, refRoDtID, refRoDtBomID, err := utils.MapNewUpdatePoDts(ctx, req, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	var soDt []map[string]interface{}
	var soDtBom []map[string]interface{}
	// var poDt []map[string]interface{}
	var poDtBom []map[string]interface{}
	var roDt []map[string]interface{}

	if len(refSoDtID) > 0 {
		soDt, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "so_dts", "sales_order_id", refSoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refSoDtBomID) > 0 {
		soDtBom, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "so_dt_boms", "sales_order_id", refSoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtID) > 0 {
		roDt, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "request_order_dts", "request_order_id", refRoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtBomID) > 0 {
		poDtBom, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "ro_dt_boms", "request_order_id", refRoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	newRefSoDt, newRefSoDtBom, newRefRoDt, newRefRoDtBom, err := utils.MapNewUpdatedRefsPo(ctx, req, soDt, soDtBom, roDt, poDtBom)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

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

	if len(newRefRoDtBom) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefRoDtBom, "po_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	// err = s.repo.ResetCreateOrUpdateStockOut(ctx, tx, req, oldPoDts, childSpan)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	tx.Rollback()
	// 	return nil, err
	// }

	return tx, nil
}

func (s *PurchaseOrderService) GetPurchaseOrderByID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.PurchaseOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPurchaseOrderByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	purchaseOrder, err := s.repo.GetPurchaseOrderByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}
	return purchaseOrder, nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-UpdatePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	purchaseOrder, err := utils.MapUpdatePurchaseOrder(ctx, req, userID, branchID, childSpan)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePurchaseOrder(tx, &purchaseOrder, childSpan); err != nil {
		return nil, err
	}

	poDtIDs, _ := utils.GetPoIDs(req)

	// params *dtos.GetPurchaseOrderPoDtParams
	isDeleted := 0
	params := &dtos.GetPurchaseOrderPoDtParams{
		PurchaseOrderID: purchaseOrder.ID,
		IsDeleted:       &isDeleted,
	}

	oldPoDts, err := s.repo.GetPurchaseOrderPoDts(ctx, tx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	tx, err = s.updateRefReverseQtyInOut(ctx, oldPoDts, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.DeletePoDtsWhereNotIn(ctx, tx, *req.ID, poDtIDs, childSpan); err != nil {
		return nil, err
	}

	poDts, err := utils.MapUpdatePoDts(ctx, req, &purchaseOrder, userID, childSpan)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePoDts(tx, poDts, childSpan); err != nil {
		return nil, err
	}

	tx, err = s.updateRefQtyInOut(ctx, req, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return &purchaseOrder, nil
}

func (s *PurchaseOrderService) updateRefReverseQtyInOut(ctx *fiber.Ctx, oldPoDts []dtos.PurchaseOrderPoDtListDTO, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-updateRefReverseQtyInOut", opentracing.ChildOf(span.Context()))

	refSoDtID, refSoDtBomID, refRoDtID, refRoDtBomID, err := utils.MapOldUpdatePoDts(ctx, oldPoDts, userID, span)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	var soDt []map[string]interface{}
	var soDtBom []map[string]interface{}
	var roDt []map[string]interface{}
	var poDtBom []map[string]interface{}

	if len(refSoDtID) > 0 {
		soDt, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "so_dts", "sales_order_id", refSoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refSoDtBomID) > 0 {
		soDtBom, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "so_dt_boms", "sales_order_id", refSoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtID) > 0 {
		roDt, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "request_order_dts", "request_order_id", refRoDtID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(refRoDtBomID) > 0 {
		poDtBom, err = s.repo.GetRefPoDtByRefDtID(ctx, tx, "ro_dt_boms", "request_order_id", refRoDtBomID, childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	newRefSoDt, newRefSoDtBom, newRefRoDt, newRefRoDtBom, err := utils.MapNewUpdatedReverseRefsPo(ctx, oldPoDts, soDt, soDtBom, roDt, poDtBom)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

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

	if len(newRefRoDtBom) > 0 {
		err = s.repo.BulkUpdateReverseInvRefDtsQty(ctx, tx, newRefRoDtBom, "po_dt_boms", childSpan)
		if err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx *fiber.Ctx, tx *gorm.DB, id uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-DeletePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	claims := utils.GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	isDeleted := 0
	params := &dtos.GetPurchaseOrderPoDtParams{
		PurchaseOrderID: id,
		IsDeleted:       &isDeleted,
	}

	oldPoDts, err := s.repo.GetPurchaseOrderPoDts(ctx, tx, params, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	tx, err = s.updateRefReverseQtyInOut(ctx, oldPoDts, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	if err := s.repo.DeletePurchaseOrder(ctx, id, childSpan); err != nil {
		return err
	}

	return nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrderStatus(ctx *fiber.Ctx, id uint, status string, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-UpdatePurchaseOrderStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := s.repo.UpdatePurchaseOrderStatus(ctx, id, status, childSpan); err != nil {
		return err
	}

	return nil
}

func (s *PurchaseOrderService) RestorePurchaseOrder(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-RestorePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := s.repo.RestorePurchaseOrder(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseOrderService) GetWidgetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetWidgetPurchaseOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	purchaseOrders, total, err := s.repo.GetWidgetPurchaseOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return purchaseOrders, total, nil
}

func (s *PurchaseOrderService) GetRefIndexSoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefPoIndexSoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetRefIndexSoDts", opentracing.ChildOf(span.Context()))

	soDts, total, err := s.repo.GetRefIndexSoDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	return soDts, total, nil
}

func (s *PurchaseOrderService) GetRefIndexRoDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefPoIndexRoDtListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetRefIndexRoDts", opentracing.ChildOf(span.Context()))

	roDts, total, err := s.repo.GetRefIndexRoDts(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	return roDts, total, nil
}
