package utils

import (
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
						"quo_dt_id":      soDt.SoDtID,
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

func MapCreateSoDts(ctx *fiber.Ctx, req dtos.CreateSalesOrderRequest, createdSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {

	soDtsModel := []models.SoDt{}

	for _, soDt := range req.SoDts {
		soDtModel := models.SoDt{
			ProductUuid:  soDt.ProductUuid,
			SalesOrderID: &createdSalesOrder.ID,
			ItemUnitID:   soDt.ItemUnitID,
			VatID:        soDt.VatID,
			RefID:        soDt.RefID,
			ItemID:       soDt.ItemID,
			RefType:      soDt.RefType,
			ItemType:     soDt.ItemType,
			// RefJSON: 	soDt.RefJSON,
			Remark:       soDt.Remark,
			VatPerc:      soDt.VatPerc,
			VatPercAm:    soDt.VatPercAm,
			QtySO:        soDt.QtySO,
			Qty:          soDt.Qty,
			PriceSell:    soDt.PriceSell,
			PriceBuy:     soDt.PriceBuy,
			SubtotalSell: soDt.SubtotalSell,
			SubtotalBuy:  soDt.SubtotalBuy,
			DiscAm:       soDt.DiscAm,
			DiscPerc:     soDt.DiscPerc,
			DiscPercNum:  soDt.DiscPercNum,
			DiscPercAm:   soDt.DiscPercAm,
			DiscFinal:    soDt.DiscFinal,
			DiscType:     soDt.DiscType,
			TotalAm:      soDt.TotalAm,
			CreatedByID:  &userID,
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
						"id":            0,
						"salesOrder_id": createdSoDt.SalesOrderID,
						"product_uuid":  reqSoDtBom.ProductUuid,
						"so_dt_id":      createdSoDt.ID,
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

	return soDtBomsModel
}

func MapCreateUpdateSoDts(ctx *fiber.Ctx, req dtos.UpdateSalesOrderRequest, updatedSalesOrder *models.SalesOrder, userID uint, span opentracing.Span) ([]models.SoDt, error) {

	soDtsModel := []models.SoDt{}

	// refJson := "{}"
	// itemJson := "{}"

	for _, reqSoDt := range req.SoDts {
		soDtID := uint(0)
		if reqSoDt.SoDtID != nil {
			soDtID = *reqSoDt.SoDtID
		}

		soDtModel := models.SoDt{
			ID:           soDtID,
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
		soDtsModel = append(soDtsModel, soDtModel)

	}

	return soDtsModel, nil
}
