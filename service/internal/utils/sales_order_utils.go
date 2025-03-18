package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func MapFilterSoDtBomsToSoDts(quoDtBoms []dtos.SalesOrderSoDtBomListDTO, quoDts []dtos.SalesOrderSoDtListDTO) []dtos.SalesOrderSoDtListDTO {
	combinedSoDts := []dtos.SalesOrderSoDtListDTO{}

	for _, quoDt := range quoDts {
		newSoDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
		for _, quoDtBom := range quoDtBoms {
			if *quoDtBom.SoDtID == *quoDt.ID {
				quoDtBoms = append(quoDtBoms, quoDtBom)
				newSoDtBoms = append(newSoDtBoms, quoDtBom)
			}
		}

		quoDt.SoDtsBoms = newSoDtBoms
		combinedSoDts = append(combinedSoDts, quoDt)
	}

	return combinedSoDts
}

// func MapFilterUpdateSoDtBomsToSoDts(ctx *fiber.Ctx, quoDts []dtos.SalesOrderSoDtListDTO, req dtos.UpdateSalesOrderRequest, quotationID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
func MapFilterUpdateSoDtBomsToSoDts(ctx *fiber.Ctx, quoDts []dtos.SalesOrderSoDtListUpdateDTO, req dtos.UpdateSalesOrderRequest, quotationID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
	childSpan := span.Tracer().StartSpan("MapFilterUpdateSoDtBomsToSoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateSoDtBoms := []map[string]interface{}{}
	// filter with ID to bulk update
	bulkUpdateSoDtBoms := []map[string]interface{}{}
	// get all ids
	quoDtBomIDs := []uint{}

	claims := GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, reqSoDt := range req.SoDts {
		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			for _, quoDt := range quoDts {
				if *reqSoDtBom.ProductUuid == *quoDt.ProductUuid {
					quoDtBomID := uint(0)
					if reqSoDtBom.SoDtBomID != nil {
						quoDtBomID = *reqSoDtBom.SoDtBomID
					}
					newSoDtBom := map[string]interface{}{
						"id":           quoDtBomID,
						"product_uuid": reqSoDtBom.ProductUuid,
						"quotation_id": quotationID,
						"quo_dt_id":    quoDt.SoDtID,
						"product_id":   reqSoDtBom.ProductID,
						"item_id":      reqSoDtBom.ItemID,
						"item_unit_id": reqSoDtBom.ItemUnitID,
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
						quoDtBomIDs = append(quoDtBomIDs, *reqSoDtBom.SoDtBomID)
					}
				}
			}
		}
	}

	return bulkCreateSoDtBoms, bulkUpdateSoDtBoms, quoDtBomIDs, nil
}

func GetSoIDs(req dtos.UpdateSalesOrderRequest) ([]*uint, []*uint, []*uint, []*uint) {
	quoDtIDs := []*uint{}
	quoDtBomIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqSoDt := range req.SoDts {
		if reqSoDt.SoDtID != nil && *reqSoDt.SoDtID > 0 {
			quoDtIDs = append(quoDtIDs, reqSoDt.SoDtID)
		}

		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			if reqSoDtBom.SoDtBomID != nil && *reqSoDtBom.SoDtBomID > 0 {
				quoDtBomIDs = append(quoDtBomIDs, reqSoDtBom.SoDtBomID)
				productIDs = append(productIDs, &reqSoDtBom.ProductID)
				productIDs = append(productIDs, reqSoDtBom.ItemID)
				itemUnitIDs = append(itemUnitIDs, reqSoDtBom.ItemUnitID)
			}
		}
	}

	return quoDtIDs, quoDtBomIDs, productIDs, itemUnitIDs
}

func MapCreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {
	quoDtsModel := []models.SoDt{}

	for _, quoDt := range req.SoDts {
		quoDtModel := models.SoDt{
			ProductUuid:  quoDt.ProductUuid,
			SalesOrderID: &createdSalesOrder.ID,
			ItemUnitID:   quoDt.ItemUnitID,
			VatID:        quoDt.VatID,
			RefID:        quoDt.RefID,
			ItemID:       quoDt.ItemID,
			RefType:      quoDt.RefType,
			ItemType:     quoDt.ItemType,
			// RefJSON: 	quoDt.RefJSON,
			Remark:       quoDt.Remark,
			VatPerc:      quoDt.VatPerc,
			VatPercAm:    quoDt.VatPercAm,
			QtySO:        quoDt.QtySO,
			Qty:          quoDt.Qty,
			PriceSell:    quoDt.PriceSell,
			PriceBuy:     quoDt.PriceBuy,
			SubtotalSell: quoDt.SubtotalSell,
			SubtotalBuy:  quoDt.SubtotalBuy,
			DiscAm:       quoDt.DiscAm,
			DiscPerc:     quoDt.DiscPerc,
			DiscPercNum:  quoDt.DiscPercNum,
			DiscPercAm:   quoDt.DiscPercAm,
			DiscFinal:    quoDt.DiscFinal,
			DiscType:     quoDt.DiscType,
			TotalAm:      quoDt.TotalAm,
			CreatedByID:  &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}

func MapCreateSoDtBoms(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSoDts []models.SoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	quoDtBomsModel := make([]map[string]interface{}, 0)

	for _, reqSoDt := range req.SoDts {
		for _, reqSoDtBom := range reqSoDt.SoDtsBoms {
			for _, createdSoDt := range createdSoDts {
				if reqSoDtBom.ProductUuid == createdSoDt.ProductUuid {
					genCode := "-"
					itemJson := "{}"
					quoDtBomsModel = append(quoDtBomsModel, map[string]interface{}{
						"id":            0,
						"quotation_id":  createdSoDt.SalesOrderID,
						"product_uuid":  reqSoDtBom.ProductUuid,
						"quo_dt_id":     createdSoDt.ID,
						"product_id":    reqSoDtBom.ProductID,
						"item_id":       reqSoDtBom.ItemID,
						"item_unit_id":  reqSoDtBom.ItemUnitID,
						"remark":        reqSoDtBom.Remark,
						"qty":           reqSoDtBom.Qty,
						"price_sell":    reqSoDtBom.PriceSell,
						"price_buy":     reqSoDtBom.PriceBuy,
						"subtotal_sell": reqSoDtBom.SubtotalSell,
						"subtotal_buy":  reqSoDtBom.SubtotalBuy,
						"gen_code":      &genCode,
						"item_json":     &itemJson,
						"created_by_id": userID,
					})
				}

			}
		}
	}

	return quoDtBomsModel
}

func MapCreateUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {

	quoDtsModel := []models.SoDt{}

	// refJson := "{}"
	// itemJson := "{}"

	for _, reqSoDt := range req.SoDts {
		quoDtID := uint(0)
		if reqSoDt.SoDtID != nil {
			quoDtID = *reqSoDt.SoDtID
		}

		quoDtModel := models.SoDt{
			ID:           quoDtID,
			ProductUuid:  reqSoDt.ProductUuid,
			SalesOrderID: &updatedSalesOrder.ID,
			ItemUnitID:   reqSoDt.ItemUnitID,
			VatID:        reqSoDt.VatID,
			RefID:        reqSoDt.RefID,
			ItemID:       reqSoDt.ItemID,
			RefType:      reqSoDt.RefType,
			ItemType:     reqSoDt.ItemType,
			// RefJSON: 	refJson,
			// ItemJSON: 	 itemJson,
			GenCode:      reqSoDt.GenCode,
			Remark:       reqSoDt.Remark,
			VatPerc:      reqSoDt.VatPerc,
			VatPercAm:    reqSoDt.VatPercAm,
			QtySO:        reqSoDt.QtySO,
			Qty:          reqSoDt.Qty,
			PriceSell:    reqSoDt.PriceSell,
			PriceBuy:     reqSoDt.PriceBuy,
			SubtotalSell: reqSoDt.SubtotalSell,
			SubtotalBuy:  reqSoDt.SubtotalBuy,
			DiscAm:       reqSoDt.DiscAm,
			DiscPerc:     reqSoDt.DiscPerc,
			DiscPercNum:  reqSoDt.DiscPercNum,
			DiscPercAm:   reqSoDt.DiscPercAm,
			DiscFinal:    reqSoDt.DiscFinal,
			DiscType:     reqSoDt.DiscType,
			TotalAm:      reqSoDt.TotalAm,
			CreatedByID:  &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}

func MapCreateSalesOrder(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, userID uint, branchID uint, span opentracing.Span) (models.SalesOrder, error) {
	quotation := models.SalesOrder{
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		QuoNo:         req.QuoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		Status:        req.Status,
		IsApproved:    req.IsApproved,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		TotalQty:      req.TotalQty,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		BranchID:      &branchID,
		CreatedByID:   &userID,
	}

	return quotation, nil
}

func MapUpdateSalesOrder(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, userID uint, branchID uint, span opentracing.Span) (models.SalesOrder, error) {
	quotation := models.SalesOrder{
		ID:            req.ID,
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		QuoNo:         req.QuoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		Status:        req.Status,
		IsApproved:    req.IsApproved,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		TotalQty:      req.TotalQty,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		BranchID:      &branchID,
		CreatedByID:   &userID,
	}

	return quotation, nil
}
