package rest

import (
	"log"
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

type MsItemController struct {
	service *service.MsItemService
	repo    *repository.MsItemRepository
	tracer  opentracing.Tracer
}

func NewMsItemController(service *service.MsItemService, repo *repository.MsItemRepository, tracer opentracing.Tracer) *MsItemController {
	return &MsItemController{service: service, repo: repo, tracer: tracer}
}

func (c *MsItemController) GetMsItems(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-GetMsItems", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("MsItemController-GetMsItems: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	msItems, total, err := c.service.GetMsItems(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, msItems, paginationMeta, "Master item fetched successfully", http.StatusOK, nil, nil)
}

func (c *MsItemController) CreateMsItem(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-CreateMsItem", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateMsItemRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewMsItemStoreRequest().Validate(&req, ctx)
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

	msItem := models.MsItem{
		Code:           req.Code,
		Name:           req.Name,
		ItemSubGroupID: &req.ItemSubGroupID,
		// ItemUnitID:     &req.ItemUnitID,
		Specification: req.Specification,
		Description:   req.Description,
		TpbCode:       req.TpbCode,
		MinimumStock:  req.MinimumStock,
		IsAllBranch:   req.IsAllBranch,
		Status:        req.Status,
		CreatedByID:   &userID,
	}

	tx := c.repo.BeginTransaction()
	createdMsItem, err := c.service.CreateMsItem(ctx, &msItem, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	// bulk create item units
	itemUnits := make([]models.ItemUnit, 0)
	for _, unit := range req.Units {
		itemUnit := models.ItemUnit{
			MsItemID:    createdMsItem.ID,
			UnitID:      unit.UnitID,
			Conversion:  &unit.Conversion,
			PriceSell:   &unit.PriceSell,
			PriceBuy:    &unit.PriceBuy,
			Status:      1,
			CreatedByID: &userID,
		}
		itemUnits = append(itemUnits, itemUnit)
	}

	err = c.service.CreateItemUnits(ctx, itemUnits, createdMsItem.ID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	// get selected item unit id by unit id
	paramsItemUnit := &dtos.GetMsItemItemUnitParams{MsItemID: createdMsItem.ID, UnitID: req.ItemUnitID}
	selectedItemUnit, err := c.service.GetItemUnitIDBySelectedItemID(ctx, paramsItemUnit, parentSpan)
	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	log.Println("selectedItemUnit:", selectedItemUnit)

	// update ms item with selected item unit id
	msItem.ItemUnitID = &selectedItemUnit.ID
	_, err = c.service.UpdateMsItem(ctx, &msItem, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetMsItemParams{ID: createdMsItem.ID}
	getMsItem, err := c.service.GetMsItemByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getMsItem}, paginationMeta, "Master item created successfully", http.StatusCreated, nil, nil)
}
func (c *MsItemController) GetMsItemByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-GetMsItemByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetMsItemByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetMsItemParams{ID: req.ID}
	msItem, err := c.service.GetMsItemByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusNotFound, err.Error(), nil)
	}

	msItemArray := []interface{}{msItem}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, msItemArray, paginationMeta, "Master item fetched successfully", http.StatusOK, nil, nil)
}

// update msItem
func (c *MsItemController) UpdateMsItem(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-UpdateMsItem", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateMsItemRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewMsItemUpdateRequest().Validate(&req, ctx)
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

	msItem := models.MsItem{
		ID:             req.ID,
		Code:           req.Code,
		Name:           req.Name,
		ItemSubGroupID: &req.ItemSubGroupID,
		ItemUnitID:     &req.ItemUnitID,
		Specification:  req.Specification,
		Description:    req.Description,
		TpbCode:        req.TpbCode,
		MinimumStock:   req.MinimumStock,
		IsAllBranch:    req.IsAllBranch,
		Status:         req.Status,
		UpdatedByID:    &userID,
	}

	tx := c.repo.BeginTransaction()

	updatedMsItem, err := c.service.UpdateMsItem(ctx, &msItem, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "msItem name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Master item already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Master item", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetMsItemParams{ID: updatedMsItem.ID}
	getMsItem, err := c.service.GetMsItemByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getMsItem}, paginationMeta, "Master item updated successfully", http.StatusOK, nil, nil)
}

// delete msItem
func (c *MsItemController) DeleteMsItem(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-DeleteMsItem", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteMsItemRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetMsItemParams{ID: req.ID}
	// GET msItem by ID
	_, err := c.service.GetMsItemByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteMsItem(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Master item", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master item deleted successfully", http.StatusOK, nil, nil)
}

// restore msItem
func (c *MsItemController) RestoreMsItem(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-RestoreMsItem", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteMsItemRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetMsItemParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET msItem by ID
	_, err := c.service.GetMsItemByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master item not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreMsItem(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Master item", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master item restored successfully", http.StatusOK, nil, nil)
}

func (c *MsItemController) ExcelGetMsItems(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("MsItemController-ExcelGetMsItems", opentracing.ChildOf(apiSpan.Context()))
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

	msItems, err := c.service.ExcelGetMsItems(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(msItems)
}

func (c *MsItemController) CsvGetMsItems(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetMsItems", opentracing.ChildOf(apiSpan.Context()))
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

	msItems, err := c.service.CsvGetMsItems(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(msItems)
}
