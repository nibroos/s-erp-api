package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
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

// Get condition for quotation
func GetQuotationCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetQuotationCondition", opentracing.ChildOf(span.Context()))
	var err error

	condition := ""
	var args []interface{}
	queryGlobal := ""
	i := 1

	filterDBColumnKey := []string{
		"q.quo_no", "q.title", "q.remark",
		"pi.name",
		"it.name",
		"qd.remark",
		"qd.gen_code",
		"qdb.remark",
		"qdb.gen_code",
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
		condition += fmt.Sprintf(" AND q.id IN (%s)", filters["ids"])
	}

	filterKey := map[string]string{
		"status":        "q.status",
		"customer_id":   "q.customer_id",
		"order_type_id": "q.order_type_id",
		"currency_id":   "q.currency_id",
		"vat_id":        "q.vat_id",
		"payment_id":    "q.payment_id",
		"pph23_id":      "q.pph23_id",
		"expired_at":    "q.expired_at",
		"due_at":        "q.due_at",
	}

	for key, valueID := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", valueID, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"customer_ids":   "q.customer_id",
		"order_type_ids": "q.order_type_id",
		"currency_ids":   "q.currency_id",
		"payment_ids":    "q.payment_id",
		"pph23_ids":      "q.pph23_id",
	}

	for key, valueID := range filterIDsKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := SplitStringArrayOfInts(ids)
			if err != nil {
				LogErrors(childSpan, err)
				return nil, i, "", "", err
			}

			condition += fmt.Sprintf(" AND %s = ANY($%d)", valueID, i)
			args = append(args, pq.Array(intIDs)) // Use pq.Array to pass the array to PostgreSQL
			i++
		}
	}

	filterIDsOrKey := map[string][]string{
		"vat_ids": {"q.vat_id", "qd.vat_id"},
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
		"product_ids": {"pi.id", "it.id"},
	}

	// Handle array OR conditions
	for key, valueIDs := range filterIDsOrArrayKey {
		if value, ok := filters[key]; ok && value != "" {
			// Split the string into an array of integers
			ids := strings.Split(value, ",")
			intIDs, err := SplitStringArrayOfInts(ids)
			if err != nil {
				LogErrors(childSpan, err)
				return nil, i, "", "", err
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
		filterDateTypeKey := map[string]string{
			"due_at":     "q.due_at",
			"expired_at": "q.expired_at",
		}

		dateTypeColumn := filterDateTypeKey[filters["date_type"]]
		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	return args, i, condition, queryGlobal, err
}

func GetQuotationDetailsIDs(quotations []dtos.QuotationDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToQuotations(quoDts []dtos.QuotationQuoDtListDTO, quotations []dtos.QuotationDetailDTO) []dtos.QuotationDetailDTO {
	// quotations := []dtos.SalesOrderAttachmentsDTO{}
	for i, quotation := range quotations {
		newQuoDts := make([]dtos.QuotationQuoDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.QuotationID == quotation.ID {
				newQuoDts = append(newQuoDts, quoDt)
			}
		}
		quotation.QuoDts = newQuoDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetQuotationDetails(quotations []dtos.QuotationDetailDTO, quoDts []dtos.QuotationQuoDtListDTO, quoDtBoms []dtos.QuotationQuoDtBomListDTO) []dtos.QuotationDetailDTO {
	// quotations := []dtos.SalesOrderAttachmentsDTO{}
	for i, quotation := range quotations {
		newQuoDts := make([]dtos.QuotationQuoDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.QuotationID == quotation.ID {
				newQuoDtBoms := make([]dtos.QuotationQuoDtBomListDTO, 0)
				for _, quoDtBom := range quoDtBoms {
					if *quoDtBom.QuoDtID == *quoDt.ID {
						newQuoDtBoms = append(newQuoDtBoms, quoDtBom)
					}
				}
				quoDt.QuoDtsBoms = newQuoDtBoms
				newQuoDts = append(newQuoDts, quoDt)
			}
		}
		quotation.QuoDts = newQuoDts
		quotations[i] = quotation
	}

	return quotations
}

// Build CSV rows, dtos.QuotationListDTO, csv pointer
func BuildQuotationAllCSVRows(quotations []dtos.QuotationListDTO, csv *string) error {
	rows := [][]string{}
	// ID,Quotation No,Title,Order Type,Customer,Expired Date,Quot Date,Currency,Total,Status,Created By,Updated By\n
	header := []string{
		"ID", "Quotation No", "Title", "Order Type", "Customer", "Expired Date", "Quot Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
	}
	rows = append(rows, header)

	for _, quotation := range quotations {
		ID := fmt.Sprintf("%d", quotation.ID)
		QuoNo := GetPtrVal(quotation.QuoNo)
		Title := quotation.Title
		OrderTypeName := GetPtrVal(quotation.OrderTypeName)
		CustomerName := GetPtrVal(quotation.CustomerName)
		ExpiredAt := GetPtrVal(quotation.ExpiredAt)
		DueAt := GetPtrVal(quotation.DueAt)
		CurrencyName := GetPtrVal(quotation.CurrencyName)
		Status := GetPtrVal(&quotation.Status)
		CreatedByName := GetPtrVal(quotation.CreatedByName)
		UpdatedByName := GetPtrVal(quotation.UpdatedByName)

		// EscapeCsvField
		ID = EscapeCsvField(ID)
		QuoNo = EscapeCsvField(QuoNo)
		Title = EscapeCsvField(Title)
		OrderTypeName = EscapeCsvField(OrderTypeName)
		CustomerName = EscapeCsvField(CustomerName)
		ExpiredAt = EscapeCsvField(ExpiredAt)
		DueAt = EscapeCsvField(DueAt)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		row := []string{
			ID, QuoNo, Title, OrderTypeName, CustomerName, ExpiredAt, DueAt,
			CurrencyName, fmt.Sprintf("%f", *quotation.GrandTotal), Status, CreatedByName, UpdatedByName,
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

// Build CSV rows, dtos.QuotationListDTO, csv pointer
func BuildQuotationDetailCSVRows(quotations []dtos.QuotationDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Quotation No", "Title", "Order Type", "Customer", "Expired Date", "Quot Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
		"Product/Item Name", "Qty", "Price", "Subtotal",
		"BOM Item Name", "BOM Qty",
	}
	rows = append(rows, header)

	for iQuotation, quotation := range quotations {

		No := fmt.Sprintf("%d", iQuotation+1)
		QuoNo := GetPtrVal(quotation.QuoNo)
		Title := quotation.Title
		OrderTypeName := GetPtrVal(quotation.OrderTypeName)
		CustomerName := GetPtrVal(quotation.CustomerName)
		ExpiredAt := GetPtrVal(quotation.ExpiredAt)
		DueAt := GetPtrVal(quotation.DueAt)
		CurrencyName := GetPtrVal(quotation.CurrencyName)
		Status := GetPtrVal(&quotation.Status)
		CreatedByName := GetPtrVal(quotation.CreatedByName)
		UpdatedByName := GetPtrVal(quotation.UpdatedByName)

		for iQuoDt, quoDt := range quotation.QuoDts {
			ProductItemName := GetPtrVal(quoDt.ItemName)
			Qty := fmt.Sprintf("%f", *quoDt.Qty)
			PriceSell := fmt.Sprintf("%f", *quoDt.PriceSell)
			TotalAm := fmt.Sprintf("%f", *quoDt.SubtotalSell)

			if len(quoDt.QuoDtsBoms) == 0 {

				if iQuoDt == 0 {
					row := []string{
						No, QuoNo, Title, OrderTypeName, CustomerName, ExpiredAt, DueAt,
						CurrencyName, fmt.Sprintf("%f", quotation.GrandTotal), Status, CreatedByName, UpdatedByName,
						ProductItemName, Qty, PriceSell, TotalAm,
						"", "", "", "",
					}
					rows = append(rows, row)
				} else {
					row := []string{
						"", "", "", "", "", "", "",
						"", "", "", "", "",
						ProductItemName, Qty, PriceSell, TotalAm,
						"", "", "", "",
					}
					rows = append(rows, row)
				}
			} else {
				for iBom, bom := range quoDt.QuoDtsBoms {
					BomItemName := GetPtrVal(bom.ItemName)
					BomQty := fmt.Sprintf("%f", bom.Qty)

					if iQuoDt == 0 && iBom == 0 {
						row := []string{
							No, QuoNo, Title, OrderTypeName, CustomerName, ExpiredAt, DueAt,
							CurrencyName, fmt.Sprintf("%f", quotation.GrandTotal), Status, CreatedByName, UpdatedByName,
							ProductItemName, Qty, PriceSell, TotalAm,
							BomItemName, BomQty,
						}
						rows = append(rows, row)
					} else if iBom == 0 {
						row := []string{
							"", "", "", "", "", "", "",
							"", "", "", "", "",
							ProductItemName, Qty, PriceSell, TotalAm,
							BomItemName, BomQty,
						}
						rows = append(rows, row)
					} else {
						row := []string{
							"", "", "", "", "", "", "",
							"", "", "", "", "",
							"", "", "", "",
							BomItemName, BomQty,
						}
						rows = append(rows, row)
					}
				}
			}
		}
		if len(quotation.QuoDts) == 0 {
			row := []string{
				No, QuoNo, Title, OrderTypeName, CustomerName, ExpiredAt, DueAt,
				CurrencyName, fmt.Sprintf("%f", quotation.GrandTotal), Status, CreatedByName, UpdatedByName,
				"", "", "", "",
				"", "",
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
