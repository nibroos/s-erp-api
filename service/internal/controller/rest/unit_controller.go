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

type UnitController struct {
	service *service.UnitService
	repo    *repository.UnitRepository
}

// func NewUnitController(service *service.UnitService) *UnitController {
func NewUnitController(service *service.UnitService, repo *repository.UnitRepository) *UnitController {
	return &UnitController{service: service, repo: repo}
}

func (c *UnitController) GetUnits(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	units, total, err := c.service.GetUnits(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, units, paginationMeta, "Unit fetched successfully", http.StatusOK, nil, nil)
}

func (c *UnitController) CreateUnit(ctx *fiber.Ctx) error {
	var req dtos.CreateUnitRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewUnitStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	unit := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.UnitID,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create unit", http.StatusInternalServerError, err.Error(), nil)
	}

	createdUnit, err := c.service.CreateUnit(ctx.Context(), &unit, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to create unit", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create unit", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetUnitParams{ID: createdUnit.ID}
	getUnit, err := c.service.GetUnitByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getUnit}, paginationMeta, "Unit created successfully", http.StatusCreated, nil, nil)
}
func (c *UnitController) GetUnitByID(ctx *fiber.Ctx) error {
	var req dtos.GetUnitByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetUnitParams{ID: req.ID}
	unit, err := c.service.GetUnitByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	}

	unitArray := []interface{}{unit}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, unitArray, paginationMeta, "Unit fetched successfully", http.StatusOK, nil, nil)
}

// update unit
func (c *UnitController) UpdateUnit(ctx *fiber.Ctx) error {
	var req dtos.UpdateUnitRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewUnitUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// // Fetch the existing unit to get the current data
	// existingUnit, err := c.service.GetUnitByID(ctx.Context(), req.ID)
	// if err != nil {
	// 	return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	// }

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	unit := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.UnitID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		// CreatedByID: &existingUnit.CreatedByID,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create unit", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedUnit, err := c.service.UpdateUnit(ctx.Context(), &unit, tx)

	if err != nil {
		tx.Rollback()
		if err.Error() == "unit name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Unit already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Unit", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update Unit", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetUnitParams{ID: updatedUnit.ID}
	getUnit, err := c.service.GetUnitByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getUnit}, paginationMeta, "Unit updated successfully", http.StatusOK, nil, nil)
}

// delete unit
func (c *UnitController) DeleteUnit(ctx *fiber.Ctx) error {
	var req dtos.DeleteUnitRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetUnitParams{ID: req.ID}
	// GET unit by ID
	_, err := c.service.GetUnitByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	err = c.service.DeleteUnit(ctx.Context(), params, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Unit", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return utils.GetResponse(ctx, nil, nil, "Unit deleted successfully", http.StatusOK, nil, nil)
}

// restore unit
func (c *UnitController) RestoreUnit(ctx *fiber.Ctx) error {
	var req dtos.DeleteUnitRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore unit", http.StatusInternalServerError, err.Error(), nil)
	}

	isDeleted := 1
	params := &dtos.GetUnitParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET unit by ID
	_, err := c.service.GetUnitByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unit not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreUnit(ctx.Context(), params, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Unit", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Unit", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Unit restored successfully", http.StatusOK, nil, nil)
}

func (c *UnitController) ExcelGetUnits(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	units, err := c.service.ExcelGetUnits(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(units)
}

func (c *UnitController) CsvGetUnits(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	units, err := c.service.CsvGetUnits(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(units)
}
