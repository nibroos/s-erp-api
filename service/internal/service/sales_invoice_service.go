package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

	uniqueInventoryIDs := make([]uint, 0)
	inventoryIDsMap := make(map[uint]bool)
	invDtIDs := make(map[uint]uint)

	for _, dt := range req.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && *dt.RefID > 0 {
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				uniqueSOIDs = append(uniqueSOIDs, *dt.RefID)
			}

			if dt.RefDtID != nil && *dt.RefDtID > 0 {
				soDtIDs[*dt.RefDtID] = *dt.RefID
			}
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && *dt.RefID > 0 {
			if !inventoryIDsMap[*dt.RefID] {
				inventoryIDsMap[*dt.RefID] = true
				uniqueInventoryIDs = append(uniqueInventoryIDs, *dt.RefID)
			}

			if dt.RefDtID != nil && *dt.RefDtID > 0 {
				invDtIDs[*dt.RefDtID] = *dt.RefID
			}
		}
	}

	if len(uniqueSOIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, uniqueSOIDs, childSpan); err != nil {
			tx.Rollback()
			return nil, tx, err
		}
	}

	if len(uniqueInventoryIDs) > 0 {
		if err := s.repo.LockInventories(tx, uniqueInventoryIDs, childSpan); err != nil {
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

	for invDtID, invID := range invDtIDs {
		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, "INVOICE", childSpan)
		if err != nil {
			tx.Rollback()
			return nil, tx, err
		}

		tx, err = s.repo.CheckAndUpdateInventoryStatus(tx, invID, childSpan)
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

// func (s *SalesInvoiceService) UpdateSalesInvoice(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesInvoice, error) {
// 	childSpan := opentracing.StartSpan("SalesInvoiceService-UpdateSalesInvoice", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	existingSoDtIDs := make(map[uint]uint)
// 	newSoDtIDs := make(map[uint]uint)
// 	allSOIDs := make([]uint, 0)
// 	uniqueSOIDs := make(map[uint]bool)

// 	existingInvDtIDs := make(map[uint]uint)
// 	newInvDtIDs := make(map[uint]uint)
// 	allInventoryIDs := make([]uint, 0)
// 	uniqueInventoryIDs := make(map[uint]bool)

// 	params := dtos.GetSalesInvoiceParams{ID: req.ID}
// 	existingSalesInvoice, err := s.GetSalesInvoiceByID(ctx, &params, tx, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
// 		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
// 			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueSOIDs[*dt.RefID] {
// 				uniqueSOIDs[*dt.RefID] = true
// 				allSOIDs = append(allSOIDs, *dt.RefID)
// 			}
// 		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
// 			existingInvDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueInventoryIDs[*dt.RefID] {
// 				uniqueInventoryIDs[*dt.RefID] = true
// 				allInventoryIDs = append(allInventoryIDs, *dt.RefID)
// 			}
// 		}
// 	}

// 	for _, dt := range req.SalesInvoiceDts {
// 		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
// 			newSoDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueSOIDs[*dt.RefID] {
// 				uniqueSOIDs[*dt.RefID] = true
// 				allSOIDs = append(allSOIDs, *dt.RefID)
// 			}
// 		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
// 			newInvDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueInventoryIDs[*dt.RefID] {
// 				uniqueInventoryIDs[*dt.RefID] = true
// 				allInventoryIDs = append(allInventoryIDs, *dt.RefID)
// 			}
// 		}
// 	}

// 	if len(allSOIDs) > 0 {
// 		if err := s.repo.LockSalesOrders(tx, allSOIDs, childSpan); err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(allInventoryIDs) > 0 {
// 		if err := s.repo.LockInventories(tx, allInventoryIDs, childSpan); err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if err := s.repo.LockSalesInvoice(tx, req.ID, childSpan); err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	removedSoDtIDs := make(map[uint]uint)
// 	for soDtID, soID := range existingSoDtIDs {
// 		if _, exists := newSoDtIDs[soDtID]; !exists {
// 			removedSoDtIDs[soDtID] = soID
// 		}
// 	}

// 	addedSoDtIDs := make(map[uint]uint)
// 	for soDtID, soID := range newSoDtIDs {
// 		if _, exists := existingSoDtIDs[soDtID]; !exists {
// 			addedSoDtIDs[soDtID] = soID
// 		}
// 	}

// 	removedInvDtIDs := make(map[uint]uint)
// 	for invDtID, invID := range existingInvDtIDs {
// 		if _, exists := newInvDtIDs[invDtID]; !exists {
// 			removedInvDtIDs[invDtID] = invID
// 		}
// 	}

// 	addedInvDtIDs := make(map[uint]uint)
// 	for invDtID, invID := range newInvDtIDs {
// 		if _, exists := existingInvDtIDs[invDtID]; !exists {
// 			addedInvDtIDs[invDtID] = invID
// 		}
// 	}

// 	for soDtID := range removedSoDtIDs {
// 		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, nil, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	for invDtID := range removedInvDtIDs {
// 		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, nil, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	salesInvoice, err := utils.MapUpdateSalesInvoice(ctx, req, userID, branchID, existingSalesInvoice.RevNo, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	tx, err = s.repo.UpdateSalesInvoice(tx, &salesInvoice, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	existingDtMap := make(map[string]*dtos.SalesInvoiceDtListDTO)
// 	for i, dt := range existingSalesInvoice.SalesInvoiceDts {
// 		if dt.ID != nil {
// 			key := fmt.Sprintf("%d-%d-%d",
// 				utils.GetValueOrDefault(dt.ProductID, 0),
// 				utils.GetValueOrDefault(dt.RefID, 0),
// 				utils.GetValueOrDefault(dt.RefDtID, 0))
// 			existingDtMap[key] = &existingSalesInvoice.SalesInvoiceDts[i]
// 		}
// 	}

// 	salesInvoiceDts, err := utils.MapUpdateSalesInvoiceDts(ctx, req, &salesInvoice, userID, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	var createSalesInvoiceDts []models.SalesInvoiceDt
// 	var updateSalesInvoiceDts []models.SalesInvoiceDt
// 	var deleteSalesInvoiceDtIDs []uint

// 	processedExistingDts := make(map[uint]bool)

// 	for i, dt := range salesInvoiceDts {
// 		key := fmt.Sprintf("%d-%d-%d",
// 			utils.GetValueOrDefault(dt.ProductID, 0),
// 			utils.GetValueOrDefault(dt.RefID, 0),
// 			utils.GetValueOrDefault(dt.RefDtID, 0))

// 		if existingDt, exists := existingDtMap[key]; exists {
// 			salesInvoiceDts[i].ID = *existingDt.ID
// 			salesInvoiceDts[i].UpdatedByID = &userID
// 			updateSalesInvoiceDts = append(updateSalesInvoiceDts, salesInvoiceDts[i])
// 			processedExistingDts[*existingDt.ID] = true
// 		} else {
// 			salesInvoiceDts[i].CreatedByID = &userID
// 			salesInvoiceDts[i].CreatedAt = time.Now()
// 			createSalesInvoiceDts = append(createSalesInvoiceDts, salesInvoiceDts[i])
// 		}
// 	}

// 	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
// 		if dt.ID != nil && !processedExistingDts[*dt.ID] {
// 			deleteSalesInvoiceDtIDs = append(deleteSalesInvoiceDtIDs, *dt.ID)
// 		}
// 	}

// 	if len(deleteSalesInvoiceDtIDs) > 0 {
// 		tx, err = s.repo.DeleteSalesInvoiceDtsByIDs(tx, deleteSalesInvoiceDtIDs, userID, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(createSalesInvoiceDts) > 0 {
// 		tx, err = s.repo.BulkCreateSalesInvoiceDts(tx, createSalesInvoiceDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(updateSalesInvoiceDts) > 0 {
// 		tx, err = s.repo.BulkUpdateSalesInvoiceDts(tx, updateSalesInvoiceDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	for soDtID := range addedSoDtIDs {
// 		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, "INVOICE", childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	for invDtID := range addedInvDtIDs {
// 		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, "INVOICE", childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	affectedSOIDs := make(map[uint]bool)
// 	for _, soID := range removedSoDtIDs {
// 		affectedSOIDs[soID] = true
// 	}
// 	for _, soID := range addedSoDtIDs {
// 		affectedSOIDs[soID] = true
// 	}

// 	for soID := range affectedSOIDs {
// 		tx, err = s.repo.CheckAndUpdateSalesOrderStatus(tx, soID, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	affectedInventoryIDs := make(map[uint]bool)
// 	for _, invID := range removedInvDtIDs {
// 		affectedInventoryIDs[invID] = true
// 	}
// 	for _, invID := range addedInvDtIDs {
// 		affectedInventoryIDs[invID] = true
// 	}

// 	for invID := range affectedInventoryIDs {
// 		tx, err = s.repo.CheckAndUpdateInventoryStatus(tx, invID, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	return &salesInvoice, nil
// }

func (s *SalesInvoiceService) UpdateSalesInvoice(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.SalesInvoice, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-UpdateSalesInvoice", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	existingSoDtIDs := make(map[uint]uint)
	newSoDtIDs := make(map[uint]uint)
	allSOIDs := make([]uint, 0)
	uniqueSOIDs := make(map[uint]bool)

	existingInvDtIDs := make(map[uint]uint)
	newInvDtIDs := make(map[uint]uint)
	allInventoryIDs := make([]uint, 0)
	uniqueInventoryIDs := make(map[uint]bool)

	params := dtos.GetSalesInvoiceParams{ID: req.ID}
	existingSalesInvoice, err := s.GetSalesInvoiceByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	isChangingToCancelled := req.Status != nil && *req.Status == "CANCELLED" &&
		(existingSalesInvoice.Status == nil || *existingSalesInvoice.Status != "CANCELLED")

	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueSOIDs[*dt.RefID] {
				uniqueSOIDs[*dt.RefID] = true
				allSOIDs = append(allSOIDs, *dt.RefID)
			}
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
			existingInvDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueInventoryIDs[*dt.RefID] {
				uniqueInventoryIDs[*dt.RefID] = true
				allInventoryIDs = append(allInventoryIDs, *dt.RefID)
			}
		}
	}

	for _, dt := range existingSalesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueSOIDs[*dt.RefID] {
				uniqueSOIDs[*dt.RefID] = true
				allSOIDs = append(allSOIDs, *dt.RefID)
			}
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
			existingInvDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueInventoryIDs[*dt.RefID] {
				uniqueInventoryIDs[*dt.RefID] = true
				allInventoryIDs = append(allInventoryIDs, *dt.RefID)
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
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
			newInvDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueInventoryIDs[*dt.RefID] {
				uniqueInventoryIDs[*dt.RefID] = true
				allInventoryIDs = append(allInventoryIDs, *dt.RefID)
			}
		}
	}

	if len(allSOIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, allSOIDs, childSpan); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(allInventoryIDs) > 0 {
		if err := s.repo.LockInventories(tx, allInventoryIDs, childSpan); err != nil {
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

	removedInvDtIDs := make(map[uint]uint)
	for invDtID, invID := range existingInvDtIDs {
		if _, exists := newInvDtIDs[invDtID]; !exists {
			removedInvDtIDs[invDtID] = invID
		}
	}

	addedInvDtIDs := make(map[uint]uint)
	for invDtID, invID := range newInvDtIDs {
		if _, exists := existingInvDtIDs[invDtID]; !exists {
			addedInvDtIDs[invDtID] = invID
		}
	}

	for soDtID := range removedSoDtIDs {
		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, nil, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	for invDtID := range removedInvDtIDs {
		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, nil, childSpan)
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

	if isChangingToCancelled {
		tx, err = s.repo.ResetReferencesForCancelled(tx, salesInvoice.ID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx, err = s.repo.UpdateSalesInvoice(tx, &salesInvoice, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if isChangingToCancelled {
		return &salesInvoice, nil
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

	for invDtID := range addedInvDtIDs {
		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, "INVOICE", childSpan)
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

	affectedInventoryIDs := make(map[uint]bool)
	for _, invID := range removedInvDtIDs {
		affectedInventoryIDs[invID] = true
	}
	for _, invID := range addedInvDtIDs {
		affectedInventoryIDs[invID] = true
	}

	for invID := range affectedInventoryIDs {
		tx, err = s.repo.CheckAndUpdateInventoryStatus(tx, invID, childSpan)
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

	invDtIDs := make(map[uint]uint)
	invIDs := make([]uint, 0)
	invIDsMap := make(map[uint]bool)

	for _, dt := range salesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			soDtIDs[*dt.RefDtID] = *dt.RefID
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				soIDs = append(soIDs, *dt.RefID)
			}
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
			invDtIDs[*dt.RefDtID] = *dt.RefID
			if !invIDsMap[*dt.RefID] {
				invIDsMap[*dt.RefID] = true
				invIDs = append(invIDs, *dt.RefID)
			}
		}
	}

	if len(soIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, soIDs, childSpan); err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(invIDs) > 0 {
		if err := s.repo.LockInventories(tx, invIDs, childSpan); err != nil {
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

	for invDtID, invID := range invDtIDs {
		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, nil, childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx, err = s.repo.CheckAndUpdateInventoryStatus(tx, invID, childSpan)
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

	invDtIDs := make(map[uint]uint)
	invIDs := make([]uint, 0)
	invIDsMap := make(map[uint]bool)

	for _, dt := range salesInvoice.SalesInvoiceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			soDtIDs[*dt.RefDtID] = *dt.RefID
			if !soIDsMap[*dt.RefID] {
				soIDsMap[*dt.RefID] = true
				soIDs = append(soIDs, *dt.RefID)
			}
		} else if dt.RefType != nil && *dt.RefType == "inv_out" && dt.RefID != nil && dt.RefDtID != nil {
			invDtIDs[*dt.RefDtID] = *dt.RefID
			if !invIDsMap[*dt.RefID] {
				invIDsMap[*dt.RefID] = true
				invIDs = append(invIDs, *dt.RefID)
			}
		}
	}

	if len(soIDs) > 0 {
		if err := s.repo.LockSalesOrders(tx, soIDs, childSpan); err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(invIDs) > 0 {
		if err := s.repo.LockInventories(tx, invIDs, childSpan); err != nil {
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

	for invDtID, invID := range invDtIDs {
		tx, err = s.repo.UpdateInvDtInvoiceStatus(tx, invDtID, "INVOICE", childSpan)
		if err != nil {
			tx.Rollback()
			return err
		}

		tx, err = s.repo.CheckAndUpdateInventoryStatus(tx, invID, childSpan)
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

func (s *SalesInvoiceService) GetRefInventoryOutDts(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefInventoryOutForInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetRefInventoryOutDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invDts, total, err := s.repo.GetRefInventoryOutDts(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}

	return invDts, total, nil
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

func (s *SalesInvoiceService) GetInvDtInvoiceStatus(ctx *fiber.Ctx, invDtIDs []uint, span opentracing.Span) (map[uint]string, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetInvDtInvoiceStatus", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(invDtIDs) == 0 {
		return make(map[uint]string), nil
	}

	statusMap, err := s.repo.GetInvDtInvoiceStatus(ctx, invDtIDs, childSpan)
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

func (s *SalesInvoiceService) ExcelGetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-ExcelGetSalesInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	salesInvoices, _, err := s.GetSalesInvoices(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "SalesInvoices"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{
		"Customer", "Invoice No", "Title", "Invoice Date", "Due Date",
		"Bank", "Currency", "Exchange Rate", "VAT", "PPh23",
		"Qty", "Sub Amount", "DP Amount", "Balance", "Grand Total", "Status", "Created By",
	}
	file.SetSheetRow(sheetName, "A1", &headers)

	file.SetColWidth(sheetName, "A", "A", 25)
	file.SetColWidth(sheetName, "B", "B", 15)
	file.SetColWidth(sheetName, "C", "C", 30)
	file.SetColWidth(sheetName, "D", "E", 15)
	file.SetColWidth(sheetName, "F", "F", 20)
	file.SetColWidth(sheetName, "G", "G", 15)
	file.SetColWidth(sheetName, "H", "H", 15)
	file.SetColWidth(sheetName, "I", "J", 15)
	file.SetColWidth(sheetName, "K", "O", 15)
	file.SetColWidth(sheetName, "P", "P", 15)
	file.SetColWidth(sheetName, "Q", "Q", 20)

	headerStyle, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DCE6F1"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	file.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)

	currencyStyle, _ := file.NewStyle(&excelize.Style{
		NumFmt: 4,
	})

	for i, invoice := range salesInvoices {
		row := i + 2
		file.SetSheetRow(sheetName, fmt.Sprintf("A%d", row), &[]interface{}{
			invoice.CustomerName,
			invoice.InvoiceNo,
			invoice.Title,
			invoice.InvoiceDate,
			invoice.DueDate,
			invoice.BankName,
			invoice.CurrencyName,
			invoice.ExchangeRate,
			invoice.VatName,
			invoice.Pph23Name,
			invoice.TotalQty,
			invoice.Subtotal,
			invoice.TotalDpProducts,
			invoice.TotalBalanceProducts,
			invoice.GrandTotal,
			invoice.Status,
			invoice.CreatedByName,
		})

		file.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), currencyStyle)
		file.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("O%d", row), currencyStyle)
	}

	file.SetActiveSheet(index)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// func (s *SalesInvoiceService) CsvGetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
// 	childSpan := opentracing.StartSpan("SalesInvoiceService-CsvGetSalesInvoices", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	filters["is_csv"] = "1"
// 	salesInvoices, _, err := s.GetSalesInvoices(ctx, filters, childSpan)
// 	if err != nil {
// 		return nil, err
// 	}

// 	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
// 	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
// 	appName := "App"
// 	if err == nil && companyProfile != nil && companyProfile.CompanyName != nil {
// 		appName = *companyProfile.CompanyName
// 	}

// 	csv := fmt.Sprintf("%s\n", appName)
// 	csv += "\n"
// 	csv += "Invoice Sales\n"
// 	csv += "\n"

// 	csv += "Customer,Invoice No,Title,Invoice Date,Due Date,Bank,Currency,Exchange Rate,VAT,PPh23,Qty,Sub Amount,DP Amount,Balance,Grand Total,Status,Created By\n"

// 	for _, invoice := range salesInvoices {
// 		customerName := utils.GetPtrVal(invoice.CustomerName)
// 		invoiceNo := utils.GetPtrVal(invoice.InvoiceNo)
// 		title := utils.GetPtrVal(invoice.Title)
// 		invoiceDate := utils.GetPtrVal(invoice.InvoiceDate)
// 		dueDate := utils.GetPtrVal(invoice.DueDate)

// 		bankName := utils.GetPtrVal(invoice.BankName)
// 		accountNumber := utils.GetPtrVal(invoice.AccountNumber)
// 		accountName := utils.GetPtrVal(invoice.AccountName)

// 		bankInfo := bankName
// 		if accountNumber != "" {
// 			if bankInfo != "" {
// 				bankInfo += " - "
// 			}
// 			bankInfo += accountNumber
// 		}
// 		if accountName != "" {
// 			if bankInfo != "" {
// 				bankInfo += " - "
// 			}
// 			bankInfo += accountName
// 		}

// 		currencyName := utils.GetPtrVal(invoice.CurrencyName)
// 		totalVat := utils.GetFloatPtrVal(invoice.TotalVat)
// 		totalPph23 := utils.GetFloatPtrVal(invoice.TotalPph23)
// 		status := utils.GetPtrVal(invoice.Status)
// 		createdByName := utils.GetPtrVal(invoice.CreatedByName)

// 		exchangeRate := utils.GetFloatPtrVal(invoice.ExchangeRate)
// 		totalQty := utils.GetFloatPtrVal(invoice.TotalQty)
// 		subtotal := utils.GetFloatPtrVal(invoice.Subtotal)
// 		totalDpProducts := utils.GetFloatPtrVal(invoice.TotalDpProducts)
// 		totalBalanceProducts := utils.GetFloatPtrVal(invoice.TotalBalanceProducts)
// 		grandTotal := utils.GetFloatPtrVal(invoice.GrandTotal)

// 		customerName = utils.EscapeCsvField(customerName)
// 		invoiceNo = utils.EscapeCsvField(invoiceNo)
// 		title = utils.EscapeCsvField(title)
// 		bankInfo = utils.EscapeCsvField(bankInfo)
// 		currencyName = utils.EscapeCsvField(currencyName)
// 		status = utils.EscapeCsvField(status)
// 		createdByName = utils.EscapeCsvField(createdByName)

// 		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s,%s\n",
// 			customerName,
// 			invoiceNo,
// 			title,
// 			invoiceDate,
// 			dueDate,
// 			bankInfo,
// 			currencyName,
// 			exchangeRate,
// 			totalVat,
// 			totalPph23,
// 			totalQty,
// 			subtotal,
// 			totalDpProducts,
// 			totalBalanceProducts,
// 			grandTotal,
// 			status,
// 			createdByName,
// 		)
// 	}

// 	return []byte(csv), nil
// }

// github.com/xuri/excelize/v2
func (s *SalesInvoiceService) CsvGetSalesInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-CsvGetSalesInvoices", opentracing.ChildOf(span.Context()))

	// filters is_csv
	filters["is_csv"] = "1"

	exportType := utils.GetStringOrDefault(filters["export_type"], "all")

	if exportType == "detail" {
		return s.CsvGetDetail(ctx, filters, childSpan)
	}

	return s.CsvGetAll(ctx, filters, childSpan)
}

// CsvGetAll
func (s *SalesInvoiceService) CsvGetAll(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-CsvGetAll", opentracing.ChildOf(span.Context()))
	salesOrders, _, err := s.GetSalesInvoices(ctx, filters, childSpan)
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
	csv += "Sales Invoices\n"
	csv += "\n"

	utils.BuildSalesInvoiceAllCSVRows(salesOrders, &csv)

	return []byte(csv), nil
}

func (s *SalesInvoiceService) CsvGetDetail(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-CsvGetDetail", opentracing.ChildOf(span.Context()))
	quotations, _, err := s.GetSalesInvoicesDetails(ctx, filters, childSpan)
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
	csv += "Sales Invoices\n"
	csv += "\n"

	// csv += "ID,Sales Invoice No,Order Type,Customer,Expired Date,Quot Date,Currency,Total,Status,Created By,Updated By\n"

	// // Build CSV rows
	// for _, quotation := range quotations {
	// 	csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%.2f,%s,%s,%s\n",
	// 		quotation.ID,
	// 		utils.GetPtrVal(quotation.Remark),
	// 	)
	// }
	utils.BuildSalesInvoiceDetailCSVRows(quotations, &csv)

	return []byte(csv), nil
}

func (s *SalesInvoiceService) GetSalesInvoicesDetails(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.SalesInvoiceDetailDTO, int, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-GetSalesInvoicesDetails", opentracing.ChildOf(span.Context()))

	salesOrders, total, err := s.repo.GetSalesInvoicesDetails(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	salesInvoiceIDs := utils.GetSalesInvoiceDetailsIDs(salesOrders)

	// Get QuoDts by salesOrder IDs
	siDts, _, err := s.repo.GetSalesInvoiceDetailsDts(ctx, filters, salesInvoiceIDs, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}

	// Get QuoDtBoms by salesOrder IDs
	siDtBoms, _, err := s.repo.GetSalesInvoiceDetailsDtBoms(ctx, filters, salesInvoiceIDs, childSpan)
	if err != nil {
		log.Println("Failed to fetch salesInvoiceDtsBoms:", err)
		defer childSpan.Finish()
		return nil, 0, err
	}

	salesOrders = utils.MapGetSalesInvoiceDetails(salesOrders, siDts, siDtBoms)

	return salesOrders, total, nil
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

func (s *SalesInvoiceService) Pdf(ctx *fiber.Ctx, req dtos.SalesInvoiceDetailDTO, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("SalesInvoiceService-Pdf", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	form := req
	var salesInvoice *dtos.SalesInvoiceDetailDTO
	var err error

	var params dtos.GetSalesInvoiceParams
	if req.IsIDOnly != nil && *req.IsIDOnly == 1 {
		params.ID = req.ID

		salesInvoice, err = s.repo.GetSalesInvoiceByID(ctx, &params, tx, childSpan)
		if err != nil {
			return nil, err
		}

		salesInvoiceDts, err := s.repo.GetSalesInvoiceDts(ctx, salesInvoice.ID, nil, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch salesInvoiceDts: %v", err)
		}

		if salesInvoice.CompanyProfileID != nil {
			companyParams := &dtos.GetCompanyProfileParams{ID: uint(*salesInvoice.CompanyProfileID)}
			company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
			if err != nil {
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to fetch company: %v", err)
			} else if company != nil {
				salesInvoice.Company = *company
			}
		}

		salesInvoice.SalesInvoiceDts = salesInvoiceDts
		form = *salesInvoice
		req.InvoiceNo = salesInvoice.InvoiceNo
	}

	var num string
	if req.InvoiceNo != nil {
		num = *req.InvoiceNo
	} else {
		num = ""
	}

	data := dtos.SalesInvoicePDFData{
		Num:  num,
		Form: form,
	}

	htmlFileName := "sales-invoice-detail"
	log.Println("Pdf-htmlFileName-si", htmlFileName)

	templateFile, err := templateFS.Open(fmt.Sprintf("templates/%s.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to open embedded template: %v", err)
		return nil, err
	}

	templateContent, err := io.ReadAll(templateFile)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Printf("Failed to read template content: %v", err)
		return nil, err
	}

	htmlFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.html", htmlFileName))
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error writing htmlFile:", err)
		return nil, err
	}
	defer os.Remove(htmlFile.Name())

	funcMap := template.FuncMap{
		"formatNumber": func(n float64, args ...int) string {
			decimals := 2
			if len(args) > 0 {
				decimals = args[0]
			}

			format := fmt.Sprintf("%%.%df", decimals)
			p := message.NewPrinter(language.English)
			return p.Sprintf(format, n)
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

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

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg:", err)
		return nil, err
	}

	headerContent, err := templateFS.ReadFile("templates/header.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	footerContent, err := templateFS.ReadFile("templates/footer.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}

	headerPath, err := createTempFileFromEmbed(string(headerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create header temp file: %w", err)
	}
	defer os.Remove(headerPath)

	footerPath, err := createTempFileFromEmbed(string(footerContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create footer temp file: %w", err)
	}
	defer os.Remove(footerPath)

	page := wkhtmltopdf.NewPage(htmlFile.Name())
	page.EnableLocalFileAccess.Set(true)
	page.HeaderHTML.Set("file://" + headerPath)
	page.FooterHTML.Set("file://" + footerPath)
	page.FooterSpacing.Set(10)

	pdfg.AddPage(page)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)

	if err := pdfg.Create(); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdfg create:", err)
		return nil, err
	}

	uploadDir := "./public/generated_pdfs"

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	replacedTitle := strings.ReplaceAll(*form.Title, "/", "_")
	form.Title = &replacedTitle

	fileName := fmt.Sprintf("%s-%s.pdf", replacedTitle, time.Now().Format("20060102150405"))
	pdfPath := filepath.Join(uploadDir, fileName)
	if err := pdfg.WriteFile(pdfPath); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdf path:", err)
		return nil, err
	}

	return &pdfPath, nil
}
