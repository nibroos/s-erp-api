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

	invoiceMaintenance := models.InvoiceMaintenance{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
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

	invoiceMaintenance := models.InvoiceMaintenance{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
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
