package rest

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
)

type InvoiceMaintenanceController struct {
	service *service.InvoiceMaintenanceService
	repo    *repository.InvoiceMaintenanceRepository
	tracer  opentracing.Tracer
}

func NewInvoiceMaintenanceController(service *service.InvoiceMaintenanceService, repo *repository.InvoiceMaintenanceRepository, tracer opentracing.Tracer) *InvoiceMaintenanceController {
	return &InvoiceMaintenanceController{service: service, repo: repo, tracer: tracer}
}

func (c *InvoiceMaintenanceController) GetInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-GetInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceMaintenanceController-GetInvoiceMaintenances: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceMaintenances, total, err := c.service.GetInvoiceMaintenances(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice maintenances", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, invoiceMaintenances, paginationMeta, "Invoice maintenances fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) GetInvoiceMaintenancesDetails(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-GetInvoiceMaintenancesDetails", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceMaintenanceController-GetInvoiceMaintenancesDetails: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceMaintenances, total, err := c.service.GetInvoiceMaintenancesDetails(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice maintenances", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, invoiceMaintenances, paginationMeta, "Invoice maintenances fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) GetInvoiceMaintenanceByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-GetInvoiceMaintenanceByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetInvoiceMaintenanceByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetInvoiceMaintenanceParams{ID: req.ID}
	invoiceMaintenance, err := c.service.GetInvoiceMaintenanceByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice maintenance", http.StatusInternalServerError)
	}

	invoiceMaintenanceArray := []interface{}{invoiceMaintenance}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, invoiceMaintenanceArray, paginationMeta, "Invoice maintenance fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) CreateInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-CreateInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateInvoiceMaintenanceRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewInvoiceMaintenanceStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create invoice maintenance", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	createdInvoiceMaintenance, tx, err := c.service.CreateInvoiceMaintenance(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create invoice maintenance", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetInvoiceMaintenanceParams{ID: createdInvoiceMaintenance.ID}
	getInvoiceMaintenance, err := c.service.GetInvoiceMaintenanceByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice maintenance", http.StatusInternalServerError)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInvoiceMaintenance}, paginationMeta, "Invoice maintenance created successfully", http.StatusCreated, nil, nil)
}

func (c *InvoiceMaintenanceController) UpdateInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-UpdateInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateInvoiceMaintenanceRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewInvoiceMaintenanceUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	updatedInvoiceMaintenance, err := c.service.UpdateInvoiceMaintenance(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update invoice maintenance", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetInvoiceMaintenanceParams{ID: updatedInvoiceMaintenance.ID}
	getInvoiceMaintenance, err := c.service.GetInvoiceMaintenanceByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInvoiceMaintenance}, paginationMeta, "Invoice maintenance updated successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) DeleteInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-DeleteInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInvoiceMaintenanceRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetInvoiceMaintenanceParams{ID: req.ID}
	_, err := c.service.GetInvoiceMaintenanceByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusNotFound, err.Error(), nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))

	err = c.service.DeleteInvoiceMaintenance(ctx, req.ID, userID, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete invoice maintenance", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Invoice maintenance deleted successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) RestoreInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-RestoreInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInvoiceMaintenanceRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetInvoiceMaintenanceParams{ID: req.ID, IsDeleted: &isDeleted}
	_, err := c.service.GetInvoiceMaintenanceByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice maintenance not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreInvoiceMaintenance(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore invoice maintenance", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Invoice maintenance restored successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) GetRefSalesOrderForInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-GetRefSalesOrderForInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceMaintenanceController-GetRefSalesOrderForInvoiceMaintenance: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	if ctx.Query("specific_ids") != "" {
		filters["specific_ids"] = ctx.Query("specific_ids")
	}

	soDts, total, err := c.service.GetRefSalesOrderForInvoiceMaintenance(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order details", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, soDts, paginationMeta, "Sales order details fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) UpdateSalesOrderStatusForInvoiceMaintenance(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-UpdateSalesOrderStatusForInvoiceMaintenance", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateSalesOrderStatusForInvoiceMaintenanceRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err := c.service.UpdateSalesOrderStatusForInvoiceMaintenance(ctx, req, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update sales order status", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Sales order status updated successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) ApproveInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-ApproveInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.ApproveInvoiceMaintenancesRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	if len(req.IDs) == 0 {
		return utils.GetResponse(ctx, nil, nil, "No invoice maintenance IDs provided", http.StatusBadRequest, "IDs are required", nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))

	tx := c.repo.BeginTransaction()

	err := c.service.ApproveInvoiceMaintenances(ctx, req.IDs, userID, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to approve invoice maintenances", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, fmt.Sprintf("Successfully approved %d invoice maintenances", len(req.IDs)), http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) CancelApproveInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-CancelApproveInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CancelApproveInvoiceMaintenancesRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	if len(req.IDs) == 0 {
		return utils.GetResponse(ctx, nil, nil, "No invoice maintenance IDs provided", http.StatusBadRequest, "IDs are required", nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))

	tx := c.repo.BeginTransaction()

	err := c.service.CancelApproveInvoiceMaintenances(ctx, req.IDs, userID, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to cancel approval of invoice maintenances", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, fmt.Sprintf("Successfully canceled approval for %d invoice maintenances", len(req.IDs)), http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) GetWidgetInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-GetWidgetInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceMaintenanceController-GetWidgetInvoiceMaintenances: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceMaintenances, total, err := c.service.GetWidgetInvoiceMaintenances(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Invoice Maintenance", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, invoiceMaintenances, paginationMeta, "Invoice Maintenance fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) RepeatInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-RepeatInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in RepeatInvoiceMaintenances: %v\n", r)
			debug.PrintStack()
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Internal server error",
			})
		}
	}()

	var req dtos.RepeatInvoiceMaintenanceRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"errors":  err.Error(),
			"message": "Invalid request",
			"status":  http.StatusBadRequest,
		})
	}

	fmt.Printf("RepeatInvoiceMaintenances request: %+v\n", req)

	if len(req.Invoices) == 0 {
		return utils.GetResponse(ctx, nil, nil, "No invoices provided for duplication", http.StatusBadRequest, "Invoices are required", nil)
	}

	for _, invoice := range req.Invoices {
		if invoice.ID == 0 {
			return utils.GetResponse(ctx, nil, nil, "Invalid invoice ID", http.StatusBadRequest, "Invoice ID is required", nil)
		}
	}

	response, err := c.service.RepeatInvoiceMaintenances(ctx, req)
	if err != nil {
		fmt.Printf("Error repeating invoices: %v\n", err)
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to repeat invoice maintenances", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(
		ctx,
		response.Results,
		nil,
		fmt.Sprintf("Successfully duplicated %d invoice maintenances", len(response.Results)),
		http.StatusOK,
		nil,
		nil,
	)
}

func (c *InvoiceMaintenanceController) ExcelGetInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-ExcelGetInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceMaintenances, err := c.service.ExcelGetInvoiceMaintenances(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename=invoice_maintenances.xlsx")

	return ctx.Send(invoiceMaintenances)
}

func (c *InvoiceMaintenanceController) CsvGetInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-CsvGetInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceMaintenances, err := c.service.CsvGetInvoiceMaintenances(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", "attachment; filename=invoice_maintenances.csv")

	return ctx.Send(invoiceMaintenances)
}

func (c *InvoiceMaintenanceController) EmailsApproveInvoiceMaintenances(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-EmailsApproveInvoiceMaintenances", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	if len(req.IDs) == 0 {
		return utils.GetResponse(ctx, nil, nil, "No invoice maintenance IDs provided", http.StatusBadRequest, "IDs are required", nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	err := c.service.PublishBulkSendEmailApproved(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to send email of invoice maintenances", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, fmt.Sprintf("Successfully send email approved for %d invoice maintenances", len(req.IDs)), http.StatusOK, nil, nil)
}

func (c *InvoiceMaintenanceController) Pdf(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceMaintenanceController-Pdf", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.InvoiceMaintenanceDetailNoBomDTO

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	tx := c.repo.BeginTransaction()
	link, err := c.service.Pdf(ctx, req, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	link = utils.MapStringToURL(link)

	return utils.GetResponse(ctx, map[string]string{"link": *link}, nil, "PDF generated successfully", http.StatusOK, nil, nil)
}
