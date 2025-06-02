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

func GetSalesInvoiceIDs(req dtos.UpdateSalesInvoiceRequest) ([]*uint, []*uint, []*uint) {
	salesInvoiceDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		if reqSalesInvoiceDt.SalesInvoiceDtID != nil && *reqSalesInvoiceDt.SalesInvoiceDtID > 0 {
			salesInvoiceDtIDs = append(salesInvoiceDtIDs, reqSalesInvoiceDt.SalesInvoiceDtID)
		}
	}

	return salesInvoiceDtIDs, productIDs, itemUnitIDs
}

func GetLockSalesInvoiceSalesOrderIDs(req dtos.CreateSalesInvoiceRequest) ([]*uint, []*uint) {
	soIDs := []*uint{}
	soDtIDs := []*uint{}

	soIDsMap := make(map[uint]bool)

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		if reqSalesInvoiceDt.RefType != nil && *reqSalesInvoiceDt.RefType == "so" {
			if reqSalesInvoiceDt.RefID != nil && *reqSalesInvoiceDt.RefID > 0 {
				if !soIDsMap[*reqSalesInvoiceDt.RefID] {
					soIDsMap[*reqSalesInvoiceDt.RefID] = true
					soIDs = append(soIDs, reqSalesInvoiceDt.RefID)
				}
			}

			if reqSalesInvoiceDt.RefDtID != nil && *reqSalesInvoiceDt.RefDtID > 0 {
				soDtIDs = append(soDtIDs, reqSalesInvoiceDt.RefDtID)
			}
		}
	}

	return soIDs, soDtIDs
}

func MapCreateSalesInvoiceDts(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, createdSalesInvoice *models.SalesInvoice, userID uint, span opentracing.Span) ([]models.SalesInvoiceDt, error) {
	salesInvoiceDtsModel := []models.SalesInvoiceDt{}

	for _, salesInvoiceDt := range req.SalesInvoiceDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		salesInvoiceDtModel := models.SalesInvoiceDt{
			ProductUuid:    salesInvoiceDt.ProductUuid,
			SalesInvoiceID: &createdSalesInvoice.ID,
			ItemUnitID:     salesInvoiceDt.ItemUnitID,
			VatID:          salesInvoiceDt.VatID,
			Pph23ID:        salesInvoiceDt.Pph23ID,
			RefID:          salesInvoiceDt.RefID,
			RefDtID:        salesInvoiceDt.RefDtID,
			ProductID:      salesInvoiceDt.ProductID,
			RefType:        salesInvoiceDt.RefType,
			ProductType:    salesInvoiceDt.ProductType,
			RefJSON:        &refJSON,
			ProductJSON:    &productJSON,
			Remark:         salesInvoiceDt.Remark,
			IsVat:          salesInvoiceDt.IsVat,
			IsPph23:        salesInvoiceDt.IsPph23,
			Qty:            salesInvoiceDt.Qty,
			Price:          salesInvoiceDt.Price,
			Subtotal:       salesInvoiceDt.Subtotal,
			Discount:       salesInvoiceDt.Discount,
			TotalAmount:    salesInvoiceDt.TotalAmount,
			TotalDp:        salesInvoiceDt.TotalDp,
			TotalBalance:   salesInvoiceDt.TotalBalance,
			CreatedByID:    &userID,
		}
		salesInvoiceDtsModel = append(salesInvoiceDtsModel, salesInvoiceDtModel)
	}

	return salesInvoiceDtsModel, nil
}

func MapUpdateSalesInvoiceDts(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, updatedSalesInvoice *models.SalesInvoice, userID uint, span opentracing.Span) ([]models.SalesInvoiceDt, error) {
	salesInvoiceDtsModel := []models.SalesInvoiceDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqSalesInvoiceDt := range req.SalesInvoiceDts {
		salesInvoiceDtID := uint(0)
		if reqSalesInvoiceDt.SalesInvoiceDtID != nil {
			salesInvoiceDtID = *reqSalesInvoiceDt.SalesInvoiceDtID
		}

		salesInvoiceDtModel := models.SalesInvoiceDt{
			ID:             salesInvoiceDtID,
			ProductUuid:    reqSalesInvoiceDt.ProductUuid,
			SalesInvoiceID: &updatedSalesInvoice.ID,
			ItemUnitID:     reqSalesInvoiceDt.ItemUnitID,
			VatID:          reqSalesInvoiceDt.VatID,
			Pph23ID:        reqSalesInvoiceDt.Pph23ID,
			RefID:          reqSalesInvoiceDt.RefID,
			RefDtID:        reqSalesInvoiceDt.RefDtID,
			ProductID:      reqSalesInvoiceDt.ProductID,
			RefType:        reqSalesInvoiceDt.RefType,
			ProductType:    reqSalesInvoiceDt.ProductType,
			RefJSON:        &refJSON,
			ProductJSON:    &productJSON,
			Remark:         reqSalesInvoiceDt.Remark,
			IsVat:          reqSalesInvoiceDt.IsVat,
			IsPph23:        reqSalesInvoiceDt.IsPph23,
			Qty:            reqSalesInvoiceDt.Qty,
			Price:          reqSalesInvoiceDt.Price,
			Subtotal:       reqSalesInvoiceDt.Subtotal,
			Discount:       reqSalesInvoiceDt.Discount,
			TotalAmount:    reqSalesInvoiceDt.TotalAmount,
			TotalDp:        reqSalesInvoiceDt.TotalDp,
			TotalBalance:   reqSalesInvoiceDt.TotalBalance,
			CreatedByID:    &userID,
		}
		salesInvoiceDtsModel = append(salesInvoiceDtsModel, salesInvoiceDtModel)
	}

	return salesInvoiceDtsModel, nil
}

func GenSalesInvoiceNo(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, orderedNumber int, span opentracing.Span) string {
	prefix := "ISL"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateSalesInvoiceNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, revNo *int, span opentracing.Span) string {
	if req.InvoiceNo == nil {
		prefix := "ISL"
		year := time.Now().Format("2006")
		month := time.Now().Format("01")
		day := time.Now().Format("02")

		return fmt.Sprintf("%s/REV-%d/%s-%s-%s", prefix, *revNo, year, month, day)
	}

	invoiceNo := *req.InvoiceNo

	if !strings.Contains(invoiceNo, "REV") {
		return fmt.Sprintf("%s/REV-%d", invoiceNo, *revNo)
	} else {
		basePart := strings.Split(invoiceNo, "/REV")[0]
		return fmt.Sprintf("%s/REV-%d", basePart, *revNo)
	}
}

func MapCreateSalesInvoice(ctx *fiber.Ctx, req dtos.CreateSalesInvoiceRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.SalesInvoice, error) {
	invoiceNo := GenSalesInvoiceNo(ctx, req, orderedNumber, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		dueDate = &parsedTime
	}

	salesInvoice := models.SalesInvoice{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		RevNo:                    new(int),
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}
	*salesInvoice.RevNo = 0

	return salesInvoice, nil
}

func MapUpdateSalesInvoice(ctx *fiber.Ctx, req dtos.UpdateSalesInvoiceRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.SalesInvoice, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := GenerateSalesInvoiceNoOnUpdate(ctx, req, &revNo, span)

	var invoiceDate *time.Time
	if req.InvoiceDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		invoiceDate = &parsedTime
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return models.SalesInvoice{}, err
		}
		dueDate = &parsedTime
	}

	salesInvoice := models.SalesInvoice{
		ID:                       req.ID,
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              invoiceDate,
		DueDate:                  dueDate,
		ExchangeRate:             req.ExchangeRate,
		Remark:                   req.Remark,
		Status:                   req.Status,
		RevNo:                    &revNo,
		Pph23Percentage:          req.Pph23Percentage,
		VatPercentage:            req.VatPercentage,
		DiscountAmount:           req.DiscountAmount,
		DiscountPercentage:       req.DiscountPercentage,
		DiscountPercentageAmount: req.DiscountPercentageAmount,
		DiscountFinal:            req.DiscountFinal,
		DiscountType:             req.DiscountType,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		TotalBalanceProducts:     req.TotalBalanceProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		UpdatedByID:              &userID,
	}

	return salesInvoice, nil
}

func GetSalesOrderDtIDs(soDts []dtos.RefSalesOrderForInvoiceListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDtsForInvoice(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderForInvoiceListDTO) []dtos.RefSalesOrderForInvoiceListDTO {
	combinedSoDts := []dtos.RefSalesOrderForInvoiceListDTO{}

	for _, soDt := range soDts {
		newSoDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
		for _, soDtBom := range soDtBoms {
			if *soDtBom.SoDtID == *soDt.ID {
				newSoDtBoms = append(newSoDtBoms, soDtBom)
			}
		}

		soDt.SoDtsBoms = newSoDtBoms
		combinedSoDts = append(combinedSoDts, soDt)
	}

	return combinedSoDts
}

func MapUpdateSalesOrderStatusForInvoice(salesOrderStatusUpdate map[string]interface{}) dtos.UpdateSalesOrderStatusForInvoiceRequest {
	params := dtos.UpdateSalesOrderStatusForInvoiceRequest{}

	params.ID = salesOrderStatusUpdate["id"].(uint)
	params.Status = salesOrderStatusUpdate["status"].(string)

	return params
}

// Get condition for quotation
func GetSalesInvoiceCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetSalesInvoiceCondition", opentracing.ChildOf(span.Context()))
	var err error

	condition := ""
	var args []interface{}
	queryGlobal := ""
	joinCondition := ""
	customCondition := ""
	i := 1

	filterDBColumnKey := []string{
		"si.invoice_no", "si.remark", "si.status", "si.title",
		"it.name",
		"p.name",
		"c.name",
		"sidt.remark",
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

	filterKey := map[string]string{
		"customer_id":     "si.customer_id",
		"currency_id":     "si.currency_id",
		"payment_term_id": "si.payment_term_id",
		"vat_id":          "si.vat_id",
		"pph23_id":        "si.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND si.status = $%d", i)
		args = append(args, value)
		i++
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "si.customer_id",
		"currency_ids":     "si.currency_id",
		"payment_term_ids": "si.payment_term_id",
		"pph23_ids":        "si.pph23_id",
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
		"vat_ids": {"si.vat_id", "sidt.vat_id"},
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

	// if date_type, start_date, end_date filled
	if filters["date_type"] != "" && filters["start_date"] != "" && filters["end_date"] != "" {

		filterDateTypeKey := map[string]string{
			"invoice_date": "si.invoice_date",
			"due_date":     "si.due_date",
		}

		dateTypeColumn := "si.invoice_date"
		for key := range filterDateTypeKey {
			if key == filters["date_type"] {
				dateTypeColumn = filterDateTypeKey[filters["date_type"]]
			}
		}

		condition += fmt.Sprintf(" AND (%s BETWEEN $%d AND $%d)", dateTypeColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	return args, i, condition, queryGlobal, joinCondition, customCondition, err
}

func GetSalesInvoiceDetailsIDs(quotations []dtos.SalesInvoiceDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToSalesInvoices(quoDts []dtos.SalesInvoiceDtListDTO, quotations []dtos.SalesInvoiceDetailDTO) []dtos.SalesInvoiceDetailDTO {
	// quotations := []dtos.SalesInvoiceAttachmentsDTO{}
	for i, quotation := range quotations {
		newDts := make([]dtos.SalesInvoiceDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.SalesInvoiceID == quotation.ID {
				newDts = append(newDts, quoDt)
			}
		}
		quotation.SalesInvoiceDts = newDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetSalesInvoiceDetails(salesOrders []dtos.SalesInvoiceDetailDTO, soDts []dtos.SalesInvoiceDtListDTO, soDtBoms []dtos.SalesOrderSoDtBomListDTO) []dtos.SalesInvoiceDetailDTO {
	// salesOrders := []dtos.SalesInvoiceAttachmentsDTO{}
	for i, salesOrder := range salesOrders {
		newDts := make([]dtos.SalesInvoiceDtListDTO, 0)
		for _, soDt := range soDts {
			if *soDt.SalesInvoiceID == salesOrder.ID {
				newDtBoms := make([]dtos.SalesOrderSoDtBomListDTO, 0)
				for _, soDtBom := range soDtBoms {
					if *soDtBom.SoDtID == *soDt.RefDtID {
						newDtBoms = append(newDtBoms, soDtBom)
					}
				}
				soDt.SoDtsBoms = newDtBoms
				newDts = append(newDts, soDt)
			}
		}
		salesOrder.SalesInvoiceDts = newDts
		salesOrders[i] = salesOrder
	}

	return salesOrders
}

// Build CSV rows, dtos.SalesInvoiceListDTO, csv pointer
func BuildSalesInvoiceAllCSVRows(salesOrders []dtos.SalesInvoiceListDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Invoice No", "Customer", "Order Type", "Title", "Invoice Date", "Due Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
	}
	rows = append(rows, header)

	for idx, salesOrder := range salesOrders {
		ID := fmt.Sprintf("%d", idx+1)
		InvoiceNo := GetPtrVal(salesOrder.InvoiceNo)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		OrderTypeName := GetPtrVal(salesOrder.OrderTypeName)
		Title := GetPtrVal(salesOrder.Title)
		InvoiceDate := GetPtrVal(salesOrder.InvoiceDate)
		DueDate := GetPtrVal(salesOrder.DueDate)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		// EscapeCsvField
		ID = EscapeCsvField(ID)
		InvoiceNo = EscapeCsvField(InvoiceNo)
		CustomerName = EscapeCsvField(CustomerName)
		OrderTypeName = EscapeCsvField(OrderTypeName)
		Title = EscapeCsvField(Title)
		InvoiceDate = EscapeCsvField(InvoiceDate)
		DueDate = EscapeCsvField(DueDate)
		CurrencyName = EscapeCsvField(CurrencyName)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		row := []string{
			ID, InvoiceNo, CustomerName, OrderTypeName, Title, InvoiceDate, DueDate,
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

// Build CSV rows, dtos.SalesInvoiceListDTO, csv pointer
func BuildSalesInvoiceDetailCSVRows(salesOrders []dtos.SalesInvoiceDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Invoice No", "Customer", "Order Type", "Title", "Invoice Date", "Due Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
		"Product/Item Name", "Qty", "Price", "Subtotal",
		"BOM Item Name", "BOM Qty",
	}
	rows = append(rows, header)

	for iSalesInvoice, salesOrder := range salesOrders {

		No := fmt.Sprintf("%d", iSalesInvoice+1)
		InvoiceNo := GetPtrVal(salesOrder.InvoiceNo)
		CustomerName := GetPtrVal(salesOrder.CustomerName)
		OrderTypeName := GetPtrVal(salesOrder.OrderTypeName)
		Title := GetPtrVal(salesOrder.Title)
		InvoiceDate := GetPtrVal(salesOrder.InvoiceDate)
		DueDate := GetPtrVal(salesOrder.DueDate)
		CurrencyName := GetPtrVal(salesOrder.CurrencyName)
		Status := GetPtrVal(salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		for iDt, quoDt := range salesOrder.SalesInvoiceDts {
			ProductItemName := GetPtrVal(quoDt.ItemName)
			Qty := fmt.Sprintf("%f", *quoDt.Qty)
			Price := fmt.Sprintf("%f", *quoDt.Price)
			TotalAm := fmt.Sprintf("%f", *quoDt.Subtotal)

			if len(quoDt.SoDtsBoms) == 0 {

				if iDt == 0 {
					row := []string{
						No, InvoiceNo, CustomerName, OrderTypeName, Title, InvoiceDate, DueDate,
						CurrencyName, fmt.Sprintf("%f", salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
						ProductItemName, Qty, Price, TotalAm,
						"", "", "", "",
					}
					rows = append(rows, row)
				} else {
					row := []string{
						"", "", "", "", "", "", "",
						"", "", "", "", "",
						ProductItemName, Qty, Price, TotalAm,
						"", "", "", "",
					}
					rows = append(rows, row)
				}
			} else {
				for iBom, bom := range quoDt.SoDtsBoms {
					BomItemName := GetPtrVal(bom.ItemName)
					BomQty := fmt.Sprintf("%f", bom.Qty)

					if iDt == 0 && iBom == 0 {
						row := []string{
							No, InvoiceNo, CustomerName, OrderTypeName, Title, InvoiceDate, DueDate,
							CurrencyName, fmt.Sprintf("%f", salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
							ProductItemName, Qty, Price, TotalAm,
							BomItemName, BomQty,
						}
						rows = append(rows, row)
					} else if iBom == 0 {
						row := []string{
							"", "", "", "", "", "", "",
							"", "", "", "", "",
							ProductItemName, Qty, Price, TotalAm,
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
		if len(salesOrder.SalesInvoiceDts) == 0 {
			row := []string{
				No, InvoiceNo, CustomerName, OrderTypeName, Title, InvoiceDate, DueDate,
				CurrencyName, fmt.Sprintf("%f", salesOrder.GrandTotal), Status, CreatedByName, UpdatedByName,
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
