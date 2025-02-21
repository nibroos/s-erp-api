package rest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
	// "github.com/opentracing/opentracing-go/ext"
)

type OrderTypeController struct {
	service *service.OrderTypeService
	repo    *repository.OrderTypeRepository
	tracer  opentracing.Tracer
}

func NewOrderTypeController(service *service.OrderTypeService, repo *repository.OrderTypeRepository, tracer opentracing.Tracer) *OrderTypeController {
	return &OrderTypeController{service: service, repo: repo, tracer: tracer}
}

func (c *OrderTypeController) GetOrderTypes(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-GetOrderTypes", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, not collecting span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("OrderTypeController-GetOrderTypes: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	orderTypes, total, err := c.service.GetOrderTypes(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, orderTypes, paginationMeta, "OrderType fetched successfully", http.StatusOK, nil, nil)
}

func (c *OrderTypeController) CreateOrderType(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-CreateOrderType", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateOrderTypeRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewOrderTypeStoreRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	orderType := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.OrderTypeID,
		Description: req.Description,
		Remark:      req.Remark,
		OrderItem:   nil,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	createdOrderType, err := c.service.CreateOrderType(ctx, &orderType, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create orderTypes", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetOrderTypeParams{ID: createdOrderType.ID}
	getOrderType, err := c.service.GetOrderTypeByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getOrderType}, paginationMeta, "OrderType created successfully", http.StatusCreated, nil, nil)
}
func (c *OrderTypeController) GetOrderTypeByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-GetOrderTypeByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetOrderTypeByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetOrderTypeParams{ID: req.ID}
	orderType, err := c.service.GetOrderTypeByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusNotFound, err.Error(), nil)
	}

	orderTypeArray := []interface{}{orderType}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, orderTypeArray, paginationMeta, "OrderType fetched successfully", http.StatusOK, nil, nil)
}

// update orderType
func (c *OrderTypeController) UpdateOrderType(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-UpdateOrderType", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateOrderTypeRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewOrderTypeUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	orderType := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.OrderTypeID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	updatedOrderType, err := c.service.UpdateOrderType(ctx, &orderType, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "orderType name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "OrderType already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update OrderType", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetOrderTypeParams{ID: updatedOrderType.ID}
	getOrderType, err := c.service.GetOrderTypeByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getOrderType}, paginationMeta, "OrderType updated successfully", http.StatusOK, nil, nil)
}

// delete orderType
func (c *OrderTypeController) DeleteOrderType(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-DeleteOrderType", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteOrderTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetOrderTypeParams{ID: req.ID}
	// GET orderType by ID
	_, err := c.service.GetOrderTypeByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteOrderType(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete OrderType", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "OrderType deleted successfully", http.StatusOK, nil, nil)
}

// restore orderType
func (c *OrderTypeController) RestoreOrderType(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-RestoreOrderType", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteOrderTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetOrderTypeParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET orderType by ID
	_, err := c.service.GetOrderTypeByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "OrderType not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreOrderType(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore OrderType", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "OrderType restored successfully", http.StatusOK, nil, nil)
}

func (c *OrderTypeController) ExcelGetOrderTypes(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-ExcelGetOrderTypes", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	orderTypes, err := c.service.ExcelGetOrderTypes(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(orderTypes)
}

func (c *OrderTypeController) CsvGetOrderTypes(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("OrderTypeController-CsvGetOrderTypes", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	orderTypes, err := c.service.CsvGetOrderTypes(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(orderTypes)
}
