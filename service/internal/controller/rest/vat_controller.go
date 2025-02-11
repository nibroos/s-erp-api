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

type VatController struct {
	service *service.VatService
	repo    *repository.VatRepository
}

// func NewVatController(service *service.VatService) *VatController {
func NewVatController(service *service.VatService, repo *repository.VatRepository) *VatController {
	return &VatController{service: service, repo: repo}
}

func (c *VatController) GetVats(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	vats, total, err := c.service.GetVats(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, vats, paginationMeta, "Vat fetched successfully", http.StatusOK, nil, nil)
}

func (c *VatController) CreateVat(ctx *fiber.Ctx) error {
	var req dtos.CreateVatRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewVatStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	vat := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.VatID,
		Description: req.Description,
		Remark:      req.Remark,
		Num:         req.Num,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create vat", http.StatusInternalServerError, err.Error(), nil)
	}

	createdVat, err := c.service.CreateVat(ctx.Context(), &vat, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to create vat", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create vat", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetVatParams{ID: createdVat.ID}
	getVat, err := c.service.GetVatByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getVat}, paginationMeta, "Vat created successfully", http.StatusCreated, nil, nil)
}
func (c *VatController) GetVatByID(ctx *fiber.Ctx) error {
	var req dtos.GetVatByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatParams{ID: req.ID}
	vat, err := c.service.GetVatByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	vatArray := []interface{}{vat}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, vatArray, paginationMeta, "Vat fetched successfully", http.StatusOK, nil, nil)
}

// update vat
func (c *VatController) UpdateVat(ctx *fiber.Ctx) error {
	var req dtos.UpdateVatRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewVatUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// // Fetch the existing vat to get the current data
	// existingVat, err := c.service.GetVatByID(ctx.Context(), req.ID)
	// if err != nil {
	// 	return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	// }

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	vat := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.VatID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Num:         req.Num,
		Status:      req.Status,
		// CreatedByID: &existingVat.CreatedByID,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create vat", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedVat, err := c.service.UpdateVat(ctx.Context(), &vat, tx)

	if err != nil {
		tx.Rollback()
		if err.Error() == "vat name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Vat already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetVatParams{ID: updatedVat.ID}
	getVat, err := c.service.GetVatByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getVat}, paginationMeta, "Vat updated successfully", http.StatusOK, nil, nil)
}

// delete vat
func (c *VatController) DeleteVat(ctx *fiber.Ctx) error {
	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatParams{ID: req.ID}
	// GET vat by ID
	_, err := c.service.GetVatByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return err
	}

	err = c.service.DeleteVat(ctx.Context(), params, tx)

	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return utils.GetResponse(ctx, nil, nil, "Vat deleted successfully", http.StatusOK, nil, nil)
}

// restore vat
func (c *VatController) RestoreVat(ctx *fiber.Ctx) error {
	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore vat", http.StatusInternalServerError, err.Error(), nil)
	}

	isDeleted := 1
	params := &dtos.GetVatParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET vat by ID
	_, err := c.service.GetVatByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreVat(ctx.Context(), params, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Vat restored successfully", http.StatusOK, nil, nil)
}

func (c *VatController) ExcelGetVats(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	vats, err := c.service.ExcelGetVats(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}

func (c *VatController) CsvGetVats(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	vats, err := c.service.CsvGetVats(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}
