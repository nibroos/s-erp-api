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

func GetInvIDs(req dtos.UpdateInventoryRequest) ([]*uint, []*uint, []*uint) {
	soDtIDs := []*uint{}
	soDtBomIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.InvDtID != nil && *reqInvDt.InvDtID > 0 {
			soDtIDs = append(soDtIDs, reqInvDt.InvDtID)
		}

		for _, reqInvDtBom := range reqInvDt.InvDtsBoms {
			if reqInvDtBom.InvDtBomID != nil && *reqInvDtBom.InvDtBomID > 0 {
				soDtBomIDs = append(soDtBomIDs, reqInvDtBom.InvDtBomID)
				productIDs = append(productIDs, &reqInvDtBom.ProductID)
				productIDs = append(productIDs, reqInvDtBom.ItemID)
				itemUnitIDs = append(itemUnitIDs, reqInvDtBom.ItemUnitID)
			}
		}
	}

	return soDtIDs, productIDs, itemUnitIDs
}

func GetLockInventoryQuoIDs(req dtos.CreateInventoryRequest) []*uint {
	soDtIDs := []*uint{}

	for _, reqInvDt := range req.InvDts {
		if reqInvDt.RefType == "so" && reqInvDt.RefID > 0 {
			soDtIDs = append(soDtIDs, &reqInvDt.RefID)
		}
	}

	return soDtIDs
}

func MapCreateInvDts(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, createdInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {
	soDtsModel := []models.InvDt{}

	for _, soDt := range req.InvDts {
		genCode := "-"
		itemJson := "{}"

		soDtModel := models.InvDt{
			ProductUuid:  soDt.ProductUuid,
			InventoryID:  createdInventory.ID,
			ItemUnitID:   soDt.ItemUnitID,
			VatID:        soDt.VatID,
			Pph23ID:      soDt.Pph23ID,
			RefID:        soDt.RefID,
			ItemID:       soDt.ItemID,
			RefType:      soDt.RefType,
			ItemType:     soDt.ItemType,
			ItemJSON:     &itemJson,
			GenCode:      &genCode,
			Remark:       soDt.Remark,
			VatPerc:      soDt.VatPerc,
			VatPercAm:    soDt.VatPercAm,
			Pph23Perc:    soDt.Pph23Perc,
			Pph23PercAm:  soDt.Pph23PercAm,
			IsVat:        soDt.IsVat,
			IsPph23:      soDt.IsPph23,
			Qty:          soDt.Qty,
			PriceSell:    soDt.PriceSell,
			PriceBuy:     soDt.PriceBuy,
			SubtotalSell: soDt.SubtotalSell,
			SubtotalBuy:  soDt.SubtotalBuy,
			TotalAm:      soDt.TotalAm,
			CreatedByID:  &userID,
		}
		soDtsModel = append(soDtsModel, soDtModel)

	}

	return soDtsModel, nil
}

func MapUpdateInvDts(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, updatedInventory *models.Inventory, userID uint, span opentracing.Span) ([]models.InvDt, error) {

	soDtsModel := []models.InvDt{}

	itemJson := "{}"

	for _, reqInvDt := range req.InvDts {
		soDtID := uint(0)
		if reqInvDt.InvDtID != nil {
			soDtID = *reqInvDt.InvDtID
		}

		soDtModel := models.InvDt{
			ID:           soDtID,
			ProductUuid:  reqInvDt.ProductUuid,
			InventoryID:  updatedInventory.ID,
			ItemUnitID:   reqInvDt.ItemUnitID,
			VatID:        reqInvDt.VatID,
			Pph23ID:      reqInvDt.Pph23ID,
			RefID:        reqInvDt.RefID,
			ItemID:       reqInvDt.ItemID,
			RefType:      reqInvDt.RefType,
			ItemType:     reqInvDt.ItemType,
			ItemJSON:     &itemJson,
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
			TotalAm:      reqInvDt.TotalAm,
			CreatedByID:  &userID,
		}
		soDtsModel = append(soDtsModel, soDtModel)

	}

	return soDtsModel, nil
}

func GenInventoryNo(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, orderedNumber int, span opentracing.Span) string {
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

func MapCreateInventory(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, userID uint, branchID uint, customerInvCreatedThisMonthNumber int, span opentracing.Span) (models.Inventory, error) {
	// order := 1
	customerInvCreatedThisMonthNumber++

	// orderNo := GenInventoryNo(ctx, req, customerInvCreatedThisMonthNumber, span)
	inventoryNo := GeneratePoBuyerNoNoOnCreateInventory(ctx, req, customerInvCreatedThisMonthNumber, span)

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

func MapUpdateInventory(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, userID uint, branchID uint, span opentracing.Span) (models.Inventory, error) {
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
	}

	return salesOrder, nil
}

func GenerateInvNoOnUpdateSo(ctx *fiber.Ctx, req dtos.UpdateInventoryRequest, inventoryNo *string, revNo int, span opentracing.Span) string {

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

func GeneratePoBuyerNoNoOnCreateInventory(ctx *fiber.Ctx, req dtos.CreateInventoryRequest, orderedNumber int, span opentracing.Span) string {
	if req.InventoryNo != nil {
		return *req.InventoryNo
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

func MapUpdateInvSoDtsQty(soDtsQtyUpdate []dtos.GetQuoDtQtyUpdateDTO, req dtos.CreateInventoryRequest) []map[string]interface{} {
	// filter with ID to bulk update
	bulkUpdateQuoDts := []map[string]interface{}{}

	for _, reqInvDt := range req.InvDts {
		for _, soDts := range soDtsQtyUpdate {

			if soDts.QtySO == nil {
				soDts.QtySO = new(float64)
			}

			newQuoDt := map[string]interface{}{
				"id":      soDts.QuoDtID,
				"qty_out": (*reqInvDt.Qty + *soDts.QtySO),
			}
			bulkUpdateQuoDts = append(bulkUpdateQuoDts, newQuoDt)
		}
	}

	return bulkUpdateQuoDts
}

func GetInvSoDtIDs(soDts []dtos.RefInvIndexSoDtListDTO) []uint {
	quotationIDs := []uint{}

	for _, soDt := range soDts {
		quotationIDs = append(quotationIDs, *soDt.SalesOrderID)
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
