package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/opentracing/opentracing-go"
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

		invDtModel := models.InvDt{
			ProductUuid:  invDt.ProductUuid,
			InventoryID:  createdInventory.ID,
			ItemUnitID:   invDt.ItemUnitID,
			VatID:        invDt.VatID,
			Pph23ID:      invDt.Pph23ID,
			RefSoDtID:    invDt.RefSoDtID,
			RefSoDtBomID: invDt.RefSoDtBomID,
			RefPoDtID:    invDt.RefPoDtID,
			RefPoDtBomID: invDt.RefPoDtBomID,
			RefInvDtID:   invDt.RefInvDtID,
			RefProductID: invDt.RefProductID,
			ItemID:       invDt.ItemID,
			RefType:      invDt.RefType,
			ItemType:     invDt.ItemType,
			ItemJSON:     string(itemJson),
			GenCode:      &genCode,
			Remark:       invDt.Remark,
			VatPerc:      invDt.VatPerc,
			VatPercAm:    invDt.VatPercAm,
			Pph23Perc:    invDt.Pph23Perc,
			Pph23PercAm:  invDt.Pph23PercAm,
			IsVat:        invDt.IsVat,
			IsPph23:      invDt.IsPph23,
			Qty:          invDt.Qty,
			PriceSell:    invDt.PriceSell,
			PriceBuy:     invDt.PriceBuy,
			SubtotalSell: invDt.SubtotalSell,
			SubtotalBuy:  invDt.SubtotalBuy,
			ExpiredAt:    invDt.ExpiredAt,
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
		if reqInvDt.InvDtID != nil {
			invDtID = *reqInvDt.InvDtID
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
			ID:           invDtID,
			ProductUuid:  reqInvDt.ProductUuid,
			InventoryID:  updatedInventory.ID,
			ItemUnitID:   reqInvDt.ItemUnitID,
			VatID:        reqInvDt.VatID,
			Pph23ID:      reqInvDt.Pph23ID,
			RefSoDtID:    reqInvDt.RefSoDtID,
			RefSoDtBomID: reqInvDt.RefSoDtBomID,
			RefPoDtID:    reqInvDt.RefPoDtID,
			RefPoDtBomID: reqInvDt.RefPoDtBomID,
			RefInvDtID:   reqInvDt.RefInvDtID,
			RefProductID: reqInvDt.RefProductID,
			ItemID:       reqInvDt.ItemID,
			RefType:      reqInvDt.RefType,
			ItemType:     reqInvDt.ItemType,
			ItemJSON:     string(itemJson),
			GenCode:      reqInvDt.GenCode,
			Remark:       reqInvDt.Remark,
			VatPerc:      reqInvDt.VatPerc,
			VatPercAm:    reqInvDt.VatPercAm,
			Pph23Perc:    reqInvDt.Pph23Perc,
			Pph23PercAm:  reqInvDt.Pph23PercAm,
			IsVat:        reqInvDt.IsVat,
			IsPph23:      reqInvDt.IsPph23,
			Qty:          reqInvDt.Qty,
			PriceSell:    reqInvDt.PriceSell,
			PriceBuy:     reqInvDt.PriceBuy,
			SubtotalSell: reqInvDt.SubtotalSell,
			SubtotalBuy:  reqInvDt.SubtotalBuy,
			ExpiredAt:    reqInvDt.ExpiredAt,
			QtyOut:       reqInvDt.QtyOut,
			QtyInvoice:   reqInvDt.QtyInvoice,
			TotalAm:      reqInvDt.TotalAm,
			CreatedByID:  &userID,
		}
		invDtsModel = append(invDtsModel, invDtModel)

	}

	return invDtsModel, nil
}

func MapOldUpdateInvDts(ctx *fiber.Ctx, req dtos.FormInventoryRequest, userID uint, span opentracing.Span) ([]uint, []uint, []uint, []uint, []uint, error) {
	refSoDtID := []uint{}
	refSoDtBomDtID := []uint{}
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

	return refSoDtID, refSoDtBomDtID, refPoDtID, refPoDtBomID, refInvDtID, nil
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
