package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func GetInvIDs(req dtos.FormInventoryRequest) ([]*uint, []*uint, []*uint) {
	invDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.InvDtID != nil && *reqInvDt.InvDtID > 0 {
			invDtIDs = append(invDtIDs, reqInvDt.InvDtID)
		}
	}

	return invDtIDs, productIDs, itemUnitIDs
}

func GetLockInventorySoIDs(req dtos.FormInventoryRequest) ([]*uint, []*uint) {
	soDtIDs := []*uint{}
	soDtBomIDs := []*uint{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefType == "so" && reqInvDt.RefSoDtID != nil && *reqInvDt.RefSoDtID > 0 {
			soDtIDs = append(soDtIDs, reqInvDt.RefSoDtID)
		}
		if reqInvDt.RefType == "so" && reqInvDt.RefSoDtBomID != nil && *reqInvDt.RefSoDtBomID > 0 {
			soDtBomIDs = append(soDtBomIDs, reqInvDt.RefSoDtBomID)
		}
	}

	return soDtIDs, soDtBomIDs
}

func MapCreateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, createdInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {
	invDtsModel := []models.InvDt{}

	for _, invDt := range req.InvDts {
		genCode := "-"

		itemObj := map[string]interface{}{
			"item_code": invDt.ItemCode,
			"item_name": invDt.ItemName,
		}

		itemJson, err := json.Marshal(itemObj)
		if err != nil {
			return invDtsModel, err
		}

		if invDt.ExpiredAt != nil && *invDt.ExpiredAt == "" {
			invDt.ExpiredAt = nil
		}

		invDtModel := models.InvDt{
			ProductUuid:     invDt.ProductUuid,
			InventoryID:     createdInventory.ID,
			ItemUnitID:      invDt.ItemUnitID,
			VatID:           invDt.VatID,
			Pph23ID:         invDt.Pph23ID,
			RefSoDtID:       invDt.RefSoDtID,
			RefSoDtBomID:    invDt.RefSoDtBomID,
			RefRoDtID:       invDt.RefRoDtID,
			RefPoDtID:       invDt.RefPoDtID,
			RefPoDtBomID:    invDt.RefPoDtBomID,
			RefInvDtID:      invDt.RefInvDtID,
			RefProductID:    invDt.RefProductID,
			RefProductBomID: invDt.RefProductBomID,
			ItemID:          invDt.ItemID,
			RefType:         invDt.RefType,
			ItemType:        invDt.ItemType,
			ItemJSON:        string(itemJson),
			GenCode:         &genCode,
			Remark:          invDt.Remark,
			VatPerc:         invDt.VatPerc,
			VatPercAm:       invDt.VatPercAm,
			Pph23Perc:       invDt.Pph23Perc,
			Pph23PercAm:     invDt.Pph23PercAm,
			IsVat:           invDt.IsVat,
			IsPph23:         invDt.IsPph23,
			Qty:             invDt.Qty,
			PriceSell:       invDt.PriceSell,
			PriceBuy:        invDt.PriceBuy,
			SubtotalSell:    invDt.SubtotalSell,
			SubtotalBuy:     invDt.SubtotalBuy,
			ExpiredAt:       invDt.ExpiredAt,
			// TotalAm:      invDt.TotalAm,
			CreatedByID: &userID,
		}
		invDtsModel = append(invDtsModel, invDtModel)

	}

	return invDtsModel, nil
}

func MapUpdateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, updatedInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {

	invDtsModel := []models.InvDt{}

	for _, reqInvDt := range req.InvDts {
		invDtID := uint(0)
		log.Println("1invDtID: ", reqInvDt.InvDtID)
		if reqInvDt.InvDtID != nil {
			invDtID = *reqInvDt.InvDtID
			log.Println("2invDtID: ", invDtID)
		}

		itemObj := map[string]interface{}{
			"item_code": reqInvDt.ItemCode,
			"item_name": reqInvDt.ItemName,
		}

		itemJson, err := json.Marshal(itemObj)
		if err != nil {
			return invDtsModel, err
		}

		invDtModel := models.InvDt{
			ID:              invDtID,
			ProductUuid:     reqInvDt.ProductUuid,
			InventoryID:     updatedInventory.ID,
			ItemUnitID:      reqInvDt.ItemUnitID,
			VatID:           reqInvDt.VatID,
			Pph23ID:         reqInvDt.Pph23ID,
			RefSoDtID:       reqInvDt.RefSoDtID,
			RefSoDtBomID:    reqInvDt.RefSoDtBomID,
			RefPoDtID:       reqInvDt.RefPoDtID,
			RefPoDtBomID:    reqInvDt.RefPoDtBomID,
			RefInvDtID:      reqInvDt.RefInvDtID,
			RefProductID:    reqInvDt.RefProductID,
			RefProductBomID: reqInvDt.RefProductBomID,
			ItemID:          reqInvDt.ItemID,
			RefType:         reqInvDt.RefType,
			ItemType:        reqInvDt.ItemType,
			ItemJSON:        string(itemJson),
			GenCode:         reqInvDt.GenCode,
			Remark:          reqInvDt.Remark,
			VatPerc:         reqInvDt.VatPerc,
			VatPercAm:       reqInvDt.VatPercAm,
			Pph23Perc:       reqInvDt.Pph23Perc,
			Pph23PercAm:     reqInvDt.Pph23PercAm,
			IsVat:           reqInvDt.IsVat,
			IsPph23:         reqInvDt.IsPph23,
			Qty:             reqInvDt.Qty,
			PriceSell:       reqInvDt.PriceSell,
			PriceBuy:        reqInvDt.PriceBuy,
			SubtotalSell:    reqInvDt.SubtotalSell,
			SubtotalBuy:     reqInvDt.SubtotalBuy,
			ExpiredAt:       reqInvDt.ExpiredAt,
			// QtyOut:       reqInvDt.QtyOut,
			QtyInvoice:  reqInvDt.QtyInvoice,
			TotalAm:     reqInvDt.TotalAm,
			CreatedByID: &userID,
		}
		invDtsModel = append(invDtsModel, invDtModel)

	}

	return invDtsModel, nil
}

func MapOldUpdateInvDts(ctx *fiber.Ctx, oldInvDts []dtos.InventoryInvDtListDTO, userID uint, span opentracing.Span) ([]uint, []uint, []uint, []uint, []uint, []uint, error) {
	refSoDtID := []uint{}
	refSoDtBomDtID := []uint{}
	refRoDtID := []uint{}
	refPoDtID := []uint{}
	refPoDtBomID := []uint{}
	refInvDtID := []uint{}

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

		if reqInvDt.RefPoDtID != nil {
			refPoDtID = append(refPoDtID, *reqInvDt.RefPoDtID)
		}

		if reqInvDt.RefPoDtBomID != nil {
			refPoDtBomID = append(refPoDtBomID, *reqInvDt.RefPoDtBomID)
		}

		if reqInvDt.RefInvDtID != nil {
			refInvDtID = append(refInvDtID, *reqInvDt.RefInvDtID)
		}
	}

	return refSoDtID, refSoDtBomDtID, refRoDtID, refPoDtID, refPoDtBomID, refInvDtID, nil
}

func MapNewUpdateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, span opentracing.Span) ([]uint, []uint, []uint, []uint, []uint, []uint, error) {
	refSoDtID := []uint{}
	refSoDtBomDtID := []uint{}
	refRoDtID := []uint{}
	refPoDtID := []uint{}
	refPoDtBomID := []uint{}
	refInvDtID := []uint{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefSoDtID != nil {
			refSoDtID = append(refSoDtID, *reqInvDt.RefSoDtID)
		}

		if reqInvDt.RefSoDtBomID != nil {
			refSoDtBomDtID = append(refSoDtBomDtID, *reqInvDt.RefSoDtBomID)
		}

		if reqInvDt.RefRoDtID != nil {
			refRoDtID = append(refRoDtID, *reqInvDt.RefRoDtID)
		}

		if reqInvDt.RefPoDtID != nil {
			refPoDtID = append(refPoDtID, *reqInvDt.RefPoDtID)
		}

		if reqInvDt.RefPoDtBomID != nil {
			refPoDtBomID = append(refPoDtBomID, *reqInvDt.RefPoDtBomID)
		}

		if reqInvDt.RefInvDtID != nil {
			refInvDtID = append(refInvDtID, *reqInvDt.RefInvDtID)
		}
	}

	return refSoDtID, refSoDtBomDtID, refRoDtID, refPoDtID, refPoDtBomID, refInvDtID, nil
}

func MapNewUpdatedReverseRefs(ctx *fiber.Ctx, oldInvDts []dtos.InventoryInvDtListDTO, soDts []map[string]interface{}, soDtBoms []map[string]interface{}, roDts []map[string]interface{}, poDts []map[string]interface{}, poDtBoms []map[string]interface{}, invDts []map[string]interface{}) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	refSoDt := []map[string]interface{}{}
	refSoDtBomDt := []map[string]interface{}{}
	refRoDt := []map[string]interface{}{}
	refPoDt := []map[string]interface{}{}
	refPoDtBom := []map[string]interface{}{}
	refInvDt := []map[string]interface{}{}

	for _, reqInvDt := range oldInvDts {
		if reqInvDt.RefSoDtID != nil && *reqInvDt.RefSoDtID > 0 {
			for _, soDt := range soDts {
				// log.Printf("soDt[\"id\"] value: %v, type: %T\n", soDt["id"], soDt["id"])

				if id, ok := soDt["id"].(int32); ok {
					// if *reqInvDt.RefSoDtID == soDt["id"] {
					if *reqInvDt.RefSoDtID == uint(id) {
						soDt["qty_out"] = soDt["qty_out"].(float64) - *reqInvDt.Qty
						refSoDt = append(refSoDt, soDt)
						break
					}
				}
			}
		}

		if reqInvDt.RefSoDtBomID != nil && *reqInvDt.RefSoDtBomID > 0 {
			for _, soDtBom := range soDtBoms {
				if id, ok := soDtBom["id"].(int32); ok {
					if *reqInvDt.RefSoDtBomID == uint(id) {
						soDtBom["qty_out"] = soDtBom["qty_out"].(float64) - *reqInvDt.Qty
						refSoDtBomDt = append(refSoDtBomDt, soDtBom)
						break
					}
				}
			}
		}

		if reqInvDt.RefRoDtID != nil && *reqInvDt.RefRoDtID > 0 {
			for _, roDt := range roDts {
				// log.Printf("roDt[\"id\"] value: %v, type: %T\n", roDt["id"], roDt["id"])

				if id, ok := roDt["id"].(int32); ok {
					// if *reqInvDt.RefRoDtID == roDt["id"] {
					if *reqInvDt.RefRoDtID == uint(id) {
						roDt["qty_out"] = roDt["qty_out"].(float64) - *reqInvDt.Qty
						refRoDt = append(refRoDt, roDt)
						break
					}
				}
			}
		}

		if reqInvDt.RefPoDtID != nil && *reqInvDt.RefPoDtID > 0 {
			for _, poDt := range poDts {
				log.Printf("soDt[\"id\"] value: %v, type: %T\n", poDt["id"], poDt["id"])
				// if id, ok := ParseMapIDToUint(poDt["id"]); ok {
				if id, ok := poDt["id"].(int32); ok {
					if *reqInvDt.RefPoDtID == uint(id) {
						// poDt["qty_in"] = poDt["qty_in"].(float64) - *reqInvDt.Qty
						// refPoDt = append(refPoDt, poDt)
						// break
						currentQty := ParseMapQtyToFloat64(poDt["qty_in"])
						poDt["qty_in"] = currentQty - *reqInvDt.Qty
						refPoDt = append(refPoDt, poDt)
					}
				}
			}
		}

		if reqInvDt.RefPoDtBomID != nil && *reqInvDt.RefPoDtBomID > 0 {
			for _, poDtBom := range poDtBoms {
				if id, ok := poDtBom["id"].(int32); ok {
					if *reqInvDt.RefPoDtBomID == uint(id) {
						poDtBom["qty_in"] = poDtBom["qty_in"].(float64) - *reqInvDt.Qty
						refPoDtBom = append(refPoDtBom, poDtBom)
						break
					}
				}
			}
		}

		if reqInvDt.RefInvDtID != nil && *reqInvDt.RefInvDtID > 0 {
			for _, invDt := range invDts {
				if id, ok := ParseMapIDToUint(invDt["id"]); ok {
					if *reqInvDt.RefInvDtID == id {
						// invDt["qty_out"] = invDt["qty_out"].(float64) - *reqInvDt.Qty
						currentQty := ParseMapQtyToFloat64(invDt["qty_out"])
						invDt["qty_out"] = currentQty - *reqInvDt.Qty
						refInvDt = append(refInvDt, invDt)
						break
					}
				}
			}
		}
	}

	return refSoDt, refSoDtBomDt, refRoDt, refPoDt, refPoDtBom, refInvDt, nil
}

func MapNewUpdatedRefs(ctx *fiber.Ctx, req dtos.FormInventoryRequest, soDts []map[string]interface{}, soDtBoms []map[string]interface{}, roDts []map[string]interface{}, poDts []map[string]interface{}, poDtBoms []map[string]interface{}, invDts []map[string]interface{}) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	refSoDt := []map[string]interface{}{}
	refSoDtBomDt := []map[string]interface{}{}
	refRoDt := []map[string]interface{}{}
	refPoDt := []map[string]interface{}{}
	refPoDtBom := []map[string]interface{}{}
	refInvDt := []map[string]interface{}{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefSoDtID != nil && *reqInvDt.RefSoDtID > 0 {
			for _, soDt := range soDts {
				// log.Printf("soDt[\"id\"] value: %v, type: %T\n", soDt["id"], soDt["id"])

				if id, ok := soDt["id"].(int32); ok {
					// if *reqInvDt.RefSoDtID == soDt["id"] {
					if *reqInvDt.RefSoDtID == uint(id) {

						if soDt["qty_out"] == nil {
							soDt["qty_out"] = *new(float64)
						}
						soDt["qty_out"] = soDt["qty_out"].(float64) + *reqInvDt.Qty
						refSoDt = append(refSoDt, soDt)
						break
					}
				}
			}
		}

		if reqInvDt.RefSoDtBomID != nil && *reqInvDt.RefSoDtBomID > 0 {
			for _, soDtBom := range soDtBoms {
				if id, ok := soDtBom["id"].(int32); ok {
					if *reqInvDt.RefSoDtBomID == uint(id) {

						if soDtBom["qty_out"] == nil {
							soDtBom["qty_out"] = *new(float64)
						}
						soDtBom["qty_out"] = soDtBom["qty_out"].(float64) + *reqInvDt.Qty
						refSoDtBomDt = append(refSoDtBomDt, soDtBom)
						break
					}
				}
			}
		}

		if reqInvDt.RefRoDtID != nil && *reqInvDt.RefRoDtID > 0 {
			for _, roDt := range roDts {
				// log.Printf("roDt[\"id\"] value: %v, type: %T\n", roDt["id"], roDt["id"])

				if id, ok := roDt["id"].(int32); ok {
					// if *reqInvDt.RefRoDtID == roDt["id"] {
					if *reqInvDt.RefRoDtID == uint(id) {

						if roDt["qty_out"] == nil {
							roDt["qty_out"] = *new(float64)
						}
						log.Println("reqInvDt.Qty: ", *reqInvDt.Qty)
						roDt["qty_out"] = roDt["qty_out"].(float64) + *reqInvDt.Qty
						refRoDt = append(refRoDt, roDt)
						break
					}
				}
			}
		}
		if reqInvDt.RefPoDtID != nil && *reqInvDt.RefPoDtID > 0 {
			for _, poDt := range poDts {
				if id, ok := poDt["id"].(int32); ok {
					if *reqInvDt.RefPoDtID == uint(id) {

						if poDt["qty_in"] == nil {
							poDt["qty_in"] = *new(float64)
						}
						poDt["qty_in"] = poDt["qty_in"].(float64) + *reqInvDt.Qty
						refPoDt = append(refPoDt, poDt)
						break
					}
				}
			}
		}

		if reqInvDt.RefPoDtBomID != nil && *reqInvDt.RefPoDtBomID > 0 {
			for _, poDtBom := range poDtBoms {
				if id, ok := poDtBom["id"].(int32); ok {
					if *reqInvDt.RefPoDtBomID == uint(id) {

						if poDtBom["qty_in"] == nil {
							poDtBom["qty_in"] = *new(float64)
						}
						poDtBom["qty_in"] = poDtBom["qty_in"].(float64) + *reqInvDt.Qty
						refPoDtBom = append(refPoDtBom, poDtBom)
						break
					}
				}
			}
		}

		if reqInvDt.RefInvDtID != nil && *reqInvDt.RefInvDtID > 0 {
			for _, invDt := range invDts {
				// if id, ok := invDt["id"].(int32); ok {
				if id, ok := ParseMapIDToUint(invDt["id"]); ok {
					if *reqInvDt.RefInvDtID == uint(id) {

						if invDt["qty_out"] == nil {
							invDt["qty_out"] = 0.0
						}

						currentQty := ParseMapQtyToFloat64(invDt["qty_out"])
						newQty := currentQty + *reqInvDt.Qty
						invDt["qty_out"] = newQty

						refInvDt = append(refInvDt, invDt)
						break
					}
				}
			}
		}
	}

	return refSoDt, refSoDtBomDt, refRoDt, refPoDt, refPoDtBom, refInvDt, nil
}

func GenInventoryNo(ctx *fiber.Ctx, req dtos.FormInventoryRequest, orderedNumber int, span opentracing.Span) string {
	if req.InventoryNo != nil {
		return *req.InventoryNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	surname := req.CustomerCode
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("IVT/%s/%s-%s-%s", order, surname, year, month)

	return str
}

func MapCreateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, branchID uint, customerInvCreatedThisMonthNumber int, span opentracing.Span) (models.Inventory, error) {
	// order := 1
	customerInvCreatedThisMonthNumber++

	// orderNo := GenInventoryNo(ctx, req, customerInvCreatedThisMonthNumber, span)
	inventoryNo := GenerateInventoryNoOnCreateInventory(ctx, req, customerInvCreatedThisMonthNumber, span)

	salesOrder := models.Inventory{
		CustomerID:     req.CustomerID,
		CurrencyID:     req.CurrencyID,
		IOTypeID:       req.IoTypeID,
		PaymentTermID:  req.PaymentTermID,
		VatID:          req.VatID,
		Pph23ID:        req.Pph23ID,
		WarehouseID:    req.WarehouseID,
		InventoryNo:    &inventoryNo,
		InventoryNoOri: &inventoryNo,
		DoNo:           req.DoNo,
		SuratJalanNo:   req.SuratJalanNo,
		InvoiceNo:      req.InvoiceNo,
		ShipDest:       req.ShipDest,
		Remark:         req.Remark,
		Status:         req.Status,
		ExchangeRate:   req.ExchangeRate,
		IsVat:          req.IsVat,
		VatPerc:        req.VatPerc,
		Pph23Perc:      req.Pph23Perc,
		TotalQty:       req.TotalQty,
		Subtotal:       req.Subtotal,
		TotalPph23:     req.TotalPph23,
		TotalVat:       req.TotalVat,
		GrandTotal:     req.GrandTotal,
		DoAt:           req.DoAt,
		IngoingAt:      req.IngoingAt,
		InvoiceAt:      req.InvoiceAt,
		BranchID:       &branchID,
		CreatedByID:    &userID,
	}

	return salesOrder, nil
}

func MapUpdateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, branchID uint, span opentracing.Span) (models.Inventory, error) {
	revNo := 0
	if req.RevNo != nil {
		revNo = *req.RevNo
	}
	revNo++

	// salesOrderNo := GenerateInvNoOnUpdateSo(ctx, req, req.InventoryNo, revNo, span)
	inventoryNo := GenerateInvNoOnUpdateSo(ctx, req, req.InventoryNo, revNo, span)

	salesOrder := models.Inventory{
		ID:             *req.ID,
		CustomerID:     req.CustomerID,
		CurrencyID:     req.CurrencyID,
		IOTypeID:       req.IoTypeID,
		PaymentTermID:  req.PaymentTermID,
		VatID:          req.VatID,
		Pph23ID:        req.Pph23ID,
		WarehouseID:    req.WarehouseID,
		InventoryNo:    &inventoryNo,
		InventoryNoOri: &inventoryNo,
		DoNo:           req.DoNo,
		SuratJalanNo:   req.SuratJalanNo,
		InvoiceNo:      req.InvoiceNo,
		ShipDest:       req.ShipDest,
		Remark:         req.Remark,
		Status:         req.Status,
		ExchangeRate:   req.ExchangeRate,
		IsVat:          req.IsVat,
		VatPerc:        req.VatPerc,
		Pph23Perc:      req.Pph23Perc,
		TotalQty:       req.TotalQty,
		Subtotal:       req.Subtotal,
		TotalPph23:     req.TotalPph23,
		TotalVat:       req.TotalVat,
		GrandTotal:     req.GrandTotal,
		DoAt:           req.DoAt,
		IngoingAt:      req.IngoingAt,
		InvoiceAt:      req.InvoiceAt,
		BranchID:       &branchID,
		UpdatedByID:    &userID,
		RevNo:          &revNo,
	}

	return salesOrder, nil
}

func GenerateInvNoOnUpdateSo(ctx *fiber.Ctx, req dtos.FormInventoryRequest, inventoryNo *string, revNo int, span opentracing.Span) string {

	// get before REV-number, full string is SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1 or SURNAME-2001-12-20
	// check if "REV" string exist (random), if not add "REV-1" else replace REV-1 change the number to increment REV-revNo
	if !strings.Contains(*inventoryNo, "REV") {
		*inventoryNo = fmt.Sprintf("%s/REV-1", *inventoryNo)
	} else {
		// remove after /REV
		// use split to get the first part of string
		*inventoryNo = strings.Split(*inventoryNo, "/REV")[0]
		*inventoryNo = fmt.Sprintf("%s/REV-%d", *inventoryNo, revNo)
	}

	return *inventoryNo
}

func GenerateInventoryNoOnCreateInventory(ctx *fiber.Ctx, req dtos.FormInventoryRequest, orderedNumber int, span opentracing.Span) string {
	if req.InventoryNo != nil {
		return *req.InventoryNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	surname := req.CustomerCode
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	order := fmt.Sprintf("%d", orderedNumber)

	// str := fmt.Sprintf("%s-%s-%s-%s", surname, year, month, order)
	str := fmt.Sprintf("IVT/%s/%s-%s-%s", surname, year, month, order)

	return str
}

func MapUpdateInvSoDtsQty(invDtsQtyUpdate []dtos.GetInvSoDtQtyUpdateDTO, req dtos.FormInventoryRequest) []map[string]interface{} {
	// filter with ID to bulk update
	bulkUpdateSoDts := []map[string]interface{}{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefType != "so" {
			continue
		}

		for _, invDts := range invDtsQtyUpdate {

			if invDts.QtyOut == nil {
				invDts.QtyOut = new(float64)
			}

			log.Println("SoDtID: ", invDts.SoDtID, "RefSoDtID: ", reqInvDt.RefSoDtID)

			if reqInvDt.RefSoDtID != nil && *reqInvDt.RefSoDtID == *invDts.SoDtID {
				newSoDt := map[string]interface{}{
					"id":      invDts.SoDtID,
					"qty_out": (*reqInvDt.Qty + *invDts.QtyOut),
				}
				bulkUpdateSoDts = append(bulkUpdateSoDts, newSoDt)
			}
		}
	}

	return bulkUpdateSoDts
}

func MapUpdateInvSoDtBomsQty(invDtsQtyUpdate []dtos.GetInvSoDtQtyUpdateDTO, req dtos.FormInventoryRequest) []map[string]interface{} {
	// filter with ID to bulk update
	bulkUpdateQuoDts := []map[string]interface{}{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefType != "so" {
			continue
		}

		for _, invDts := range invDtsQtyUpdate {

			if invDts.QtyOut == nil {
				invDts.QtyOut = new(float64)
			}

			log.Println("SoDtBomID: ", invDts.SoDtBomID, "RefSoDtBomID: ", reqInvDt.RefSoDtBomID)

			if reqInvDt.RefSoDtBomID != nil && *reqInvDt.RefSoDtBomID == *invDts.SoDtBomID {
				newQuoDt := map[string]interface{}{
					"id":      invDts.SoDtBomID,
					"qty_out": (*reqInvDt.Qty + *invDts.QtyOut),
				}
				bulkUpdateQuoDts = append(bulkUpdateQuoDts, newQuoDt)
			}
		}
	}

	return bulkUpdateQuoDts
}

func GetInvSoDtIDs(invDts []dtos.RefInvIndexSoDtListDTO) []uint {
	quotationIDs := []uint{}

	for _, invDt := range invDts {
		quotationIDs = append(quotationIDs, *invDt.SalesOrderID)
	}

	return quotationIDs
}

// func MapInvRefSoDtBomsToQuoDts(quoDtBoms []dtos.InvSalesOrderQuoDtBomListDTO, quoDts []dtos.RefInvIndexSoDtListDTO) []dtos.RefInvIndexSoDtListDTO {
// 	combinedQuoDts := []dtos.RefInvIndexSoDtListDTO{}

// 	for _, quoDt := range quoDts {
// 		newQuoDtBoms := make([]dtos.InvSalesOrderQuoDtBomListDTO, 0)
// 		for _, quoDtBom := range quoDtBoms {
// 			if *quoDtBom.QuoDtID == *quoDt.ID {
// 				quoDtBoms = append(quoDtBoms, quoDtBom)
// 				newQuoDtBoms = append(newQuoDtBoms, quoDtBom)
// 			}
// 		}

// 		quoDt.QuoDtsBoms = newQuoDtBoms
// 		combinedQuoDts = append(combinedQuoDts, quoDt)
// 	}

// 	return combinedQuoDts
// }

func MapInvRefHeadUpdateStatus(heads []map[string]interface{}, refType string) []map[string]interface{} {
	updatedHeads := []map[string]interface{}{}
	for _, head := range heads {
		if refType == "po" {
			status := determinePoStatusInvRefHead(head)
			updatedHeads = append(updatedHeads, map[string]interface{}{
				"id":     head["id"],
				"status": status,
			})

		}
	}

	return updatedHeads
}

func determinePoStatusInvRefHead(head map[string]interface{}) string {
	// Default status
	status := "PROCESS"

	qtyIn, totalQty, ok := getInvRefHeadQuantities(head)
	if !ok {
		return status
	}

	switch {
	case qtyIn == 0:
		status = "PROCESS"
	case qtyIn > 0 && qtyIn < totalQty:
		status = "PARTIAL"
	case qtyIn >= totalQty:
		status = "FINISH"
	}

	return status
}

func getInvRefHeadQuantities(head map[string]interface{}) (qtyIn float64, totalQty float64, ok bool) {
	// Check if both quantities exist
	if head["qty_in"] == nil || head["total_qty"] == nil {
		return 0, 0, false
	}

	qtyIn = head["qty_in"].(float64)
	totalQty = head["total_qty"].(float64)

	return qtyIn, totalQty, true
}

func GetInventoryDtItemIDs(inventoryItems []dtos.InventoryStatusDTO) []uint {
	itemIDs := []uint{}

	for _, item := range inventoryItems {
		itemIDs = append(itemIDs, item.ItemID)
	}

	return itemIDs
}

func MapInventoryStatusList(items []dtos.InventoryStatusDTO, inventoryItems []dtos.InventoryStatusDtDTO) []dtos.InventoryStatusDtDTO {
	mappedInvDt := []dtos.InventoryStatusDtDTO{}

	// type InventoryStatusDtDTO struct {
	// 	ID uint `json:"id" db:"id"`
	// 	IngoingAt *string `json:"ingoing_at" db:"ingoing_at"`

	// 	IoType           *string  `json:"io_type" db:"io_type"`
	// 	ItemGroupName    *string  `json:"item_group_name" db:"item_group_name"`
	// 	ItemSubGroupName *string  `json:"item_sub_group_name" db:"item_sub_group_name"`
	// 	WarehouseName    *string  `json:"warehouse_name" db:"warehouse_name"`
	// 	IoTypeName       *string  `json:"io_type_name" db:"io_type_name"`
	// 	CustomerName     *string  `json:"customer_name" db:"customer_name"`
	// 	ItemCode         *string  `json:"item_code" db:"item_code"`
	// 	ItemName         *string  `json:"item_name" db:"item_name"`
	// 	UnitName         *string  `json:"unit_name" db:"unit_name"`
	// 	CurrencyName     *string  `json:"currency_name" db:"currency_name"`
	// 	PriceSell        *float64 `json:"price_sell" db:"price_sell"`
	// 	PriceBuy         *float64 `json:"price_buy" db:"price_buy"`
	// 	QtyIn            *float64 `json:"qty_in" db:"qty_in"`
	// 	QtyOut           *float64 `json:"qty_out" db:"qty_out"`
	// 	Balance          *float64 `json:"balance" db:"balance"`
	// }

	// get the inv dt list, add total each "items"

	for _, inv := range items {
		var balance float64 = 0
		var totalQtyIn float64 = 0
		var totalQtyOut float64 = 0
		for _, invDt := range inventoryItems {
			if inv.ItemID == invDt.ItemID {
				if invDt.IoType == "INVENTORY_IN" {
					invDt.QtyIn = invDt.Qty
					invDt.Price = invDt.PriceBuy
					balance += invDt.Qty
					totalQtyIn += invDt.Qty
				} else if invDt.IoType == "INVENTORY_OUT" {
					invDt.QtyOut = invDt.Qty
					invDt.Price = invDt.PriceSell
					balance -= invDt.Qty
					totalQtyOut += invDt.Qty
				}
				invDt.Balance = balance

				mappedInvDt = append(mappedInvDt, invDt)
			}
		}

		var totalDt dtos.InventoryStatusDtDTO
		totalDt.ID = 0
		totalDt.IsTotal = true
		totalDt.OutTotal = totalQtyOut
		totalDt.InTotal = totalQtyIn
		totalDt.BalanceTotal = totalQtyIn - totalQtyOut
		mappedInvDt = append(mappedInvDt, totalDt)
	}
	return mappedInvDt
}

func PublishSyncCreateStockClosings(ctx *fiber.Ctx, rabbitmq *config.RabbitMQ, data dtos.SyncStockInventoryRequest) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rabbitmq.Channel.PublishWithContext(ctx.Context(),
		"",                 // exchange
		"sync_stock_queue", // routing key (queue name)
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func GetDatesBetween(req dtos.SyncStockInventoryRequest, latestInvDate *string) ([]string, error) {
	const dateLayout = "2006-01-02"
	var dates []time.Time

	// Parse end_closing_at (required field)
	endDate, err := time.Parse(dateLayout, req.EndClosingAt)
	if err != nil {
		return nil, fmt.Errorf("invalid end_closing_at format: %v", err)
	}
	dates = append(dates, endDate)

	// Parse start_closing_at if exists
	if req.StartClosingAt != nil {
		startDate, err := time.Parse(dateLayout, *req.StartClosingAt)
		if err != nil {
			return nil, fmt.Errorf("invalid start_closing_at format: %v", err)
		}
		dates = append(dates, startDate)
	}

	// Parse latest inventory date if exists
	if latestInvDate != nil {
		latestDate, err := time.Parse(dateLayout, *latestInvDate)
		if err != nil {
			return nil, fmt.Errorf("invalid latest_inv_date format: %v", err)
		}
		dates = append(dates, latestDate)
	}

	// Find min and max dates
	if len(dates) == 0 {
		return []string{}, nil
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	minDate := dates[0]
	maxDate := dates[len(dates)-1]

	// Generate all dates between min and max (inclusive)
	var result []string
	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
		result = append(result, d.Format(dateLayout))
	}
	log.Println("dates: ", dates)
	log.Println("result: ", result)

	return result, nil
}

// func GetDatesBetween(req dtos.SyncStockInventoryRequest, latestInvDate *string) []string {
// 	// sort by start closing at to end closing at, get all dates between
// 	var dates []string
// 	endClosingAt, _ := time.Parse("2006-01-02", req.EndClosingAt)

// 	// If startClosingAt is nil, use the same date as endClosingAt
// 	if req.StartClosingAt == nil {
// 		dates = append(dates, endClosingAt.Format("2006-01-02"))
// 		return dates
// 	}

// 	// If startClosingAt has value, get dates between
// 	startClosingAt, _ := time.Parse("2006-01-02", *req.StartClosingAt)

// 	if startClosingAt.After(endClosingAt) {
// 		startClosingAt, endClosingAt = endClosingAt, startClosingAt
// 	}

// 	for d := startClosingAt; d.Before(endClosingAt) || d.Equal(endClosingAt); d = d.AddDate(0, 0, 1) {
// 		dates = append(dates, d.Format("2006-01-02"))
// 	}

// 	log.Println("dates: ", dates)
// 	return dates
// }

// func GetDatesBetween(req dtos.SyncStockInventoryRequest) []string {
// 	// sort by start closing at to end closing at, get all dates between
// 	var dates []string
// 	endClosingAt, _ := time.Parse("2006-01-02", req.EndClosingAt)

// 	// If startClosingAt is nil, use the same date as endClosingAt
// 	if req.StartClosingAt == nil {
// 		dates = append(dates, endClosingAt.Format("2006-01-02"))
// 		return dates
// 	}

// 	// If startClosingAt has value, get dates between
// 	startClosingAt, _ := time.Parse("2006-01-02", *req.StartClosingAt)
// 	for d := startClosingAt; d.Before(endClosingAt) || d.Equal(endClosingAt); d = d.AddDate(0, 0, 1) {
// 		dates = append(dates, d.Format("2006-01-02"))
// 	}

// 	log.Println("dates: ", dates)
// 	return dates
// }

// Get condition for quotation
func GetInventoryCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetInventoryCondition", opentracing.ChildOf(span.Context()))
	var err error

	joinCondition := ""
	customCondition := ""
	condition := ""
	var args []interface{}
	queryGlobal := ""
	i := 1

	filterDBColumnKey := []string{
		"iv.inventory_no", "iv.do_no", "iv.surat_jalan_no", "iv.invoice_no", "iv.remark", "iv.ship_dest",
		"pi.name",
		"so.po_buyer_no",
		"so.sales_order_no",
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
		condition += fmt.Sprintf(" AND iv.id IN (%s)", filters["ids"])
	}

	if filters["io_type"] != "" {
		condition += fmt.Sprintf(" AND ot.options_json->>'io_type' = '%s'", filters["io_type"])
	}

	filterKey := map[string]string{
		"status":          "iv.status",
		"customer_id":     "iv.customer_id",
		"io_type_id":      "iv.io_type_id",
		"currency_id":     "iv.currency_id",
		"vat_id":          "iv.vat_id",
		"payment_term_id": "iv.payment_term_id",
		"pph23_id":        "iv.pph23_id",
		"warehouse_id":    "iv.warehouse_id",
		"due_at":          "iv.due_at",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "iv.customer_id",
		"io_type_ids":      "iv.io_type_id",
		"currency_ids":     "iv.currency_id",
		"payment_term_ids": "iv.payment_term_id",
		"pph23_ids":        "iv.pph23_id",
		"warehouse_ids":    "iv.warehouse_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := SplitStringArrayOfInts(ids)
			if err != nil {
				LogErrors(childSpan, err)
				return nil, 0, "", "", "", "", err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs)) // Use pq.Array to pass the array to PostgreSQL
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": {"iv.vat_id", "sd.vat_id"},
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
		"product_ids": {"ivd.ref_product_id", "ivd.item_id"},
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

	filterDBColumnLikeKey := map[string]string{
		"do_no":          "iv.do_no",
		"invoice_no":     "iv.invoice_no",
		"surat_jalan_no": "iv.surat_jalan_no",
		"inventory_no":   "iv.inventory_no",
		"po_buyer_no":    "so.po_buyer_no",
		"sales_order_no": "so.sales_order_no",
		"ship_dest":      "iv.ship_dest",
		"remark":         "iv.remark",
		"item_name":      "pi.name",
	}

	for key, value := range filters {
		if value != "" {
			for keyLike, column := range filterDBColumnLikeKey {
				if filters[key] != "" && key == keyLike {
					condition += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
					args = append(args, "%"+value+"%")
					i++
				}
			}
		}
	}

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"ingoing_at":  "iv.ingoing_at",
			"invoice_at":  "iv.invoice_at",
			"do_at":       "iv.do_at",
			"shipping_at": "so.shipping_at",
			"expired_at":  "ivd.expired_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	return args, i, condition, queryGlobal, joinCondition, customCondition, err
}

func GetInventoryDetailsIDs(quotations []dtos.InventoryDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToInventorys(quoDts []dtos.InventoryInvDtListDTO, quotations []dtos.InventoryDetailDTO) []dtos.InventoryDetailDTO {
	// quotations := []dtos.InventoryAttachmentsDTO{}
	for i, quotation := range quotations {
		newSoDts := make([]dtos.InventoryInvDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.InventoryID == quotation.ID {
				newSoDts = append(newSoDts, quoDt)
			}
		}
		quotation.InvDts = newSoDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetInventoryDetails(salesOrders []dtos.InventoryDetailDTO, soDts []dtos.InventoryInvDtListDTO) []dtos.InventoryDetailDTO {
	// salesOrders := []dtos.InventoryAttachmentsDTO{}
	for i, salesOrder := range salesOrders {
		newSoDts := make([]dtos.InventoryInvDtListDTO, 0)
		for _, soDt := range soDts {
			if *soDt.InventoryID == salesOrder.ID {
				newSoDts = append(newSoDts, soDt)
			}
		}
		salesOrder.InvDts = newSoDts
		salesOrders[i] = salesOrder
	}

	return salesOrders
}

// Build CSV rows, dtos.InventoryListDTO, csv pointer
func BuildInventoryAllCSVRows(salesOrders []dtos.InventoryListDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Inventory No", "DO No", "Invoice No", "IO Type", "Customer", "I/O Date", "DO Date", "Invoice Date",
		"Currency", "Status", "Created By", "Updated By",
	}
	rows = append(rows, header)

	for idx, salesOrder := range salesOrders {
		ID := fmt.Sprintf("%d", idx+1)
		PoNo := GetPtrVal(&salesOrder.InventoryNo)
		DoNo := GetPtrVal(salesOrder.DoNo)
		InvoiceNo := GetPtrVal(salesOrder.InvoiceNo)
		IoType := GetPtrVal(salesOrder.IoTypeName)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		PoDate := GetPtrVal(salesOrder.IngoingAt)
		DeliveryDate := GetPtrVal(salesOrder.DoAt)
		InvoiceDate := GetPtrVal(salesOrder.InvoiceAt)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(&salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		// EscapeCsvField
		PoNo = EscapeCsvField(PoNo)
		DoNo = EscapeCsvField(DoNo)
		InvoiceNo = EscapeCsvField(InvoiceNo)
		IoType = EscapeCsvField(IoType)
		CustomerName = EscapeCsvField(CustomerName)
		PoDate = EscapeCsvField(PoDate)
		DeliveryDate = EscapeCsvField(DeliveryDate)
		InvoiceDate = GetPtrVal(salesOrder.InvoiceAt)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		row := []string{
			ID, PoNo, DoNo, InvoiceNo, IoType, CustomerName, PoDate, DeliveryDate, InvoiceDate,
			CurrencyName, Status, CreatedByName, UpdatedByName,
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

// Build CSV rows, dtos.InventoryListDTO, csv pointer
func BuildInventoryDetailCSVRows(salesOrders []dtos.InventoryDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Inventory No", "DO No", "Invoice No", "IO Type", "Customer", "I/O Date", "DO Date", "Invoice Date",
		"Currency", "Status", "Created By", "Updated By",
		"Product/Item Name", "Qty",
	}
	rows = append(rows, header)

	for idx, salesOrder := range salesOrders {
		ID := fmt.Sprintf("%d", idx+1)
		PoNo := GetPtrVal(&salesOrder.InventoryNo)
		DoNo := GetPtrVal(salesOrder.DoNo)
		InvoiceNo := GetPtrVal(salesOrder.InvoiceNo)
		IoType := GetPtrVal(salesOrder.IoTypeName)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		PoDate := GetPtrVal(salesOrder.IngoingAt)
		DeliveryDate := GetPtrVal(salesOrder.DoAt)
		InvoiceDate := GetPtrVal(salesOrder.InvoiceAt)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(&salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		// EscapeCsvField
		PoNo = EscapeCsvField(PoNo)
		DoNo = EscapeCsvField(DoNo)
		InvoiceNo = EscapeCsvField(InvoiceNo)
		IoType = EscapeCsvField(IoType)
		CustomerName = EscapeCsvField(CustomerName)
		PoDate = EscapeCsvField(PoDate)
		DeliveryDate = EscapeCsvField(DeliveryDate)
		InvoiceDate = GetPtrVal(salesOrder.InvoiceAt)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		if len(salesOrder.InvDts) == 0 {
			row := []string{
				ID, PoNo, DoNo, InvoiceNo, IoType, CustomerName, PoDate, DeliveryDate, InvoiceDate,
				CurrencyName, Status, CreatedByName, UpdatedByName,
				"", "", "", "",
			}
			rows = append(rows, row)
		} else {
			for iSoDt, quoDt := range salesOrder.InvDts {
				ProductItemName := GetPtrVal(quoDt.ItemName)
				ProductItemName = EscapeCsvField(ProductItemName)
				Qty := fmt.Sprintf("%f", *quoDt.Qty)

				if iSoDt == 0 {
					row := []string{
						ID, PoNo, DoNo, InvoiceNo, IoType, CustomerName, PoDate, DeliveryDate, InvoiceDate,
						CurrencyName, Status, CreatedByName, UpdatedByName,
						ProductItemName, Qty,
					}
					rows = append(rows, row)
				} else {
					row := []string{
						"", "", "", "", "", "", "", "", "",
						"", "", "", "",
						ProductItemName, Qty,
					}
					rows = append(rows, row)
				}
			}
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
