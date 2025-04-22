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

func GetSalesInvoiceIDs(req dtos.UpdateSalesInvoiceRequest) ([]*uint, []*uint, []*uint) {
	salesInvoiceDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		if reqSalesInvoiceDt.SalesInvoiceDtID != nil && *reqSalesInvoiceDt.SalesInvoiceDtID > 0 {
			salesInvoiceDtIDs = append(salesInvoiceDtIDs, reqSalesInvoiceDt.SalesInvoiceDtID)
		}
	}

	return salesInvoiceDtIDs, productIDs, itemUnitIDs
}

func GetLockSalesInvoiceSalesOrderIDs(req dtos.CreateSalesInvoiceRequest) []*uint {
	soDtIDs := []*uint{}

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		if reqSalesInvoiceDt.RefType != nil && *reqSalesInvoiceDt.RefType == "sales_orders" && reqSalesInvoiceDt.RefID != nil && *reqSalesInvoiceDt.RefID > 0 {
			soDtIDs = append(soDtIDs, reqSalesInvoiceDt.RefID)
		}
	}

	return soDtIDs
}

func MapCreateSalesInvoiceDts(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, createdSalesInvoice *models.SalesInvoice, userID uint, span opentracing.Span) ([]models.SalesInvoiceDt, error) {
	salesInvoiceDtsModel := []models.SalesInvoiceDt{}

	for _, salesInvoiceDt := range req.SalesInvoiceDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		salesInvoiceDtModel := models.SalesInvoiceDt{
			ProductUuid:    salesInvoiceDt.ProductUuid,
			SalesInvoiceID: &createdSalesInvoice.ID,
			ItemUnitID:     salesInvoiceDt.ItemUnitID,
			VatID:          salesInvoiceDt.VatID,
			Pph23ID:        salesInvoiceDt.Pph23ID,
			RefID:          salesInvoiceDt.RefID,
			RefDtID:        salesInvoiceDt.RefDtID,
			ProductID:      salesInvoiceDt.ProductID,
			RefType:        salesInvoiceDt.RefType,
			ProductType:    salesInvoiceDt.ProductType,
			RefJSON:        &refJSON,
			ProductJSON:    &productJSON,
			Remark:         salesInvoiceDt.Remark,
			IsVat:          salesInvoiceDt.IsVat,
			IsPph23:        salesInvoiceDt.IsPph23,
			Qty:            salesInvoiceDt.Qty,
			Price:          salesInvoiceDt.Price,
			Subtotal:       salesInvoiceDt.Subtotal,
			Discount:       salesInvoiceDt.Discount,
			TotalAmount:    salesInvoiceDt.TotalAmount,
			TotalDp:        salesInvoiceDt.TotalDp,
			TotalBalance:   salesInvoiceDt.TotalBalance,
			CreatedByID:    &userID,
		}
		salesInvoiceDtsModel = append(salesInvoiceDtsModel, salesInvoiceDtModel)
	}

	return salesInvoiceDtsModel, nil
}

func MapUpdateSalesInvoiceDts(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, updatedSalesInvoice *models.SalesInvoice, userID uint, span opentracing.Span) ([]models.SalesInvoiceDt, error) {
	salesInvoiceDtsModel := []models.SalesInvoiceDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		salesInvoiceDtID := uint(0)
		if reqSalesInvoiceDt.SalesInvoiceDtID != nil {
			salesInvoiceDtID = *reqSalesInvoiceDt.SalesInvoiceDtID
		}

		salesInvoiceDtModel := models.SalesInvoiceDt{
			ID:             salesInvoiceDtID,
			ProductUuid:    reqSalesInvoiceDt.ProductUuid,
			SalesInvoiceID: &updatedSalesInvoice.ID,
			ItemUnitID:     reqSalesInvoiceDt.ItemUnitID,
			VatID:          reqSalesInvoiceDt.VatID,
			Pph23ID:        reqSalesInvoiceDt.Pph23ID,
			RefID:          reqSalesInvoiceDt.RefID,
			RefDtID:        reqSalesInvoiceDt.RefDtID,
			ProductID:      reqSalesInvoiceDt.ProductID,
			RefType:        reqSalesInvoiceDt.RefType,
			ProductType:    reqSalesInvoiceDt.ProductType,
			RefJSON:        &refJSON,
			ProductJSON:    &productJSON,
			Remark:         reqSalesInvoiceDt.Remark,
			IsVat:          reqSalesInvoiceDt.IsVat,
			IsPph23:        reqSalesInvoiceDt.IsPph23,
			Qty:            reqSalesInvoiceDt.Qty,
			Price:          reqSalesInvoiceDt.Price,
			Subtotal:       reqSalesInvoiceDt.Subtotal,
			Discount:       reqSalesInvoiceDt.Discount,
			TotalAmount:    reqSalesInvoiceDt.TotalAmount,
			TotalDp:        reqSalesInvoiceDt.TotalDp,
			TotalBalance:   reqSalesInvoiceDt.TotalBalance,
			CreatedByID:    &userID,
		}
		salesInvoiceDtsModel = append(salesInvoiceDtsModel, salesInvoiceDtModel)
	}

	return salesInvoiceDtsModel, nil
}

func GenSalesInvoiceNo(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, orderedNumber int, span opentracing.Span) string {
	if req.InvoiceNo != nil {
		return *req.InvoiceNo
	}

	prefix := "ISL"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateSalesInvoiceNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, revNo *int, span opentracing.Span) string {
	if req.InvoiceNo == nil {
		prefix := "ISL"
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

func MapCreateSalesInvoice(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.SalesInvoice, error) {
	invoiceNo := GenSalesInvoiceNo(ctx, req, orderedNumber, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		invoiceDate = &parsedTime
	}

	salesInvoice := models.SalesInvoice{
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
	*salesInvoice.RevNo = 0

	return salesInvoice, nil
}

func MapUpdateSalesInvoice(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.SalesInvoice, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := GenerateSalesInvoiceNoOnUpdate(ctx, req, &revNo, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		invoiceDate = &parsedTime
	}

	salesInvoice := models.SalesInvoice{
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

	return salesInvoice, nil
}

func GetSalesOrderDtIDs(soDts []dtos.RefSalesOrderForInvoiceListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDtsForInvoice(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderForInvoiceListDTO) []dtos.RefSalesOrderForInvoiceListDTO {
	combinedSoDts := []dtos.RefSalesOrderForInvoiceListDTO{}

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

func MapUpdateSalesOrderStatusForInvoice(salesOrderStatusUpdate map[string]interface{}) dtos.UpdateSalesOrderStatusForInvoiceRequest {
	params := dtos.UpdateSalesOrderStatusForInvoiceRequest{}

	params.ID = salesOrderStatusUpdate["id"].(uint)
	params.Status = salesOrderStatusUpdate["status"].(string)

	return params
}
