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

func GetInvoiceDpIDs(req dtos.UpdateInvoiceDpRequest) ([]*uint, []*uint, []*uint) {
	invoiceDpDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.InvoiceDpDtID != nil && *reqInvoiceDpDt.InvoiceDpDtID > 0 {
			invoiceDpDtIDs = append(invoiceDpDtIDs, reqInvoiceDpDt.InvoiceDpDtID)
		}
	}

	return invoiceDpDtIDs, productIDs, itemUnitIDs
}

func GetLockInvoiceDpSalesOrderIDs(req dtos.CreateInvoiceDpRequest) []*uint {
	soDtIDs := []*uint{}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.RefType != nil && *reqInvoiceDpDt.RefType == "sales_orders" && reqInvoiceDpDt.RefID != nil && *reqInvoiceDpDt.RefID > 0 {
			soDtIDs = append(soDtIDs, reqInvoiceDpDt.RefID)
		}
	}

	return soDtIDs
}

func MapCreateInvoiceDpDts(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, createdInvoiceDp *models.InvoiceDp, userID uint, span opentracing.Span) ([]models.InvoiceDpDt, []map[string]interface{}, error) {
	invoiceDpDtsModel := []models.InvoiceDpDt{}
	updateSalesOrderIDs := []map[string]interface{}{}
	// filter unique sales order IDs

	for _, invoiceDpDt := range req.InvoiceDpDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		invoiceDpDtModel := models.InvoiceDpDt{
			ProductUuid:  invoiceDpDt.ProductUuid,
			InvoiceDpID:  &createdInvoiceDp.ID,
			ItemUnitID:   invoiceDpDt.ItemUnitID,
			VatID:        invoiceDpDt.VatID,
			Pph23ID:      invoiceDpDt.Pph23ID,
			RefID:        invoiceDpDt.RefID,
			RefDtID:      invoiceDpDt.RefDtID,
			ProductID:    invoiceDpDt.ProductID,
			RefType:      invoiceDpDt.RefType,
			ProductType:  invoiceDpDt.ProductType,
			RefJSON:      &refJSON,
			ProductJSON:  &productJSON,
			Remark:       invoiceDpDt.Remark,
			DpPercentage: invoiceDpDt.DpPercentage,
			IsVat:        invoiceDpDt.IsVat,
			IsPph23:      invoiceDpDt.IsPph23,
			Qty:          invoiceDpDt.Qty,
			Price:        invoiceDpDt.Price,
			Subtotal:     invoiceDpDt.Subtotal,
			Discount:     invoiceDpDt.Discount,
			TotalAmount:  invoiceDpDt.TotalAmount,
			TotalDp:      invoiceDpDt.TotalDp,
			CreatedByID:  &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
		updateSalesOrderIDs = append(updateSalesOrderIDs, map[string]interface{}{
			"id":     invoiceDpDt.RefID,
			"status": "INVOICE",
		})
	}

	uniqueSalesOrderIDs := []map[string]interface{}{}
	seen := make(map[uint]bool)

	for _, entry := range updateSalesOrderIDs {
		idPtr, ok := entry["id"].(*uint)
		if !ok || idPtr == nil {
			continue
		}
		id := *idPtr
		if !seen[id] {
			uniqueSalesOrderIDs = append(uniqueSalesOrderIDs, entry)
			seen[id] = true
		}
	}

	return invoiceDpDtsModel, uniqueSalesOrderIDs, nil
}

func MapUpdateInvoiceDpDts(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, updatedInvoiceDp *models.InvoiceDp, userID uint, span opentracing.Span) ([]models.InvoiceDpDt, error) {
	invoiceDpDtsModel := []models.InvoiceDpDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		invoiceDpDtID := uint(0)
		if reqInvoiceDpDt.InvoiceDpDtID != nil {
			invoiceDpDtID = *reqInvoiceDpDt.InvoiceDpDtID
		}

		invoiceDpDtModel := models.InvoiceDpDt{
			ID:           invoiceDpDtID,
			ProductUuid:  reqInvoiceDpDt.ProductUuid,
			InvoiceDpID:  &updatedInvoiceDp.ID,
			ItemUnitID:   reqInvoiceDpDt.ItemUnitID,
			VatID:        reqInvoiceDpDt.VatID,
			Pph23ID:      reqInvoiceDpDt.Pph23ID,
			RefID:        reqInvoiceDpDt.RefID,
			RefDtID:      reqInvoiceDpDt.RefDtID,
			ProductID:    reqInvoiceDpDt.ProductID,
			RefType:      reqInvoiceDpDt.RefType,
			ProductType:  reqInvoiceDpDt.ProductType,
			RefJSON:      &refJSON,
			ProductJSON:  &productJSON,
			Remark:       reqInvoiceDpDt.Remark,
			DpPercentage: reqInvoiceDpDt.DpPercentage,
			IsVat:        reqInvoiceDpDt.IsVat,
			IsPph23:      reqInvoiceDpDt.IsPph23,
			Qty:          reqInvoiceDpDt.Qty,
			Price:        reqInvoiceDpDt.Price,
			Subtotal:     reqInvoiceDpDt.Subtotal,
			Discount:     reqInvoiceDpDt.Discount,
			TotalAmount:  reqInvoiceDpDt.TotalAmount,
			TotalDp:      reqInvoiceDpDt.TotalDp,
			CreatedByID:  &userID,
		}
		invoiceDpDtsModel = append(invoiceDpDtsModel, invoiceDpDtModel)
	}

	return invoiceDpDtsModel, nil
}

func GenInvoiceDpNo(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, orderedNumber int, span opentracing.Span) string {
	// if req.InvoiceNo != nil {
	// 	return *req.InvoiceNo
	// }

	prefix := "IDP"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateInvoiceDpNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, revNo *int, span opentracing.Span) string {
	invoiceNo := req.InvoiceNo

	if !strings.Contains(*invoiceNo, "REV") {
		*invoiceNo = fmt.Sprintf("%s/REV-%d", *invoiceNo, *revNo)
	} else {
		*invoiceNo = strings.Split(*invoiceNo, "/REV")[0]
		*invoiceNo = fmt.Sprintf("%s/REV-%d", *invoiceNo, *revNo)
	}

	return *invoiceNo
}

func MapCreateInvoiceDp(ctx *fiber.Ctx, req dtos.CreateInvoiceDpRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.InvoiceDp, error) {
	invoiceNo := GenInvoiceDpNo(ctx, req, orderedNumber, span)

	invoiceDp := models.InvoiceDp{
		CustomerID:               req.CustomerID,
		CurrencyID:               req.CurrencyID,
		PaymentTermID:            req.PaymentTermID,
		VatID:                    req.VatID,
		Pph23ID:                  req.Pph23ID,
		BranchID:                 &branchID,
		BankID:                   req.BankID,
		Title:                    req.Title,
		InvoiceNo:                &invoiceNo,
		InvoiceDate:              req.InvoiceDate,
		DueDate:                  req.DueDate,
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
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		CreatedByID:              &userID,
	}
	*invoiceDp.RevNo = 0

	return invoiceDp, nil
}

func MapUpdateInvoiceDp(ctx *fiber.Ctx, req dtos.UpdateInvoiceDpRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.InvoiceDp, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	invoiceNo := *req.InvoiceNo
	invoiceNo = GenerateInvoiceDpNoOnUpdate(ctx, req, &revNo, span)

	invoiceDp := models.InvoiceDp{
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
		InvoiceDate:              req.InvoiceDate,
		DueDate:                  req.DueDate,
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
		DpPercentage:             req.DpPercentage,
		TotalAmountProducts:      req.TotalAmountProducts,
		TotalDpProducts:          req.TotalDpProducts,
		Subtotal:                 req.Subtotal,
		TotalQty:                 req.TotalQty,
		TotalDiscount:            req.TotalDiscount,
		TotalPph23:               req.TotalPph23,
		TotalVat:                 req.TotalVat,
		GrandTotal:               req.GrandTotal,
		UpdatedByID:              &userID,
	}

	return invoiceDp, nil
}

func GetSoDtIDs(soDts []dtos.RefSalesOrderDtListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

func MapRefSoDtBomsToSoDts(soDtBoms []dtos.SalesOrderSoDtBomListDTO, soDts []dtos.RefSalesOrderDtListDTO) []dtos.RefSalesOrderDtListDTO {
	combinedSoDts := []dtos.RefSalesOrderDtListDTO{}

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

func MapUpdateSoDtsQtyForInvoice(soDtsQtyUpdate []dtos.GetSoDtQtyUpdateForInvoiceDTO, req dtos.CreateInvoiceDpRequest) []map[string]interface{} {
	bulkUpdateSoDts := []map[string]interface{}{}

	soDtsQtyMap := make(map[uint]*dtos.GetSoDtQtyUpdateForInvoiceDTO)
	for i := range soDtsQtyUpdate {
		if soDtsQtyUpdate[i].SoDtID != nil {
			soDtsQtyMap[*soDtsQtyUpdate[i].SoDtID] = &soDtsQtyUpdate[i]
		}
	}

	for _, reqInvoiceDpDt := range req.InvoiceDpDts {
		if reqInvoiceDpDt.RefID != nil {
			soDtID := *reqInvoiceDpDt.RefID
			if soDt, exists := soDtsQtyMap[soDtID]; exists {
				if soDt.QtyInvoiced == nil {
					soDt.QtyInvoiced = new(float64)
				}

				newSoDt := map[string]interface{}{
					"id":           soDtID,
					"qty_invoiced": (*reqInvoiceDpDt.Qty + *soDt.QtyInvoiced),
				}
				bulkUpdateSoDts = append(bulkUpdateSoDts, newSoDt)
			}
		}
	}

	return bulkUpdateSoDts
}

// func MapUpdateSalesOrderStatusForInvoice(salesOrderStatusUpdate map[string]interface{}, req dtos.CreateInvoiceDpRequest) dtos.UpdateSalesOrderStatusForInvoiceRequest {
// 	params := dtos.UpdateSalesOrderStatusForInvoiceRequest{}

// 	params.ID = salesOrderStatusUpdate["id"].(uint)
// 	params.Status = salesOrderStatusUpdate["status"].(string)

// 	return params
// }

// Get condition for quotation
func GetInvoiceDpCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetInvoiceDpCondition", opentracing.ChildOf(span.Context()))
	var err error

	condition := ""
	var args []interface{}
	queryGlobal := ""
	joinCondition := ""
	customCondition := ""
	i := 1

	filterDBColumnKey := []string{
		"idp.invoice_no", "idp.remark", "idp.status", "idp.title",
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
		"customer_id":     "idp.customer_id",
		"currency_id":     "idp.currency_id",
		"payment_term_id": "idp.payment_term_id",
		"vat_id":          "idp.vat_id",
		"pph23_id":        "idp.pph23_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND idp.status = $%d", i)
		args = append(args, value)
		i++
	}

	filterIDsKey := map[string]string{
		"customer_ids":     "idp.customer_id",
		"currency_ids":     "idp.currency_id",
		"payment_term_ids": "idp.payment_term_id",
		"pph23_ids":        "idp.pph23_id",
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
		"vat_ids": {"idp.vat_id", "sidt.vat_id"},
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
			"invoice_date": "idp.invoice_date",
			"due_date":     "idp.due_date",
		}

		dateTypeColumn := "idp.invoice_date"
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

func GetInvoiceDpDetailsIDs(quotations []dtos.InvoiceDpDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToInvoiceDps(quoDts []dtos.InvoiceDpDtListDTO, quotations []dtos.InvoiceDpDetailDTO) []dtos.InvoiceDpDetailDTO {
	// quotations := []dtos.InvoiceDpAttachmentsDTO{}
	for i, quotation := range quotations {
		newDts := make([]dtos.InvoiceDpDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.InvoiceDpID == quotation.ID {
				newDts = append(newDts, quoDt)
			}
		}
		quotation.InvoiceDpDts = newDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetInvoiceDpDetails(salesOrders []dtos.InvoiceDpDetailDTO, soDts []dtos.InvoiceDpDtListDTO, soDtBoms []dtos.SalesOrderSoDtBomListDTO) []dtos.InvoiceDpDetailDTO {
	// salesOrders := []dtos.InvoiceDpAttachmentsDTO{}
	for i, salesOrder := range salesOrders {
		newDts := make([]dtos.InvoiceDpDtListDTO, 0)
		for _, soDt := range soDts {
			if *soDt.InvoiceDpID == salesOrder.ID {
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
		salesOrder.InvoiceDpDts = newDts
		salesOrders[i] = salesOrder
	}

	return salesOrders
}

// Build CSV rows, dtos.InvoiceDpListDTO, csv pointer
func BuildInvoiceDpAllCSVRows(salesOrders []dtos.InvoiceDpListDTO, csv *string) error {
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

// Build CSV rows, dtos.InvoiceDpListDTO, csv pointer
func BuildInvoiceDpDetailCSVRows(salesOrders []dtos.InvoiceDpDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Invoice No", "Customer", "Order Type", "Title", "Invoice Date", "Due Date",
		"Currency", "Total", "Status", "Created By", "Updated By",
		"Product/Item Name", "Qty", "Price", "Subtotal",
		"BOM Item Name", "BOM Qty",
	}
	rows = append(rows, header)

	for iInvoiceDp, salesOrder := range salesOrders {

		No := fmt.Sprintf("%d", iInvoiceDp+1)
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

		for iDt, quoDt := range salesOrder.InvoiceDpDts {
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
		if len(salesOrder.InvoiceDpDts) == 0 {
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
