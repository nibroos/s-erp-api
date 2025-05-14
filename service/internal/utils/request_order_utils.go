package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
