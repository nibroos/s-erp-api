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

type SalesInvoiceService struct {
	repo     *repository.SalesInvoiceRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewSalesInvoiceService(repo *repository.SalesInvoiceRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *SalesInvoiceService {
	return &SalesInvoiceService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *SalesInvoiceService) GetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetSalesInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoices, total, err := s.repo.GetSalesInvoices(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return salesInvoices, total, nil
}

func (s *SalesInvoiceService) CreateSalesInvoice(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesInvoice, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-CreateSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	uniqueSOIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)
	soDtIDs := make(map[uint]uint)

	for _, dt := range req.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && *dt.RefID > 0 {
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				uniqueSOIDs = append(uniqueSOIDs, *dt.RefID)
			}

			if dt.RefDtID != nil && *dt.RefDtID > 0 {
				soDtIDs[*dt.RefDtID] = *dt.RefID
			}
		}
	}

	if len(uniqueSOIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, uniqueSOIDs, childSpan); err != nil {
			tx.Rollback()
			return nil, tx, err
		}
	}

	salesInvoiceCreatedThisMonthNumber, err := s.repo.GetSalesInvoiceCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := salesInvoiceCreatedThisMonthNumber + 1

	salesInvoice, err := utils.MapCreateSalesInvoice(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateSalesInvoice(tx, &salesInvoice, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, _, err = s.CreateSalesInvoiceDts(ctx, req, userID, &salesInvoice, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	for soDtID, soID := range soDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, "INVOICE", childSpan)
		if err != nil {
			tx.Rollback()
			return nil, tx, err
		}

		tx, err = s.repo.CheckAndUpdateSalesOrderStatus(tx, soID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, tx, err
		}
	}

	return &salesInvoice, tx, nil
}

func (s *SalesInvoiceService) GetSalesInvoiceByID(ctx *fiber.Ctx, params *dtos.GetSalesInvoiceParams, tx *gorm.DB, span opentracing.Span) (*dtos.SalesInvoiceDetailDTO, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetSalesInvoiceByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoice, err := s.repo.GetSalesInvoiceByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	salesInvoiceDts, err := s.repo.GetSalesInvoiceDts(ctx, salesInvoice.ID, params.IsDeleted, childSpan)
	if err != nil {
		return nil, err
	}

	var soDtIDs []uint
	for _, dt := range salesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
			soDtIDs = append(soDtIDs, *dt.RefDtID)
		}
	}

	if len(soDtIDs) > 0 {
		soDtBoms, err := s.repo.GetSoDtBoms(ctx, soDtIDs, childSpan)
		if err != nil {
			return nil, err
		}

		for i, dt := range salesInvoiceDts {
			if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
				var dtBoms []dtos.SalesOrderSoDtBomListDTO
				for _, bom := range soDtBoms {
					if bom.SoDtID != nil && *bom.SoDtID == *dt.RefDtID {
						dtBoms = append(dtBoms, bom)
					}
				}
				salesInvoiceDts[i].SoDtsBoms = dtBoms
			}
		}
	}

	salesInvoice.SalesInvoiceDts = salesInvoiceDts

	return salesInvoice, nil
}

func (s *SalesInvoiceService) UpdateSalesInvoice(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesInvoice, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-UpdateSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	existingSoDtIDs := make(map[uint]uint)
	newSoDtIDs := make(map[uint]uint)
	allSOIDs := make([]uint, 0)
	uniqueSOIDs := make(map[uint]bool)

	params := dtos.GetSalesInvoiceParams{ID: req.ID}
	existingSalesInvoice, err := s.GetSalesInvoiceByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueSOIDs[*dt.RefID] {
				uniqueSOIDs[*dt.RefID] = true
				allSOIDs = append(allSOIDs, *dt.RefID)
			}
		}
	}

	for _, dt := range req.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			newSoDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueSOIDs[*dt.RefID] {
				uniqueSOIDs[*dt.RefID] = true
				allSOIDs = append(allSOIDs, *dt.RefID)
			}
		}
	}

	if len(allSOIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, allSOIDs, childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := s.repo.LockSalesInvoice(tx, req.ID, childSpan); err != nil {
		tx.Rollback()
		return nil, err
	}

	removedSoDtIDs := make(map[uint]uint)
	for soDtID, soID := range existingSoDtIDs {
		if _, exists := newSoDtIDs[soDtID]; !exists {
			removedSoDtIDs[soDtID] = soID
		}
	}

	addedSoDtIDs := make(map[uint]uint)
	for soDtID, soID := range newSoDtIDs {
		if _, exists := existingSoDtIDs[soDtID]; !exists {
			addedSoDtIDs[soDtID] = soID
		}
	}

	for soDtID := range removedSoDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, nil, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	salesInvoice, err := utils.MapUpdateSalesInvoice(ctx, req, userID, branchID, existingSalesInvoice.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.UpdateSalesInvoice(tx, &salesInvoice, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	existingDtMap := make(map[string]*dtos.SalesInvoiceDtListDTO)
	for i, dt := range existingSalesInvoice.SalesInvoiceDts {
		if dt.ID != nil {
			key := fmt.Sprintf("%d-%d-%d",
				utils.GetValueOrDefault(dt.ProductID, 0),
				utils.GetValueOrDefault(dt.RefID, 0),
				utils.GetValueOrDefault(dt.RefDtID, 0))
			existingDtMap[key] = &existingSalesInvoice.SalesInvoiceDts[i]
		}
	}

	salesInvoiceDts, err := utils.MapUpdateSalesInvoiceDts(ctx, req, &salesInvoice, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var createSalesInvoiceDts []models.SalesInvoiceDt
	var updateSalesInvoiceDts []models.SalesInvoiceDt
	var deleteSalesInvoiceDtIDs []uint

	processedExistingDts := make(map[uint]bool)

	for i, dt := range salesInvoiceDts {
		key := fmt.Sprintf("%d-%d-%d",
			utils.GetValueOrDefault(dt.ProductID, 0),
			utils.GetValueOrDefault(dt.RefID, 0),
			utils.GetValueOrDefault(dt.RefDtID, 0))

		if existingDt, exists := existingDtMap[key]; exists {
			salesInvoiceDts[i].ID = *existingDt.ID
			salesInvoiceDts[i].UpdatedByID = &userID
			updateSalesInvoiceDts = append(updateSalesInvoiceDts, salesInvoiceDts[i])
			processedExistingDts[*existingDt.ID] = true
		} else {
			salesInvoiceDts[i].CreatedByID = &userID
			salesInvoiceDts[i].CreatedAt = time.Now()
			createSalesInvoiceDts = append(createSalesInvoiceDts, salesInvoiceDts[i])
		}
	}

	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
		if dt.ID != nil && !processedExistingDts[*dt.ID] {
			deleteSalesInvoiceDtIDs = append(deleteSalesInvoiceDtIDs, *dt.ID)
		}
	}

	if len(deleteSalesInvoiceDtIDs) > 0 {
		tx, err = s.repo.DeleteSalesInvoiceDtsByIDs(tx, deleteSalesInvoiceDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createSalesInvoiceDts) > 0 {
		tx, err = s.repo.BulkCreateSalesInvoiceDts(tx, createSalesInvoiceDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(updateSalesInvoiceDts) > 0 {
		tx, err = s.repo.BulkUpdateSalesInvoiceDts(tx, updateSalesInvoiceDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	for soDtID := range addedSoDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, "INVOICE", childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	affectedSOIDs := make(map[uint]bool)
	for _, soID := range removedSoDtIDs {
		affectedSOIDs[soID] = true
	}
	for _, soID := range addedSoDtIDs {
		affectedSOIDs[soID] = true
	}

	for soID := range affectedSOIDs {
		tx, err = s.repo.CheckAndUpdateSalesOrderStatus(tx, soID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &salesInvoice, nil
}

func (s *SalesInvoiceService) DeleteSalesInvoice(ctx *fiber.Ctx, salesInvoiceID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesInvoiceService-DeleteSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetSalesInvoiceParams{ID: salesInvoiceID}
	salesInvoice, err := s.GetSalesInvoiceByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	soDtIDs := make(map[uint]uint)
	soIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)

	for _, dt := range salesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			soDtIDs[*dt.RefDtID] = *dt.RefID
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				soIDs = append(soIDs, *dt.RefID)
			}
		}
	}

	if len(soIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, soIDs, childSpan); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := s.repo.LockSalesInvoice(tx, salesInvoiceID, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	for soDtID, soID := range soDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, nil, childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx, err = s.repo.CheckAndUpdateSalesOrderStatus(tx, soID, childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx, err = s.repo.DeleteSalesInvoice(tx, salesInvoiceID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *SalesInvoiceService) RestoreSalesInvoice(ctx *fiber.Ctx, params *dtos.GetSalesInvoiceParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("SalesInvoiceService-RestoreSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	isDeleted := 1
	getParams := &dtos.GetSalesInvoiceParams{
		ID:        params.ID,
		IsDeleted: &isDeleted,
	}

	salesInvoice, err := s.GetSalesInvoiceByID(ctx, getParams, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	soDtIDs := make(map[uint]uint)
	soIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)

	for _, dt := range salesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			soDtIDs[*dt.RefDtID] = *dt.RefID
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				soIDs = append(soIDs, *dt.RefID)
			}
		}
	}

	if len(soIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, soIDs, childSpan); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := s.repo.RestoreSalesInvoice(ctx, params, tx, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	for soDtID, soID := range soDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, "INVOICE", childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx, err = s.repo.CheckAndUpdateSalesOrderStatus(tx, soID, childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (s *SalesInvoiceService) CreateSalesInvoiceDts(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, userID uint, createdSalesInvoice *models.SalesInvoice, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.SalesInvoiceDt, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-CreateSalesInvoiceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoiceDts, err := utils.MapCreateSalesInvoiceDts(ctx, req, createdSalesInvoice, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx, createdSalesInvoiceDts, err := s.repo.CreateSalesInvoiceDts(tx, salesInvoiceDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	return tx, createdSalesInvoiceDts, nil
}

func (s *SalesInvoiceService) GetRefSalesOrderDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetRefSalesOrderDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDts, total, err := s.repo.GetRefSalesOrderDts(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	if len(soDts) > 0 {
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
				soDts = utils.MapRefSoDtBomsToSoDtsForInvoice(soDtBoms, soDts)
			}
		}
	}

	return soDts, total, nil
}

func (s *SalesInvoiceService) GetSoDtInvoiceStatus(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) (map[uint]string, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetSoDtInvoiceStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(soDtIDs) == 0 {
		return make(map[uint]string), nil
	}

	statusMap, err := s.repo.GetSoDtInvoiceStatus(ctx, soDtIDs, childSpan)
	if err != nil {
		return nil, err
	}

	return statusMap, nil
}

func (s *SalesInvoiceService) GetWidgetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesInvoiceStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetWidgetSalesInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	salesInvoices, total, err := s.repo.GetWidgetSalesInvoices(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return salesInvoices, total, nil
}

func (s *SalesInvoiceService) BeginTransaction() *gorm.DB {
	return s.repo.BeginTransaction()
}

func (s *SalesInvoiceService) Commit(tx *gorm.DB) error {
	return s.repo.Commit(tx)
}

func (s *SalesInvoiceService) Rollback(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}
