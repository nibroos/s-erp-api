package rest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
	// "github.com/opentracing/opentracing-go/ext"
)

type SalesOrderController struct {
	service *service.SalesOrderService
	repo    *repository.SalesOrderRepository
	tracer  opentracing.Tracer
}

func NewSalesOrderController(service *service.SalesOrderService, repo *repository.SalesOrderRepository, tracer opentracing.Tracer) *SalesOrderController {
	return &SalesOrderController{service: service, repo: repo, tracer: tracer}
}

func (c *SalesOrderController) GetSalesOrders(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-GetSalesOrders", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("SalesOrderController-GetSalesOrders: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	salesOrders, total, err := c.service.GetSalesOrders(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	salesOrderIDs := make([]uint, 0)
	for _, salesOrder := range salesOrders {
		salesOrderIDs = append(salesOrderIDs, uint(salesOrder.ID))
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, salesOrders, paginationMeta, "Sales Order fetched successfully", http.StatusOK, nil, nil)
}

func (c *SalesOrderController) GetSalesOrderByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-GetSalesOrderByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetSalesOrderByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "sales order not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "sales order not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetSalesOrderParams{ID: req.ID}
	salesOrder, err := c.service.GetSalesOrderByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	salesOrder.Schedule, err = c.service.GetScheduleBySalesOrderID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	salesOrder.Attachments, err = c.service.GetAttachmentsBySalesOrderID(ctx, tx, salesOrder.ID, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	createdSalesOrderIDs := make([]uint, 0)
	createdSalesOrderIDs = append(createdSalesOrderIDs, salesOrder.ID)

	// get soDts
	soDts, err := c.service.GetSoDtsBySalesOrderIDs(ctx, tx, createdSalesOrderIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	filters := ctx.Locals("filters").(map[string]string)
	// get soDtBoms
	soDtBoms, err := c.service.GetSoDtsBomBySalesOrders(ctx, filters, createdSalesOrderIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Sales Order", http.StatusInternalServerError)
	}

	soDts = c.service.MapFilterSoDtBomsToSoDts(ctx, soDtBoms, soDts, parentSpan)

	salesOrder.SoDts = soDts

	salesOrderArray := []interface{}{salesOrder}

	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, salesOrderArray, paginationMeta, "Sales Order fetched successfully", http.StatusOK, nil, nil)
}

func (c *SalesOrderController) CreateSalesOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-CreateSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateSalesOrderRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewSalesOrderStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create Sales Order", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	// Lock the rows for update
	tx, err := c.service.LockCreateSalesOrderTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	// create header salesOrder
	createdSalesOrder, tx, err := c.service.CreateSalesOrder(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Sales Order", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetSalesOrderParams{ID: createdSalesOrder.ID}
	getSalesOrder, err := c.service.GetSalesOrderByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order", http.StatusInternalServerError)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getSalesOrder}, paginationMeta, "Sales Order created successfully", http.StatusCreated, nil, nil)
}

// update salesOrder
func (c *SalesOrderController) UpdateSalesOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-UpdateSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateSalesOrderRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewSalesOrderUpdateRequest().Validate(&req, ctx)
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
	tx, err := c.service.LockSalesOrderTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedSalesOrder, err := c.service.UpdateSalesOrder(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetSalesOrderParams{ID: updatedSalesOrder.ID}
	getSalesOrder, err := c.service.GetSalesOrderByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getSalesOrder}, paginationMeta, "Sales Order updated successfully", http.StatusOK, nil, nil)
}

// delete salesOrder
func (c *SalesOrderController) DeleteSalesOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-DeleteSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteSalesOrderRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusBadRequest, "ID is required", nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()

	params := &dtos.GetSalesOrderParams{ID: req.ID}
	// GET salesOrder by ID
	_, err := c.service.GetSalesOrderByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusNotFound, err.Error(), nil)
	}

	// DELETE soDtBoms by SalesOrder ID
	err = c.service.DeleteSoDtBomsBySalesOrderID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	// DELETE soDts by SalesOrder ID
	err = c.service.DeleteSoDtsBySalesOrderID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	err = c.service.DeleteSalesOrder(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Sales Order deleted successfully", http.StatusOK, nil, nil)
}

// restore salesOrder
func (c *SalesOrderController) RestoreSalesOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-RestoreSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteSalesOrderRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetSalesOrderParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET salesOrder by ID
	_, err := c.service.GetSalesOrderByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Sales Order not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreSalesOrder(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Sales Order", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Sales Order restored successfully", http.StatusOK, nil, nil)
}

func (c *SalesOrderController) ExcelGetSalesOrders(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-ExcelGetSalesOrders", opentracing.ChildOf(apiSpan.Context()))
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

	salesOrders, err := c.service.ExcelGetSalesOrders(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(salesOrders)
}

func (c *SalesOrderController) CsvGetSalesOrders(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetSalesOrders", opentracing.ChildOf(apiSpan.Context()))
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

	salesOrders, err := c.service.CsvGetSalesOrders(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(salesOrders)
}

func (c *SalesOrderController) GetRefIndexQuoDts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-GetRefIndexQuoDts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("SalesOrderController-GetRefIndexQuoDts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quoDts, total, err := c.service.GetRefIndexQuoDts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quoDts, paginationMeta, "Quotation fetched successfully", http.StatusOK, nil, nil)
}

// UpdateScheduleSalesOrder updates the schedule of a sales order
func (c *SalesOrderController) UpdateScheduleSalesOrder(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-UpdateScheduleSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateSalesOrderScheduleRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewScheduleUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	_, err := c.service.UpdateSalesOrderSchedule(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Schedule", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{}, paginationMeta, "Schedule updated successfully", http.StatusOK, nil, nil)
}

// UpdateScheduleSalesOrder updates the schedule of a sales order
func (c *SalesOrderController) UpdateScheduleSalesOrderApp(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("SalesOrderController-UpdateScheduleSalesOrder", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateSalesOrderScheduleAppRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewScheduleUpdateAppRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	err := c.service.UpdateSalesOrderScheduleApp(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update task", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{}, paginationMeta, "Task updated successfully", http.StatusOK, nil, nil)
}
