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

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))

	purchaseOrder, err := s.MapCreatePurchaseOrder(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, tx, err
	}

	if tx, err := s.repo.CreatePurchaseOrder(tx, &purchaseOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	var poDts []models.PurchaseOrderDt
	tx, poDts, err = s.CreatePoDts(ctx, req, userID, &purchaseOrder, tx, childSpan)

	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.CreatePoDtBoms(ctx, poDts, req, &purchaseOrder, userID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	return &purchaseOrder, tx, nil
}

func (s *PurchaseOrderService) MapCreatePurchaseOrder(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, userID uint, branchID uint, span opentracing.Span) (models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-MapCreatePurchaseOrder", opentracing.ChildOf(span.Context()))

	purchaseOrderModel, err := utils.MapCreatePurchaseOrder(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return purchaseOrderModel, err
	}

	return purchaseOrderModel, nil
}

func (s *PurchaseOrderService) CreatePoDts(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, userID uint, createdPurchaseOrder *models.PurchaseOrder, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.PurchaseOrderDt, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePoDts", opentracing.ChildOf(span.Context()))

	poDts, err := s.MapCreatePoDts(ctx, req, createdPurchaseOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, poDts, err
	}

	poDtsModel := []models.PurchaseOrderDt{}

	tx, poDtsModel, err = s.repo.CreatePoDts(tx, poDts, createdPurchaseOrder.ID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, poDtsModel, err
	}

	return tx, poDtsModel, nil
}

func (s *PurchaseOrderService) MapCreatePoDts(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, createdPurchaseOrder *models.PurchaseOrder, userID uint, span opentracing.Span) ([]models.PurchaseOrderDt, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-MapCreatePoDts", opentracing.ChildOf(span.Context()))

	poDtsModel, err := utils.MapCreatePoDts(ctx, req, createdPurchaseOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return poDtsModel, nil
}

func (s *PurchaseOrderService) MapCreatePoDtBoms(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, createdPoDts []models.PurchaseOrderDt, userID uint, span opentracing.Span) []map[string]interface{} {
	childSpan := opentracing.StartSpan("PurchaseOrderService-MapCreatePoDtBoms", opentracing.ChildOf(span.Context()))

	poDtBomsModel := utils.MapCreatePoDtBoms(ctx, req, createdPoDts, userID, childSpan)

	return poDtBomsModel
}

func (s *PurchaseOrderService) CreatePoDtBoms(ctx *fiber.Ctx, poDts []models.PurchaseOrderDt, req dtos.CreatePurchaseOrderRequest, createdPurchaseOrder *models.PurchaseOrder, userID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePoDtBoms", opentracing.ChildOf(span.Context()))

	poDtBoms := s.MapCreatePoDtBoms(ctx, req, poDts, userID, childSpan)

	if tx, err := s.repo.CreatePoDtBoms(tx, poDtBoms, createdPurchaseOrder.ID, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func (s *PurchaseOrderService) GetPurchaseOrderByID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.PurchaseOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPurchaseOrderByID", opentracing.ChildOf(span.Context()))

	purchaseOrder, err := s.repo.GetPurchaseOrderByID(ctx, params, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	purchaseOrderIDs := make([]uint, 0)
	purchaseOrderIDs = append(purchaseOrderIDs, purchaseOrder.ID)

	poDts, err := s.GetPoDtsByPurchaseOrderIDs(ctx, tx, purchaseOrderIDs, childSpan)
	if err != nil {
		utils.LogErrors(childSpan, err)
		defer childSpan.Finish()
		return nil, err
	}

	filters := ctx.Locals("filters").(map[string]string)

	poDtBoms, err := s.GetPoDtsBomByPurchaseOrders(ctx, filters, purchaseOrderIDs, childSpan)
	if err != nil {
		utils.LogErrors(childSpan, err)
		defer childSpan.Finish()
		return nil, err
	}

	poDts = s.MapFilterPoDtBomsToPoDts(ctx, poDtBoms, poDts, childSpan)

	purchaseOrder.PoDts = poDts

	return purchaseOrder, nil
}

func (s *PurchaseOrderService) GetPoDtsByPurchaseOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtListDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPoDtsByPurchaseOrderIDs", opentracing.ChildOf(span.Context()))

	poDts, err := s.repo.GetPoDtsByPurchaseOrderIDs(ctx, tx, purchaseOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return poDts, nil
}

func (s *PurchaseOrderService) GetPoDtsBomByPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtBomListDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPoDtsBomByPurchaseOrders", opentracing.ChildOf(span.Context()))

	poDtBoms, err := s.repo.GetPoDtsBomByPurchaseOrders(ctx, filters, purchaseOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return poDtBoms, nil
}

func (s *PurchaseOrderService) MapFilterPoDtBomsToPoDts(ctx *fiber.Ctx, poDtBoms []dtos.PurchaseOrderPoDtBomListDTO, poDts []dtos.PurchaseOrderPoDtListDTO, span opentracing.Span) []dtos.PurchaseOrderPoDtListDTO {
	// childSpan := opentracing.StartSpan("PurchaseOrderService-MapFilterPoDtBomsToPoDts", opentracing.ChildOf(span.Context()))

	return utils.MapFilterPoDtBomsToPoDts(poDtBoms, poDts)
}

func (s *PurchaseOrderService) GetPurchaseOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.PurchaseOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetPurchaseOrders", opentracing.ChildOf(span.Context()))

	purchaseOrders, total, err := s.repo.GetPurchaseOrders(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return purchaseOrders, total, nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx *fiber.Ctx, req dtos.UpdatePurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-UpdatePurchaseOrder", opentracing.ChildOf(span.Context()))

	purchaseOrder, err := utils.MapUpdatePurchaseOrder(ctx, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	if err := s.repo.UpdatePurchaseOrder(tx, &purchaseOrder, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	poDts, err := s.MapCreateUpdatePoDts(ctx, req, &purchaseOrder, userID, childSpan)

	tx, err = s.BulkCreateUpdatePoDts(ctx, req, &purchaseOrder, userID, poDts, purchaseOrder.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	updatedPurchaseOrderIDs := make([]uint, 0)
	updatedPurchaseOrderIDs = append(updatedPurchaseOrderIDs, purchaseOrder.ID)

	updatedPoDts, err := s.GetUpdatedPoDtsByPurchaseOrderIDs(ctx, tx, updatedPurchaseOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	err = s.BulkCreateUpdatePoDtBoms(ctx, updatedPoDts, req, purchaseOrder.ID, tx, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	return &purchaseOrder, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-DeletePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePurchaseOrder(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseOrderService) RestorePurchaseOrder(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-RestorePurchaseOrder", opentracing.ChildOf(span.Context()))

	if err := s.repo.RestorePurchaseOrder(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseOrderService) MapCreateUpdatePoDts(ctx *fiber.Ctx, req dtos.UpdatePurchaseOrderRequest, updatedPurchaseOrder *models.PurchaseOrder, userID uint, span opentracing.Span) ([]models.PurchaseOrderDt, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-MapCreateUpdatePoDts", opentracing.ChildOf(span.Context()))

	poDtsModel, err := utils.MapCreateUpdatePoDts(ctx, req, updatedPurchaseOrder, userID, span)

	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return poDtsModel, nil
}

func (s *PurchaseOrderService) GetUpdatedPoDtsByPurchaseOrderIDs(ctx *fiber.Ctx, tx *gorm.DB, purchaseOrderIDs []uint, span opentracing.Span) ([]dtos.PurchaseOrderPoDtListUpdateDTO, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-GetUpdatedPoDtsByPurchaseOrderIDs", opentracing.ChildOf(span.Context()))

	poDts, err := s.repo.GetUpdatedPoDtsByPurchaseOrderIDs(ctx, tx, purchaseOrderIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}
	return poDts, nil
}

func (s *PurchaseOrderService) BulkCreateUpdatePoDts(ctx *fiber.Ctx, req dtos.UpdatePurchaseOrderRequest, updatedPurchaseOrder *models.PurchaseOrder, userID uint, poDts []models.PurchaseOrderDt, purchaseOrderID uint, tx *gorm.DB, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-BulkCreateUpdatePoDts", opentracing.ChildOf(span.Context()))

	poDts, err := s.MapCreateUpdatePoDts(ctx, req, updatedPurchaseOrder, userID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, err
	}

	bulkCreatePoDts := []models.PurchaseOrderDt{}
	bulkUpdatePoDts := []models.PurchaseOrderDt{}
	poDtIDs := []uint{}

	for _, poDt := range poDts {
		if poDt.ID == 0 {
			poDt.CreatedByID = &userID
			poDt.CreatedAt = time.Now()
			bulkCreatePoDts = append(bulkCreatePoDts, poDt)
		} else {
			bulkUpdatePoDts = append(bulkUpdatePoDts, poDt)
			poDtIDs = append(poDtIDs, poDt.ID)
		}
	}

	if len(poDtIDs) > 0 {
		if tx, err := s.repo.DeletePoDtsWhereNotIn(ctx, tx, purchaseOrderID, poDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkCreatePoDts) > 0 {
		if tx, _, err := s.repo.CreatePoDts(tx, bulkCreatePoDts, purchaseOrderID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(bulkUpdatePoDts) > 0 {
		if tx, err := s.repo.UpdatePoDts(tx, bulkUpdatePoDts, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	return tx, nil
}

func (s *PurchaseOrderService) BulkCreateUpdatePoDtBoms(ctx *fiber.Ctx, poDts []dtos.PurchaseOrderPoDtListUpdateDTO, req dtos.UpdatePurchaseOrderRequest, purchaseOrderID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-BulkCreateUpdatePoDtBoms", opentracing.ChildOf(span.Context()))

	bulkCreatePoDtBoms, bulkUpdatePoDtBoms, poDtBomIDs, err := utils.MapFilterUpdatePoDtBomsToPoDts(ctx, poDts, req, purchaseOrderID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	if len(poDtBomIDs) > 0 {
		if err := s.repo.DeletePoDtBomsWhereNotIn(ctx, tx, purchaseOrderID, poDtBomIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkCreatePoDtBoms) > 0 {
		if tx, err := s.repo.CreatePoDtBoms(tx, bulkCreatePoDtBoms, purchaseOrderID, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	if len(bulkUpdatePoDtBoms) > 0 {
		if err := s.repo.UpdatePoDtBoms(tx, bulkUpdatePoDtBoms, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *PurchaseOrderService) DeletePoDtBomsByPurchaseOrderID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-DeletePoDtBomsByPurchaseOrderID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePoDtBomsByPurchaseOrderID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseOrderService) DeletePoDtsByPurchaseOrderID(ctx *fiber.Ctx, params *dtos.GetPurchaseOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-DeletePoDtsByPurchaseOrderID", opentracing.ChildOf(span.Context()))

	if err := s.repo.DeletePoDtsByPurchaseOrderID(tx, params, childSpan); err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return err
	}

	return nil
}

func (s *PurchaseOrderService) LockPurchaseOrderTable(ctx *fiber.Ctx, tx *gorm.DB, req dtos.UpdatePurchaseOrderRequest, span opentracing.Span) (*gorm.DB, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-LockPurchaseOrderTable", opentracing.ChildOf(span.Context()))

	if req.ID > 0 {
		if err := s.repo.LockPurchaseOrderHeader(ctx, tx, req, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	poDtIDs, poDtBomIDs, productIDs, itemUnitIDs := utils.GetPoIDs(req)

	if len(poDtIDs) > 0 {
		if err := s.repo.LockPoDts(ctx, tx, poDtIDs, childSpan); err != nil {
			defer childSpan.Finish()
			tx.Rollback()
			return nil, err
		}
	}

	if len(poDtBomIDs) > 0 {
		if err := s.repo.LockPoDtBoms(ctx, tx, poDtBomIDs, childSpan); err != nil {
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
