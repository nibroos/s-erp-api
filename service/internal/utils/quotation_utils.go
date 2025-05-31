package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
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

func MapCreateQuoDts(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuotation *models.Quotation, userID uint, span opentracing.Span) ([]models.QuoDt, error) {
	quoDtsModel := []models.QuoDt{}

	for _, quoDt := range req.QuoDts {
		quoDtModel := models.QuoDt{
			ProductUuid: quoDt.ProductUuid,
			QuotationID: &createdQuotation.ID,
			ItemUnitID:  quoDt.ItemUnitID,
			VatID:       quoDt.VatID,
			Pph23ID:     quoDt.Pph23ID,
			RefID:       quoDt.RefID,
			ItemID:      quoDt.ItemID,
			RefType:     quoDt.RefType,
			ItemType:    quoDt.ItemType,
			// RefJSON: 	quoDt.RefJSON,
			Remark:          quoDt.Remark,
			VatPerc:         quoDt.VatPerc,
			VatPercAm:       quoDt.VatPercAm,
			Pph23Perc:       quoDt.Pph23Perc,
			Pph23PercAm:     quoDt.Pph23PercAm,
			MarkupPerc:      quoDt.MarkupPerc,
			MarkupPercAm:    quoDt.MarkupPercAm,
			IsVat:           quoDt.IsVat,
			IsPph23:         quoDt.IsPph23,
			IsLockMarkup:    quoDt.IsLockMarkup,
			IsLockPriceSell: quoDt.IsLockPriceSell,
			IsLockPriceBuy:  quoDt.IsLockPriceBuy,
			Qty:             quoDt.Qty,
			PriceSell:       quoDt.PriceSell,
			PriceBuy:        quoDt.PriceBuy,
			SubtotalSell:    quoDt.SubtotalSell,
			SubtotalBuy:     quoDt.SubtotalBuy,
			DiscAm:          quoDt.DiscAm,
			DiscPerc:        quoDt.DiscPerc,
			DiscPercNum:     quoDt.DiscPercNum,
			DiscPercAm:      quoDt.DiscPercAm,
			DiscType:        quoDt.DiscType,
			DiscFinal:       quoDt.DiscFinal,
			TotalAm:         quoDt.TotalAm,
			CreatedByID:     &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}

func MapCreateQuoDtBoms(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, createdQuoDts []models.QuoDt, userID uint, span opentracing.Span) []map[string]interface{} {
	quoDtBomsModel := make([]map[string]interface{}, 0)

	for _, reqQuoDt := range req.QuoDts {
		for _, reqQuoDtBom := range reqQuoDt.QuoDtsBoms {
			for _, createdQuoDt := range createdQuoDts {
				if reqQuoDtBom.ProductUuid == createdQuoDt.ProductUuid {
					genCode := "-"
					itemJson := "{}"
					quoDtBomsModel = append(quoDtBomsModel, map[string]interface{}{
						"id":            0,
						"quotation_id":  createdQuoDt.QuotationID,
						"product_uuid":  reqQuoDtBom.ProductUuid,
						"quo_dt_id":     createdQuoDt.ID,
						"product_id":    reqQuoDtBom.ProductID,
						"item_id":       reqQuoDtBom.ItemID,
						"item_unit_id":  reqQuoDtBom.ItemUnitID,
						"remark":        reqQuoDtBom.Remark,
						"qty":           reqQuoDtBom.Qty,
						"price_sell":    reqQuoDtBom.PriceSell,
						"price_buy":     reqQuoDtBom.PriceBuy,
						"subtotal_sell": reqQuoDtBom.SubtotalSell,
						"subtotal_buy":  reqQuoDtBom.SubtotalBuy,
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

func MapCreateUpdateQuoDts(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, updatedQuotation *models.Quotation, userID uint, span opentracing.Span) ([]models.QuoDt, error) {

	quoDtsModel := []models.QuoDt{}

	// refJson := "{}"
	// itemJson := "{}"

	for _, reqQuoDt := range req.QuoDts {
		quoDtID := uint(0)
		if reqQuoDt.QuoDtID != nil {
			quoDtID = *reqQuoDt.QuoDtID
		}

		quoDtModel := models.QuoDt{
			ID:          quoDtID,
			ProductUuid: reqQuoDt.ProductUuid,
			QuotationID: &updatedQuotation.ID,
			ItemUnitID:  reqQuoDt.ItemUnitID,
			Pph23ID:     reqQuoDt.Pph23ID,
			VatID:       reqQuoDt.VatID,
			RefID:       reqQuoDt.RefID,
			ItemID:      reqQuoDt.ItemID,
			RefType:     reqQuoDt.RefType,
			ItemType:    reqQuoDt.ItemType,
			// RefJSON: 	refJson,
			// ItemJSON: 	 itemJson,
			GenCode:         reqQuoDt.GenCode,
			Remark:          reqQuoDt.Remark,
			VatPerc:         reqQuoDt.VatPerc,
			VatPercAm:       reqQuoDt.VatPercAm,
			Pph23Perc:       reqQuoDt.Pph23Perc,
			Pph23PercAm:     reqQuoDt.Pph23PercAm,
			MarkupPerc:      reqQuoDt.MarkupPerc,
			MarkupPercAm:    reqQuoDt.MarkupPercAm,
			IsVat:           reqQuoDt.IsVat,
			IsPph23:         reqQuoDt.IsPph23,
			IsLockMarkup:    reqQuoDt.IsLockMarkup,
			IsLockPriceSell: reqQuoDt.IsLockPriceSell,
			IsLockPriceBuy:  reqQuoDt.IsLockPriceBuy,
			Qty:             reqQuoDt.Qty,
			PriceSell:       reqQuoDt.PriceSell,
			PriceBuy:        reqQuoDt.PriceBuy,
			SubtotalSell:    reqQuoDt.SubtotalSell,
			SubtotalBuy:     reqQuoDt.SubtotalBuy,
			DiscAm:          reqQuoDt.DiscAm,
			DiscPerc:        reqQuoDt.DiscPerc,
			DiscPercNum:     reqQuoDt.DiscPercNum,
			DiscPercAm:      reqQuoDt.DiscPercAm,
			DiscType:        reqQuoDt.DiscType,
			DiscFinal:       reqQuoDt.DiscFinal,
			TotalAm:         reqQuoDt.TotalAm,
			CreatedByID:     &userID,
		}
		quoDtsModel = append(quoDtsModel, quoDtModel)

	}

	return quoDtsModel, nil
}

func GenQuoNo() string {
	return "QUO-" + time.Now().Format("20060102-150405")
}

func MapCreateQuotation(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, userID uint, branchID uint, orderedNumber int, orderedNumberGlobal int, span opentracing.Span) (models.Quotation, error) {
	revNo := 0

	orderNumber := orderedNumber
	orderNumber++

	orderNumberGlobal := orderedNumberGlobal
	orderNumberGlobal++

	quoNo := GenerateQuoNoOnCreateQuotation(ctx, req, orderNumber, orderNumberGlobal, span)

	quotation := models.Quotation{
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		RevNo:         &revNo,
		QuoNo:         &quoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		TermDesc:      req.TermDesc,
		LicenseDesc:   req.LicenseDesc,
		Status:        req.Status,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		MarkupPerc:    req.MarkupPerc,
		IsVat:         req.IsVat,
		IsPph23:       req.IsPph23,
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
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		BranchID:      &branchID,
		CreatedByID:   &userID,
	}

	return quotation, nil
}

func MapUpdateQuotation(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, userID uint, branchID uint, span opentracing.Span) (models.Quotation, error) {
	// quo_no track last number on string, REV-1, REV-2, REV-3
	revNo := *req.RevNo
	revNo++

	quoNo := GenerateQuoNoOnUpdateQuotation(ctx, req, &revNo, span)

	quotation := models.Quotation{
		ID:            req.ID,
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		RevNo:         &revNo,
		QuoNo:         &quoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		TermDesc:      req.TermDesc,
		LicenseDesc:   req.LicenseDesc,
		Status:        req.Status,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		MarkupPerc:    req.MarkupPerc,
		IsVat:         req.IsVat,
		IsPph23:       req.IsPph23,
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
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		BranchID:      &branchID,
		UpdatedByID:   &userID,
	}

	return quotation, nil
}

func GenerateQuoNoOnUpdateQuotation(ctx *fiber.Ctx, req dtos.UpdateQuotationRequest, revNo *int, span opentracing.Span) string {
	quoNo := req.QuoNo

	// get before REV-number, full string is SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1
	// check if "REV" string exist (random), if not add "REV-1" else change the number to incremented number
	if !strings.Contains(*quoNo, "REV") {
		*quoNo = fmt.Sprintf("%s/REV-%d", *quoNo, *revNo)
	} else {
		// remove after /REV
		// use split to get the first part of string
		*quoNo = strings.Split(*quoNo, "/REV")[0]
		*quoNo = fmt.Sprintf("%s/REV-%d", *quoNo, *revNo)
	}

	return *quoNo
}

func GenerateQuoNoOnCreateQuotation(ctx *fiber.Ctx, req dtos.CreateQuotationRequest, orderedNumber int, orderedNumberGlobal int, span opentracing.Span) string {
	if req.QuoNo != nil {
		return *req.QuoNo
	}

	// SURNAME-YEAR-MONTH-ORDER-REV-(NUM) -> SURNAME-2001-12-20-REV-1

	if req.CustomerCode == nil {
		return ""
	}

	// surname := *req.CustomerCode
	surname := "Yubi"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	orderGlobal := fmt.Sprintf("%d", orderedNumberGlobal)
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("QUO/%s/%s-%s-%s-%s", orderGlobal, surname, year, month, order)

	return str
}

// add url before path file by env
func MapQuotationsToURL(quotations []dtos.QuotationListDTO) []dtos.QuotationListDTO {
	// quotations := []dtos.SalesOrderAttachmentsDTO{}
	for i, quotation := range quotations {
		if quotation.PdfPath != nil {
			// remove first letter from file url
			newPdfPath := fmt.Sprintf("%s%s", os.Getenv("APP_HOST"), (*quotation.PdfPath)[1:])
			// newPdfPath := fmt.Sprintf("%s%s", os.Getenv("APP_HOST"), *quotation.PdfPath)
			quotations[i].PdfPathUrl = &newPdfPath
		}
	}

	return quotations
}

// add url before path file by env
func MapStringToURL(path *string) *string {
	if path != nil {
		// remove first letter from file url
		newPath := fmt.Sprintf("%s/%s", os.Getenv("APP_HOST"), *path)
		path = &newPath
	}
	return path
}

// add url before path file by env
func MapDotStringToURL(path *string) *string {
	if path != nil {
		// remove first letter from file url
		newPath := fmt.Sprintf("%s/%s", os.Getenv("APP_HOST"), (*path)[1:])
		path = &newPath
	}
	return path
}
