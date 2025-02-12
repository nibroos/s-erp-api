package rest

import (
	"encoding/json"
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

type CustomerTypeController struct {
	service *service.CustomerTypeService
	repo    *repository.CustomerTypeRepository
	tracer  opentracing.Tracer
}

func NewCustomerTypeController(service *service.CustomerTypeService, repo *repository.CustomerTypeRepository, tracer opentracing.Tracer) *CustomerTypeController {
	return &CustomerTypeController{service: service, repo: repo, tracer: tracer}
}

func (c *CustomerTypeController) GetCustomerTypes(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-GetCustomerTypes")
	defer func() {
		// If no error, delete span
		if ctx.Response().StatusCode() >= 400 {
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		parentSpan.LogKV("response_body", string("CustomerTypeController-GetCustomerTypes: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	customerTypes, total, err := c.service.GetCustomerTypes(ctx.Context(), filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, customerTypes, paginationMeta, "Customer type fetched successfully", http.StatusOK, nil, nil)
}

func (c *CustomerTypeController) CreateCustomerType(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-CreateCustomerType")
	defer func() {
		// If no error, delete span
		if ctx.Response().StatusCode() >= 400 {
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateCustomerTypeRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewCustomerTypeStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		stringReqValidator, _ := json.Marshal(reqValidator)
		parentSpan.LogKV("response_body", string(stringReqValidator))
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusUnauthorized)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	customerType := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.CustomerTypeID,
		Description: req.Description,
		Remark:      req.Remark,
		OrderItem:   nil,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	createdCustomerType, err := c.service.CreateCustomerType(ctx.Context(), &customerType, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: createdCustomerType.ID}
	getCustomerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCustomerType}, paginationMeta, "Customer type created successfully", http.StatusCreated, nil, nil)
}
func (c *CustomerTypeController) GetCustomerTypeByID(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-GetCustomerTypeByID")
	defer func() {
		// If no error, delete span
		if ctx.Response().StatusCode() >= 400 {
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetCustomerTypeByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: req.ID}
	customerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		responseJSON, _ := json.Marshal(response)
		parentSpan.LogKV("response_body", string(responseJSON))
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	customerTypeArray := []interface{}{customerType}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, customerTypeArray, paginationMeta, "Customer type fetched successfully", http.StatusOK, nil, nil)
}

// update customerType
func (c *CustomerTypeController) UpdateCustomerType(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-UpdateCustomerType")
	defer func() {
		// If no error, delete span
		if ctx.Response().StatusCode() >= 400 {
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateCustomerTypeRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewCustomerTypeUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		stringReqValidator, _ := json.Marshal(reqValidator)
		parentSpan.LogKV("response_body", string(stringReqValidator))
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	customerType := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.CustomerTypeID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	updatedCustomerType, err := c.service.UpdateCustomerType(ctx.Context(), &customerType, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		if err.Error() == "customerType name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Customer type already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetCustomerTypeParams{ID: updatedCustomerType.ID}
	getCustomerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCustomerType}, paginationMeta, "Customer type updated successfully", http.StatusOK, nil, nil)
}

// delete customerType
func (c *CustomerTypeController) DeleteCustomerType(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-DeleteCustomerType")
	var req dtos.DeleteCustomerTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: req.ID}
	// GET customerType by ID
	_, err := c.service.GetCustomerTypeByID(ctx.Context(), params, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteCustomerType(ctx.Context(), params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Customer type deleted successfully", http.StatusOK, nil, nil)
}

// restore customerType
func (c *CustomerTypeController) RestoreCustomerType(ctx *fiber.Ctx) error {
	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-RestoreCustomerType")
	var req dtos.DeleteCustomerTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetCustomerTypeParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET customerType by ID
	_, err := c.service.GetCustomerTypeByID(ctx.Context(), params, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreCustomerType(ctx.Context(), params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Customer type restored successfully", http.StatusOK, nil, nil)
}

func (c *CustomerTypeController) ExcelGetCustomerTypes(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-ExcelGetCustomerTypes")
	defer parentSpan.Finish()
	customerTypes, err := c.service.ExcelGetCustomerTypes(ctx.Context(), filters, parentSpan)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(customerTypes)
}

func (c *CustomerTypeController) CsvGetCustomerTypes(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	parentSpan := utils.StartSpanFromController(ctx, c.tracer, "CustomerTypeController-CsvGetCustomerTypes")
	defer parentSpan.Finish()
	customerTypes, err := c.service.CsvGetCustomerTypes(ctx.Context(), filters, parentSpan)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(customerTypes)
}
