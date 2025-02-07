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

type ItemGroupController struct {
	service *service.ItemGroupService
}

func NewItemGroupController(service *service.ItemGroupService) *ItemGroupController {
	return &ItemGroupController{service: service}
}

func (c *ItemGroupController) GetItemGroups(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	itemGroups, total, err := c.service.GetItemGroups(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, itemGroups, paginationMeta, "Item group fetched successfully", http.StatusOK, nil, nil)
}
func (c *ItemGroupController) CreateItemGroup(ctx *fiber.Ctx) error {
	var req dtos.CreateItemGroupRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewItemGroupStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	itemGroup := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.ItemGroupID,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	createdItemGroup, err := c.service.CreateItemGroup(ctx.Context(), &itemGroup)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create item group", http.StatusInternalServerError, err.Error(), nil)
	}

	getItemGroup, err := c.service.GetItemGroupByID(ctx.Context(), createdItemGroup.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getItemGroup}, paginationMeta, "Item group created successfully", http.StatusCreated, nil, nil)
}
func (c *ItemGroupController) GetItemGroupByID(ctx *fiber.Ctx) error {
	var req dtos.GetItemGroupByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, "ID is required", nil)
	}

	itemGroup, err := c.service.GetItemGroupByID(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	}

	itemGroupArray := []interface{}{itemGroup}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, itemGroupArray, paginationMeta, "Item group fetched successfully", http.StatusOK, nil, nil)
}

// update itemGroup
func (c *ItemGroupController) UpdateItemGroup(ctx *fiber.Ctx) error {
	var req dtos.UpdateItemGroupRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewItemGroupUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// // Fetch the existing itemGroup to get the current data
	// existingItemGroup, err := c.service.GetItemGroupByID(ctx.Context(), req.ID)
	// if err != nil {
	// 	return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	// }

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	itemGroup := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.ItemGroupID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		// CreatedByID: &existingItemGroup.CreatedByID,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	updatedItemGroup, err := c.service.UpdateItemGroup(ctx.Context(), &itemGroup)
	if err != nil {
		if err.Error() == "itemGroup name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Item group already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Item group", http.StatusInternalServerError, err.Error(), nil)
	}

	getItemGroup, err := c.service.GetItemGroupByID(ctx.Context(), updatedItemGroup.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getItemGroup}, paginationMeta, "Item group updated successfully", http.StatusOK, nil, nil)
}

// delete itemGroup
func (c *ItemGroupController) DeleteItemGroup(ctx *fiber.Ctx) error {
	var req dtos.DeleteItemGroupRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, "ID is required", nil)
	}

	// GET itemGroup by ID
	_, err := c.service.GetItemGroupByID(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.DeleteItemGroup(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Item group", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Item group deleted successfully", http.StatusOK, nil, nil)
}

// restore itemGroup
func (c *ItemGroupController) RestoreItemGroup(ctx *fiber.Ctx) error {
	var req dtos.DeleteItemGroupRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusBadRequest, "ID is required", nil)
	}

	// GET itemGroup by ID
	_, err := c.service.GetItemGroupByID(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Item group not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreItemGroup(ctx.Context(), req.ID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Item group", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Item group restored successfully", http.StatusOK, nil, nil)
}
