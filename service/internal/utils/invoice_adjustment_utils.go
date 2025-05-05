package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func GetInvoiceAdjustmentIDs(req dtos.UpdateInvoiceAdjustmentRequest) ([]*uint, []*uint) {
	adjustmentDtIDs := []*uint{}
	refIDs := []*uint{}

	for _, reqAdjustmentDt := range req.AdjustmentDts {
		if reqAdjustmentDt.ID != nil && *reqAdjustmentDt.ID > 0 {
			adjustmentDtIDs = append(adjustmentDtIDs, reqAdjustmentDt.ID)
		}
		if reqAdjustmentDt.RefID != nil && *reqAdjustmentDt.RefID > 0 {
			refIDs = append(refIDs, reqAdjustmentDt.RefID)
		}
	}

	return adjustmentDtIDs, refIDs
}

func GetLockInvoiceAdjustmentReferenceIDs(req dtos.CreateInvoiceAdjustmentRequest) []*uint {
	refIDs := []*uint{}

	for _, reqAdjustmentDt := range req.AdjustmentDts {
		if reqAdjustmentDt.RefID != nil && *reqAdjustmentDt.RefID > 0 {
			refIDs = append(refIDs, reqAdjustmentDt.RefID)
		}
	}

	return refIDs
}

func MapCreateInvoiceAdjustmentDts(ctx *fiber.Ctx, req dtos.CreateInvoiceAdjustmentRequest, createdInvoiceAdjustment *models.InvoiceAdjustment, userID uint, span opentracing.Span) ([]models.InvoiceAdjustmentDt, error) {
	adjustmentDtsModel := []models.InvoiceAdjustmentDt{}

	for _, adjustmentDt := range req.AdjustmentDts {
		refJSONStr := "{}"
		refJSON := json.RawMessage(refJSONStr)

		var invoiceDate *time.Time
		if adjustmentDt.InvoiceDate != nil {
			parsedTime, err := time.Parse("2006-01-02", *adjustmentDt.InvoiceDate)
			if err != nil {
				return []models.InvoiceAdjustmentDt{}, err
			}
			invoiceDate = &parsedTime
		}

		adjustmentDtModel := models.InvoiceAdjustmentDt{
			InvoiceUUID:         adjustmentDt.InvoiceUUID,
			InvoiceAdjustmentID: &createdInvoiceAdjustment.ID,
			RefID:               adjustmentDt.RefID,
			RefType:             adjustmentDt.RefType,
			RefJSON:             &refJSON,
			InvoiceNo:           adjustmentDt.InvoiceNo,
			InvoiceDate:         invoiceDate,
			InvoiceAmount:       adjustmentDt.InvoiceAmount,
			TotalAdjustment:     adjustmentDt.TotalAdjustment,
			BalanceAmount:       adjustmentDt.BalanceAmount,
			AdjustmentAmount:    adjustmentDt.AdjustmentAmount,
			AdminBank:           adjustmentDt.AdminBank,
			TotalAmount:         adjustmentDt.TotalAmount,
			CreatedByID:         &userID,
		}
		adjustmentDtsModel = append(adjustmentDtsModel, adjustmentDtModel)
	}

	return adjustmentDtsModel, nil
}

func MapUpdateInvoiceAdjustmentDts(ctx *fiber.Ctx, req dtos.UpdateInvoiceAdjustmentRequest, updatedInvoiceAdjustment *models.InvoiceAdjustment, userID uint, span opentracing.Span) ([]models.InvoiceAdjustmentDt, error) {
	adjustmentDtsModel := []models.InvoiceAdjustmentDt{}

	refJSONStr := "{}"
	refJSON := json.RawMessage(refJSONStr)

	for _, reqAdjustmentDt := range req.AdjustmentDts {
		adjustmentDtID := uint(0)
		if reqAdjustmentDt.ID != nil {
			adjustmentDtID = *reqAdjustmentDt.ID
		}

		var invoiceDate *time.Time
		if reqAdjustmentDt.InvoiceDate != nil {
			parsedTime, err := time.Parse("2006-01-02", *reqAdjustmentDt.InvoiceDate)
			if err != nil {
				return []models.InvoiceAdjustmentDt{}, err
			}
			invoiceDate = &parsedTime
		}

		adjustmentDtModel := models.InvoiceAdjustmentDt{
			ID:                  adjustmentDtID,
			InvoiceUUID:         reqAdjustmentDt.InvoiceUUID,
			InvoiceAdjustmentID: &updatedInvoiceAdjustment.ID,
			RefID:               reqAdjustmentDt.RefID,
			RefType:             reqAdjustmentDt.RefType,
			RefJSON:             &refJSON,
			InvoiceNo:           reqAdjustmentDt.InvoiceNo,
			InvoiceDate:         invoiceDate,
			InvoiceAmount:       reqAdjustmentDt.InvoiceAmount,
			TotalAdjustment:     reqAdjustmentDt.TotalAdjustment,
			BalanceAmount:       reqAdjustmentDt.BalanceAmount,
			AdjustmentAmount:    reqAdjustmentDt.AdjustmentAmount,
			AdminBank:           reqAdjustmentDt.AdminBank,
			TotalAmount:         reqAdjustmentDt.TotalAmount,
			CreatedByID:         &userID,
		}
		adjustmentDtsModel = append(adjustmentDtsModel, adjustmentDtModel)
	}

	return adjustmentDtsModel, nil
}

func GenInvoiceAdjustmentNo(ctx *fiber.Ctx, req dtos.CreateInvoiceAdjustmentRequest, orderedNumber int, span opentracing.Span) string {
	// if req.InvoiceNo != nil {
	// 	return *req.InvoiceNo
	// }

	prefix := "IAD"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateInvoiceAdjustmentNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateInvoiceAdjustmentRequest, revNo *int, span opentracing.Span) string {
	if req.InvoiceNo == nil {
		prefix := "IAD"
		year := time.Now().Format("2006")
		month := time.Now().Format("01")
		day := time.Now().Format("02")

		return fmt.Sprintf("%s/REV-%d/%s-%s-%s", prefix, *revNo, year, month, day)
	}

	invoiceNo := *req.InvoiceNo

	if !strings.Contains(invoiceNo, "REV") {
		return fmt.Sprintf("%s/REV-%d", invoiceNo, *revNo)
	} else {
		basePart := strings.Split(invoiceNo, "/REV")[0]
		return fmt.Sprintf("%s/REV-%d", basePart, *revNo)
	}
}

func MapCreateInvoiceAdjustment(ctx *fiber.Ctx, req dtos.CreateInvoiceAdjustmentRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceAdjustment, error) {
	invoiceNo := GenInvoiceAdjustmentNo(ctx, req, orderedNumber, span)

	var paymentDate *time.Time
	if req.PaymentDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		paymentDate = &parsedTime
	}

	var refStartDate *time.Time
	if req.RefStartDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RefStartDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		refStartDate = &parsedTime
	}

	var refEndDate *time.Time
	if req.RefEndDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RefEndDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		refEndDate = &parsedTime
	}

	invoiceAdjustment := models.InvoiceAdjustment{
		CustomerID:      req.CustomerID,
		CurrencyID:      req.CurrencyID,
		BranchID:        &branchID,
		BankID:          req.BankID,
		InvoiceNo:       &invoiceNo,
		PaymentDate:     paymentDate,
		PaymentAmount:   req.PaymentAmount,
		ExchangeRate:    req.ExchangeRate,
		Reference:       req.Reference,
		RefStartDate:    refStartDate,
		RefEndDate:      refEndDate,
		Remark:          req.Remark,
		RevNo:           new(int),
		TotalInvoice:    req.TotalInvoice,
		TotalAdjustment: req.TotalAdjustment,
		TotalBalance:    req.TotalBalance,
		TotalAdminBank:  req.TotalAdminBank,
		GrandTotal:      req.GrandTotal,
		CreatedByID:     &userID,
	}
	*invoiceAdjustment.RevNo = 0

	return invoiceAdjustment, nil
}

func MapUpdateInvoiceAdjustment(ctx *fiber.Ctx, req dtos.UpdateInvoiceAdjustmentRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.InvoiceAdjustment, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := GenerateInvoiceAdjustmentNoOnUpdate(ctx, req, &revNo, span)

	var paymentDate *time.Time
	if req.PaymentDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		paymentDate = &parsedTime
	}

	var refStartDate *time.Time
	if req.RefStartDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RefStartDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		refStartDate = &parsedTime
	}

	var refEndDate *time.Time
	if req.RefEndDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RefEndDate)
		if err != nil {
			return models.InvoiceAdjustment{}, err
		}
		refEndDate = &parsedTime
	}

	invoiceAdjustment := models.InvoiceAdjustment{
		ID:              req.ID,
		CustomerID:      req.CustomerID,
		CurrencyID:      req.CurrencyID,
		BranchID:        &branchID,
		BankID:          req.BankID,
		InvoiceNo:       &invoiceNo,
		PaymentDate:     paymentDate,
		PaymentAmount:   req.PaymentAmount,
		ExchangeRate:    req.ExchangeRate,
		Reference:       req.Reference,
		RefStartDate:    refStartDate,
		RefEndDate:      refEndDate,
		Remark:          req.Remark,
		RevNo:           &revNo,
		TotalInvoice:    req.TotalInvoice,
		TotalAdjustment: req.TotalAdjustment,
		TotalBalance:    req.TotalBalance,
		TotalAdminBank:  req.TotalAdminBank,
		GrandTotal:      req.GrandTotal,
		UpdatedByID:     &userID,
	}

	return invoiceAdjustment, nil
}

func GetReferenceInvoiceIDs(refInvoices []dtos.ReferenceInvoiceListDTO) []uint {
	refInvoiceIDs := []uint{}

	for _, refInvoice := range refInvoices {
		if refInvoice.ID != nil {
			refInvoiceIDs = append(refInvoiceIDs, *refInvoice.ID)
		}
	}

	return refInvoiceIDs
}
