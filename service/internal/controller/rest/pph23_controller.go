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

type Pph23Controller struct {
	service *service.Pph23Service
	repo    *repository.Pph23Repository
}

// func NewPph23Controller(service *service.Pph23Service) *Pph23Controller {
func NewPph23Controller(service *service.Pph23Service, repo *repository.Pph23Repository) *Pph23Controller {
	return &Pph23Controller{service: service, repo: repo}
}

func (c *Pph23Controller) GetPph23s(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	pph23s, total, err := c.service.GetPph23s(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, pph23s, paginationMeta, "Pph23 fetched successfully", http.StatusOK, nil, nil)
}

func (c *Pph23Controller) CreatePph23(ctx *fiber.Ctx) error {
	var req dtos.CreatePph23Request

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewPph23StoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	pph23 := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.Pph23ID,
		Description: req.Description,
		Remark:      req.Remark,
		Num:         req.Num,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	createdPph23, err := c.service.CreatePph23(ctx.Context(), &pph23, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to create pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetPph23Params{ID: createdPph23.ID}
	getPph23, err := c.service.GetPph23ByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getPph23}, paginationMeta, "Pph23 created successfully", http.StatusCreated, nil, nil)
}
func (c *Pph23Controller) GetPph23ByID(ctx *fiber.Ctx) error {
	var req dtos.GetPph23ByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetPph23Params{ID: req.ID}
	pph23, err := c.service.GetPph23ByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	}

	pph23Array := []interface{}{pph23}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, pph23Array, paginationMeta, "Pph23 fetched successfully", http.StatusOK, nil, nil)
}

// update pph23
func (c *Pph23Controller) UpdatePph23(ctx *fiber.Ctx) error {
	var req dtos.UpdatePph23Request

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewPph23UpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// // Fetch the existing pph23 to get the current data
	// existingPph23, err := c.service.GetPph23ByID(ctx.Context(), req.ID)
	// if err != nil {
	// 	return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	// }

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	pph23 := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.Pph23ID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Num:         req.Num,
		Status:      req.Status,
		// CreatedByID: &existingPph23.CreatedByID,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedPph23, err := c.service.UpdatePph23(ctx.Context(), &pph23, tx)

	if err != nil {
		tx.Rollback()
		if err.Error() == "pph23 name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Pph23 already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update Pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetPph23Params{ID: updatedPph23.ID}
	getPph23, err := c.service.GetPph23ByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getPph23}, paginationMeta, "Pph23 updated successfully", http.StatusOK, nil, nil)
}

// delete pph23
func (c *Pph23Controller) DeletePph23(ctx *fiber.Ctx) error {
	var req dtos.DeletePph23Request

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetPph23Params{ID: req.ID}
	// GET pph23 by ID
	_, err := c.service.GetPph23ByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	err = c.service.DeletePph23(ctx.Context(), params, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return utils.GetResponse(ctx, nil, nil, "Pph23 deleted successfully", http.StatusOK, nil, nil)
}

// restore pph23
func (c *Pph23Controller) RestorePph23(ctx *fiber.Ctx) error {
	var req dtos.DeletePph23Request

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	isDeleted := 1
	params := &dtos.GetPph23Params{ID: req.ID, IsDeleted: &isDeleted}
	// GET pph23 by ID
	_, err := c.service.GetPph23ByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Pph23 not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestorePph23(ctx.Context(), params, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Pph23", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Pph23 restored successfully", http.StatusOK, nil, nil)
}

func (c *Pph23Controller) ExcelGetPph23s(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	pph23s, err := c.service.ExcelGetPph23s(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(pph23s)
}

func (c *Pph23Controller) CsvGetPph23s(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	pph23s, err := c.service.CsvGetPph23s(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(pph23s)
}
