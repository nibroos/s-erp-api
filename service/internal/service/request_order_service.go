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
	"gorm.io/gorm"
)

type RequestOrderService struct {
	repo     *repository.RequestOrderRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewRequestOrderService(repo *repository.RequestOrderRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *RequestOrderService {
	return &RequestOrderService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *RequestOrderService) GetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRequestOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrders, total, err := s.repo.GetRequestOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return requestOrders, total, nil
}

func (s *RequestOrderService) CreateRequestOrder(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.RequestOrder, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CreateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderCreatedThisMonthNumber, err := s.repo.GetRequestOrderCreatedThisMonth(ctx, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := requestOrderCreatedThisMonthNumber + 1

	requestOrder, err := utils.MapCreateRequestOrder(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateRequestOrder(tx, &requestOrder, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, _, err = s.CreateRequestOrderDts(ctx, req, userID, &requestOrder, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	return &requestOrder, tx, nil
}

func (s *RequestOrderService) GetRequestOrderByID(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) (*dtos.RequestOrderDetailDTO, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRequestOrderByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrder, err := s.repo.GetRequestOrderByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	requestOrderDts, err := s.repo.GetRequestOrderDts(ctx, requestOrder.ID, params.IsDeleted, childSpan)
	if err != nil {
		return nil, err
	}

	requestOrder.RequestOrderDts = requestOrderDts

	return requestOrder, nil
}

func (s *RequestOrderService) UpdateRequestOrder(ctx *fiber.Ctx, req dtos.UpdateRequestOrderRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.RequestOrder, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-UpdateRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetRequestOrderParams{ID: req.ID}
	existingRequestOrder, err := s.GetRequestOrderByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.repo.LockRequestOrder(tx, req.ID, childSpan); err != nil {
		tx.Rollback()
		return nil, err
	}

	requestOrder, err := utils.MapUpdateRequestOrder(ctx, req, userID, branchID, existingRequestOrder.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.UpdateRequestOrder(tx, &requestOrder, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	existingDtMap := make(map[string]*dtos.RequestOrderDtListDTO)
	for i, dt := range existingRequestOrder.RequestOrderDts {
		if dt.ID != nil {
			key := fmt.Sprintf("%d-%d-%d",
				utils.GetValueOrDefault(dt.ProductID, 0),
				utils.GetValueOrDefault(dt.RefID, 0),
				utils.GetValueOrDefault(dt.ItemID, 0))
			existingDtMap[key] = &existingRequestOrder.RequestOrderDts[i]
		}
	}

	requestOrderDts, err := utils.MapUpdateRequestOrderDts(ctx, req, &requestOrder, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var createRequestOrderDts []models.RequestOrderDt
	var updateRequestOrderDts []models.RequestOrderDt
	var deleteRequestOrderDtIDs []uint

	processedExistingDts := make(map[uint]bool)

	for i, dt := range requestOrderDts {
		key := fmt.Sprintf("%d-%d-%d",
			utils.GetValueOrDefault(dt.ProductID, 0),
			utils.GetValueOrDefault(dt.RefID, 0),
			utils.GetValueOrDefault(dt.ItemID, 0))

		if existingDt, exists := existingDtMap[key]; exists {
			requestOrderDts[i].ID = *existingDt.ID
			requestOrderDts[i].UpdatedByID = &userID
			updateRequestOrderDts = append(updateRequestOrderDts, requestOrderDts[i])
			processedExistingDts[*existingDt.ID] = true
		} else {
			requestOrderDts[i].CreatedByID = &userID
			requestOrderDts[i].CreatedAt = time.Now()
			createRequestOrderDts = append(createRequestOrderDts, requestOrderDts[i])
		}
	}

	for _, dt := range existingRequestOrder.RequestOrderDts {
		if dt.ID != nil && !processedExistingDts[*dt.ID] {
			deleteRequestOrderDtIDs = append(deleteRequestOrderDtIDs, *dt.ID)
		}
	}

	if len(deleteRequestOrderDtIDs) > 0 {
		tx, err = s.repo.DeleteRequestOrderDtsByIDs(tx, deleteRequestOrderDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createRequestOrderDts) > 0 {
		tx, err = s.repo.BulkCreateRequestOrderDts(tx, createRequestOrderDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(updateRequestOrderDts) > 0 {
		tx, err = s.repo.BulkUpdateRequestOrderDts(tx, updateRequestOrderDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &requestOrder, nil
}

func (s *RequestOrderService) DeleteRequestOrder(ctx *fiber.Ctx, requestOrderID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderService-DeleteRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetRequestOrderParams{ID: requestOrderID}
	_, err := s.GetRequestOrderByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.repo.LockRequestOrder(tx, requestOrderID, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	tx, err = s.repo.DeleteRequestOrder(tx, requestOrderID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *RequestOrderService) RestoreRequestOrder(ctx *fiber.Ctx, params *dtos.GetRequestOrderParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("RequestOrderService-RestoreRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	isDeleted := 1
	getParams := &dtos.GetRequestOrderParams{
		ID:        params.ID,
		IsDeleted: &isDeleted,
	}

	_, err := s.GetRequestOrderByID(ctx, getParams, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.repo.RestoreRequestOrder(ctx, params, tx, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *RequestOrderService) CreateRequestOrderDts(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, userID uint, createdRequestOrder *models.RequestOrder, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.RequestOrderDt, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-CreateRequestOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrderDts, err := utils.MapCreateRequestOrderDts(ctx, req, createdRequestOrder, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx, createdRequestOrderDts, err := s.repo.CreateRequestOrderDts(tx, requestOrderDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	return tx, createdRequestOrderDts, nil
}

func (s *RequestOrderService) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDts, total, err := s.repo.GetRefSalesOrderDts(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	return soDts, total, nil
}

func (s *RequestOrderService) GetRefProductForRequestOrder(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefProductForRequestOrderListDTO, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetRefProductForRequestOrder", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	products, total, err := s.repo.GetRefProductForRequestOrder(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (s *RequestOrderService) GetWidgetRequestOrders(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RequestOrderStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("RequestOrderService-GetWidgetRequestOrders", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	requestOrders, total, err := s.repo.GetWidgetRequestOrders(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return requestOrders, total, nil
}

func (s *RequestOrderService) BeginTransaction() *gorm.DB {
	return s.repo.BeginTransaction()
}

func (s *RequestOrderService) Commit(tx *gorm.DB) error {
	return s.repo.Commit(tx)
}

func (s *RequestOrderService) Rollback(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}
