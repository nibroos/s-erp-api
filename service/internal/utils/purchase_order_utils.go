package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
)

func GenPurchaseOrderNo() string {
	return "PO-" + time.Now().Format("20060102-150405")
}

func MapCreatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, branchID uint, customerSoCreatedThisMonthNumber int, span opentracing.Span) (models.PurchaseOrder, error) {

	poNo := GeneratePurchaseOrderNoOnCreatePurchaseOrder(ctx, req, customerSoCreatedThisMonthNumber, span)

	purchaseOrder := models.PurchaseOrder{
		CustomerID:               req.CustomerID,
		PurchaseTypeID:           req.PurchaseTypeID,
		CurrencyID:               req.CurrencyID,
		VatID:                    req.VatID,
		PaymentTermID:            req.PaymentTermID,
		ShippingTermID:           req.ShippingTermID,
		Pph23ID:                  req.Pph23ID,
		PaymentID:                req.PaymentID,
		IsVat:                    req.IsVat,
		PoNo:                     &poNo,
		PoNoOri:                  &poNo,
		PoDate:                   req.PoDate,
		DeliveryDate:             req.DeliveryDate,
		ShippingDestination:      req.ShippingDestination,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ExchangeRate:             req.ExchangeRate,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinalHeader:      req.DiscountFinalHeader,
		DiscountAmountProduct:    req.DiscountAmountProduct,
		DiscountType:             req.DiscountType,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		BranchID:                 &branchID,
		CreatedByID:              &userID,
	}

	return purchaseOrder, nil
}

func GeneratePurchaseOrderNoOnCreatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, orderedNumber int, span opentracing.Span) string {
	if req.PoNo != nil {
		return *req.PoNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	// surname := req.CustomerCode
	surname := "Yubi"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	orderedNumber++
	order := fmt.Sprintf("%d", orderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("PO/%s-%s-%s-%s", surname, year, month, order)

	return str
}

func GeneratePurchaseOrderNoOnUpdatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, purchaseOrderNo *string, revNo int, span opentracing.Span) string {

	// get before REV-number, full string is SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1 or SURNAME-2001-12-20
	// check if "REV" string exist (random), if not add "REV-1" else replace REV-1 change the number to increment REV-revNo
	if !strings.Contains(*purchaseOrderNo, "REV") {
		*purchaseOrderNo = fmt.Sprintf("%s/REV-1", *purchaseOrderNo)
	} else {
		// remove after /REV
		// use split to get the first part of string
		*purchaseOrderNo = strings.Split(*purchaseOrderNo, "/REV")[0]
		*purchaseOrderNo = fmt.Sprintf("%s/REV-%d", *purchaseOrderNo, revNo)
	}

	return *purchaseOrderNo
}

func MapUpdatePurchaseOrder(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, branchID uint, span opentracing.Span) (models.PurchaseOrder, error) {
	revNo := 0
	if req.RevNo != nil {
		revNo = *req.RevNo
	}
	revNo++

	poNo := GeneratePurchaseOrderNoOnUpdatePurchaseOrder(ctx, req, req.PoNo, revNo, span)

	purchaseOrder := models.PurchaseOrder{
		ID:                       *req.ID,
		CustomerID:               req.CustomerID,
		PurchaseTypeID:           req.PurchaseTypeID,
		CurrencyID:               req.CurrencyID,
		VatID:                    req.VatID,
		PaymentTermID:            req.PaymentTermID,
		ShippingTermID:           req.ShippingTermID,
		Pph23ID:                  req.Pph23ID,
		PaymentID:                req.PaymentID,
		IsVat:                    req.IsVat,
		PoNo:                     &poNo,
		RevNo:                    &revNo,
		PoNoOri:                  req.PoNoOri,
		PoDate:                   req.PoDate,
		DeliveryDate:             req.DeliveryDate,
		ShippingDestination:      req.ShippingDestination,
		Remark:                   req.Remark,
		Status:                   req.Status,
		ExchangeRate:             req.ExchangeRate,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinalHeader:      req.DiscountFinalHeader,
		DiscountAmountProduct:    req.DiscountAmountProduct,
		DiscountType:             req.DiscountType,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		BranchID:                 &branchID,
		UpdatedByID:              &userID,
	}

	return purchaseOrder, nil
}

func MapCreatePoDts(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, createdPurchaseOrder *models.PurchaseOrder, userID uint, span opentracing.Span) ([]models.PoDt, []map[string]interface{}, error) {
	poDtsModel := []models.PoDt{}
	updatedItemUnits := []map[string]interface{}{}

	for _, poDt := range req.PoDts {
		genCode := "-"
		productJsonStr := "{}"
		refJSONStr := "{}"

		productJson := json.RawMessage(productJsonStr)
		refJSON := json.RawMessage(refJSONStr)

		productID := poDt.ProductID

		var isVat, isPph23 *uint
		if poDt.IsVat != nil {
			isVatUint := uint(*poDt.IsVat)
			isVat = &isVatUint
		}

		if poDt.IsPph23 != nil {
			isPph23Uint := uint(*poDt.IsPph23)
			isPph23 = &isPph23Uint
		}

		margin := *poDt.PriceSell - *poDt.Price

		updatedItemUnits = append(updatedItemUnits, map[string]interface{}{
			"id":            poDt.ItemUnitID,
			"product_id":    poDt.ProductID,
			"unit_id":       poDt.ItemUnitUnitID,
			"price_buy":     poDt.Price,
			"conversion":    poDt.ItemUnitConversion,
			"margin":        margin,
			"status":        1,
			"updated_by_id": userID,
			"updated_at":    time.Now(),
		})

		poDtModel := models.PoDt{
			ProductUuid:              poDt.ProductUuid,
			PoID:                     &createdPurchaseOrder.ID,
			ItemUnitID:               poDt.ItemUnitID,
			VatID:                    poDt.VatID,
			Pph23ID:                  poDt.Pph23ID,
			RefID:                    poDt.RefID,
			RefSoDtID:                poDt.RefSoDtID,
			RefSoDtBomID:             poDt.RefSoDtBomID,
			RefProductID:             poDt.RefProductID,
			RefProductBomID:          poDt.RefProductBomID,
			ProductID:                &productID,
			BomID:                    poDt.BomID,
			ProductType:              poDt.ProductType,
			ProductJSON:              &productJson,
			RefType:                  poDt.RefType,
			RefJSON:                  &refJSON,
			GenCode:                  &genCode,
			Remark:                   poDt.Remark,
			NeedQty:                  poDt.NeedQty,
			Qty:                      poDt.Qty,
			QtyIn:                    new(float64),
			Price:                    poDt.Price,
			Subtotal:                 poDt.Subtotal,
			DiscountAmount:           poDt.DiscountAmount,
			DiscountPercentage:       poDt.DiscountPercentage,
			DiscountPercentageNum:    poDt.DiscountPercentageNum,
			DiscountPercentageAmount: poDt.DiscountPercentageAmount,
			DiscountFinal:            poDt.DiscountFinal,
			DiscountType:             poDt.DiscountType,
			VatPerc:                  poDt.VatPerc,
			VatPercAm:                poDt.VatPercAm,
			Pph23Perc:                poDt.Pph23Perc,
			Pph23PercAm:              poDt.Pph23PercAm,
			IsVat:                    isVat,
			IsPph23:                  isPph23,
			TotalAmount:              poDt.TotalAmount,
			CreatedByID:              &userID,
		}
		poDtsModel = append(poDtsModel, poDtModel)
	}

	return poDtsModel, updatedItemUnits, nil
}

func MapUpdatePoDts(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, updatedPurchaseOrder *models.PurchaseOrder, userID uint, span opentracing.Span) ([]models.PoDt, error) {
	poDtsModel := []models.PoDt{}

	for _, reqPoDt := range req.PoDts {
		poDtID := uint(0)
		if reqPoDt.PoDtID != nil {
			poDtID = *reqPoDt.PoDtID
		}

		productJsonStr := "{}"
		refJsonStr := "{}"
		productJson := json.RawMessage(productJsonStr)
		refJson := json.RawMessage(refJsonStr)

		productID := reqPoDt.ProductID

		var isVat, isPph23 *uint
		if reqPoDt.IsVat != nil {
			isVatUint := uint(*reqPoDt.IsVat)
			isVat = &isVatUint
		}

		if reqPoDt.IsPph23 != nil {
			isPph23Uint := uint(*reqPoDt.IsPph23)
			isPph23 = &isPph23Uint
		}

		poDtModel := models.PoDt{
			ID:                       poDtID,
			ProductUuid:              reqPoDt.ProductUuid,
			PoID:                     &updatedPurchaseOrder.ID,
			ItemUnitID:               reqPoDt.ItemUnitID,
			VatID:                    reqPoDt.VatID,
			Pph23ID:                  reqPoDt.Pph23ID,
			RefID:                    reqPoDt.RefID,
			RefSoDtID:                reqPoDt.RefSoDtID,
			RefSoDtBomID:             reqPoDt.RefSoDtBomID,
			RefRoDtID:                reqPoDt.RefRoDtID,
			RefProductID:             reqPoDt.RefProductID,
			RefProductBomID:          reqPoDt.RefProductBomID,
			ProductID:                &productID,
			BomID:                    reqPoDt.BomID,
			ProductType:              reqPoDt.ProductType,
			ProductJSON:              &productJson,
			RefType:                  reqPoDt.RefType,
			RefJSON:                  &refJson,
			GenCode:                  reqPoDt.GenCode,
			Remark:                   reqPoDt.Remark,
			NeedQty:                  reqPoDt.NeedQty,
			Qty:                      reqPoDt.Qty,
			Price:                    reqPoDt.Price,
			Subtotal:                 reqPoDt.Subtotal,
			DiscountAmount:           reqPoDt.DiscountAmount,
			DiscountPercentage:       reqPoDt.DiscountPercentage,
			DiscountPercentageNum:    reqPoDt.DiscountPercentageNum,
			DiscountPercentageAmount: reqPoDt.DiscountPercentageAmount,
			DiscountFinal:            reqPoDt.DiscountFinal,
			DiscountType:             reqPoDt.DiscountType,
			VatPerc:                  reqPoDt.VatPerc,
			VatPercAm:                reqPoDt.VatPercAm,
			Pph23Perc:                reqPoDt.Pph23Perc,
			Pph23PercAm:              reqPoDt.Pph23PercAm,
			IsVat:                    isVat,
			IsPph23:                  isPph23,
			TotalAmount:              reqPoDt.TotalAmount,
			UpdatedByID:              &userID,
		}
		poDtsModel = append(poDtsModel, poDtModel)
	}

	return poDtsModel, nil
}

func GetPoIDs(req dtos.FormPurchaseOrderRequest) ([]*uint, []*uint) {
	poDtIDs := []*uint{}
	productIDs := []*uint{}

	for _, reqPoDt := range req.PoDts {
		if reqPoDt.PoDtID != nil && *reqPoDt.PoDtID > 0 {
			poDtIDs = append(poDtIDs, reqPoDt.PoDtID)
		}

		if reqPoDt.ProductID > 0 {
			productID := reqPoDt.ProductID
			productIDs = append(productIDs, &productID)
		}
	}

	return poDtIDs, productIDs
}

func MapFilterPoDts(poDts []dtos.PurchaseOrderPoDtListDTO) []dtos.PurchaseOrderPoDtListDTO {
	combinedPoDts := []dtos.PurchaseOrderPoDtListDTO{}
	combinedPoDts = append(combinedPoDts, poDts...)
	return combinedPoDts
}

func MapNewUpdatePoDts(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, userID uint, span opentracing.Span) ([]uint, []uint, []uint, []uint, error) {
	refSoDtID := []uint{}
	refSoDtBomDtID := []uint{}
	refRoDtID := []uint{}
	refRoDtBomID := []uint{}

	for _, reqPoDt := range req.PoDts {
		if reqPoDt.RefSoDtID != nil {
			refSoDtID = append(refSoDtID, *reqPoDt.RefSoDtID)
		}

		if reqPoDt.RefSoDtBomID != nil {
			refSoDtBomDtID = append(refSoDtBomDtID, *reqPoDt.RefSoDtBomID)
		}

		if reqPoDt.RefRoDtID != nil {
			refRoDtID = append(refRoDtID, *reqPoDt.RefRoDtID)
		}

		// if reqPoDt.RefRoDtBomID != nil {
		// 	refRoDtBomID = append(refRoDtBomID, *reqPoDt.RefRoDtBomID)
		// }
	}

	return refSoDtID, refSoDtBomDtID, refRoDtID, refRoDtBomID, nil
}

func MapNewUpdatedReverseRefsPo(ctx *fiber.Ctx, oldInvDts []dtos.PurchaseOrderPoDtListDTO, soDts []map[string]interface{}, soDtBoms []map[string]interface{}, poDts []map[string]interface{}, poDtBoms []map[string]interface{}) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	refSoDt := []map[string]interface{}{}
	refSoDtBomDt := []map[string]interface{}{}
	refRoDt := []map[string]interface{}{}
	refRoDtBom := []map[string]interface{}{}

	for _, reqPoDt := range oldInvDts {
		if reqPoDt.RefSoDtID != nil && *reqPoDt.RefSoDtID > 0 {
			for _, soDt := range soDts {
				// log.Printf("soDt[\"id\"] value: %v, type: %T\n", soDt["id"], soDt["id"])

				if id, ok := soDt["id"].(int32); ok {
					// if *reqPoDt.RefSoDtID == soDt["id"] {
					if *reqPoDt.RefSoDtID == uint(id) {
						soDt["qty_po"] = soDt["qty_po"].(float64) - *reqPoDt.Qty
						refSoDt = append(refSoDt, soDt)
						break
					}
				}
			}
		}

		if reqPoDt.RefSoDtBomID != nil && *reqPoDt.RefSoDtBomID > 0 {
			for _, soDtBom := range soDtBoms {
				if id, ok := soDtBom["id"].(int32); ok {
					if *reqPoDt.RefSoDtBomID == uint(id) {
						soDtBom["qty_po"] = soDtBom["qty_po"].(float64) - *reqPoDt.Qty
						refSoDtBomDt = append(refSoDtBomDt, soDtBom)
						break
					}
				}
			}
		}

		if reqPoDt.RefRoDtID != nil && *reqPoDt.RefRoDtID > 0 {
			for _, poDt := range poDts {
				if id, ok := poDt["id"].(int32); ok {
					if *reqPoDt.RefRoDtID == uint(id) {
						poDt["qty_po"] = poDt["qty_po"].(float64) - *reqPoDt.Qty
						refRoDt = append(refRoDt, poDt)
						break
					}
				}
			}
		}

		if reqPoDt.RefRoDtBomID != nil && *reqPoDt.RefRoDtBomID > 0 {
			for _, poDtBom := range poDtBoms {
				if id, ok := poDtBom["id"].(int32); ok {
					if *reqPoDt.RefRoDtBomID == uint(id) {
						poDtBom["qty_in"] = poDtBom["qty_in"].(float64) - *reqPoDt.Qty
						refRoDtBom = append(refRoDtBom, poDtBom)
						break
					}
				}
			}
		}
	}

	return refSoDt, refSoDtBomDt, refRoDt, refRoDtBom, nil
}

func MapNewUpdatedRefsPo(ctx *fiber.Ctx, req dtos.FormPurchaseOrderRequest, soDts []map[string]interface{}, soDtBoms []map[string]interface{}, poDts []map[string]interface{}, poDtBoms []map[string]interface{}) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	refSoDt := []map[string]interface{}{}
	refSoDtBomDt := []map[string]interface{}{}
	refRoDt := []map[string]interface{}{}
	refRoDtBom := []map[string]interface{}{}

	for _, reqPoDt := range req.PoDts {
		if reqPoDt.RefSoDtID != nil && *reqPoDt.RefSoDtID > 0 {
			for _, soDt := range soDts {
				// log.Printf("soDt[\"id\"] value: %v, type: %T\n", soDt["id"], soDt["id"])

				if id, ok := soDt["id"].(int32); ok {
					// if *reqPoDt.RefSoDtID == soDt["id"] {
					if *reqPoDt.RefSoDtID == uint(id) {

						if soDt["qty_po"] == nil {
							soDt["qty_po"] = *new(float64)
						}
						soDt["qty_po"] = soDt["qty_po"].(float64) + *reqPoDt.Qty
						refSoDt = append(refSoDt, soDt)
						break
					}
				}
			}
		}

		if reqPoDt.RefSoDtBomID != nil && *reqPoDt.RefSoDtBomID > 0 {
			for _, soDtBom := range soDtBoms {
				if id, ok := soDtBom["id"].(int32); ok {
					if *reqPoDt.RefSoDtBomID == uint(id) {

						if soDtBom["qty_po"] == nil {
							soDtBom["qty_po"] = *new(float64)
						}
						soDtBom["qty_po"] = soDtBom["qty_po"].(float64) + *reqPoDt.Qty
						refSoDtBomDt = append(refSoDtBomDt, soDtBom)
						break
					}
				}
			}
		}

		if reqPoDt.RefRoDtID != nil && *reqPoDt.RefRoDtID > 0 {
			for _, poDt := range poDts {
				if id, ok := poDt["id"].(int32); ok {
					if *reqPoDt.RefRoDtID == uint(id) {

						if poDt["qty_po"] == nil {
							poDt["qty_po"] = *new(float64)
						}
						poDt["qty_po"] = poDt["qty_po"].(float64) + *reqPoDt.Qty
						refRoDt = append(refRoDt, poDt)
						break
					}
				}
			}
		}

		// if reqPoDt.RefRoDtBomID != nil && *reqPoDt.RefRoDtBomID > 0 {
		// 	for _, poDtBom := range poDtBoms {
		// 		if id, ok := poDtBom["id"].(int32); ok {
		// 			if *reqPoDt.RefRoDtBomID == uint(id) {

		// 				if poDtBom["qty_in"] == nil {
		// 					poDtBom["qty_in"] = *new(float64)
		// 				}
		// 				poDtBom["qty_in"] = poDtBom["qty_in"].(float64) + *reqPoDt.Qty
		// 				refRoDtBom = append(refRoDtBom, poDtBom)
		// 				break
		// 			}
		// 		}
		// 	}
		// }
	}

	return refSoDt, refSoDtBomDt, refRoDt, refRoDtBom, nil
}

func MapOldUpdatePoDts(ctx *fiber.Ctx, oldInvDts []dtos.PurchaseOrderPoDtListDTO, userID uint, span opentracing.Span) ([]uint, []uint, []uint, []uint, error) {
	refSoDtID := []uint{}
	refSoDtBomDtID := []uint{}
	refRoDtID := []uint{}
	refRoDtBomID := []uint{}

	for _, reqInvDt := range oldInvDts {
		if reqInvDt.RefSoDtID != nil {
			refSoDtID = append(refSoDtID, *reqInvDt.RefSoDtID)
		}

		if reqInvDt.RefSoDtBomID != nil {
			refSoDtBomDtID = append(refSoDtBomDtID, *reqInvDt.RefSoDtBomID)
		}

		if reqInvDt.RefRoDtID != nil {
			refRoDtID = append(refRoDtID, *reqInvDt.RefRoDtID)
		}

		if reqInvDt.RefRoDtBomID != nil {
			refRoDtBomID = append(refRoDtBomID, *reqInvDt.RefRoDtBomID)
		}

	}

	return refSoDtID, refSoDtBomDtID, refRoDtID, refRoDtBomID, nil
}

// Get condition for quotation
func GetPurchaseOrderCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetPurchaseOrderCondition", opentracing.ChildOf(span.Context()))
	var err error

	joinCondition := ""
	customCondition := ""
	condition := ""
	var args []interface{}
	queryGlobal := ""
	i := 1

	filterDBColumnKey := []string{
		"po.po_no", "po.remark", "po.shipping_destination",
		"p.name",
		"pd.remark",
		"pd.gen_code",
	}

	if value, ok := filters["global"]; ok && value != "" {

		queryGlobal = " AND ("
		for idx, column := range filterDBColumnKey {
			if idx > 0 {
				queryGlobal += " OR"
			}
			queryGlobal += fmt.Sprintf(" %s ILIKE $%d", column, i)
			args = append(args, "%"+value+"%")
			i++
		}
		queryGlobal += ")"
	}

	if filters["ids"] != "" {
		condition += fmt.Sprintf(" AND id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":           "po.status",
		"customer_id":      "po.customer_id",
		"purchase_type_id": "po.purchase_type_id",
		"currency_id":      "po.currency_id",
		"vat_id":           "po.vat_id",
		"payment_term_id":  "po.payment_term_id",
		"shipping_term_id": "po.shipping_term_id",
		"pph23_id":         "po.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":      "po.customer_id",
		"purchase_type_ids": "po.purchase_type_id",
		"currency_ids":      "po.currency_id",
		"payment_term_ids":  "po.payment_term_id",
		"shipping_term_ids": "po.shipping_term_id",
		"pph23_ids":         "po.pph23_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			ids := strings.Split(value, ",")
			intIDs, err := SplitStringArrayOfInts(ids)
			if err != nil {
				LogErrors(childSpan, err)
				return nil, 0, "", "", "", "", err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs))
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": {"po.vat_id", "pd.vat_id"},
	}

	for key, valueIDs := range filterIDsOrKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s IN ($%d)", valueID, i)
				args = append(args, value)
			}
			condition += ")"
		}
	}

	// And one for array conditions with OR
	filterIDsOrArrayKey := map[string][]string{
		"product_ids": {"p.id"},
	}

	// Handle array OR conditions
	for key, valueIDs := range filterIDsOrArrayKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := SplitStringArrayOfInts(ids)
			if err != nil {
				LogErrors(childSpan, err)
				return nil, 0, "", "", "", "", err
			}

			condition += " AND ("
			for idx, valueID := range valueIDs {
				if idx > 0 {
					condition += " OR"
				}
				condition += fmt.Sprintf(" %s = ANY($%d)", valueID, i)
				args = append(args, pq.Array(intIDs))
				i++
			}
			condition += ")"
		}
	}

	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		dateColumn := ""
		switch filters["date_type"] {
		case "1":
			dateColumn = "po.po_date"
		case "2":
			dateColumn = "po.delivery_date"
		default:
			dateColumn = "po.po_date"
		}

		condition += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", dateColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	return args, i, condition, queryGlobal, joinCondition, customCondition, err
}

func GetPurchaseOrderDetailsIDs(quotations []dtos.PurchaseOrderDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToPurchaseOrders(quoDts []dtos.PurchaseOrderPoDtListDTO, quotations []dtos.PurchaseOrderDetailDTO) []dtos.PurchaseOrderDetailDTO {
	// quotations := []dtos.PurchaseOrderAttachmentsDTO{}
	for i, quotation := range quotations {
		newSoDts := make([]dtos.PurchaseOrderPoDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.PoID == quotation.ID {
				newSoDts = append(newSoDts, quoDt)
			}
		}
		quotation.PoDts = newSoDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetPurchaseOrderDetails(salesOrders []dtos.PurchaseOrderDetailDTO, soDts []dtos.PurchaseOrderPoDtListDTO) []dtos.PurchaseOrderDetailDTO {
	// salesOrders := []dtos.PurchaseOrderAttachmentsDTO{}
	for i, salesOrder := range salesOrders {
		newSoDts := make([]dtos.PurchaseOrderPoDtListDTO, 0)
		for _, soDt := range soDts {
			if *soDt.PoID == salesOrder.ID {
				newSoDts = append(newSoDts, soDt)
			}
		}
		salesOrder.PoDts = newSoDts
		salesOrders[i] = salesOrder
	}

	return salesOrders
}

// Build CSV rows, dtos.PurchaseOrderListDTO, csv pointer
func BuildPurchaseOrderAllCSVRows(salesOrders []dtos.PurchaseOrderListDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"ID", "Purchase No", "Supplier", "PO Date", "Delivery Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
	}
	rows = append(rows, header)

	for _, salesOrder := range salesOrders {
		ID := fmt.Sprintf("%d", salesOrder.ID)
		PoNo := GetPtrVal(salesOrder.PoNo)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		PoDate := GetPtrVal(salesOrder.PoDate)
		DeliveryDate := GetPtrVal(salesOrder.DeliveryDate)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(&salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		// EscapeCsvField
		ID = EscapeCsvField(ID)
		PoNo = EscapeCsvField(PoNo)
		CustomerName = EscapeCsvField(CustomerName)
		PoDate = EscapeCsvField(PoDate)
		DeliveryDate = EscapeCsvField(DeliveryDate)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		row := []string{
			ID, PoNo, CustomerName, PoDate, DeliveryDate,
			CurrencyName, fmt.Sprintf("%f", *salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
		}

		rows = append(rows, row)
	}

	// Convert rows to CSV format
	csvContent := ""
	for _, row := range rows {
		csvContent += strings.Join(row, ";") + "\n"
	}

	*csv = csvContent

	return nil
}

// Build CSV rows, dtos.PurchaseOrderListDTO, csv pointer
func BuildPurchaseOrderDetailCSVRows(salesOrders []dtos.PurchaseOrderDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Purchase No", "Customer", "PO Date", "Delivery Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
		"Product/Item Name", "Qty", "Price", "Subtotal",
	}
	rows = append(rows, header)

	for iPurchaseOrder, salesOrder := range salesOrders {

		No := fmt.Sprintf("%d", iPurchaseOrder+1)
		PoNo := GetPtrVal(salesOrder.PoNo)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		PoDate := GetPtrVal(salesOrder.PoDate)
		DeliveryDate := GetPtrVal(salesOrder.DeliveryDate)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(&salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		No = EscapeCsvField(No)
		PoNo = EscapeCsvField(PoNo)
		CustomerName = EscapeCsvField(CustomerName)
		PoDate = EscapeCsvField(PoDate)
		DeliveryDate = EscapeCsvField(DeliveryDate)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		for iSoDt, quoDt := range salesOrder.PoDts {
			ProductItemName := GetPtrVal(quoDt.ItemName)
			ProductItemName = EscapeCsvField(ProductItemName)
			Qty := fmt.Sprintf("%f", *quoDt.Qty)
			PriceSell := fmt.Sprintf("%f", *quoDt.Price)
			TotalAm := fmt.Sprintf("%f", *quoDt.Subtotal)

			if iSoDt == 0 {
				row := []string{
					No, PoNo, CustomerName, PoDate, DeliveryDate,
					CurrencyName, fmt.Sprintf("%f", salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
					ProductItemName, Qty, PriceSell, TotalAm,
				}
				rows = append(rows, row)
			} else {
				row := []string{
					"", "", "", "", "", "",
					"", "", "", "",
					ProductItemName, Qty, PriceSell, TotalAm,
				}
				rows = append(rows, row)
			}
		}
		if len(salesOrder.PoDts) == 0 {
			row := []string{
				No, PoNo, CustomerName, PoDate, DeliveryDate,
				CurrencyName, fmt.Sprintf("%f", salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
				"", "", "", "",
			}
			rows = append(rows, row)
		}
	}

	// Convert rows to CSV format
	csvContent := ""
	for _, row := range rows {
		csvContent += strings.Join(row, ";") + "\n"
	}

	*csv = csvContent

	return nil
}
