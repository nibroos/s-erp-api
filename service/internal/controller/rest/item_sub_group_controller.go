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

type ItemSubGroupController struct {
	service *service.ItemSubGroupService
	repo    *repository.ItemSubGroupRepository
}

// func NewItemSubGroupController(service *service.ItemSubGroupService) *ItemSubGroupController {
func NewItemSubGroupController(service *service.ItemSubGroupService, repo *repository.ItemSubGroupRepository) *ItemSubGroupController {
	return &ItemSubGroupController{service: service, repo: repo}
}

func (c *ItemSubGroupController) GetItemSubGroups(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	itemSubGroups, total, err := c.service.GetItemSubGroups(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, itemSubGroups, paginationMeta, "Item sub groups fetched successfully", http.StatusOK, nil, nil)
}
func (c *ItemSubGroupController) CreateItemSubGroup(ctx *fiber.Ctx) error {
	var req dtos.CreateItemSubGroupRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewItemSubGroupStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	itemSubGroup := models.MixValue{
		GroupID:     utils.ItemSubGroupID,
		ParentID:    &req.ItemGroupID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		OptionsJSON: "{}",
		CreatedByID: &userID,
	}

	tx := c.repo.BeginTransaction()
	createdItemSubGroup, err := c.service.CreateItemSubGroup(ctx.Context(), &itemSubGroup, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to create item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetItemSubGroupParams{ID: createdItemSubGroup.ID}
	getItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getItemSubGroup}, paginationMeta, "Item sub group created successfully", http.StatusCreated, nil, nil)
}
func (c *ItemSubGroupController) GetItemSubGroupByID(ctx *fiber.Ctx) error {
	var req dtos.GetItemSubGroupByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetItemSubGroupParams{ID: req.ID}
	itemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	itemSubGroupArray := []interface{}{itemSubGroup}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, itemSubGroupArray, paginationMeta, "Item sub group fetched successfully", http.StatusOK, nil, nil)
}

// update itemSubGroup
func (c *ItemSubGroupController) UpdateItemSubGroup(ctx *fiber.Ctx) error {
	var req dtos.UpdateItemSubGroupRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewItemSubGroupUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	params := &dtos.GetItemSubGroupParams{ID: req.ID}
	// Fetch the existing itemSubGroup to get the current data
	existingItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	itemSubGroup := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.ItemSubGroupID,
		ParentID:    &req.ItemGroupID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		OptionsJSON: "{}",
		CreatedByID: &existingItemSubGroup.CreatedByID,
		UpdatedByID: &userID,
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedItemSubGroup, err := c.service.UpdateItemSubGroup(ctx.Context(), &itemSubGroup, tx)
	if err != nil {
		tx.Rollback()
		if err.Error() == "itemSubGroup name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Item sub group already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	params = &dtos.GetItemSubGroupParams{ID: updatedItemSubGroup.ID}
	getItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getItemSubGroup}, paginationMeta, "Item sub group updated successfully", http.StatusOK, nil, nil)
}

// delete itemSubGroup
func (c *ItemSubGroupController) DeleteItemSubGroup(ctx *fiber.Ctx) error {
	var req dtos.DeleteItemSubGroupRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetItemSubGroupParams{ID: req.ID}
	// GET itemSubGroup by ID
	_, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()
	if err := tx.Error; err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to delete item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	err = c.service.DeleteItemSubGroup(ctx.Context(), req.ID, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to delete item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Item sub group deleted successfully", http.StatusOK, nil, nil)
}

// restore itemSubGroup
func (c *ItemSubGroupController) RestoreItemSubGroup(ctx *fiber.Ctx) error {
	var req dtos.DeleteItemSubGroupRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusBadRequest, "ID is required", nil)
	}

	isDeleted := 1
	params := &dtos.GetItemSubGroupParams{ID: req.ID, IsDeleted: &isDeleted}
	_, err := c.service.GetItemSubGroupByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()
	err = c.service.RestoreItemSubGroup(ctx.Context(), params.ID, tx)
	if err != nil {
		tx.Rollback()
		return utils.GetResponse(ctx, nil, nil, "Failed to restore sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Item sub group restored successfully", http.StatusOK, nil, nil)
}

func (c *ItemSubGroupController) ExcelGetItemSubGroups(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	itemSubGroups, err := c.service.ExcelGetItemSubGroups(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(itemSubGroups)
}

func (c *ItemSubGroupController) CsvGetItemSubGroups(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	itemSubGroups, err := c.service.CsvGetItemSubGroups(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(itemSubGroups)
}
