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
	"github.com/opentracing/opentracing-go"
)

type PaymentTermController struct {
	service *service.PaymentTermService
	repo    *repository.PaymentTermRepository
	tracer  opentracing.Tracer
}

func NewPaymentTermController(service *service.PaymentTermService, repo *repository.PaymentTermRepository, tracer opentracing.Tracer) *PaymentTermController {
	return &PaymentTermController{service: service, repo: repo, tracer: tracer}
}

func (c *PaymentTermController) GetPaymentTerms(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-GetPaymentTerms", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("PaymentTermController-GetPaymentTerms: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	paymentTerms, total, err := c.service.GetPaymentTerms(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, paymentTerms, paginationMeta, "Payment terms fetched successfully", http.StatusOK, nil, nil)
}

func (c *PaymentTermController) GetPaymentTermByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-GetPaymentTermByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetPaymentTermByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetPaymentTermParams{ID: req.ID}
	term, err := c.service.GetPaymentTermByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusNotFound, err.Error(), nil)
	}

	termArray := []interface{}{term}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, termArray, paginationMeta, "Payment term fetched successfully", http.StatusOK, nil, nil)
}

func (c *PaymentTermController) CreatePaymentTerm(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-CreatePaymentTerm", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreatePaymentTermRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewPaymentTermStoreRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	term := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.PaymentTermID,
		Description: req.Description,
		Remark:      req.Remark,
		OrderItem:   nil,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	createdTerm, err := c.service.CreatePaymentTerm(ctx, &term, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create payment term", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetPaymentTermParams{ID: createdTerm.ID}
	getTerm, err := c.service.GetPaymentTermByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getTerm}, paginationMeta, "Payment term created successfully", http.StatusCreated, nil, nil)
}

func (c *PaymentTermController) UpdatePaymentTerm(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-UpdatePaymentTerm", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdatePaymentTermRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	reqValidator, isValid := form_requests.NewPaymentTermUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	term := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.PaymentTermID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	updatedTerm, err := c.service.UpdatePaymentTerm(ctx, &term, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "payment term name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Payment term already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update payment term", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetPaymentTermParams{ID: updatedTerm.ID}
	getTerm, err := c.service.GetPaymentTermByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getTerm}, paginationMeta, "Payment term updated successfully", http.StatusOK, nil, nil)
}

func (c *PaymentTermController) DeletePaymentTerm(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-DeletePaymentTerm", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeletePaymentTermRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, "ID is required", nil)
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	params := &dtos.GetPaymentTermParams{
		ID:          req.ID,
		DeletedByID: &userID,
	}

	_, err = c.service.GetPaymentTermByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()
	err = c.service.DeletePaymentTerm(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete payment term", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Payment term deleted successfully", http.StatusOK, nil, nil)
}

func (c *PaymentTermController) RestorePaymentTerm(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-RestorePaymentTerm", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeletePaymentTermRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetPaymentTermParams{ID: req.ID, IsDeleted: &isDeleted}

	_, err := c.service.GetPaymentTermByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Payment term not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestorePaymentTerm(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore payment term", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Payment term restored successfully", http.StatusOK, nil, nil)
}

func (c *PaymentTermController) ExcelGetPaymentTerms(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-ExcelGetPaymentTerms", opentracing.ChildOf(apiSpan.Context()))
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

	paymentTerms, err := c.service.ExcelGetPaymentTerms(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(paymentTerms)
}

func (c *PaymentTermController) CsvGetPaymentTerms(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("PaymentTermController-CsvGetPaymentTerms", opentracing.ChildOf(apiSpan.Context()))
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

	paymentTerms, err := c.service.CsvGetPaymentTerms(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(paymentTerms)
}
