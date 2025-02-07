package rest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
)

type ItemSubGroupController struct {
	service *service.ItemSubGroupService
}

func NewItemSubGroupController(service *service.ItemSubGroupService) *ItemSubGroupController {
	return &ItemSubGroupController{service: service}
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
		CreatedByID: &userID,
	}

	createdItemSubGroup, err := c.service.CreateItemSubGroup(ctx.Context(), &itemSubGroup)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	getItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), createdItemSubGroup.ID)
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

	itemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), req.ID)
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

	// Fetch the existing itemSubGroup to get the current data
	existingItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), req.ID)
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
		CreatedByID: &existingItemSubGroup.CreatedByID,
		UpdatedByID: &userID,
	}

	updatedItemSubGroup, err := c.service.UpdateItemSubGroup(ctx.Context(), &itemSubGroup)
	if err != nil {
		if err.Error() == "itemSubGroup name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Item sub group already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	getItemSubGroup, err := c.service.GetItemSubGroupByID(ctx.Context(), updatedItemSubGroup.ID)
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

	// GET itemSubGroup by ID
	_, err := c.service.GetItemSubGroupByID(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.DeleteItemSubGroup(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to delete item sub group", http.StatusInternalServerError, err.Error(), nil)
	}

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

	// GET itemSubGroup by ID
	_, err := c.service.GetItemSubGroupByID(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item sub group not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreItemSubGroup(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore sub group", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Item sub group restored successfully", http.StatusOK, nil, nil)
}
