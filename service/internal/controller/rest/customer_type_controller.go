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
)

type CustomerTypeController struct {
	service *service.CustomerTypeService
	repo    *repository.CustomerTypeRepository
}

// func NewCustomerTypeController(service *service.CustomerTypeService) *CustomerTypeController {
func NewCustomerTypeController(service *service.CustomerTypeService, repo *repository.CustomerTypeRepository) *CustomerTypeController {
	return &CustomerTypeController{service: service, repo: repo}
}

func (c *CustomerTypeController) GetCustomerTypes(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	customerTypes, total, err := c.service.GetCustomerTypes(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, customerTypes, paginationMeta, "Customer type fetched successfully", http.StatusOK, nil, nil)
}

func (c *CustomerTypeController) CreateCustomerType(ctx *fiber.Ctx) error {
	var req dtos.CreateCustomerTypeRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewCustomerTypeStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	customerType := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.CustomerTypeID,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	createdCustomerType, err := c.service.CreateCustomerType(ctx.Context(), &customerType, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: createdCustomerType.ID}
	getCustomerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCustomerType}, paginationMeta, "Customer type created successfully", http.StatusCreated, nil, nil)
}
func (c *CustomerTypeController) GetCustomerTypeByID(ctx *fiber.Ctx) error {
	var req dtos.GetCustomerTypeByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: req.ID}
	customerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	customerTypeArray := []interface{}{customerType}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, customerTypeArray, paginationMeta, "Customer type fetched successfully", http.StatusOK, nil, nil)
}

// update customerType
func (c *CustomerTypeController) UpdateCustomerType(ctx *fiber.Ctx) error {
	var req dtos.UpdateCustomerTypeRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewCustomerTypeUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// // Fetch the existing customerType to get the current data
	// existingCustomerType, err := c.service.GetCustomerTypeByID(ctx.Context(), req.ID)
	// if err != nil {
	// 	return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	// }

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
		// CreatedByID: &existingCustomerType.CreatedByID,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedCustomerType, err := c.service.UpdateCustomerType(ctx.Context(), &customerType, tx)

	if err != nil {
		tx.Rollback()
		if err.Error() == "customerType name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Customer type already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: updatedCustomerType.ID}
	getCustomerType, err := c.service.GetCustomerTypeByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCustomerType}, paginationMeta, "Customer type updated successfully", http.StatusOK, nil, nil)
}

// delete customerType
func (c *CustomerTypeController) DeleteCustomerType(ctx *fiber.Ctx) error {
	var req dtos.DeleteCustomerTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCustomerTypeParams{ID: req.ID}
	// GET customerType by ID
	_, err := c.service.GetCustomerTypeByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	err = c.service.DeleteCustomerType(ctx.Context(), params, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return utils.GetResponse(ctx, nil, nil, "Customer type deleted successfully", http.StatusOK, nil, nil)
}

// restore customerType
func (c *CustomerTypeController) RestoreCustomerType(ctx *fiber.Ctx) error {
	var req dtos.DeleteCustomerTypeRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore customer types", http.StatusInternalServerError, err.Error(), nil)
	}

	isDeleted := 1
	params := &dtos.GetCustomerTypeParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET customerType by ID
	_, err := c.service.GetCustomerTypeByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Customer type not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreCustomerType(ctx.Context(), params, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Customer type", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Customer type restored successfully", http.StatusOK, nil, nil)
}

func (c *CustomerTypeController) ExcelGetCustomerTypes(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	customerTypes, err := c.service.ExcelGetCustomerTypes(ctx.Context(), filters)
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

	customerTypes, err := c.service.CsvGetCustomerTypes(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(customerTypes)
}
