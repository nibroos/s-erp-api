package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/auth"
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

type InvoiceMaintenanceService struct {
	repo     *repository.InvoiceMaintenanceRepository
	utilRepo *repository.UtilRepository
	rabbitmq *config.RabbitMQ
	tracer   opentracing.Tracer
}

func NewInvoiceMaintenanceService(repo *repository.InvoiceMaintenanceRepository, utilRepo *repository.UtilRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *InvoiceMaintenanceService {
	return &InvoiceMaintenanceService{
		repo:     repo,
		utilRepo: utilRepo,
		rabbitmq: rabbitmq,
		tracer:   tracer,
	}
}

func (s *InvoiceMaintenanceService) GetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetInvoiceMaintenances", opentracing.ChildOf(span.Context()))

	invoiceMaintenances, total, err := s.repo.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, 0, err
	}
	return invoiceMaintenances, total, nil
}

func (s *InvoiceMaintenanceService) CreateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceMaintenance, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-CreateInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	uniqueSOIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)
	soDtIDs := make(map[uint]uint)

	for _, dt := range req.InvoiceMaintenanceDts {
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

	invoiceMaintenanceCreatedThisMonthNumber, err := s.repo.GetInvoiceMaintenanceCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := invoiceMaintenanceCreatedThisMonthNumber + 1

	invoiceMaintenance, err := utils.MapCreateInvoiceMaintenance(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		defer childSpan.Finish()
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateInvoiceMaintenance(tx, &invoiceMaintenance, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, _, err = s.CreateInvoiceMaintenanceDts(ctx, req, userID, &invoiceMaintenance, tx, childSpan)
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

	return &invoiceMaintenance, tx, nil
}

func (s *InvoiceMaintenanceService) GetInvoiceMaintenanceByID(ctx *fiber.Ctx, params *dtos.GetInvoiceMaintenanceParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceMaintenanceDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetInvoiceMaintenanceByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenance, err := s.repo.GetInvoiceMaintenanceByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	invoiceMaintenanceDts, err := s.repo.GetInvoiceMaintenanceDts(ctx, invoiceMaintenance.ID, params.IsDeleted, childSpan)
	if err != nil {
		return nil, err
	}

	var soDtIDs []uint
	for _, dt := range invoiceMaintenanceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
			soDtIDs = append(soDtIDs, *dt.RefDtID)
		}
	}

	if len(soDtIDs) > 0 {
		soDtBoms, err := s.repo.GetSoDtBoms(ctx, soDtIDs, childSpan)
		if err != nil {
			return nil, err
		}

		for i, dt := range invoiceMaintenanceDts {
			if dt.RefType != nil && *dt.RefType == "so" && dt.ProductType != nil && *dt.ProductType == "product" && dt.RefDtID != nil {
				var dtBoms []dtos.SalesOrderSoDtBomListDTO
				for _, bom := range soDtBoms {
					if bom.SoDtID != nil && *bom.SoDtID == *dt.RefDtID {
						dtBoms = append(dtBoms, bom)
					}
				}
				invoiceMaintenanceDts[i].SoDtsBoms = dtBoms
			}
		}
	}

	invoiceMaintenance.InvoiceMaintenanceDts = invoiceMaintenanceDts

	return invoiceMaintenance, nil
}

func (s *InvoiceMaintenanceService) GetInvoiceMaintenanceDtsByID(ctx *fiber.Ctx, filters map[string]string, tx *gorm.DB, span opentracing.Span) ([]dtos.InvoiceMaintenanceDtListNoBomDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetInvoiceMaintenanceByID", opentracing.ChildOf(span.Context()))

	invoiceMaintenance, _, err := s.repo.GetInvoiceMaintenanceDtsRawByIDs(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		return nil, err
	}

	return invoiceMaintenance, nil
}

// func (s *InvoiceMaintenanceService) UpdateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceMaintenance, error) {
// 	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-UpdateInvoiceMaintenance", opentracing.ChildOf(span.Context()))
// 	defer childSpan.Finish()

// 	existingSoDtIDs := make(map[uint]uint)
// 	newSoDtIDs := make(map[uint]uint)
// 	allSOIDs := make([]uint, 0)
// 	uniqueSOIDs := make(map[uint]bool)

// 	params := dtos.GetInvoiceMaintenanceParams{ID: req.ID}
// 	existingInvoiceMaintenance, err := s.GetInvoiceMaintenanceByID(ctx, &params, tx, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	for _, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
// 		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
// 			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueSOIDs[*dt.RefID] {
// 				uniqueSOIDs[*dt.RefID] = true
// 				allSOIDs = append(allSOIDs, *dt.RefID)
// 			}
// 		}
// 	}

// 	for _, dt := range req.InvoiceMaintenanceDts {
// 		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
// 			newSoDtIDs[*dt.RefDtID] = *dt.RefID
// 			if !uniqueSOIDs[*dt.RefID] {
// 				uniqueSOIDs[*dt.RefID] = true
// 				allSOIDs = append(allSOIDs, *dt.RefID)
// 			}
// 		}
// 	}

// 	if len(allSOIDs) > 0 {
// 		if err := s.repo.LockSalesOrders(tx, allSOIDs, childSpan); err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if _, err := s.repo.GetInvoiceMaintenanceForUpdate(tx, req.ID, childSpan); err != nil {
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

// 	for soDtID := range removedSoDtIDs {
// 		tx, err = s.repo.UpdateSoDtInvoiceStatus(tx, soDtID, nil, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	invoiceMaintenance, err := utils.MapUpdateInvoiceMaintenance(ctx, req, userID, branchID, existingInvoiceMaintenance.RevNo, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	tx, err = s.repo.UpdateInvoiceMaintenance(tx, &invoiceMaintenance, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	existingDtMap := make(map[string]*dtos.InvoiceMaintenanceDtListDTO)
// 	for i, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
// 		if dt.ID != nil {
// 			key := fmt.Sprintf("%d-%d-%d",
// 				utils.GetValueOrDefault(dt.ProductID, 0),
// 				utils.GetValueOrDefault(dt.RefID, 0),
// 				utils.GetValueOrDefault(dt.RefDtID, 0))
// 			existingDtMap[key] = &existingInvoiceMaintenance.InvoiceMaintenanceDts[i]
// 		}
// 	}

// 	invoiceMaintenanceDts, err := utils.MapUpdateInvoiceMaintenanceDts(ctx, req, &invoiceMaintenance, userID, childSpan)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	var createInvoiceMaintenanceDts []models.InvoiceMaintenanceDt
// 	var updateInvoiceMaintenanceDts []models.InvoiceMaintenanceDt
// 	var deleteInvoiceMaintenanceDtIDs []uint

// 	processedExistingDts := make(map[uint]bool)

// 	for i, dt := range invoiceMaintenanceDts {
// 		key := fmt.Sprintf("%d-%d-%d",
// 			utils.GetValueOrDefault(dt.ProductID, 0),
// 			utils.GetValueOrDefault(dt.RefID, 0),
// 			utils.GetValueOrDefault(dt.RefDtID, 0))

// 		if existingDt, exists := existingDtMap[key]; exists {
// 			invoiceMaintenanceDts[i].ID = *existingDt.ID
// 			invoiceMaintenanceDts[i].UpdatedByID = &userID
// 			updateInvoiceMaintenanceDts = append(updateInvoiceMaintenanceDts, invoiceMaintenanceDts[i])
// 			processedExistingDts[*existingDt.ID] = true
// 		} else {
// 			invoiceMaintenanceDts[i].CreatedByID = &userID
// 			invoiceMaintenanceDts[i].CreatedAt = time.Now()
// 			createInvoiceMaintenanceDts = append(createInvoiceMaintenanceDts, invoiceMaintenanceDts[i])
// 		}
// 	}

// 	for _, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
// 		if dt.ID != nil && !processedExistingDts[*dt.ID] {
// 			deleteInvoiceMaintenanceDtIDs = append(deleteInvoiceMaintenanceDtIDs, *dt.ID)
// 		}
// 	}

// 	if len(deleteInvoiceMaintenanceDtIDs) > 0 {
// 		tx, err = s.repo.DeleteInvoiceMaintenanceDtsByIDs(tx, deleteInvoiceMaintenanceDtIDs, userID, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(createInvoiceMaintenanceDts) > 0 {
// 		tx, err = s.repo.BulkCreateInvoiceMaintenanceDts(tx, createInvoiceMaintenanceDts, childSpan)
// 		if err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}
// 	}

// 	if len(updateInvoiceMaintenanceDts) > 0 {
// 		tx, err = s.repo.BulkUpdateInvoiceMaintenanceDts(tx, updateInvoiceMaintenanceDts, childSpan)
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

// 	return &invoiceMaintenance, nil
// }

func (s *InvoiceMaintenanceService) UpdateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceMaintenance, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-UpdateInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	existingSoDtIDs := make(map[uint]uint)
	newSoDtIDs := make(map[uint]uint)
	allSOIDs := make([]uint, 0)
	uniqueSOIDs := make(map[uint]bool)

	params := dtos.GetInvoiceMaintenanceParams{ID: req.ID}
	existingInvoiceMaintenance, err := s.GetInvoiceMaintenanceByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	isChangingToCancelled := req.Status != nil && *req.Status == "CANCELLED" &&
		(existingInvoiceMaintenance.Status == nil || *existingInvoiceMaintenance.Status != "CANCELLED")

	for _, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
		if dt.RefType != nil && *dt.RefType == "so" && dt.RefID != nil && dt.RefDtID != nil {
			existingSoDtIDs[*dt.RefDtID] = *dt.RefID
			if !uniqueSOIDs[*dt.RefID] {
				uniqueSOIDs[*dt.RefID] = true
				allSOIDs = append(allSOIDs, *dt.RefID)
			}
		}
	}

	for _, dt := range req.InvoiceMaintenanceDts {
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

	if _, err := s.repo.GetInvoiceMaintenanceForUpdate(tx, req.ID, childSpan); err != nil {
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

	invoiceMaintenance, err := utils.MapUpdateInvoiceMaintenance(ctx, req, userID, branchID, existingInvoiceMaintenance.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if isChangingToCancelled {
		tx, err = s.repo.ResetReferencesForCancelled(tx, invoiceMaintenance.ID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx, err = s.repo.UpdateInvoiceMaintenance(tx, &invoiceMaintenance, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if isChangingToCancelled {
		return &invoiceMaintenance, nil
	}

	existingDtMap := make(map[string]*dtos.InvoiceMaintenanceDtListDTO)
	for i, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
		if dt.ID != nil {
			key := fmt.Sprintf("%d-%d-%d",
				utils.GetValueOrDefault(dt.ProductID, 0),
				utils.GetValueOrDefault(dt.RefID, 0),
				utils.GetValueOrDefault(dt.RefDtID, 0))
			existingDtMap[key] = &existingInvoiceMaintenance.InvoiceMaintenanceDts[i]
		}
	}

	invoiceMaintenanceDts, err := utils.MapUpdateInvoiceMaintenanceDts(ctx, req, &invoiceMaintenance, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var createInvoiceMaintenanceDts []models.InvoiceMaintenanceDt
	var updateInvoiceMaintenanceDts []models.InvoiceMaintenanceDt
	var deleteInvoiceMaintenanceDtIDs []uint

	processedExistingDts := make(map[uint]bool)

	for i, dt := range invoiceMaintenanceDts {
		key := fmt.Sprintf("%d-%d-%d",
			utils.GetValueOrDefault(dt.ProductID, 0),
			utils.GetValueOrDefault(dt.RefID, 0),
			utils.GetValueOrDefault(dt.RefDtID, 0))

		if existingDt, exists := existingDtMap[key]; exists {
			invoiceMaintenanceDts[i].ID = *existingDt.ID
			invoiceMaintenanceDts[i].UpdatedByID = &userID
			updateInvoiceMaintenanceDts = append(updateInvoiceMaintenanceDts, invoiceMaintenanceDts[i])
			processedExistingDts[*existingDt.ID] = true
		} else {
			invoiceMaintenanceDts[i].CreatedByID = &userID
			invoiceMaintenanceDts[i].CreatedAt = time.Now()
			createInvoiceMaintenanceDts = append(createInvoiceMaintenanceDts, invoiceMaintenanceDts[i])
		}
	}

	for _, dt := range existingInvoiceMaintenance.InvoiceMaintenanceDts {
		if dt.ID != nil && !processedExistingDts[*dt.ID] {
			deleteInvoiceMaintenanceDtIDs = append(deleteInvoiceMaintenanceDtIDs, *dt.ID)
		}
	}

	if len(deleteInvoiceMaintenanceDtIDs) > 0 {
		tx, err = s.repo.DeleteInvoiceMaintenanceDtsByIDs(tx, deleteInvoiceMaintenanceDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createInvoiceMaintenanceDts) > 0 {
		tx, err = s.repo.BulkCreateInvoiceMaintenanceDts(tx, createInvoiceMaintenanceDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(updateInvoiceMaintenanceDts) > 0 {
		tx, err = s.repo.BulkUpdateInvoiceMaintenanceDts(tx, updateInvoiceMaintenanceDts, childSpan)
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

	return &invoiceMaintenance, nil
}

func (s *InvoiceMaintenanceService) DeleteInvoiceMaintenance(ctx *fiber.Ctx, invoiceMaintenanceID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-DeleteInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetInvoiceMaintenanceParams{ID: invoiceMaintenanceID}
	invoiceMaintenance, err := s.GetInvoiceMaintenanceByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	soDtIDs := make(map[uint]uint)
	soIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)

	for _, dt := range invoiceMaintenance.InvoiceMaintenanceDts {
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

	if err := s.repo.LockInvoiceMaintenance(tx, invoiceMaintenanceID, childSpan); err != nil {
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

	tx, err = s.repo.DeleteInvoiceMaintenance(tx, invoiceMaintenanceID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceMaintenanceService) RestoreInvoiceMaintenance(ctx *fiber.Ctx, params *dtos.GetInvoiceMaintenanceParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-RestoreInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	isDeleted := 1
	getParams := &dtos.GetInvoiceMaintenanceParams{
		ID:        params.ID,
		IsDeleted: &isDeleted,
	}

	invoiceMaintenance, err := s.GetInvoiceMaintenanceByID(ctx, getParams, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	soDtIDs := make(map[uint]uint)
	soIDs := make([]uint, 0)
	soIDsMap := make(map[uint]bool)

	for _, dt := range invoiceMaintenance.InvoiceMaintenanceDts {
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

	if err := s.repo.RestoreInvoiceMaintenance(ctx, params, tx, childSpan); err != nil {
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

func (s *InvoiceMaintenanceService) CreateInvoiceMaintenanceDts(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, userID uint, createdInvoiceMaintenance *models.InvoiceMaintenance, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.InvoiceMaintenanceDt, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-CreateInvoiceMaintenanceDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenanceDts, err := utils.MapCreateInvoiceMaintenanceDts(ctx, req, createdInvoiceMaintenance, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx, createdInvoiceMaintenanceDts, err := s.repo.CreateInvoiceMaintenanceDts(tx, invoiceMaintenanceDts, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	return tx, createdInvoiceMaintenanceDts, nil
}

func (s *InvoiceMaintenanceService) GetRefSalesOrderForInvoiceMaintenance(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.RefSalesOrderForInvoiceMaintenanceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetRefSalesOrderForInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	soDts, total, err := s.repo.GetRefSalesOrderForInvoiceMaintenance(ctx, filters, childSpan)
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
				soDts = utils.MapRefSoDtBomsToSoDtsForInvoiceMaintenance(soDtBoms, soDts)
			}
		}
	}

	return soDts, total, nil
}

func (s *InvoiceMaintenanceService) UpdateSalesOrderStatusForInvoiceMaintenance(ctx *fiber.Ctx, req dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-UpdateSalesOrderStatusForInvoiceMaintenance", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if err := s.repo.LockSalesOrders(tx, []uint{req.ID}, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	_, err := s.repo.UpdateSalesOrdersStatusForInvoiceMaintenance(tx, req, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceMaintenanceService) ApproveInvoiceMaintenances(ctx *fiber.Ctx, ids []uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-ApproveInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(ids) == 0 {
		return nil
	}

	var count int64
	if err := tx.Model(&models.InvoiceMaintenance{}).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Count(&count).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	if int(count) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("one or more invoice maintenances not found or already deleted")
	}

	tx, err := s.repo.BulkApproveInvoiceMaintenances(tx, ids, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceMaintenanceService) CancelApproveInvoiceMaintenances(ctx *fiber.Ctx, ids []uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-CancelApproveInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	if len(ids) == 0 {
		return nil
	}

	var count int64
	if err := tx.Model(&models.InvoiceMaintenance{}).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Count(&count).Error; err != nil {
		tx.Rollback()
		utils.LogErrors(childSpan, err)
		return err
	}

	if int(count) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("one or more invoice maintenances not found or already deleted")
	}

	tx, err := s.repo.BulkCancelApproveInvoiceMaintenances(tx, ids, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *InvoiceMaintenanceService) GetSoDtInvoiceStatus(ctx *fiber.Ctx, soDtIDs []uint, span opentracing.Span) (map[uint]string, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetSoDtInvoiceStatus", opentracing.ChildOf(span.Context()))
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

func (s *InvoiceMaintenanceService) GetWidgetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceStatusWidget, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetWidgetInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenances, total, err := s.repo.GetWidgetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return invoiceMaintenances, total, nil
}

func (s *InvoiceMaintenanceService) RepeatInvoiceMaintenances(ctx *fiber.Ctx, req dtos.RepeatInvoiceMaintenanceRequest) (*dtos.RepeatInvoiceMaintenanceResponse, error) {
	span := s.tracer.StartSpan("InvoiceMaintenanceService-RepeatInvoiceMaintenances")
	defer span.Finish()

	claims, err := auth.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(span, err)
		return nil, err
	}

	var userID uint
	if uid, ok := claims["user_id"]; ok {
		switch v := uid.(type) {
		case uint:
			userID = v
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		case int64:
			userID = uint(v)
		default:
			utils.LogErrors(span, fmt.Errorf("invalid user ID type: %T", uid))
			return nil, fmt.Errorf("invalid user ID type")
		}
	} else {
		utils.LogErrors(span, fmt.Errorf("user ID not found in claims"))
		return nil, fmt.Errorf("user ID not found in claims")
	}

	results, err := s.repo.RepeatInvoiceMaintenances(ctx, req, userID, span)
	if err != nil {
		utils.LogErrors(span, err)
		return nil, err
	}

	response := &dtos.RepeatInvoiceMaintenanceResponse{
		Results: results,
	}

	return response, nil
}

func (s *InvoiceMaintenanceService) BeginTransaction() *gorm.DB {
	return s.repo.BeginTransaction()
}

func (s *InvoiceMaintenanceService) Commit(tx *gorm.DB) error {
	return s.repo.Commit(tx)
}

func (s *InvoiceMaintenanceService) Rollback(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}

func (s *InvoiceMaintenanceService) ExcelGetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-ExcelGetInvoiceMaintenances", opentracing.ChildOf(span.Context()))

	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceMaintenances, _, err := s.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "InvoiceMaintenances"
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

	for i, invoice := range invoiceMaintenances {
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

func (s *InvoiceMaintenanceService) CsvGetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-CsvGetInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceMaintenances, _, err := s.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}
	companyProfile, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	appName := "App"
	if err == nil && companyProfile != nil && companyProfile.CompanyName != nil {
		appName = *companyProfile.CompanyName
	}

	csv := fmt.Sprintf("%s\n", appName)
	csv += "\n"
	csv += "Invoice Maintenances\n"
	csv += "\n"

	csv += "Customer,Invoice No,Title,Invoice Date,Due Date,Bank,Currency,Exchange Rate,VAT,PPh23,Qty,Sub Amount,DP Amount,Balance,Grand Total,Status,Created By\n"

	for _, invoice := range invoiceMaintenances {
		customerName := utils.GetPtrVal(invoice.CustomerName)
		invoiceNo := utils.GetPtrVal(invoice.InvoiceNo)
		title := utils.GetPtrVal(invoice.Title)
		invoiceDate := utils.GetPtrVal(invoice.InvoiceDate)
		dueDate := utils.GetPtrVal(invoice.DueDate)

		bankName := utils.GetPtrVal(invoice.BankName)
		accountNumber := utils.GetPtrVal(invoice.AccountNumber)
		accountName := utils.GetPtrVal(invoice.AccountName)

		bankInfo := bankName
		if accountNumber != "" {
			if bankInfo != "" {
				bankInfo += " - "
			}
			bankInfo += accountNumber
		}
		if accountName != "" {
			if bankInfo != "" {
				bankInfo += " - "
			}
			bankInfo += accountName
		}

		currencyName := utils.GetPtrVal(invoice.CurrencyName)
		totalVat := utils.GetFloatPtrVal(invoice.TotalVat)
		totalPph23 := utils.GetFloatPtrVal(invoice.TotalPph23)
		status := utils.GetPtrVal(invoice.Status)
		createdByName := utils.GetPtrVal(invoice.CreatedByName)

		exchangeRate := utils.GetFloatPtrVal(invoice.ExchangeRate)
		totalQty := utils.GetFloatPtrVal(invoice.TotalQty)
		subtotal := utils.GetFloatPtrVal(invoice.Subtotal)
		totalDpProducts := utils.GetFloatPtrVal(invoice.TotalDpProducts)
		totalBalanceProducts := utils.GetFloatPtrVal(invoice.TotalBalanceProducts)
		grandTotal := utils.GetFloatPtrVal(invoice.GrandTotal)

		customerName = utils.EscapeCsvField(customerName)
		invoiceNo = utils.EscapeCsvField(invoiceNo)
		title = utils.EscapeCsvField(title)
		bankInfo = utils.EscapeCsvField(bankInfo)
		currencyName = utils.EscapeCsvField(currencyName)
		status = utils.EscapeCsvField(status)
		createdByName = utils.EscapeCsvField(createdByName)

		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s,%s\n",
			customerName,
			invoiceNo,
			title,
			invoiceDate,
			dueDate,
			bankInfo,
			currencyName,
			exchangeRate,
			totalVat,
			totalPph23,
			totalQty,
			subtotal,
			totalDpProducts,
			totalBalanceProducts,
			grandTotal,
			status,
			createdByName,
		)
	}

	return []byte(csv), nil
}

func (s *InvoiceMaintenanceService) PublishBulkSendEmailApproved(ctx *fiber.Ctx, req dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-PublishBulkSendEmailApproved", opentracing.ChildOf(span.Context()))

	strIDs := make([]string, len(req.IDs))
	for i, id := range req.IDs {
		strIDs[i] = fmt.Sprint(id)
	}
	filters := map[string]string{
		// "invoice_maintenance_ids": req.IDs,
		"invoice_maintenance_ids": strings.Join(strIDs, ","),
		"is_csv":                  "1",
	}
	log.Println("PublishBulkSendEmailApproved-filters", filters)
	invoiceMaintenances, _, err := s.repo.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		return err
	}

	refType := "invoice_maintenances"

	// MapFormSentEmailSolution
	emails, refIDs, err := utils.MapBulkSendEmailApproved(invoiceMaintenances, req, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return err
	}

	// delete all emails with the same refIDs
	if tx, err := s.utilRepo.DeleteSentEmailsByRefIDs(ctx, tx, refIDs, refType, childSpan); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return err
	}

	// create sent_emails
	if tx, err := s.utilRepo.CreateSentEmails(ctx, tx, emails, childSpan); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return err
	}

	mappedEmails, err := utils.MapBulkSendEmailApprovedModelToDTO(ctx, emails, userID, branchID, childSpan)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return err
	}
	req.SentEmails = mappedEmails
	req.SenderID = userID
	// log.Println("PublishSendEmailSolutionTicket-req.SentEmailID", req.SentEmailID)
	// log.Println("PublishSendEmailSolutionTicket-email.ID", email.ID)

	err = utils.PublishBulkSendEmailApprovedInvoiceMaintenance(ctx, s.rabbitmq, req)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		return err
	}

	tx.Commit()

	fmt.Println("PublishBulkSendEmailApproved-emails", emails)
	fmt.Println("PublishBulkSendEmailApproved-refIDs", refIDs)
	fmt.Println("PublishBulkSendEmailApproved", req)

	return nil
}

func (s *InvoiceMaintenanceService) ConsumeBulkSendEmailApprovedInvoiceMaintenance(req dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest) error {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-ConsumeBulkSendEmailApprovedInvoiceMaintenance")

	// Start transaction early to ensure we can update status
	tx := s.repo.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	address := host + ":" + port

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	companyProfileParams := dtos.GetCompanyProfileParams{ID: 1}

	company, err := s.utilRepo.GetCompanyProfileByID(ctx, &companyProfileParams)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
	}
	fromEmail := *company.CompanyEmail
	fromEmailPassword := company.CompanyEmailPassword

	strIDs := make([]string, len(req.IDs))
	for i, id := range req.IDs {
		strIDs[i] = fmt.Sprint(id)
	}
	filters := map[string]string{
		"invoice_maintenance_ids": strings.Join(strIDs, ","),
		"is_csv":                  "1",
	}
	invoiceMaintenances, _, err := s.repo.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
		return err
	}

	log.Println("consumer-invoiceMaintenances", invoiceMaintenances)

	invoiceMaintenancesIDs := utils.MapGetInvoiceMaintenancesIDs(invoiceMaintenances)
	log.Println("consumer-invoiceMaintenancesIDs", invoiceMaintenancesIDs)
	if len(invoiceMaintenancesIDs) == 0 {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, fmt.Errorf("no invoice maintenances found"))
		tx.Rollback()
		return fmt.Errorf("no invoice maintenances found")
	}

	invoiceMaintenanceDts, _, err := s.repo.GetInvoiceMaintenanceDtsRawByIDs(ctx, filters, childSpan)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		tx.Rollback()
		return err
	}
	log.Println("consumer-invoiceMaintenanceDts", invoiceMaintenanceDts)

	invoiceMaintenances = utils.MapInvoiceMaintenancesDts(invoiceMaintenances, invoiceMaintenanceDts)
	// foreach the selected invoice maintenances
	log.Println("consumer-invoiceMaintenances2", invoiceMaintenances)

	for _, invoiceMaintenance := range invoiceMaintenances {
		// Create the multipart email message
		email := utils.GetSelectedEmailInvoiceMaintenance(invoiceMaintenance, req)
		selectedDts := utils.GetSelectedDtsInvoiceMaintenance(invoiceMaintenance, invoiceMaintenanceDts)

		isIDOnly := 1
		invoiceMaintenanceParam := dtos.InvoiceMaintenanceDetailNoBomDTO{
			ID:       invoiceMaintenance.ID,
			IsIDOnly: &isIDOnly,
		}
		pdfLink, err := s.Pdf(ctx, invoiceMaintenanceParam, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			utils.LogErrors(childSpan, err)
		}
		pdfLink = utils.MapStringToURL(pdfLink)

		attachments := []dtos.BulkSendEmailApprovedInvoiceMaintenancesEmailAttachment{}
		attachments = append(attachments, dtos.BulkSendEmailApprovedInvoiceMaintenancesEmailAttachment{
			Label: "Invoice Maintenance - " + *invoiceMaintenance.Title + ".pdf",
			Path:  *pdfLink,
		})

		to := email.ToEmail
		subject := ""
		subject = "Invoice Maintenance: " + *invoiceMaintenance.Title
		fromString := fmt.Sprintf("From: %s <%s>\r\n", *company.CompanyName, fromEmail)
		toString := fmt.Sprintf("To: Me <%s>\r\n", to)
		subjectString := fmt.Sprintf("Subject: %s\r\n", subject)

		// Update status function
		updateEmailStatus := func(status string, err error) error {

			refType := "invoice_maintenances"
			emailObject := dtos.FormSentEmailRequest{
				ID:           email.ID,
				RefID:        &invoiceMaintenance.ID,
				SenderID:     &req.SenderID,
				RefType:      &refType,
				FromEmail:    &fromEmail,
				ToEmail:      to,
				Subject:      &subject,
				ErrorMessage: utils.ErrorToStringPtr(err),
				Status:       &status,
				UpdatedByID:  &req.SenderID,
			}

			// MapFormSentEmailSolution
			emailModel := utils.MapBulkSendEmailApprovedSingle(&emailObject, req.SenderID, *invoiceMaintenance.BranchID, childSpan)
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
			}

			emailModel.Status = status
			if err != nil {
				errMsg := err.Error()
				emailModel.ErrorMessage = &errMsg
			}

			if _, err := s.utilRepo.UpdateSentEmail(tx, &emailModel, childSpan); err != nil {
				utils.LogErrors(childSpan, err)
				tx.Rollback()
				return err
			}

			return tx.Commit().Error
		}

		// Handle email sending process with error handling
		if err := func() error {

			// attachments := utils.MapAttachmentsTicket(ctx, req.SolutionAttachments, req.SelectedSolutionAttachments)

			sentAt := time.Now().Format("2006-01-02 15:04:05")
			data := dtos.BulkSendEmailApprovedInvoiceMaintenancesEmailData{
				Subject:     subject,
				Req:         invoiceMaintenance,
				Dts:         selectedDts,
				SentAt:      sentAt,
				Company:     company,
				Attachments: attachments,
			}

			// Read the embedded template file
			templateFile, err := templateFS.Open("templates/send-email-approved-invoice-maintenance.html")
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to open embedded template: %v", err)
				return err
			}
			log.Println("templateFile", templateFile)
			defer templateFile.Close()

			// Read the template content
			templateContent, err := io.ReadAll(templateFile)
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to read template content: %v", err)
				return err
			}

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
				},
				"inc": func(i int) int {
					return i + 1
				},
			}

			// Parse the template
			tmpl, err := template.New("email").Funcs(funcMap).Parse(string(templateContent))
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to parse template: %v", err)
				return err
			}

			// // Execute template with data
			var body bytes.Buffer
			if err := tmpl.Execute(&body, data); err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to execute template: %v", err)
				return err
			}

			var msg bytes.Buffer
			msg.WriteString(fromString)
			msg.WriteString(toString)
			msg.WriteString(subjectString)
			msg.WriteString("MIME-Version: 1.0\r\n")
			msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
			msg.WriteString("\r\n")
			msg.Write(body.Bytes())

			// TLS config (important for port 465)
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true, // For testing ONLY (remove in production)
				ServerName:         host,
			}

			// Set up authentication
			auth := smtp.PlainAuth("", fromEmail, *fromEmailPassword, host)
			log.Println("auth", auth)
			// Connect to the server via TLS
			conn, err := tls.Dial("tcp", address, tlsConfig)
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				return err
			}

			// Create a new SMTP client
			client, err := smtp.NewClient(conn, host)
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to create SMTP client: %v", err)
				return err
			}
			defer client.Quit()

			if err := client.Auth(auth); err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("SMTP authentication failed: %v", err)
				return err
			}

			// Set sender and recipient
			if err := client.Mail(fromEmail); err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to set sender: %v", err)
				return err
			}
			if err := client.Rcpt(to); err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to set recipient: %v", err)
				return err
			}

			// Write the email data
			w, err := client.Data()
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to get Data writer: %v", err)
				return err
			}
			defer w.Close()

			_, err = w.Write(msg.Bytes())
			if err != nil {
				defer childSpan.Finish()
				utils.LogErrors(childSpan, err)
				log.Printf("Failed to write email data: %v", err)
				return err
			}
			// return w.Close()
			// // Send the email
			// err = smtp.SendMail(
			// 	host+":"+port,
			// 	auth,
			// 	from,
			// 	[]string{to},
			// 	msg.Bytes(),
			// )

			log.Println("Email sent successfully!")
			return nil
		}(); err != nil {
			// If there's an error, update status to FAILED
			if updateErr := updateEmailStatus("FAILED", err); updateErr != nil {
				return fmt.Errorf("failed to update email status: %v (original error: %v)", updateErr, err)
			}
			return err
		}
		// If successful, update status to SUCCESS
		if err := updateEmailStatus("SUCCESS", nil); err != nil {
			return fmt.Errorf("failed to update success status: %v", err)
		}
	}

	return nil
}

// PdfGetQuotations
func (s *InvoiceMaintenanceService) Pdf(ctx *fiber.Ctx, req dtos.InvoiceMaintenanceDetailNoBomDTO, tx *gorm.DB, span opentracing.Span) (*string, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-Pdf", opentracing.ChildOf(span.Context()))

	form := req
	var invoiceMaintenance *dtos.InvoiceMaintenanceDetailNoBomDTO
	var err error

	// req.IsIDOnly != nil
	var params dtos.GetInvoiceMaintenanceParams
	if req.IsIDOnly != nil && *req.IsIDOnly == 1 {
		params.ID = req.ID

		invoiceMaintenance, err = s.repo.GetInvoiceMaintenanceByNoBomID(ctx, &params, tx, childSpan)
		if err != nil {
			defer childSpan.Finish()
			return nil, err
		}

		// createdInvoiceMaintenanceIDs := make([]uint, 0)
		// createdInvoiceMaintenanceIDs = append(createdInvoiceMaintenanceIDs, invoiceMaintenance.ID)
		dtsFilters := map[string]string{
			"invoice_maintenance_ids": fmt.Sprintf("%d", invoiceMaintenance.ID),
			"is_csv":                  "1",
		}

		invoiceMaintenanceDts, err := s.GetInvoiceMaintenanceDtsByID(ctx, dtsFilters, tx, childSpan)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch invoiceMaintenanceDts: %v", err)
		}

		companyParams := &dtos.GetCompanyProfileParams{ID: uint(*invoiceMaintenance.CompanyProfileID)}
		invoiceMaintenance.InvoiceMaintenanceDts = invoiceMaintenanceDts
		company, err := s.utilRepo.GetCompanyProfileByID(ctx, companyParams)
		if err != nil {
			utils.LogErrors(childSpan, err)
			log.Printf("Failed to fetch company: %v", err)
		}

		invoiceMaintenance.Company = *company

		form = *invoiceMaintenance
		req.InvoiceNo = invoiceMaintenance.InvoiceNo
	}

	var num string
	if req.InvoiceNo != nil {
		num = *req.InvoiceNo
	} else {
		num = ""
	}

	uploadDir := "./public/generated_pdfs"
	// Ensure the directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	replacedTitle := strings.ReplaceAll(*form.Title, "/", "_")
	form.Title = &replacedTitle

	fileName := fmt.Sprintf("%s-%s.pdf", *form.Title, time.Now().Format("20060102150405"))
	pdfPath := filepath.Join(uploadDir, fileName)
	pdfPublicPath := utils.MapStringToURL(&pdfPath)

	uploadDir = "./public/barcodes"
	// Ensure the directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		utils.LogErrors(childSpan, err)
		log.Println("Error mkdirall:", err)
		return nil, err
	}

	// Create the barcode
	qrCodeFilename := fmt.Sprintf("barcodes-%s-%s.png", *form.Title, time.Now().Format("20060102150405"))
	qrCodePath := filepath.Join(uploadDir, qrCodeFilename)
	qrCodePublicPath := utils.MapStringToURL(&qrCodePath)
	// log.Println("qrCodePublicPath", *qrCodePublicPath)

	grandTotal := utils.FormatNumberSeparator(form.GrandTotal)
	qrCodeContent := fmt.Sprintf("To: %s\nInvoice No: %s\nGrand Total: %s.%s\nLink: %s",
		*form.CustomerName,
		num,
		*form.CurrencyName,
		grandTotal,
		*pdfPublicPath,
	)

	qrCode, _ := qr.Encode(qrCodeContent, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 600, 600)

	// create the output file
	qrCodeImg, err := os.Create(qrCodePath)
	if err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error creating qrCodeImg:", err)
		return nil, err
	}
	// Only defer Close() after we know the file was created successfully
	defer func() {
		if qrCodeImg != nil {
			qrCodeImg.Close()
		}
	}()

	// encode the barcode as png
	png.Encode(qrCodeImg, qrCode)

	if form.ApprovedStatus != nil && *form.ApprovedStatus != "APPROVED" {
		qrCodePublicPath = utils.EmptyStringPointer(qrCodePublicPath)
	}

	data := dtos.InvoiceMaintenancePDFData{
		Num:    num,
		Form:   form,
		QrCode: *qrCodePublicPath,
	}

	htmlFileName := "invoice-maintenance-detail"
	// log.Println("Pdf-htmlFileName-im", htmlFileName)

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

	if err := pdfg.WriteFile(pdfPath); err != nil {
		defer childSpan.Finish()
		utils.LogErrors(childSpan, err)
		log.Println("Error pdf path:", err)
		return nil, err
	}

	return &pdfPath, nil
}
