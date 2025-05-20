package service

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type InvoiceMaintenanceService struct {
	repo     *repository.InvoiceMaintenanceRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewInvoiceMaintenanceService(repo *repository.InvoiceMaintenanceRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *InvoiceMaintenanceService {
	return &InvoiceMaintenanceService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *InvoiceMaintenanceService) GetInvoiceMaintenances(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceMaintenanceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceMaintenanceService-GetInvoiceMaintenances", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceMaintenances, total, err := s.repo.GetInvoiceMaintenances(ctx, filters, childSpan)
	if err != nil {
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

	tx, err = s.repo.UpdateInvoiceMaintenance(tx, &invoiceMaintenance, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
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

func (s *InvoiceMaintenanceService) PublishBulkSendEmailSolutionTicket(ctx *fiber.Ctx, req dtos.FormTicketRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) error {
	// childSpan := opentracing.StartSpan("TicketSeInvoiceMaintenanceServicervice-PublishBulkSendEmailSolutionTicket", opentracing.ChildOf(span.Context()))

	// log sent_emails
	// make *dtos.FormSentEmailRequest

	// refType := "tickets"
	// status := "PROCESS"
	// emailObject := dtos.FormSentEmailRequest{
	// 	RefType: &refType,
	// 	Status:  &status,
	// }

	// // MapFormSentEmailSolution
	// email, err := utils.MapFormSentEmailSolution(ctx, req, &emailObject, userID, branchID, childSpan)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	utils.LogErrors(childSpan, err)
	// 	tx.Rollback()
	// 	return err
	// }

	// // create sent_emails
	// if tx, err := s.repo.CreateSentEmail(tx, email, childSpan); err != nil {
	// 	defer childSpan.Finish()
	// 	utils.LogErrors(childSpan, err)
	// 	tx.Rollback()
	// 	return err
	// }

	// req.SentEmailID = &email.ID
	// log.Println("PublishSendEmailSolutionTicket-req.SentEmailID", req.SentEmailID)
	// log.Println("PublishSendEmailSolutionTicket-email.ID", email.ID)

	// err = utils.PublishSendEmailSolutionTicket(ctx, s.rabbitmq, req)
	// if err != nil {
	// 	defer childSpan.Finish()
	// 	utils.LogErrors(childSpan, err)
	// 	return err
	// }

	// tx.Commit()

	// fmt.Println("PublishSendEmailSolutionTicket", req)

	return nil
}
