package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/opentracing/opentracing-go"
)

func MapFilterQuoDtBomsToQuoDts(quoDtBoms []dtos.QuotationQuoDtBomListDTO, quoDts []dtos.QuotationQuoDtListDTO) []dtos.QuotationQuoDtListDTO {
	combinedQuoDts := []dtos.QuotationQuoDtListDTO{}

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

// func MapFilterUpdateQuoDtBomsToQuoDts(ctx *fiber.Ctx, quoDts []dtos.QuotationQuoDtListDTO, req dtos.UpdateQuotationRequest, quotationID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
func MapFilterUpdateQuoDtBomsToQuoDts(ctx *fiber.Ctx, quoDts []dtos.QuotationQuoDtListUpdateDTO, req dtos.UpdateQuotationRequest, quotationID uint, span opentracing.Span) ([]map[string]interface{}, []map[string]interface{}, []uint, error) {
	childSpan := span.Tracer().StartSpan("MapFilterUpdateQuoDtBomsToQuoDts", opentracing.ChildOf(span.Context()))

	// filter without ID to bulk create
	bulkCreateQuoDtBoms := []map[string]interface{}{}
	// filter with ID to bulk update
	bulkUpdateQuoDtBoms := []map[string]interface{}{}
	// get all ids
	quoDtBomIDs := []uint{}

	claims := GetClaims(ctx, childSpan)
	userID := uint(claims["user_id"].(float64))

	for _, reqQuoDt := range req.QuoDts {
		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
			for _, quoDt := range quoDts {
				if *reqQuoDtBom.ProductUuid == *quoDt.ProductUuid {
					quoDtBomID := uint(0)
					if reqQuoDtBom.QuoDtBomID != nil {
						quoDtBomID = *reqQuoDtBom.QuoDtBomID
					}
					newQuoDtBom := map[string]interface{}{
						"id":           quoDtBomID,
						"product_uuid": reqQuoDtBom.ProductUuid,
						"quotation_id": quotationID,
						"quo_dt_id":    quoDt.QuoDtID,
						"product_id":   reqQuoDtBom.ProductID,
						"item_id":      reqQuoDtBom.ItemID,
						"item_unit_id": reqQuoDtBom.ItemUnitID,
						// "item_json":     reqQuoDtBom.ItemJSON,
						"gen_code":      reqQuoDtBom.GenCode,
						"remark":        reqQuoDtBom.Remark,
						"qty":           reqQuoDtBom.Qty,
						"price_sell":    reqQuoDtBom.PriceSell,
						"price_buy":     reqQuoDtBom.PriceBuy,
						"subtotal_sell": reqQuoDtBom.SubtotalSell,
						"subtotal_buy":  reqQuoDtBom.SubtotalBuy,
					}

					if reqQuoDtBom.QuoDtBomID == nil {
						newQuoDtBom["created_by_id"] = userID
						newQuoDtBom["created_at"] = time.Now()
						bulkCreateQuoDtBoms = append(bulkCreateQuoDtBoms, newQuoDtBom)
					} else {
						newQuoDtBom["updated_by_id"] = userID
						newQuoDtBom["updated_at"] = time.Now()
						bulkUpdateQuoDtBoms = append(bulkUpdateQuoDtBoms, newQuoDtBom)
						quoDtBomIDs = append(quoDtBomIDs, *reqQuoDtBom.QuoDtBomID)
					}
				}
			}
		}
	}

	return bulkCreateQuoDtBoms, bulkUpdateQuoDtBoms, quoDtBomIDs, nil
}

func GetQuoIDs(req dtos.UpdateQuotationRequest) ([]*uint, []*uint, []*uint, []*uint) {
	quoDtIDs := []*uint{}
	quoDtBomIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqQuoDt := range req.QuoDts {
		if reqQuoDt.QuoDtID != nil && *reqQuoDt.QuoDtID > 0 {
			quoDtIDs = append(quoDtIDs, reqQuoDt.QuoDtID)
		}

		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
			if reqQuoDtBom.QuoDtBomID != nil && *reqQuoDtBom.QuoDtBomID > 0 {
				quoDtBomIDs = append(quoDtBomIDs, reqQuoDtBom.QuoDtBomID)
				productIDs = append(productIDs, &reqQuoDtBom.ProductID)
				productIDs = append(productIDs, reqQuoDtBom.ItemID)
				itemUnitIDs = append(itemUnitIDs, reqQuoDtBom.ItemUnitID)
			}
		}
	}

	return quoDtIDs, quoDtBomIDs, productIDs, itemUnitIDs
}
