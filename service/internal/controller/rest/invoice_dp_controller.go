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
)

type InvoiceDpController struct {
	service *service.InvoiceDpService
	repo    *repository.InvoiceDpRepository
	tracer  opentracing.Tracer
}

func NewInvoiceDpController(service *service.InvoiceDpService, repo *repository.InvoiceDpRepository, tracer opentracing.Tracer) *InvoiceDpController {
	return &InvoiceDpController{service: service, repo: repo, tracer: tracer}
}

func (c *InvoiceDpController) GetInvoiceDps(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-GetInvoiceDps", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceDpController-GetInvoiceDps: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceDps, total, err := c.service.GetInvoiceDps(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice DP", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, invoiceDps, paginationMeta, "Invoice DP fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) GetInvoiceDpByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-GetInvoiceDpByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetInvoiceDpByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetInvoiceDpParams{ID: req.ID}
	invoiceDp, err := c.service.GetInvoiceDpByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice DP", http.StatusInternalServerError)
	}

	invoiceDpArray := []interface{}{invoiceDp}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, invoiceDpArray, paginationMeta, "Invoice DP fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) CreateInvoiceDp(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-CreateInvoiceDp", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateInvoiceDpRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewInvoiceDpStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create invoice DP", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	tx, err := c.service.LockCreateInvoiceDpTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to create invoice DP", http.StatusInternalServerError, err.Error(), nil)
	}

	createdInvoiceDp, tx, err := c.service.CreateInvoiceDp(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create invoice DP", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetInvoiceDpParams{ID: createdInvoiceDp.ID}
	getInvoiceDp, err := c.service.GetInvoiceDpByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch invoice DP", http.StatusInternalServerError)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInvoiceDp}, paginationMeta, "Invoice DP created successfully", http.StatusCreated, nil, nil)
}

func (c *InvoiceDpController) UpdateInvoiceDp(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-UpdateInvoiceDp", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateInvoiceDpRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewInvoiceDpUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	tx, err := c.service.LockInvoiceDpTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update invoice DP", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedInvoiceDp, err := c.service.UpdateInvoiceDp(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update invoice DP", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetInvoiceDpParams{ID: updatedInvoiceDp.ID}
	getInvoiceDp, err := c.service.GetInvoiceDpByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getInvoiceDp}, paginationMeta, "Invoice DP updated successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) DeleteInvoiceDp(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-DeleteInvoiceDp", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInvoiceDpRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetInvoiceDpParams{ID: req.ID}
	_, err := c.service.GetInvoiceDpByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusNotFound, err.Error(), nil)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))

	err = c.service.DeleteInvoiceDp(ctx, req.ID, userID, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete invoice DP", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Invoice DP deleted successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) RestoreInvoiceDp(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-RestoreInvoiceDp", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteInvoiceDpRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetInvoiceDpParams{ID: req.ID, IsDeleted: &isDeleted}
	_, err := c.service.GetInvoiceDpByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Invoice DP not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreInvoiceDp(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore invoice DP", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Invoice DP restored successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) GetRefSalesOrderDts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-GetRefSalesOrderDts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceDpController-GetRefSalesOrderDts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	if ctx.Query("specific_ids") != "" {
		filters["specific_ids"] = ctx.Query("specific_ids")
	}

	soDts, total, err := c.service.GetRefSalesOrderDts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch sales order details", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, soDts, paginationMeta, "Sales order details fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) GetWidgetInvoiceDps(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-GetWidgetInvoiceDps", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("InvoiceDpController-GetWidgetInvoiceDps: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	invoiceDps, total, err := c.service.GetWidgetInvoiceDps(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Invoice DP", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, invoiceDps, paginationMeta, "Invoice DP fetched successfully", http.StatusOK, nil, nil)
}

func (c *InvoiceDpController) ExcelGetInvoiceDps(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-ExcelGetInvoiceDps", opentracing.ChildOf(apiSpan.Context()))
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

	invoiceDps, err := c.service.ExcelGetInvoiceDps(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename=invoice_dps.xlsx")

	return ctx.Send(invoiceDps)
}

func (c *InvoiceDpController) CsvGetInvoiceDps(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("InvoiceDpController-CsvGetInvoiceDps", opentracing.ChildOf(apiSpan.Context()))
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

	invoiceDps, err := c.service.CsvGetInvoiceDps(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", "attachment; filename=invoice_dps.csv")

	return ctx.Send(invoiceDps)
}
