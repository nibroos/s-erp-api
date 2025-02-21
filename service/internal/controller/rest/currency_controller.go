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
	// "github.com/opentracing/opentracing-go/ext"
)

type CurrencyController struct {
	service *service.CurrencyService
	repo    *repository.CurrencyRepository
	tracer  opentracing.Tracer
}

func NewCurrencyController(service *service.CurrencyService, repo *repository.CurrencyRepository, tracer opentracing.Tracer) *CurrencyController {
	return &CurrencyController{service: service, repo: repo, tracer: tracer}
}

func (c *CurrencyController) GetCurrencies(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-GetCurrencies", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("CurrencyController-GetCurrencies: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	currencies, total, err := c.service.GetCurrencies(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, currencies, paginationMeta, "Currency fetched successfully", http.StatusOK, nil, nil)
}

func (c *CurrencyController) CreateCurrency(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-CreateCurrency", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateCurrencyRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewCurrencyStoreRequest().Validate(&req, ctx)
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

	currency := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.CurrencyID,
		Description: req.Description,
		Remark:      req.Remark,
		OrderItem:   nil,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	createdCurrency, err := c.service.CreateCurrency(ctx, &currency, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create currencies", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetCurrencyParams{ID: createdCurrency.ID}
	getCurrency, err := c.service.GetCurrencyByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCurrency}, paginationMeta, "Currency created successfully", http.StatusCreated, nil, nil)
}
func (c *CurrencyController) GetCurrencyByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-GetCurrencyByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetCurrencyByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCurrencyParams{ID: req.ID}
	currency, err := c.service.GetCurrencyByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusNotFound, err.Error(), nil)
	}

	currencyArray := []interface{}{currency}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, currencyArray, paginationMeta, "Currency fetched successfully", http.StatusOK, nil, nil)
}

// update currency
func (c *CurrencyController) UpdateCurrency(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-UpdateCurrency", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateCurrencyRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewCurrencyUpdateRequest().Validate(&req, ctx)
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

	currency := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.CurrencyID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	updatedCurrency, err := c.service.UpdateCurrency(ctx, &currency, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "currency name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Currency already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Currency", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetCurrencyParams{ID: updatedCurrency.ID}
	getCurrency, err := c.service.GetCurrencyByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCurrency}, paginationMeta, "Currency updated successfully", http.StatusOK, nil, nil)
}

// delete currency
func (c *CurrencyController) DeleteCurrency(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-DeleteCurrency", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteCurrencyRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCurrencyParams{ID: req.ID}
	// GET currency by ID
	_, err := c.service.GetCurrencyByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteCurrency(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Currency", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Currency deleted successfully", http.StatusOK, nil, nil)
}

// restore currency
func (c *CurrencyController) RestoreCurrency(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-RestoreCurrency", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteCurrencyRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetCurrencyParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET currency by ID
	_, err := c.service.GetCurrencyByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Currency not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreCurrency(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Currency", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Currency restored successfully", http.StatusOK, nil, nil)
}

func (c *CurrencyController) ExcelGetCurrencies(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CurrencyController-ExcelGetCurrencies", opentracing.ChildOf(apiSpan.Context()))
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

	currencies, err := c.service.ExcelGetCurrencies(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(currencies)
}

func (c *CurrencyController) CsvGetCurrencies(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetCurrencies", opentracing.ChildOf(apiSpan.Context()))
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

	currencies, err := c.service.CsvGetCurrencies(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(currencies)
}
