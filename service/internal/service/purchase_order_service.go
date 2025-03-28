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

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx *fiber.Ctx, req dtos.CreatePurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, error) {
	childSpan := opentracing.StartSpan("PurchaseOrderService-CreatePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	purchaseOrder, err := utils.MapCreatePurchaseOrder(ctx, req, userID, branchID, childSpan)
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

	return createdPurchaseOrder, nil
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

func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx *fiber.Ctx, req dtos.UpdatePurchaseOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.PurchaseOrder, error) {
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

	if err := s.repo.DeletePoDtsWhereNotIn(ctx, tx, req.ID, poDtIDs, childSpan); err != nil {
		return nil, err
	}

	poDts, err := utils.MapUpdatePoDts(ctx, req, &purchaseOrder, userID, childSpan)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePoDts(tx, poDts, childSpan); err != nil {
		return nil, err
	}

	return &purchaseOrder, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx *fiber.Ctx, id uint, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("PurchaseOrderService-DeletePurchaseOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

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
