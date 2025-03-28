package utils

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func MapFilterInvoiceDpDtBomsToInvoiceDpDts(invoiceDpDtBoms []dtos.InvoiceDpDtBomListDTO, invoiceDpDts []dtos.InvoiceDpDtListDTO) []dtos.InvoiceDpDtListDTO {
	combinedInvoiceDpDts := []dtos.InvoiceDpDtListDTO{}

	for _, invoiceDpDt := range invoiceDpDts {
		newInvoiceDpDtBoms := make([]dtos.InvoiceDpDtBomListDTO, 0)
		for _, invoiceDpDtBom := range invoiceDpDtBoms {
			if *invoiceDpDtBom.InvoiceDpDtID == *invoiceDpDt.ID {
				invoiceDpDtBoms = append(invoiceDpDtBoms, invoiceDpDtBom)
				newInvoiceDpDtBoms = append(newInvoiceDpDtBoms, invoiceDpDtBom)
			}
		}

		invoiceDpDt.InvoiceDpDtBoms = newInvoiceDpDtBoms
		combinedInvoiceDpDts = append(combinedInvoiceDpDts, invoiceDpDt)
	}

	return combinedInvoiceDpDts
}

func MapFilterUpdateInvoiceDpDtBomsToInvoiceDpDts(ctx *fiber.Ctx, invoiceDpDts []dtos.InvoiceDpDtListUpdateDTO, req dtos.UpdateInvoiceDpRequest, invoiceDpID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
	childSpan := span.Tracer().StartSpan("MapFilterUpdateInvoiceDpDtBomsToInvoiceDpDts", opentracing.ChildOf(span.Context()))

	bulkCreateInvoiceDpDtBoms := []map[string]interface{}{}

	bulkUpdateInvoiceDpDtBoms := []map[string]interface{}{}

	invoiceDpDtBomIDs := []uint{}

	claims := GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		for _, reqInvoiceDpDtBom := range reqInvoiceDpDt.InvoiceDpDtBoms {
			for _, invoiceDpDt := range invoiceDpDts {
				if *reqInvoiceDpDtBom.ProductUuid == *invoiceDpDt.ProductUuid {
					invoiceDpDtBomID := uint(0)
					if reqInvoiceDpDtBom.InvoiceDpDtBomID != nil {
						invoiceDpDtBomID = *reqInvoiceDpDtBom.InvoiceDpDtBomID
					}
					newInvoiceDpDtBom := map[string]interface{}{
						"id":               invoiceDpDtBomID,
						"product_uuid":     reqInvoiceDpDtBom.ProductUuid,
						"invoice_dp_id":    invoiceDpID,
						"invoice_dp_dt_id": invoiceDpDt.InvoiceDpDtID,
						"product_id":       reqInvoiceDpDtBom.ProductID,
						"bom_id":           reqInvoiceDpDtBom.BomID,
						"item_unit_id":     reqInvoiceDpDtBom.ItemUnitID,
						"remark":           reqInvoiceDpDtBom.Remark,
						"qty":              reqInvoiceDpDtBom.Qty,
						"price":            reqInvoiceDpDtBom.Price,
						"subtotal":         reqInvoiceDpDtBom.Subtotal,
					}

					if reqInvoiceDpDtBom.InvoiceDpDtBomID == nil {
						newInvoiceDpDtBom["created_by_id"] = userID
						newInvoiceDpDtBom["created_at"] = time.Now()
						bulkCreateInvoiceDpDtBoms = append(bulkCreateInvoiceDpDtBoms, newInvoiceDpDtBom)
					} else {
						newInvoiceDpDtBom["updated_by_id"] = userID
						newInvoiceDpDtBom["updated_at"] = time.Now()
						bulkUpdateInvoiceDpDtBoms = append(bulkUpdateInvoiceDpDtBoms, newInvoiceDpDtBom)
						invoiceDpDtBomIDs = append(invoiceDpDtBomIDs, *reqInvoiceDpDtBom.InvoiceDpDtBomID)
					}
				}
			}
		}
	}

	return bulkCreateInvoiceDpDtBoms, bulkUpdateInvoiceDpDtBoms, invoiceDpDtBomIDs, nil
}

func GetInvoiceDpIDs(req dtos.UpdateInvoiceDpRequest) ([]*uint, []*uint, []*uint, []*uint) {
	invoiceDpDtIDs := []*uint{}
	invoiceDpDtBomIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.InvoiceDpDtID != nil && *reqInvoiceDpDt.InvoiceDpDtID > 0 {
			invoiceDpDtIDs = append(invoiceDpDtIDs, reqInvoiceDpDt.InvoiceDpDtID)
		}

		for _, reqInvoiceDpDtBom := range reqInvoiceDpDt.InvoiceDpDtBoms {
			if reqInvoiceDpDtBom.InvoiceDpDtBomID != nil && *reqInvoiceDpDtBom.InvoiceDpDtBomID > 0 {
				invoiceDpDtBomIDs = append(invoiceDpDtBomIDs, reqInvoiceDpDtBom.InvoiceDpDtBomID)
				productIDs = append(productIDs, reqInvoiceDpDtBom.ProductID)
				itemUnitIDs = append(itemUnitIDs, reqInvoiceDpDtBom.ItemUnitID)
			}
		}
	}

	return invoiceDpDtIDs, invoiceDpDtBomIDs, productIDs, itemUnitIDs
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
			ProductUuid:              invoiceDpDt.ProductUuid,
			InvoiceDpID:              &createdInvoiceDp.ID,
			ItemUnitID:               invoiceDpDt.ItemUnitID,
			VatID:                    invoiceDpDt.VatID,
			Pph23ID:                  invoiceDpDt.Pph23ID,
			RefID:                    invoiceDpDt.RefID,
			ProductID:                invoiceDpDt.ProductID,
			RefType:                  invoiceDpDt.RefType,
			ProductType:              invoiceDpDt.ProductType,
			RefJSON:                  &refJSON,
			ProductJSON:              &productJSON,
			Remark:                   invoiceDpDt.Remark,
			DpPercentage:             invoiceDpDt.DpPercentage,
			IsVat:                    invoiceDpDt.IsVat,
			IsPph23:                  invoiceDpDt.IsPph23,
			Qty:                      invoiceDpDt.Qty,
			Price:                    invoiceDpDt.Price,
			Subtotal:                 invoiceDpDt.Subtotal,
			DiscountAmount:           invoiceDpDt.DiscountAmount,
			DiscountPercentage:       invoiceDpDt.DiscountPercentage,
			DiscountPercentageNum:    invoiceDpDt.DiscountPercentageNum,
			DiscountPercentageAmount: invoiceDpDt.DiscountPercentageAmount,
			DiscountFinal:            invoiceDpDt.DiscountFinal,
			DiscountType:             invoiceDpDt.DiscountType,
			TotalAmount:              invoiceDpDt.TotalAmount,
			TotalDp:                  invoiceDpDt.TotalDp,
			CreatedByID:              &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
	}

	return invoiceDpDtsModel, nil
}

func MapCreateInvoiceDpDtBoms(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, createdInvoiceDpDts []models.InvoiceDpDt, userID uint, span opentracing.Span) []map[string]interface{} {
	invoiceDpDtBomsModel := make([]map[string]interface{}, 0)
	productJSON := "{}"

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		for _, reqInvoiceDpDtBom := range reqInvoiceDpDt.InvoiceDpDtBoms {
			for _, createdInvoiceDpDt := range createdInvoiceDpDts {
				if reqInvoiceDpDtBom.ProductUuid == createdInvoiceDpDt.ProductUuid {
					invoiceDpDtBomsModel = append(invoiceDpDtBomsModel, map[string]interface{}{
						"id":               0,
						"product_uuid":     reqInvoiceDpDtBom.ProductUuid,
						"invoice_dp_id":    createdInvoiceDpDt.InvoiceDpID,
						"invoice_dp_dt_id": createdInvoiceDpDt.ID,
						"product_id":       reqInvoiceDpDtBom.ProductID,
						"item_unit_id":     reqInvoiceDpDtBom.ItemUnitID,
						"remark":           reqInvoiceDpDtBom.Remark,
						"qty":              reqInvoiceDpDtBom.Qty,
						"price":            reqInvoiceDpDtBom.Price,
						"subtotal":         reqInvoiceDpDtBom.Subtotal,
						"product_json":     &productJSON,
						"created_by_id":    userID,
					})
				}
			}
		}
	}

	return invoiceDpDtBomsModel
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
			ID:                       invoiceDpDtID,
			ProductUuid:              reqInvoiceDpDt.ProductUuid,
			InvoiceDpID:              &updatedInvoiceDp.ID,
			ItemUnitID:               reqInvoiceDpDt.ItemUnitID,
			VatID:                    reqInvoiceDpDt.VatID,
			Pph23ID:                  reqInvoiceDpDt.Pph23ID,
			RefID:                    reqInvoiceDpDt.RefID,
			ProductID:                reqInvoiceDpDt.ProductID,
			RefType:                  reqInvoiceDpDt.RefType,
			ProductType:              reqInvoiceDpDt.ProductType,
			RefJSON:                  &refJSON,
			ProductJSON:              &productJSON,
			Remark:                   reqInvoiceDpDt.Remark,
			DpPercentage:             reqInvoiceDpDt.DpPercentage,
			IsVat:                    reqInvoiceDpDt.IsVat,
			IsPph23:                  reqInvoiceDpDt.IsPph23,
			Qty:                      reqInvoiceDpDt.Qty,
			Price:                    reqInvoiceDpDt.Price,
			Subtotal:                 reqInvoiceDpDt.Subtotal,
			DiscountAmount:           reqInvoiceDpDt.DiscountAmount,
			DiscountPercentage:       reqInvoiceDpDt.DiscountPercentage,
			DiscountPercentageNum:    reqInvoiceDpDt.DiscountPercentageNum,
			DiscountPercentageAmount: reqInvoiceDpDt.DiscountPercentageAmount,
			DiscountFinal:            reqInvoiceDpDt.DiscountFinal,
			DiscountType:             reqInvoiceDpDt.DiscountType,
			TotalAmount:              reqInvoiceDpDt.TotalAmount,
			TotalDp:                  reqInvoiceDpDt.TotalDp,
			CreatedByID:              &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
	}

	return invoiceDpDtsModel, nil
}

func GenInvoiceDpNo() string {
	return "INV-DP-" + time.Now().Format("20060102-150405")
}

func MapCreateInvoiceDp(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, userID uint, branchID uint, span opentracing.Span) (models.InvoiceDp, error) {
	invoiceNo := GenInvoiceDpNo()
	if req.InvoiceNo != nil {
		invoiceNo = *req.InvoiceNo
	}

	invoiceDp := models.InvoiceDp{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              req.InvoiceDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}

	return invoiceDp, nil
}

func MapUpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, span opentracing.Span) (models.InvoiceDp, error) {
	invoiceDp := models.InvoiceDp{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		InvoiceNo:                req.InvoiceNo,
		InvoiceDate:              req.InvoiceDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
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
