package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func GetInvoiceMaintenanceIDs(req dtos.UpdateInvoiceMaintenanceRequest) ([]*uint, []*uint, []*uint) {
	invoiceMaintenanceDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		if reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID != nil && *reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID > 0 {
			invoiceMaintenanceDtIDs = append(invoiceMaintenanceDtIDs, reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID)
		}
	}

	return invoiceMaintenanceDtIDs, productIDs, itemUnitIDs
}

func GetLockInvoiceMaintenanceSalesOrderIDs(req dtos.CreateInvoiceMaintenanceRequest) ([]*uint, []*uint) {
	soIDs := []*uint{}
	soDtIDs := []*uint{}

	soIDsMap := make(map[uint]bool)

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		if reqInvoiceMaintenanceDt.RefType != nil && *reqInvoiceMaintenanceDt.RefType == "so" {
			if reqInvoiceMaintenanceDt.RefID != nil && *reqInvoiceMaintenanceDt.RefID > 0 {
				if !soIDsMap[*reqInvoiceMaintenanceDt.RefID] {
					soIDsMap[*reqInvoiceMaintenanceDt.RefID] = true
					soIDs = append(soIDs, reqInvoiceMaintenanceDt.RefID)
				}
			}

			if reqInvoiceMaintenanceDt.RefDtID != nil && *reqInvoiceMaintenanceDt.RefDtID > 0 {
				soDtIDs = append(soDtIDs, reqInvoiceMaintenanceDt.RefDtID)
			}
		}
	}

	return soIDs, soDtIDs
}

func MapCreateInvoiceMaintenanceDts(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, createdInvoiceMaintenance *models.InvoiceMaintenance, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	invoiceMaintenanceDtsModel := []models.InvoiceMaintenanceDt{}

	for _, invoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		invoiceMaintenanceDtModel := models.InvoiceMaintenanceDt{
			ProductUuid:          invoiceMaintenanceDt.ProductUuid,
			InvoiceMaintenanceID: &createdInvoiceMaintenance.ID,
			ItemUnitID:           invoiceMaintenanceDt.ItemUnitID,
			VatID:                invoiceMaintenanceDt.VatID,
			Pph23ID:              invoiceMaintenanceDt.Pph23ID,
			RefID:                invoiceMaintenanceDt.RefID,
			RefDtID:              invoiceMaintenanceDt.RefDtID,
			ProductID:            invoiceMaintenanceDt.ProductID,
			RefType:              invoiceMaintenanceDt.RefType,
			ProductType:          invoiceMaintenanceDt.ProductType,
			RefJSON:              &refJSON,
			ProductJSON:          &productJSON,
			Remark:               invoiceMaintenanceDt.Remark,
			IsVat:                invoiceMaintenanceDt.IsVat,
			IsPph23:              invoiceMaintenanceDt.IsPph23,
			Qty:                  invoiceMaintenanceDt.Qty,
			Price:                invoiceMaintenanceDt.Price,
			Subtotal:             invoiceMaintenanceDt.Subtotal,
			Discount:             invoiceMaintenanceDt.Discount,
			TotalAmount:          invoiceMaintenanceDt.TotalAmount,
			TotalDp:              invoiceMaintenanceDt.TotalDp,
			TotalBalance:         invoiceMaintenanceDt.TotalBalance,
			CreatedByID:          &userID,
		}
		invoiceMaintenanceDtsModel = append(invoiceMaintenanceDtsModel, invoiceMaintenanceDtModel)
	}

	return invoiceMaintenanceDtsModel, nil
}

func MapUpdateInvoiceMaintenanceDts(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, updatedInvoiceMaintenance *models.InvoiceMaintenance, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	invoiceMaintenanceDtsModel := []models.InvoiceMaintenanceDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqInvoiceMaintenanceDt := range req.InvoiceMaintenanceDts {
		invoiceMaintenanceDtID := uint(0)
		if reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID != nil {
			invoiceMaintenanceDtID = *reqInvoiceMaintenanceDt.InvoiceMaintenanceDtID
		}

		invoiceMaintenanceDtModel := models.InvoiceMaintenanceDt{
			ID:                   invoiceMaintenanceDtID,
			ProductUuid:          reqInvoiceMaintenanceDt.ProductUuid,
			InvoiceMaintenanceID: &updatedInvoiceMaintenance.ID,
			ItemUnitID:           reqInvoiceMaintenanceDt.ItemUnitID,
			VatID:                reqInvoiceMaintenanceDt.VatID,
			Pph23ID:              reqInvoiceMaintenanceDt.Pph23ID,
			RefID:                reqInvoiceMaintenanceDt.RefID,
			RefDtID:              reqInvoiceMaintenanceDt.RefDtID,
			ProductID:            reqInvoiceMaintenanceDt.ProductID,
			RefType:              reqInvoiceMaintenanceDt.RefType,
			ProductType:          reqInvoiceMaintenanceDt.ProductType,
			RefJSON:              &refJSON,
			ProductJSON:          &productJSON,
			Remark:               reqInvoiceMaintenanceDt.Remark,
			IsVat:                reqInvoiceMaintenanceDt.IsVat,
			IsPph23:              reqInvoiceMaintenanceDt.IsPph23,
			Qty:                  reqInvoiceMaintenanceDt.Qty,
			Price:                reqInvoiceMaintenanceDt.Price,
			Subtotal:             reqInvoiceMaintenanceDt.Subtotal,
			Discount:             reqInvoiceMaintenanceDt.Discount,
			TotalAmount:          reqInvoiceMaintenanceDt.TotalAmount,
			TotalDp:              reqInvoiceMaintenanceDt.TotalDp,
			TotalBalance:         reqInvoiceMaintenanceDt.TotalBalance,
			CreatedByID:          &userID,
		}
		invoiceMaintenanceDtsModel = append(invoiceMaintenanceDtsModel, invoiceMaintenanceDtModel)
	}

	return invoiceMaintenanceDtsModel, nil
}

func GenInvoiceMaintenanceNo(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, orderedNumber int, span opentracing.Span) string {
	prefix := "IMT"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateInvoiceMaintenanceNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, revNo *int, span opentracing.Span) string {
	if req.InvoiceNo == nil {
		prefix := "IMT"
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

func MapCreateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.CreateInvoiceMaintenanceRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	invoiceNo := GenInvoiceMaintenanceNo(ctx, req, orderedNumber, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		dueDate = &parsedTime
	}

	invoiceMaintenance := models.InvoiceMaintenance{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ApprovedStatus:           req.ApprovedStatus,
		RevNo:                    new(int),
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}
	*invoiceMaintenance.RevNo = 0

	return invoiceMaintenance, nil
}

func MapUpdateInvoiceMaintenance(ctx *fiber.Ctx, req dtos.UpdateInvoiceMaintenanceRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := GenerateInvoiceMaintenanceNoOnUpdate(ctx, req, &revNo, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, err
		}
		dueDate = &parsedTime
	}

	invoiceMaintenance := models.InvoiceMaintenance{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ApprovedStatus:           req.ApprovedStatus,
		RevNo:                    &revNo,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		UpdatedByID:              &userID,
	}

	return invoiceMaintenance, nil
}

func GetSalesOrderDtIDsForInvoiceMaintenance(soDts []dtos.RefSalesOrderForInvoiceMaintenanceListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDtsForInvoiceMaintenance(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderForInvoiceMaintenanceListDTO) []dtos.RefSalesOrderForInvoiceMaintenanceListDTO {
	combinedSoDts := []dtos.RefSalesOrderForInvoiceMaintenanceListDTO{}

	for _, soDt := range soDts {
		newSoDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
		for _, soDtBom := range soDtBoms {
			if *soDtBom.SoDtID == *soDt.ID {
				newSoDtBoms = append(newSoDtBoms, soDtBom)
			}
		}

		soDt.SoDtsBoms = newSoDtBoms
		combinedSoDts = append(combinedSoDts, soDt)
	}

	return combinedSoDts
}

func MapUpdateSalesOrderStatusForInvoiceMaintenance(salesOrderStatusUpdate map[string]interface{}) dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest {
	params := dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest{}

	params.ID = salesOrderStatusUpdate["id"].(uint)
	params.Status = salesOrderStatusUpdate["status"].(string)

	return params
}

func MapRepeatInvoiceMaintenance(ctx *fiber.Ctx, originalInvoice *models.InvoiceMaintenance, item dtos.RepeatInvoiceMaintenanceItem, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceMaintenance, error) {
	invoiceNo := GenInvoiceMaintenanceNo(ctx, dtos.CreateInvoiceMaintenanceRequest{}, orderedNumber, span)

	var invoiceDate *time.Time
	if item.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *item.InvoiceDate)
		if err != nil {
			return models.InvoiceMaintenance{}, fmt.Errorf("invalid invoice date format: %v", err)
		}
		invoiceDate = &parsedTime
	} else if originalInvoice.InvoiceDate != nil {
		// Use original invoice date if not provided
		invoiceDate = originalInvoice.InvoiceDate
	}

	var dueDate *time.Time
	if item.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *item.DueDate)
		if err != nil {
			return models.InvoiceMaintenance{}, fmt.Errorf("invalid due date format: %v", err)
		}
		dueDate = &parsedTime
	} else if originalInvoice.DueDate != nil {
		// Use original due date if not provided
		dueDate = originalInvoice.DueDate
	}

	// Use provided title or original title
	title := originalInvoice.Title
	if item.Title != nil {
		title = item.Title
	}

	// Use provided remark or original one
	remark := originalInvoice.Remark
	if item.Remark != nil {
		remark = item.Remark
	}

	// Set default status to "UNPAID"
	defaultStatus := "UNPAID"
	defaultApprovedStatus := "PENDING"
	defaultRevNo := 0

	invoiceMaintenance := models.InvoiceMaintenance{
		CustomerID:               originalInvoice.CustomerID,
		CurrencyID:               originalInvoice.CurrencyID,
		PaymentTermID:            originalInvoice.PaymentTermID,
		VatID:                    originalInvoice.VatID,
		Pph23ID:                  originalInvoice.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   originalInvoice.BankID,
		Title:                    title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             originalInvoice.ExchangeRate,
		Remark:                   remark,
		Status:                   &defaultStatus,
		ApprovedStatus:           &defaultApprovedStatus,
		RevNo:                    &defaultRevNo,
		Pph23Percentage:          originalInvoice.Pph23Percentage,
		VatPercentage:            originalInvoice.VatPercentage,
		DiscountAmount:           originalInvoice.DiscountAmount,
		DiscountPercentage:       originalInvoice.DiscountPercentage,
		DiscountPercentageAmount: originalInvoice.DiscountPercentageAmount,
		DiscountFinal:            originalInvoice.DiscountFinal,
		DiscountType:             originalInvoice.DiscountType,
		TotalAmountProducts:      originalInvoice.TotalAmountProducts,
		TotalDpProducts:          originalInvoice.TotalDpProducts,
		TotalBalanceProducts:     originalInvoice.TotalBalanceProducts,
		Subtotal:                 originalInvoice.Subtotal,
		TotalQty:                 originalInvoice.TotalQty,
		TotalDiscount:            originalInvoice.TotalDiscount,
		TotalPph23:               originalInvoice.TotalPph23,
		TotalVat:                 originalInvoice.TotalVat,
		GrandTotal:               originalInvoice.GrandTotal,
		CreatedByID:              &userID,
	}

	return invoiceMaintenance, nil
}

func MapRepeatInvoiceMaintenanceDts(ctx *fiber.Ctx, originalDts []models.InvoiceMaintenanceDt, newInvoiceMaintenanceID uint, userID uint, span opentracing.Span) ([]models.InvoiceMaintenanceDt, error) {
	newDts := make([]models.InvoiceMaintenanceDt, 0, len(originalDts))

	for _, dt := range originalDts {
		newProductUuid := uuid.New().String()

		newDt := models.InvoiceMaintenanceDt{
			ProductUuid:          newProductUuid,
			InvoiceMaintenanceID: &newInvoiceMaintenanceID,
			ItemUnitID:           dt.ItemUnitID,
			VatID:                dt.VatID,
			Pph23ID:              dt.Pph23ID,
			RefID:                dt.RefID,
			RefDtID:              dt.RefDtID,
			ProductID:            dt.ProductID,
			RefType:              dt.RefType,
			ProductType:          dt.ProductType,
			RefJSON:              dt.RefJSON,
			ProductJSON:          dt.ProductJSON,
			Remark:               dt.Remark,
			IsVat:                dt.IsVat,
			IsPph23:              dt.IsPph23,
			Qty:                  dt.Qty,
			Price:                dt.Price,
			Subtotal:             dt.Subtotal,
			Discount:             dt.Discount,
			TotalAmount:          dt.TotalAmount,
			TotalDp:              dt.TotalDp,
			TotalBalance:         dt.TotalBalance,
			CreatedByID:          &userID,
		}
		newDts = append(newDts, newDt)
	}

	return newDts, nil
}
