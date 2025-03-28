package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func MapFilterSoDtBomsToSoDts(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.SalesOrderSoDtListDTO) []dtos.SalesOrderSoDtListDTO {
	combinedSoDts := []dtos.SalesOrderSoDtListDTO{}

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

// func MapFilterUpdateSoDtBomsToSoDts(ctx *fiber.Ctx, soDts []dtos.SalesOrderSoDtListDTO, req dtos.UpdateSalesOrderRequest, salesOrderID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
func MapFilterUpdateSoDtBomsToSoDts(ctx *fiber.Ctx, soDts []dtos.SalesOrderSoDtListUpdateDTO, req dtos.UpdateSalesOrderRequest, salesOrderID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
	childSpan := span.Tracer().StartSpan("MapFilterUpdateSoDtBomsToSoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateSoDtBoms := []map[string]interface{}{}
	// filter with ID to bulk update
	bulkUpdateSoDtBoms := []map[string]interface{}{}
	// get all ids
	soDtBomIDs := []uint{}

	claims := GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, reqSoDt := range req.SoDts {
		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			for _, soDt := range soDts {
				if *reqSoDtBom.ProductUuid == *soDt.ProductUuid {
					soDtBomID := uint(0)
					if reqSoDtBom.SoDtBomID != nil {
						soDtBomID = *reqSoDtBom.SoDtBomID
					}
					newSoDtBom := map[string]interface{}{
						"id":             soDtBomID,
						"product_uuid":   reqSoDtBom.ProductUuid,
						"sales_order_id": salesOrderID,
						"so_dt_id":       soDt.SoDtID,
						"product_id":     reqSoDtBom.ProductID,
						"item_id":        reqSoDtBom.ItemID,
						"item_unit_id":   reqSoDtBom.ItemUnitID,
						// "item_json":     reqSoDtBom.ItemJSON,
						"gen_code":      reqSoDtBom.GenCode,
						"remark":        reqSoDtBom.Remark,
						"qty":           reqSoDtBom.Qty,
						"price_sell":    reqSoDtBom.PriceSell,
						"price_buy":     reqSoDtBom.PriceBuy,
						"subtotal_sell": reqSoDtBom.SubtotalSell,
						"subtotal_buy":  reqSoDtBom.SubtotalBuy,
					}

					if reqSoDtBom.SoDtBomID == nil {
						newSoDtBom["created_by_id"] = userID
						newSoDtBom["created_at"] = time.Now()
						bulkCreateSoDtBoms = append(bulkCreateSoDtBoms, newSoDtBom)
					} else {
						newSoDtBom["updated_by_id"] = userID
						newSoDtBom["updated_at"] = time.Now()
						bulkUpdateSoDtBoms = append(bulkUpdateSoDtBoms, newSoDtBom)
						soDtBomIDs = append(soDtBomIDs, *reqSoDtBom.SoDtBomID)
					}
				}
			}
		}
	}

	return bulkCreateSoDtBoms, bulkUpdateSoDtBoms, soDtBomIDs, nil
}

func GetSoIDs(req dtos.UpdateSalesOrderRequest) ([]*uint, []*uint, []*uint, []*uint) {
	soDtIDs := []*uint{}
	soDtBomIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqSoDt := range req.SoDts {
		if reqSoDt.SoDtID != nil && *reqSoDt.SoDtID > 0 {
			soDtIDs = append(soDtIDs, reqSoDt.SoDtID)
		}

		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			if reqSoDtBom.SoDtBomID != nil && *reqSoDtBom.SoDtBomID > 0 {
				soDtBomIDs = append(soDtBomIDs, reqSoDtBom.SoDtBomID)
				productIDs = append(productIDs, &reqSoDtBom.ProductID)
				productIDs = append(productIDs, reqSoDtBom.ItemID)
				itemUnitIDs = append(itemUnitIDs, reqSoDtBom.ItemUnitID)
			}
		}
	}

	return soDtIDs, soDtBomIDs, productIDs, itemUnitIDs
}

func GetLockSalesOrderQuoIDs(req dtos.CreateSalesOrderRequest) []*uint {
	quoDtIDs := []*uint{}

	for _, reqSoDt := range req.SoDts {
		if reqSoDt.RefType != nil && *reqSoDt.RefType == "quotations" && reqSoDt.RefID != nil && *reqSoDt.RefID > 0 {
			quoDtIDs = append(quoDtIDs, reqSoDt.RefID)
		}
	}

	return quoDtIDs
}

func MapCreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	soDtsModel := []models.SoDt{}

	for _, soDt := range req.SoDts {
		genCode := "-"
		itemJson := "{}"
		refJSON := "{}"

		soDtModel := models.SoDt{
			ProductUuid:     soDt.ProductUuid,
			SalesOrderID:    &createdSalesOrder.ID,
			ItemUnitID:      soDt.ItemUnitID,
			VatID:           soDt.VatID,
			Pph23ID:         soDt.Pph23ID,
			RefID:           soDt.RefID,
			ItemID:          soDt.ItemID,
			RefType:         soDt.RefType,
			ItemType:        soDt.ItemType,
			RefJSON:         &refJSON,
			ItemJSON:        &itemJson,
			GenCode:         &genCode,
			Remark:          soDt.Remark,
			VatPerc:         soDt.VatPerc,
			VatPercAm:       soDt.VatPercAm,
			Pph23Perc:       soDt.Pph23Perc,
			Pph23PercAm:     soDt.Pph23PercAm,
			MarkupPerc:      soDt.MarkupPerc,
			MarkupPercAm:    soDt.MarkupPercAm,
			IsVat:           soDt.IsVat,
			IsPph23:         soDt.IsPph23,
			IsLockMarkup:    soDt.IsLockMarkup,
			IsLockPriceSell: soDt.IsLockPriceSell,
			Qty:             soDt.Qty,
			PriceSell:       soDt.PriceSell,
			PriceBuy:        soDt.PriceBuy,
			SubtotalSell:    soDt.SubtotalSell,
			SubtotalBuy:     soDt.SubtotalBuy,
			DiscAm:          soDt.DiscAm,
			DiscPerc:        soDt.DiscPerc,
			DiscPercNum:     soDt.DiscPercNum,
			DiscPercAm:      soDt.DiscPercAm,
			DiscFinal:       soDt.DiscFinal,
			DiscType:        soDt.DiscType,
			TotalAm:         soDt.TotalAm,
			CreatedByID:     &userID,
		}
		soDtsModel = append(soDtsModel, soDtModel)

	}

	return soDtsModel, nil
}

func MapCreateSoDtBoms(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSoDts []models.SoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	soDtBomsModel := make([]map[string]interface{}, 0)

	for _, reqSoDt := range req.SoDts {
		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			for _, createdSoDt := range createdSoDts {
				if reqSoDtBom.ProductUuid == createdSoDt.ProductUuid {
					genCode := "-"
					itemJson := "{}"
					soDtBomsModel = append(soDtBomsModel, map[string]interface{}{
						"id":             0,
						"product_uuid":   reqSoDtBom.ProductUuid,
						"sales_order_id": createdSoDt.SalesOrderID,
						"so_dt_id":       createdSoDt.ID,
						"product_id":     reqSoDtBom.ProductID,
						"item_id":        reqSoDtBom.ItemID,
						"item_unit_id":   reqSoDtBom.ItemUnitID,
						"remark":         reqSoDtBom.Remark,
						"qty":            reqSoDtBom.Qty,
						"price_sell":     reqSoDtBom.PriceSell,
						"price_buy":      reqSoDtBom.PriceBuy,
						"subtotal_sell":  reqSoDtBom.SubtotalSell,
						"subtotal_buy":   reqSoDtBom.SubtotalBuy,
						"gen_code":       &genCode,
						"item_json":      &itemJson,
						"created_by_id":  userID,
					})
				}

			}
		}
	}

	return soDtBomsModel
}

func MapUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {

	soDtsModel := []models.SoDt{}

	refJson := "{}"
	itemJson := "{}"

	for _, reqSoDt := range req.SoDts {
		soDtID := uint(0)
		if reqSoDt.SoDtID != nil {
			soDtID = *reqSoDt.SoDtID
		}

		soDtModel := models.SoDt{
			ID:              soDtID,
			ProductUuid:     reqSoDt.ProductUuid,
			SalesOrderID:    &updatedSalesOrder.ID,
			ItemUnitID:      reqSoDt.ItemUnitID,
			VatID:           reqSoDt.VatID,
			Pph23ID:         reqSoDt.Pph23ID,
			RefID:           reqSoDt.RefID,
			ItemID:          reqSoDt.ItemID,
			RefType:         reqSoDt.RefType,
			ItemType:        reqSoDt.ItemType,
			RefJSON:         &refJson,
			ItemJSON:        &itemJson,
			GenCode:         reqSoDt.GenCode,
			Remark:          reqSoDt.Remark,
			VatPerc:         reqSoDt.VatPerc,
			VatPercAm:       reqSoDt.VatPercAm,
			Pph23Perc:       reqSoDt.Pph23Perc,
			Pph23PercAm:     reqSoDt.Pph23PercAm,
			MarkupPerc:      reqSoDt.MarkupPerc,
			MarkupPercAm:    reqSoDt.MarkupPercAm,
			IsVat:           reqSoDt.IsVat,
			IsPph23:         reqSoDt.IsPph23,
			IsLockMarkup:    reqSoDt.IsLockMarkup,
			IsLockPriceSell: reqSoDt.IsLockPriceSell,
			Qty:             reqSoDt.Qty,
			PriceSell:       reqSoDt.PriceSell,
			PriceBuy:        reqSoDt.PriceBuy,
			SubtotalSell:    reqSoDt.SubtotalSell,
			SubtotalBuy:     reqSoDt.SubtotalBuy,
			DiscAm:          reqSoDt.DiscAm,
			DiscPerc:        reqSoDt.DiscPerc,
			DiscPercNum:     reqSoDt.DiscPercNum,
			DiscPercAm:      reqSoDt.DiscPercAm,
			DiscFinal:       reqSoDt.DiscFinal,
			DiscType:        reqSoDt.DiscType,
			TotalAm:         reqSoDt.TotalAm,
			CreatedByID:     &userID,
		}
		soDtsModel = append(soDtsModel, soDtModel)

	}

	return soDtsModel, nil
}

func GenSalesOrderNo(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, orderedNumber int, span opentracing.Span) string {
	if req.PoBuyerNo != nil {
		return *req.PoBuyerNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	surname := req.CustomerCode
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("SO/%s/%s-%s-%s", order, surname, year, month)

	return str
}

func MapCreateSalesOrder(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, userID uint, branchID uint, customerSoCreatedThisMonthNumber int, span opentracing.Span) (models.SalesOrder, error) {
	// order := 1
	customerSoCreatedThisMonthNumber++

	orderNo := GenSalesOrderNo(ctx, req, customerSoCreatedThisMonthNumber, span)
	poBuyerNo := GeneratePoBuyerNoNoOnCreateSalesOrder(ctx, req, customerSoCreatedThisMonthNumber, span)

	salesOrder := models.SalesOrder{
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		WarehouseID:   req.WarehouseID,
		PoBuyerNo:     poBuyerNo,
		SalesOrderNo:  &orderNo,
		ShipDest:      req.ShipDest,
		Remark:        req.Remark,
		Status:        req.Status,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		MarkupPerc:    req.MarkupPerc,
		TotalQty:      req.TotalQty,
		DiscAm:        req.DiscAm,
		DiscPerc:      req.DiscPerc,
		DiscPercAm:    req.DiscPercAm,
		DiscFinal:     req.DiscFinal,
		DiscType:      req.DiscType,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		OrderAt:       req.OrderAt,
		ShippingAt:    req.ShippingAt,
		AgreeAt:       req.AgreeAt,
		DueAt:         req.DueAt,
		BranchID:      &branchID,
		CreatedByID:   &userID,
	}

	return salesOrder, nil
}

func MapUpdateSalesOrder(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, userID uint, branchID uint, span opentracing.Span) (models.SalesOrder, error) {
	revNo := 0
	if req.RevNo != nil {
		revNo = *req.RevNo
	}
	revNo++

	salesOrderNo := GenerateSoNoOnUpdateQuotation(ctx, req, revNo, span)
	poBuyerNo := GenerateSoNoOnUpdateQuotation(ctx, req, revNo, span)

	salesOrder := models.SalesOrder{
		ID:            req.ID,
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		WarehouseID:   req.WarehouseID,
		RevNo:         &revNo,
		PoBuyerNo:     poBuyerNo,
		SalesOrderNo:  &salesOrderNo,
		ShipDest:      req.ShipDest,
		Remark:        req.Remark,
		Status:        req.Status,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		MarkupPerc:    req.MarkupPerc,
		TotalQty:      req.TotalQty,
		DiscAm:        req.DiscAm,
		DiscPerc:      req.DiscPerc,
		DiscPercAm:    req.DiscPercAm,
		DiscFinal:     req.DiscFinal,
		DiscType:      req.DiscType,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		OrderAt:       req.OrderAt,
		ShippingAt:    req.ShippingAt,
		AgreeAt:       req.AgreeAt,
		DueAt:         req.DueAt,
		BranchID:      &branchID,
		CreatedByID:   &userID,
	}

	return salesOrder, nil
}

func GetQuoDtIDs(quoDts []dtos.RefIndexQuoDtListDTO) []uint {
	quotationIDs := []uint{}

	for _, quoDt := range quoDts {
		quotationIDs = append(quotationIDs, *quoDt.QuotationID)
	}

	return quotationIDs
}

func MapRefQuoDtBomsToQuoDts(quoDtBoms []dtos.QuotationQuoDtBomListDTO, quoDts []dtos.RefIndexQuoDtListDTO) []dtos.RefIndexQuoDtListDTO {
	combinedQuoDts := []dtos.RefIndexQuoDtListDTO{}

	for _, quoDt := range quoDts {
		newQuoDtBoms := make([]dtos.QuotationQuoDtBomListDTO, 0)
		for _, quoDtBom := range quoDtBoms {
			if *quoDtBom.QuoDtID == *quoDt.ID {
				quoDtBoms = append(quoDtBoms, quoDtBom)
				newQuoDtBoms = append(newQuoDtBoms, quoDtBom)
			}
		}

		quoDt.QuoDtsBoms = newQuoDtBoms
		combinedQuoDts = append(combinedQuoDts, quoDt)
	}

	return combinedQuoDts
}

func MapUpdateQuoDtsQty(quoDtsQtyUpdate []dtos.GetQuoDtQtyUpdateDTO, req dtos.CreateSalesOrderRequest) []map[string]interface{} {
	// filter with ID to bulk update
	bulkUpdateQuoDts := []map[string]interface{}{}

	for _, reqSoDt := range req.SoDts {
		for _, quoDts := range quoDtsQtyUpdate {
			if reqSoDt.RefID != nil && quoDts.QuoDtID != nil && *reqSoDt.RefID == *quoDts.QuoDtID {

				if quoDts.QtySO == nil {
					quoDts.QtySO = new(float64)
				}

				newQuoDt := map[string]interface{}{
					"id":     quoDts.QuoDtID,
					"qty_so": (*reqSoDt.Qty + *quoDts.QtySO),
				}
				bulkUpdateQuoDts = append(bulkUpdateQuoDts, newQuoDt)
			}
		}
	}

	return bulkUpdateQuoDts
}

func MapUpdateQuoDtsStatus(quoDtsStatusUpdate map[string]interface{}, req dtos.CreateSalesOrderRequest) dtos.UpdateQuotationStatusRequest {
	// get quotation id
	params := dtos.UpdateQuotationStatusRequest{}

	params.ID = quoDtsStatusUpdate["id"].(uint)
	params.Status = quoDtsStatusUpdate["status"].(string)

	return params
}

func GenerateSoNoOnUpdateQuotation(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, poBuyerNo string, revNo int, span opentracing.Span) string {

	// get before REV-number, full string is SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1 or SURNAME-2001-12-20
	// check if "REV" string exist (random), if not add "REV-1" else replace REV-1 change the number to increment REV-revNo
	if !strings.Contains(poBuyerNo, "REV") {
		poBuyerNo = fmt.Sprintf("%s/REV-1", poBuyerNo)
	} else {
		// remove after /REV
		// use split to get the first part of string
		poBuyerNo = strings.Split(poBuyerNo, "/REV")[0]
		poBuyerNo = fmt.Sprintf("%s/REV-%d", poBuyerNo, revNo)
	}

	return poBuyerNo
}

func GeneratePoBuyerNoNoOnCreateSalesOrder(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, orderedNumber int, span opentracing.Span) string {
	if req.PoBuyerNo != nil {
		return *req.PoBuyerNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	surname := req.CustomerCode
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("%s/%s-%s-%s", surname, year, month, order)

	return str
}
