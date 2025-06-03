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

func GetRequestOrderIDs(req dtos.UpdateRequestOrderRequest) ([]*uint, []*uint, []*uint) {
	requestOrderDtIDs := []*uint{}
	productIDs := []*uint{}
	itemUnitIDs := []*uint{}

	for _, reqRequestOrderDt := range req.RequestOrderDts {
		if reqRequestOrderDt.RequestOrderDtID != nil && *reqRequestOrderDt.RequestOrderDtID > 0 {
			requestOrderDtIDs = append(requestOrderDtIDs, reqRequestOrderDt.RequestOrderDtID)
		}
	}

	return requestOrderDtIDs, productIDs, itemUnitIDs
}

func GetLockRequestOrderSalesOrderIDs(req dtos.CreateRequestOrderRequest) ([]*uint, []*uint) {
	soIDs := []*uint{}
	soDtIDs := []*uint{}

	soIDsMap := make(map[uint]bool)

	for _, reqRequestOrderDt := range req.RequestOrderDts {
		if reqRequestOrderDt.RefType != nil && *reqRequestOrderDt.RefType == "so" {
			if reqRequestOrderDt.RefID != nil && *reqRequestOrderDt.RefID > 0 {
				if !soIDsMap[*reqRequestOrderDt.RefID] {
					soIDsMap[*reqRequestOrderDt.RefID] = true
					soIDs = append(soIDs, reqRequestOrderDt.RefID)
				}
			}

			if reqRequestOrderDt.RefID != nil && *reqRequestOrderDt.RefID > 0 {
				soDtIDs = append(soDtIDs, reqRequestOrderDt.RefID)
			}
		}
	}

	return soIDs, soDtIDs
}

func MapCreateRequestOrderDts(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, createdRequestOrder *models.RequestOrder, userID uint, span opentracing.Span) ([]models.RequestOrderDt, error) {
	requestOrderDtsModel := []models.RequestOrderDt{}

	for _, requestOrderDt := range req.RequestOrderDts {
		refJSONStr := "{}"
		productJSONStr := "{}"

		refJSON := json.RawMessage(refJSONStr)
		productJSON := json.RawMessage(productJSONStr)

		requestOrderDtModel := models.RequestOrderDt{
			ProductUuid:     requestOrderDt.ProductUuid,
			RequestOrderID:  &createdRequestOrder.ID,
			ItemUnitID:      requestOrderDt.ItemUnitID,
			RefID:           requestOrderDt.RefID,
			ProductID:       requestOrderDt.ProductID,
			ItemID:          requestOrderDt.ItemID,
			RefType:         requestOrderDt.RefType,
			ProductType:     requestOrderDt.ProductType,
			RefJSON:         &refJSON,
			ProductJSON:     &productJSON,
			ProductName:     requestOrderDt.ProductName,
			ItemName:        requestOrderDt.ItemName,
			UnitName:        requestOrderDt.UnitName,
			PriceSell:       requestOrderDt.PriceSell,
			Remark:          requestOrderDt.Remark,
			OrderProductQty: requestOrderDt.OrderProductQty,
			OrderItemQty:    requestOrderDt.OrderItemQty,
			WhQty:           requestOrderDt.WhQty,
			ReqQty:          requestOrderDt.ReqQty,
			CreatedByID:     &userID,
		}
		requestOrderDtsModel = append(requestOrderDtsModel, requestOrderDtModel)
	}

	return requestOrderDtsModel, nil
}

func MapUpdateRequestOrderDts(ctx *fiber.Ctx, req dtos.UpdateRequestOrderRequest, updatedRequestOrder *models.RequestOrder, userID uint, span opentracing.Span) ([]models.RequestOrderDt, error) {
	requestOrderDtsModel := []models.RequestOrderDt{}

	refJSONStr := "{}"
	productJSONStr := "{}"

	refJSON := json.RawMessage(refJSONStr)
	productJSON := json.RawMessage(productJSONStr)

	for _, reqRequestOrderDt := range req.RequestOrderDts {
		requestOrderDtID := uint(0)
		if reqRequestOrderDt.RequestOrderDtID != nil {
			requestOrderDtID = *reqRequestOrderDt.RequestOrderDtID
		}

		requestOrderDtModel := models.RequestOrderDt{
			ID:              requestOrderDtID,
			ProductUuid:     reqRequestOrderDt.ProductUuid,
			RequestOrderID:  &updatedRequestOrder.ID,
			ItemUnitID:      reqRequestOrderDt.ItemUnitID,
			RefID:           reqRequestOrderDt.RefID,
			ProductID:       reqRequestOrderDt.ProductID,
			ItemID:          reqRequestOrderDt.ItemID,
			RefType:         reqRequestOrderDt.RefType,
			ProductType:     reqRequestOrderDt.ProductType,
			RefJSON:         &refJSON,
			ProductJSON:     &productJSON,
			ProductName:     reqRequestOrderDt.ProductName,
			ItemName:        reqRequestOrderDt.ItemName,
			UnitName:        reqRequestOrderDt.UnitName,
			PriceSell:       reqRequestOrderDt.PriceSell,
			Remark:          reqRequestOrderDt.Remark,
			OrderProductQty: reqRequestOrderDt.OrderProductQty,
			OrderItemQty:    reqRequestOrderDt.OrderItemQty,
			WhQty:           reqRequestOrderDt.WhQty,
			ReqQty:          reqRequestOrderDt.ReqQty,
			CreatedByID:     &userID,
		}
		requestOrderDtsModel = append(requestOrderDtsModel, requestOrderDtModel)
	}

	return requestOrderDtsModel, nil
}

func GenRequestOrderNo(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, orderedNumber int, span opentracing.Span) string {
	prefix := "RO"
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	order := fmt.Sprintf("%d", orderedNumber)

	str := fmt.Sprintf("%s/%s/%s-%s-%s", prefix, order, year, month, day)

	return str
}

func GenerateRequestOrderNoOnUpdate(ctx *fiber.Ctx, req dtos.UpdateRequestOrderRequest, revNo *int, span opentracing.Span) string {
	if req.RequestNo == nil {
		prefix := "RO"
		year := time.Now().Format("2006")
		month := time.Now().Format("01")
		day := time.Now().Format("02")

		return fmt.Sprintf("%s/REV-%d/%s-%s-%s", prefix, *revNo, year, month, day)
	}

	requestNo := *req.RequestNo

	if !strings.Contains(requestNo, "REV") {
		return fmt.Sprintf("%s/REV-%d", requestNo, *revNo)
	} else {
		basePart := strings.Split(requestNo, "/REV")[0]
		return fmt.Sprintf("%s/REV-%d", basePart, *revNo)
	}
}

func MapCreateRequestOrder(ctx *fiber.Ctx, req dtos.CreateRequestOrderRequest, userID uint, branchID uint, orderedNumber int, span opentracing.Span) (models.RequestOrder, error) {
	requestNo := GenRequestOrderNo(ctx, req, orderedNumber, span)

	var requestDate *time.Time
	if req.RequestDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RequestDate)
		if err != nil {
			return models.RequestOrder{}, err
		}
		requestDate = &parsedTime
	}

	requestOrder := models.RequestOrder{
		BranchID:                  &branchID,
		WarehouseID:               req.WarehouseID,
		RequestNo:                 &requestNo,
		RequestDate:               requestDate,
		Remark:                    req.Remark,
		Requested:                 req.Requested,
		RevNo:                     new(int),
		Status:                    req.Status,
		GrandTotalOrderProductQty: req.GrandTotalOrderProductQty,
		GrandTotalOrderItemQty:    req.GrandTotalOrderItemQty,
		GrandTotalWhQty:           req.GrandTotalWhQty,
		GrandTotalReqQty:          req.GrandTotalReqQty,
		CreatedByID:               &userID,
	}
	*requestOrder.RevNo = 0

	return requestOrder, nil
}

func MapUpdateRequestOrder(ctx *fiber.Ctx, req dtos.UpdateRequestOrderRequest, userID uint, branchID uint, existingRevNo *int, span opentracing.Span) (models.RequestOrder, error) {
	revNo := 0
	if existingRevNo != nil {
		revNo = *existingRevNo + 1
	} else {
		revNo = 1
	}

	requestNo := GenerateRequestOrderNoOnUpdate(ctx, req, &revNo, span)

	var requestDate *time.Time
	if req.RequestDate != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.RequestDate)
		if err != nil {
			return models.RequestOrder{}, err
		}
		requestDate = &parsedTime
	}

	requestOrder := models.RequestOrder{
		ID:                        req.ID,
		BranchID:                  &branchID,
		WarehouseID:               req.WarehouseID,
		RequestNo:                 &requestNo,
		RequestDate:               requestDate,
		Remark:                    req.Remark,
		Requested:                 req.Requested,
		RevNo:                     &revNo,
		Status:                    req.Status,
		GrandTotalOrderProductQty: req.GrandTotalOrderProductQty,
		GrandTotalOrderItemQty:    req.GrandTotalOrderItemQty,
		GrandTotalWhQty:           req.GrandTotalWhQty,
		GrandTotalReqQty:          req.GrandTotalReqQty,
		UpdatedByID:               &userID,
	}

	return requestOrder, nil
}

func GetSalesOrderDtIDsForRequestOrder(soDts []dtos.RefSalesOrderForRequestOrderListDTO) []uint {
	salesOrderIDs := []uint{}

	for _, soDt := range soDts {
		if soDt.SalesOrderID != nil {
			salesOrderIDs = append(salesOrderIDs, *soDt.SalesOrderID)
		}
	}

	return salesOrderIDs
}

// Get condition for quotation
func GetRequestOrderCondition(ctx *fiber.Ctx, filters map[string]string, span opentracing.Span) ([]interface{}, int, string, string, string, string, error) {
	childSpan := span.Tracer().StartSpan("quotation_utils-GetRequestOrderCondition", opentracing.ChildOf(span.Context()))
	var err error

	joinCondition := ""
	customCondition := ""
	condition := ""
	var args []interface{}
	queryGlobal := ""
	i := 1

	filterDBColumnKey := []string{
		"ro.request_no", "ro.remark", "ro.requested",
		"b.name",
		"p.name",
		"w.name",
		"rodt.remark",
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
		condition += fmt.Sprintf(" AND ro.id IN (%s)", filters["ids"])
	}

	if value, ok := filters["status"]; ok && value != "" {
		condition += fmt.Sprintf(" AND ro.status = $%d", i)
		args = append(args, value)
		i++
	}

	filterKey := map[string]string{
		"warehouse_id": "ro.warehouse_id",
		"customer_id":  "so.customer_id",
	}

	for key, col := range filterKey {
		if value, ok := filters[key]; ok && value != "" {
			condition += fmt.Sprintf(" AND %s = $%d", col, i)
			args = append(args, value)
			i++
		}
	}

	filterIDsKey := map[string]string{
		"warehouse_ids": "ro.warehouse_id",
		"customer_ids":  "so.customer_id",
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

	// And one for array conditions with OR
	filterIDsOrArrayKey := map[string][]string{
		"product_ids": {"rodt.item_id", "rodt.product_id"},
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

	dateType := GetStringOrDefault(filters["date_type"], "request_date")

	if dateType != "" && filters["start_date"] != "" && filters["end_date"] != "" {
		dateColumn := ""
		switch dateType {
		default:
			dateColumn = "ro.request_date"
		}

		condition += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", dateColumn, i, i+1)
		args = append(args, filters["start_date"], filters["end_date"])
		i += 2
	}

	return args, i, condition, queryGlobal, joinCondition, customCondition, err
}

func GetRequestOrderDetailsIDs(quotations []dtos.RequestOrderDetailDTO) []uint {
	quotationsIDs := []uint{}
	for _, quotation := range quotations {
		if quotation.ID > 0 {
			quotationsIDs = append(quotationsIDs, quotation.ID)
		}
	}
	return quotationsIDs
}

func MapFilterQuoDtToRequestOrders(quoDts []dtos.RequestOrderDtListDTO, quotations []dtos.RequestOrderDetailDTO) []dtos.RequestOrderDetailDTO {
	// quotations := []dtos.RequestOrderAttachmentsDTO{}
	for i, quotation := range quotations {
		newSoDts := make([]dtos.RequestOrderDtListDTO, 0)
		for _, quoDt := range quoDts {
			if *quoDt.RequestOrderID == quotation.ID {
				newSoDts = append(newSoDts, quoDt)
			}
		}
		quotation.RequestOrderDts = newSoDts
		quotations[i] = quotation
	}

	return quotations
}

func MapGetRequestOrderDetails(salesOrders []dtos.RequestOrderDetailDTO, soDts []dtos.RequestOrderDtListDTO) []dtos.RequestOrderDetailDTO {
	// salesOrders := []dtos.RequestOrderAttachmentsDTO{}
	for i, salesOrder := range salesOrders {
		newSoDts := make([]dtos.RequestOrderDtListDTO, 0)
		for _, soDt := range soDts {
			if *soDt.RequestOrderID == salesOrder.ID {
				newSoDts = append(newSoDts, soDt)
			}
		}
		salesOrder.RequestOrderDts = newSoDts
		salesOrders[i] = salesOrder
	}

	return salesOrders
}

// Build CSV rows, dtos.RequestOrderListDTO, csv pointer
func BuildRequestOrderAllCSVRows(salesOrders []dtos.RequestOrderListDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"ID", "Request No", "RO Date",
		"Status", "Created By", "Updated By",
	}
	rows = append(rows, header)

	for _, salesOrder := range salesOrders {
		ID := fmt.Sprintf("%d", salesOrder.ID)
		PoNo := GetPtrVal(salesOrder.RequestNo)
		PoDate := GetPtrVal(salesOrder.RequestDate)
		Status := GetPtrVal(salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		// EscapeCsvField
		ID = EscapeCsvField(ID)
		PoNo = EscapeCsvField(PoNo)
		PoDate = EscapeCsvField(PoDate)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		row := []string{
			ID, PoNo, PoDate,
			Status, CreatedByName, UpdatedByName,
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

// Build CSV rows, dtos.RequestOrderListDTO, csv pointer
func BuildRequestOrderDetailCSVRows(salesOrders []dtos.RequestOrderDetailDTO, csv *string) error {
	rows := [][]string{}
	header := []string{
		"No", "Request No", "RO Date",
		"Status", "Created By", "Updated By",
		"Product/Item Name", "Qty",
	}
	rows = append(rows, header)

	for iRequestOrder, salesOrder := range salesOrders {

		No := fmt.Sprintf("%d", iRequestOrder+1)
		PoNo := GetPtrVal(salesOrder.RequestNo)
		PoDate := GetPtrVal(salesOrder.RequestDate)
		Status := GetPtrVal(salesOrder.Status)
		CreatedByName := GetPtrVal(salesOrder.CreatedByName)
		UpdatedByName := GetPtrVal(salesOrder.UpdatedByName)

		No = EscapeCsvField(No)
		PoNo = EscapeCsvField(PoNo)
		PoDate = EscapeCsvField(PoDate)
		Status = EscapeCsvField(Status)
		CreatedByName = EscapeCsvField(CreatedByName)
		UpdatedByName = EscapeCsvField(UpdatedByName)

		for iSoDt, quoDt := range salesOrder.RequestOrderDts {
			ProductItemName := GetPtrVal(quoDt.ItemName)
			ProductItemName = EscapeCsvField(ProductItemName)
			Qty := fmt.Sprintf("%f", *quoDt.ReqQty)

			if iSoDt == 0 {
				row := []string{
					No, PoNo, PoDate,
					Status, CreatedByName, UpdatedByName,
					ProductItemName, Qty,
				}
				rows = append(rows, row)
			} else {
				row := []string{
					"", "", "",
					"", "", "",
					ProductItemName, Qty,
				}
				rows = append(rows, row)
			}
		}
		if len(salesOrder.RequestOrderDts) == 0 {
			row := []string{
				No, PoNo, PoDate,
				Status, CreatedByName, UpdatedByName,
				"", "", "",
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
