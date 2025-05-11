package rest

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
	// "github.com/opentracing/opentracing-go/ext"
)

type InventoryController struct {
	service  *service.InventoryService
	repo     *repository.InventoryRepository
	tracer   opentracing.Tracer
	rabbitmq *config.RabbitMQ
}

func NewInventoryController(service *service.InventoryService, repo *repository.InventoryRepository, rabbitmq *config.RabbitMQ, tracer opentracing.Tracer) *InventoryController {
	return &InventoryController{service: service, repo: repo, rabbitmq: rabbitmq, tracer: tracer}
}

func (c *InventoryController) GetInventories(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetInventories", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	log.Println("filters2", ctx.Locals("filters"))
	filters, ok := ctx.Locals("filters").(map[string]string)

	log.Println("filters", filters)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetInventories: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	inventories, total, err := c.service.GetInventories(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	inventoryIDs := make([]uint, 0)
	for _, inventory := range inventories {
		inventoryIDs = append(inventoryIDs, uint(inventory.ID))
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, inventories, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) GetInventoryByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetInventoryByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetInventoryByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "inventory not found", http.StatusBadRequest, err.Error(), nil)
	}

	ID, _ := utils.ParseInterfaceToUint(req.ID)

	if ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "inventory not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetInventoryParams{ID: ID}
	inventory, err := c.service.GetInventoryByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	createdInventoryIDs := make([]uint, 0)
	createdInventoryIDs = append(createdInventoryIDs, inventory.ID)

	// get soDts
	soDts, err := c.service.GetInvDtsByInventoryIDs(ctx, tx, createdInventoryIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	filters := ctx.Locals("filters").(map[string]string)

	inventory.InvDts = soDts

	inventoryArray := []interface{}{inventory}

	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, inventoryArray, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) CreateInventory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-CreateInventory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.FormInventoryRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewInventoryStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create Inventory", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	// Lock the rows for update
	tx, err := c.service.LockCreateInventoryTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	// create header inventory
	createdInventory, tx, err := c.service.CreateInventory(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Inventory", http.StatusInternalServerError)
	}

	params := &dtos.GetInventoryParams{ID: createdInventory.ID}
	getInventory, err := c.service.GetInventoryByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order", http.StatusInternalServerError)
	}

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInventory}, paginationMeta, "Inventory created successfully", http.StatusCreated, nil, nil)
}

// update inventory
func (c *InventoryController) UpdateInventory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-UpdateInventory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.FormInventoryRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewInventoryUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	// Lock the rows for update
	tx, err := c.service.LockInventoryTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedInventory, err := c.service.UpdateInventory(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetInventoryParams{ID: updatedInventory.ID}
	getInventory, err := c.service.GetInventoryByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusNotFound, err.Error(), nil)
	}

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInventory}, paginationMeta, "Inventory updated successfully", http.StatusOK, nil, nil)
}

// delete inventory
func (c *InventoryController) DeleteInventory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-DeleteInventory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInventoryRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusBadRequest, "ID is required", nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))

	// Transaction handling
	tx := c.repo.BeginTransaction()

	params := &dtos.GetInventoryParams{ID: req.ID}
	// GET inventory by ID
	inventory, err := c.service.GetInventoryByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusNotFound, err.Error(), nil)
	}

	// DELETE soDts by Inventory ID
	err = c.service.DeleteInvDtsByInventoryID(ctx, params, tx, inventory, userID, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	err = c.service.DeleteInventory(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Inventory deleted successfully", http.StatusOK, nil, nil)
}

// restore inventory
func (c *InventoryController) RestoreInventory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-RestoreInventory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInventoryRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetInventoryParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET inventory by ID
	_, err := c.service.GetInventoryByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Inventory not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreInventory(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Inventory", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Inventory restored successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) ExcelGetInventories(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-ExcelGetInventories", opentracing.ChildOf(apiSpan.Context()))
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

	inventories, err := c.service.ExcelGetInventories(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(inventories)
}

func (c *InventoryController) CsvGetInventories(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetInventories", opentracing.ChildOf(apiSpan.Context()))
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

	inventories, err := c.service.CsvGetInventories(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(inventories)
}

func (c *InventoryController) GetRefIndexSoDts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetRefIndexSoDts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetRefIndexSoDts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quoDts, total, err := c.service.GetRefIndexSoDts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quoDts, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) GetRefIndexPoDts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetRefIndexPoDts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetRefIndexPoDts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quoDts, total, err := c.service.GetRefIndexPoDts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quoDts, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) GetRefIndexInvDts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetRefIndexInvDts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetRefIndexInvDts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quoDts, total, err := c.service.GetRefIndexInvDts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quoDts, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) GetStocks(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetStocks", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetStocks: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	stocks, total, err := c.service.GetStocks(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, stocks, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) GetStockClosings(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetStockClosings", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetStockClosing: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	stocks, total, err := c.service.GetStockClosings(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, stocks, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}

func (c *InventoryController) CreateStockClosings(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-CreateStockClosings", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.FormClosingStockStoreRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewClosingStockStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create Inventory", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	// create background sync stock closing
	c.service.BackgroundSyncCreateStockClosingsByRangeDate(ctx, tx, req, userID, branchID, parentSpan)

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{}, paginationMeta, "Stock Closing processed successfully", http.StatusCreated, nil, nil)
}

func (c *InventoryController) GetInventoriesStatus(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InventoryController-GetInventoriesStatus", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	log.Println("filters2", ctx.Locals("filters"))
	filters, ok := ctx.Locals("filters").(map[string]string)

	log.Println("filters", filters)

	if !ok {
		apiSpan.LogKV("response_body", string("InventoryController-GetInventoriesStatus: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	inventories, total, err := c.service.GetInventoriesStatus(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Inventory", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, inventories, paginationMeta, "Inventory fetched successfully", http.StatusOK, nil, nil)
}
