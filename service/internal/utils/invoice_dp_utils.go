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

func GetInvoiceDpIDs(req dtos.UpdateInvoiceDpRequest) ([]*uint, []*uint, []*uint) {
	invoiceDpDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.InvoiceDpDtID != nil && *reqInvoiceDpDt.InvoiceDpDtID > 0 {
			invoiceDpDtIDs = append(invoiceDpDtIDs, reqInvoiceDpDt.InvoiceDpDtID)
		}
	}

	return invoiceDpDtIDs, productIDs, itemUnitIDs
}

func GetLockInvoiceDpSalesOrderIDs(req dtos.CreateInvoiceDpRequest) []*uint {
	soDtIDs := []*uint{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.RefType != nil && *reqInvoiceDpDt.RefType == "sales_orders" && reqInvoiceDpDt.RefID != nil && *reqInvoiceDpDt.RefID > 0 {
			soDtIDs = append(soDtIDs, reqInvoiceDpDt.RefID)
		}
	}

	return soDtIDs
}

func MapCreateInvoiceDpDts(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, createdInvoiceDp *models.InvoiceDp, userID uint, span opentracing.Span) ([]models.InvoiceDpDt, error) {
	invoiceDpDtsModel := []models.InvoiceDpDt{}

	for _, invoiceDpDt := range req.InvoiceDpDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		invoiceDpDtModel := models.InvoiceDpDt{
			ProductUuid:  invoiceDpDt.ProductUuid,
			InvoiceDpID:  &createdInvoiceDp.ID,
			ItemUnitID:   invoiceDpDt.ItemUnitID,
			VatID:        invoiceDpDt.VatID,
			Pph23ID:      invoiceDpDt.Pph23ID,
			RefID:        invoiceDpDt.RefID,
			RefDtID:      invoiceDpDt.RefDtID,
			ProductID:    invoiceDpDt.ProductID,
			RefType:      invoiceDpDt.RefType,
			ProductType:  invoiceDpDt.ProductType,
			RefJSON:      &refJSON,
			ProductJSON:  &productJSON,
			Remark:       invoiceDpDt.Remark,
			DpPercentage: invoiceDpDt.DpPercentage,
			IsVat:        invoiceDpDt.IsVat,
			IsPph23:      invoiceDpDt.IsPph23,
			Qty:          invoiceDpDt.Qty,
			Price:        invoiceDpDt.Price,
			Subtotal:     invoiceDpDt.Subtotal,
			Discount:     invoiceDpDt.Discount,
			TotalAmount:  invoiceDpDt.TotalAmount,
			TotalDp:      invoiceDpDt.TotalDp,
			CreatedByID:  &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
	}

	return invoiceDpDtsModel, nil
}

func MapUpdateInvoiceDpDts(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, updatedInvoiceDp *models.InvoiceDp, userID uint, span opentracing.Span) ([]models.InvoiceDpDt, error) {
	invoiceDpDtsModel := []models.InvoiceDpDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		invoiceDpDtID := uint(0)
		if reqInvoiceDpDt.InvoiceDpDtID != nil {
			invoiceDpDtID = *reqInvoiceDpDt.InvoiceDpDtID
		}

		invoiceDpDtModel := models.InvoiceDpDt{
			ID:           invoiceDpDtID,
			ProductUuid:  reqInvoiceDpDt.ProductUuid,
			InvoiceDpID:  &updatedInvoiceDp.ID,
			ItemUnitID:   reqInvoiceDpDt.ItemUnitID,
			VatID:        reqInvoiceDpDt.VatID,
			Pph23ID:      reqInvoiceDpDt.Pph23ID,
			RefID:        reqInvoiceDpDt.RefID,
			RefDtID:      reqInvoiceDpDt.RefDtID,
			ProductID:    reqInvoiceDpDt.ProductID,
			RefType:      reqInvoiceDpDt.RefType,
			ProductType:  reqInvoiceDpDt.ProductType,
			RefJSON:      &refJSON,
			ProductJSON:  &productJSON,
			Remark:       reqInvoiceDpDt.Remark,
			DpPercentage: reqInvoiceDpDt.DpPercentage,
			IsVat:        reqInvoiceDpDt.IsVat,
			IsPph23:      reqInvoiceDpDt.IsPph23,
			Qty:          reqInvoiceDpDt.Qty,
			Price:        reqInvoiceDpDt.Price,
			Subtotal:     reqInvoiceDpDt.Subtotal,
			Discount:     reqInvoiceDpDt.Discount,
			TotalAmount:  reqInvoiceDpDt.TotalAmount,
			TotalDp:      reqInvoiceDpDt.TotalDp,
			CreatedByID:  &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
	}

	return invoiceDpDtsModel, nil
}

func GenInvoiceDpNo(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, orderedNumber int, span opentracing.Span) string {
	if req.InvoiceNo != nil {
		return *req.InvoiceNo
	}

	prefix := "IDP"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateInvoiceDpNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, revNo *int, span opentracing.Span) string {
	invoiceNo := req.InvoiceNo

	if !strings.Contains(*invoiceNo, "REV") {
		*invoiceNo = fmt.Sprintf("%s/REV-%d", *invoiceNo, *revNo)
	} else {
		*invoiceNo = strings.Split(*invoiceNo, "/REV")[0]
		*invoiceNo = fmt.Sprintf("%s/REV-%d", *invoiceNo, *revNo)
	}

	return *invoiceNo
}

func MapCreateInvoiceDp(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceDp, error) {
	invoiceNo := GenInvoiceDpNo(ctx, req, orderedNumber, span)

	invoiceDp := models.InvoiceDp{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              req.InvoiceDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		RevNo:                    new(int),
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}
	*invoiceDp.RevNo = 0

	return invoiceDp, nil
}

func MapUpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.InvoiceDp, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := *req.InvoiceNo
	invoiceNo = GenerateInvoiceDpNoOnUpdate(ctx, req, &revNo, span)

	invoiceDp := models.InvoiceDp{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              req.InvoiceDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		RevNo:                    &revNo,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		UpdatedByID:              &userID,
	}

	return invoiceDp, nil
}

func GetSoDtIDs(soDts []dtos.RefSalesOrderDtListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDts(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderDtListDTO) []dtos.RefSalesOrderDtListDTO {
	combinedSoDts := []dtos.RefSalesOrderDtListDTO{}

	for _, soDt := range soDts {
		newSoDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
		for _, soDtBom := range soDtBoms {
			if *soDtBom.SoDtID == *soDt.ID {
				soDtBoms = append(soDtBoms, soDtBom)
				newSoDtBoms = append(newSoDtBoms, soDtBom)
			}
		}

		soDt.SoDtsBoms = newSoDtBoms
		combinedSoDts = append(combinedSoDts, soDt)
	}

	return combinedSoDts
}

func MapUpdateSoDtsQtyForInvoice(soDtsQtyUpdate []dtos.GetSoDtQtyUpdateForInvoiceDTO, req dtos.CreateInvoiceDpRequest) []map[string]interface{} {
	bulkUpdateSoDts := []map[string]interface{}{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		for _, soDts := range soDtsQtyUpdate {
			if reqInvoiceDpDt.RefID != nil && soDts.SoDtID != nil && *reqInvoiceDpDt.RefID == *soDts.SoDtID {
				if soDts.QtyInvoiced == nil {
					soDts.QtyInvoiced = new(float64)
				}

				newSoDt := map[string]interface{}{
					"id":           soDts.SoDtID,
					"qty_invoiced": (*reqInvoiceDpDt.Qty + *soDts.QtyInvoiced),
				}
				bulkUpdateSoDts = append(bulkUpdateSoDts, newSoDt)
			}
		}
	}

	return bulkUpdateSoDts
}

func MapUpdateSalesOrderStatusForInvoice(salesOrderStatusUpdate map[string]interface{}, req dtos.CreateInvoiceDpRequest) dtos.UpdateSalesOrderStatusForInvoiceRequest {
	params := dtos.UpdateSalesOrderStatusForInvoiceRequest{}

	params.ID = salesOrderStatusUpdate["id"].(uint)
	params.Status = salesOrderStatusUpdate["status"].(string)

	return params
}
