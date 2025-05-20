package service

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type InvoiceAdjustmentService struct {
	repo     *repository.InvoiceAdjustmentRepository
	utilRepo *repository.UtilRepository
	tracer   opentracing.Tracer
}

func NewInvoiceAdjustmentService(repo *repository.InvoiceAdjustmentRepository, utilRepo *repository.UtilRepository, tracer opentracing.Tracer) *InvoiceAdjustmentService {
	return &InvoiceAdjustmentService{
		repo:     repo,
		utilRepo: utilRepo,
		tracer:   tracer,
	}
}

func (s *InvoiceAdjustmentService) GetInvoiceAdjustments(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.InvoiceAdjustmentListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-GetInvoiceAdjustments", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceAdjustments, total, err := s.repo.GetInvoiceAdjustments(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return invoiceAdjustments, total, nil
}

func (s *InvoiceAdjustmentService) CreateInvoiceAdjustment(ctx *fiber.Ctx, req dtos.CreateInvoiceAdjustmentRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceAdjustment, *gorm.DB, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-CreateInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	refIDsByType := make(map[string][]uint)
	for _, dt := range req.AdjustmentDts {
		if dt.RefType != nil && dt.RefID != nil && *dt.RefID > 0 {
			refType := *dt.RefType
			if _, exists := refIDsByType[refType]; !exists {
				refIDsByType[refType] = make([]uint, 0)
			}
			refIDsByType[refType] = append(refIDsByType[refType], *dt.RefID)
		}
	}

	for refType, refIDs := range refIDsByType {
		if err := s.repo.LockReferenceInvoices(tx, refIDs, refType, childSpan); err != nil {
			tx.Rollback()
			return nil, tx, err
		}
	}

	invoiceAdjustmentCreatedThisMonthNumber, err := s.repo.GetInvoiceAdjustmentCreatedThisMonth(ctx, tx, *req.CustomerID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	orderedNumber := invoiceAdjustmentCreatedThisMonthNumber + 1

	invoiceAdjustment, err := utils.MapCreateInvoiceAdjustment(ctx, req, userID, branchID, orderedNumber, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, err = s.repo.CreateInvoiceAdjustment(tx, &invoiceAdjustment, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	tx, _, err = s.CreateInvoiceAdjustmentDts(ctx, req, userID, &invoiceAdjustment, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, tx, err
	}

	for _, dt := range req.AdjustmentDts {
		if dt.RefType != nil && dt.RefID != nil && dt.AdjustmentAmount != nil && *dt.AdjustmentAmount > 0 {
			tx, err = s.repo.UpdateReferenceInvoiceAdjustmentAmount(tx, *dt.RefID, *dt.RefType, *dt.AdjustmentAmount, childSpan)
			if err != nil {
				tx.Rollback()
				return nil, tx, err
			}
		}
	}

	return &invoiceAdjustment, tx, nil
}

func (s *InvoiceAdjustmentService) GetInvoiceAdjustmentByID(ctx *fiber.Ctx, params *dtos.GetInvoiceAdjustmentParams, tx *gorm.DB, span opentracing.Span) (*dtos.InvoiceAdjustmentDetailDTO, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-GetInvoiceAdjustmentByID", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceAdjustment, err := s.repo.GetInvoiceAdjustmentByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	adjustmentDts, err := s.repo.GetInvoiceAdjustmentDts(ctx, invoiceAdjustment.ID, params.IsDeleted, childSpan)
	if err != nil {
		return nil, err
	}

	invoiceAdjustment.AdjustmentDts = adjustmentDts

	return invoiceAdjustment, nil
}

func (s *InvoiceAdjustmentService) UpdateInvoiceAdjustment(ctx *fiber.Ctx, req dtos.UpdateInvoiceAdjustmentRequest, userID uint, branchID uint, tx *gorm.DB, span opentracing.Span) (*models.InvoiceAdjustment, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-UpdateInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.NewGetInvoiceAdjustmentParams(req.ID)
	existingInvoiceAdjustment, err := s.GetInvoiceAdjustmentByID(ctx, params, tx, childSpan)
	if err != nil {
		return nil, err
	}

	if err := s.repo.LockInvoiceAdjustment(tx, req.ID, childSpan); err != nil {
		return nil, err
	}

	existingAdjustmentDts, err := s.repo.GetInvoiceAdjustmentDts(ctx, req.ID, nil, childSpan)
	if err != nil {
		return nil, err
	}

	adjustmentDtIDs, _ := utils.GetInvoiceAdjustmentIDs(req)

	refIDsByType := make(map[string][]uint)
	for _, dt := range req.AdjustmentDts {
		if dt.RefID != nil && dt.RefType != nil {
			refIDsByType[*dt.RefType] = append(refIDsByType[*dt.RefType], *dt.RefID)
		}
	}

	for refType, ids := range refIDsByType {
		if err := s.repo.LockReferenceInvoices(tx, ids, refType, childSpan); err != nil {
			return nil, err
		}
	}

	deleteAdjustmentDtIDs := []uint{}
	for _, existingDt := range existingAdjustmentDts {
		found := false
		for _, reqDtID := range adjustmentDtIDs {
			if existingDt.ID != nil && *existingDt.ID == *reqDtID {
				found = true
				break
			}
		}
		if !found && existingDt.ID != nil {
			deleteAdjustmentDtIDs = append(deleteAdjustmentDtIDs, *existingDt.ID)
		}
	}

	if len(deleteAdjustmentDtIDs) > 0 {
		tx, err = s.repo.DeleteInvoiceAdjustmentDtsByIDs(tx, deleteAdjustmentDtIDs, userID, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	invoiceAdjustment, err := utils.MapUpdateInvoiceAdjustment(ctx, req, userID, branchID, existingInvoiceAdjustment.RevNo, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.repo.UpdateInvoiceAdjustment(tx, &invoiceAdjustment, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updateAdjustmentDts := []models.InvoiceAdjustmentDt{}
	createAdjustmentDts := []models.InvoiceAdjustmentDt{}

	adjustmentDtsModel, err := utils.MapUpdateInvoiceAdjustmentDts(ctx, req, &invoiceAdjustment, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, dt := range adjustmentDtsModel {
		if dt.ID > 0 {
			updateAdjustmentDts = append(updateAdjustmentDts, dt)
		} else {
			createAdjustmentDts = append(createAdjustmentDts, dt)
		}
	}

	if len(updateAdjustmentDts) > 0 {
		tx, err = s.repo.BulkUpdateInvoiceAdjustmentDts(tx, updateAdjustmentDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if len(createAdjustmentDts) > 0 {
		tx, err = s.repo.BulkCreateInvoiceAdjustmentDts(tx, createAdjustmentDts, childSpan)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	for _, dtID := range deleteAdjustmentDtIDs {
		var detail dtos.InvoiceAdjustmentDtListDTO
		for _, dt := range existingAdjustmentDts {
			if dt.ID != nil && *dt.ID == dtID {
				detail = dt
				break
			}
		}

		if detail.RefID != nil && detail.RefType != nil && detail.AdjustmentAmount != nil && *detail.AdjustmentAmount > 0 {
			negativeAmount := -*detail.AdjustmentAmount
			tx, err = s.repo.UpdateReferenceInvoiceAdjustmentAmount(tx, *detail.RefID, *detail.RefType, negativeAmount, childSpan)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	for _, dt := range req.AdjustmentDts {
		if dt.RefID != nil && dt.RefType != nil && dt.AdjustmentAmount != nil {
			var existingAmount float64 = 0
			var isExisting bool = false

			for _, existingDt := range existingAdjustmentDts {
				if existingDt.ID != nil && dt.ID != nil && *existingDt.ID == *dt.ID {
					if existingDt.AdjustmentAmount != nil {
						existingAmount = *existingDt.AdjustmentAmount
					}
					isExisting = true
					break
				}
			}

			var adjustmentDifference float64
			if isExisting {
				adjustmentDifference = *dt.AdjustmentAmount - existingAmount
			} else {
				adjustmentDifference = *dt.AdjustmentAmount
			}

			if adjustmentDifference != 0 {
				tx, err = s.repo.UpdateReferenceInvoiceAdjustmentAmount(tx, *dt.RefID, *dt.RefType, adjustmentDifference, childSpan)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}
	}

	return &invoiceAdjustment, nil
}

func (s *InvoiceAdjustmentService) DeleteInvoiceAdjustment(ctx *fiber.Ctx, invoiceAdjustmentID uint, userID uint, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-DeleteInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	params := dtos.GetInvoiceAdjustmentParams{ID: invoiceAdjustmentID}
	invoiceAdjustment, err := s.GetInvoiceAdjustmentByID(ctx, &params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	refIDsByType := make(map[string][]uint)
	for _, dt := range invoiceAdjustment.AdjustmentDts {
		if dt.RefType != nil && dt.RefID != nil {
			refType := *dt.RefType
			if _, exists := refIDsByType[refType]; !exists {
				refIDsByType[refType] = make([]uint, 0)
			}
			refIDsByType[refType] = append(refIDsByType[refType], *dt.RefID)
		}
	}

	for refType, refIDs := range refIDsByType {
		if err := s.repo.LockReferenceInvoices(tx, refIDs, refType, childSpan); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := s.repo.LockInvoiceAdjustment(tx, invoiceAdjustmentID, childSpan); err != nil {
		tx.Rollback()
		return err
	}

	tx, err = s.repo.DeleteInvoiceAdjustment(tx, invoiceAdjustmentID, userID, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, dt := range invoiceAdjustment.AdjustmentDts {
		if dt.RefType != nil && dt.RefID != nil && dt.AdjustmentAmount != nil && *dt.AdjustmentAmount > 0 {
			negativeAmount := -*dt.AdjustmentAmount
			tx, err = s.repo.UpdateReferenceInvoiceAdjustmentAmount(tx, *dt.RefID, *dt.RefType, negativeAmount, childSpan)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func (s *InvoiceAdjustmentService) RestoreInvoiceAdjustment(ctx *fiber.Ctx, params *dtos.GetInvoiceAdjustmentParams, tx *gorm.DB, span opentracing.Span) error {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-RestoreInvoiceAdjustment", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	invoiceAdjustment, err := s.repo.RestoreInvoiceAdjustment(ctx, params, tx, childSpan)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, dt := range invoiceAdjustment.AdjustmentDts {
		if dt.RefType != nil && dt.RefID != nil {
			if err := s.repo.LockReferenceInvoices(tx, []uint{*dt.RefID}, *dt.RefType, childSpan); err != nil {
				tx.Rollback()
				return err
			}

			tx, err = s.repo.RestoreReferenceInvoiceAdjustmentAmount(tx, *dt.RefID, *dt.RefType, childSpan)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func (s *InvoiceAdjustmentService) CreateInvoiceAdjustmentDts(ctx *fiber.Ctx, req dtos.CreateInvoiceAdjustmentRequest, userID uint, invoiceAdjustment *models.InvoiceAdjustment, tx *gorm.DB, span opentracing.Span) (*gorm.DB, []models.InvoiceAdjustmentDt, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-CreateInvoiceAdjustmentDts", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	adjustmentDts, err := utils.MapCreateInvoiceAdjustmentDts(ctx, req, invoiceAdjustment, userID, childSpan)
	if err != nil {
		return tx, nil, err
	}

	tx, createdAdjustmentDts, err := s.repo.CreateInvoiceAdjustmentDts(tx, adjustmentDts, childSpan)
	if err != nil {
		return tx, nil, err
	}

	return tx, createdAdjustmentDts, nil
}

func (s *InvoiceAdjustmentService) GetReferenceInvoices(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]dtos.ReferenceInvoiceListDTO, int, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-GetReferenceInvoices", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	referenceInvoices, total, err := s.repo.GetReferenceInvoices(ctx, filters, childSpan)
	if err != nil {
		return nil, 0, err
	}
	return referenceInvoices, total, nil
}

func (s *InvoiceAdjustmentService) ExcelGetInvoiceAdjustments(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-ExcelGetInvoiceAdjustments", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceAdjustments, _, err := s.GetInvoiceAdjustments(ctx, filters, childSpan)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()

	sheetName := "InvoiceAdjustments"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	headers := []string{
		"Customer", "Invoice No", "Title", "Adjustment Date", "Payment Date",
		"Bank", "Currency", "Exchange Rate", "Payment Amount", "Total Invoice",
		"Total Adjustment", "Total Balance", "Admin Bank", "Grand Total", "Created By",
	}
	file.SetSheetRow(sheetName, "A1", &headers)

	file.SetColWidth(sheetName, "A", "A", 25)
	file.SetColWidth(sheetName, "B", "B", 15)
	file.SetColWidth(sheetName, "C", "C", 30)
	file.SetColWidth(sheetName, "D", "E", 15)
	file.SetColWidth(sheetName, "F", "F", 20)
	file.SetColWidth(sheetName, "G", "G", 15)
	file.SetColWidth(sheetName, "H", "H", 15)
	file.SetColWidth(sheetName, "I", "N", 15)
	file.SetColWidth(sheetName, "O", "O", 20)

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

	for i, adjustment := range invoiceAdjustments {
		row := i + 2
		file.SetSheetRow(sheetName, fmt.Sprintf("A%d", row), &[]interface{}{
			adjustment.CustomerName,
			adjustment.InvoiceNo,
			adjustment.Title,
			adjustment.AdjustmentDate,
			adjustment.PaymentDate,
			adjustment.BankName,
			adjustment.CurrencyName,
			adjustment.ExchangeRate,
			adjustment.PaymentAmount,
			adjustment.TotalInvoice,
			adjustment.TotalAdjustment,
			adjustment.TotalBalance,
			adjustment.TotalAdminBank,
			adjustment.GrandTotal,
			adjustment.CreatedByName,
		})

		file.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), currencyStyle)
		file.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("N%d", row), currencyStyle)
	}

	file.SetActiveSheet(index)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *InvoiceAdjustmentService) CsvGetInvoiceAdjustments(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]byte, error) {
	childSpan := opentracing.StartSpan("InvoiceAdjustmentService-CsvGetInvoiceAdjustments", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	filters["is_csv"] = "1"
	invoiceAdjustments, _, err := s.GetInvoiceAdjustments(ctx, filters, childSpan)
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
	csv += "Invoice Adjustments\n"
	csv += "\n"

	csv += "Customer,Invoice No,Title,Adjustment Date,Payment Date,Bank,Currency,Exchange Rate,Payment Amount,Total Invoice,Total Adjustment,Total Balance,Admin Bank,Grand Total,Created By\n"

	for _, adjustment := range invoiceAdjustments {
		customerName := utils.GetPtrVal(adjustment.CustomerName)
		invoiceNo := utils.GetPtrVal(adjustment.InvoiceNo)
		title := utils.GetPtrVal(adjustment.Title)
		adjustmentDate := utils.GetPtrVal(adjustment.AdjustmentDate)
		paymentDate := utils.GetPtrVal(adjustment.PaymentDate)

		bankName := utils.GetPtrVal(adjustment.BankName)
		accountNumber := utils.GetPtrVal(adjustment.AccountNumber)
		accountName := utils.GetPtrVal(adjustment.AccountName)

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

		currencyName := utils.GetPtrVal(adjustment.CurrencyName)
		adminBank := utils.GetFloatPtrVal(adjustment.TotalAdminBank)
		createdByName := utils.GetPtrVal(adjustment.CreatedByName)

		exchangeRate := utils.GetFloatPtrVal(adjustment.ExchangeRate)
		paymentAmount := utils.GetFloatPtrVal(adjustment.PaymentAmount)
		totalInvoice := utils.GetFloatPtrVal(adjustment.TotalInvoice)
		totalAdjustment := utils.GetFloatPtrVal(adjustment.TotalAdjustment)
		totalBalance := utils.GetFloatPtrVal(adjustment.TotalBalance)
		grandTotal := utils.GetFloatPtrVal(adjustment.GrandTotal)

		customerName = utils.EscapeCsvField(customerName)
		invoiceNo = utils.EscapeCsvField(invoiceNo)
		title = utils.EscapeCsvField(title)
		bankInfo = utils.EscapeCsvField(bankInfo)
		currencyName = utils.EscapeCsvField(currencyName)
		createdByName = utils.EscapeCsvField(createdByName)

		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s\n",
			customerName,
			invoiceNo,
			title,
			adjustmentDate,
			paymentDate,
			bankInfo,
			currencyName,
			exchangeRate,
			paymentAmount,
			totalInvoice,
			totalAdjustment,
			totalBalance,
			adminBank,
			grandTotal,
			createdByName,
		)
	}

	return []byte(csv), nil
}

func (s *InvoiceAdjustmentService) BeginTransaction() *gorm.DB {
	return s.repo.BeginTransaction()
}

func (s *InvoiceAdjustmentService) Commit(tx *gorm.DB) error {
	return s.repo.Commit(tx)
}

func (s *InvoiceAdjustmentService) Rollback(tx *gorm.DB) {
	tx.Rollback()
}
