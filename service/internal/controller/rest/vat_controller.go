package rest

import (
	"net/http"
	"time"

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

type VatController struct {
	service *service.VatService
	repo    *repository.VatRepository
	tracer  opentracing.Tracer
}

func NewVatController(service *service.VatService, repo *repository.VatRepository, tracer opentracing.Tracer) *VatController {
	return &VatController{service: service, repo: repo, tracer: tracer}
}

func (c *VatController) GetVats(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-GetVats", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("VatController-GetVats: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	vats, total, err := c.service.GetVats(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, vats, paginationMeta, "Vat fetched successfully", http.StatusOK, nil, nil)
}

func (c *VatController) CreateVat(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-CreateVat", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateVatRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewVatStoreRequest().Validate(&req, ctx)
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

	vat := models.MixValue{
		Name:        req.Name,
		GroupID:     utils.VatID,
		Description: req.Description,
		Remark:      req.Remark,
		Num:         req.Num,
		OrderItem:   nil,
		Status:      req.Status,
		CreatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()
	createdVat, err := c.service.CreateVat(ctx, &vat, tx, parentSpan)

	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create vats", http.StatusInternalServerError, err.Error(), nil)
	}

	createdHistory := models.VatHistory{
		VatID:       &createdVat.ID,
		Num:         &req.Num,
		Divider:     req.Divider,
		Multiplier:  req.Multiplier,
		ChangedAt:   *req.ChangedAt,
		Status:      &req.Status,
		Remark:      req.Remark,
		CreatedByID: &userID,
	}

	_, err = c.service.CreateVatHistory(ctx, &createdHistory, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create vats", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetVatParams{ID: createdVat.ID}
	getVat, err := c.service.GetVatByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getVat}, paginationMeta, "Vat created successfully", http.StatusCreated, nil, nil)
}

func (c *VatController) GetVatByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-GetVatByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetVatByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatParams{ID: req.ID}
	vat, err := c.service.GetVatByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	vatArray := []interface{}{vat}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, vatArray, paginationMeta, "Vat fetched successfully", http.StatusOK, nil, nil)
}

// update vat
func (c *VatController) UpdateVat(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-UpdateVat", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateVatRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewVatUpdateRequest().Validate(&req, ctx)
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

	vat := models.MixValue{
		ID:          req.ID,
		GroupID:     utils.VatID,
		Name:        req.Name,
		Description: req.Description,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	isLatest := 1
	vatHistoryParams := &dtos.GetVatHistoryParams{VatID: &req.ID, IsLatest: &isLatest}
	vatLatestHistory, err := c.service.GetVatHistoryByID(ctx, vatHistoryParams, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		// return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	if req.ChangedAt == nil {
		// current date
		changedAt := time.Now().Format("2006-01-02 15:04:05")
		req.ChangedAt = &changedAt
	}

	// if changed at < latest changed at, not update, create new history
	if vatLatestHistory != nil && vatLatestHistory.ChangedAt != nil && *req.ChangedAt >= *vatLatestHistory.ChangedAt {
		_, err := c.service.UpdateVat(ctx, &vat, tx, parentSpan)

		if err != nil {
			utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
			if err.Error() == "vat name already exists" {
				return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Vat already exists", "status": http.StatusConflict})
			}
			return utils.GetResponse(ctx, nil, nil, "Failed to update Vat", http.StatusInternalServerError, err.Error(), nil)
		}
	}

	createdHistory := models.VatHistory{
		VatID:       &req.ID,
		Num:         &req.Num,
		Divider:     req.Divider,
		Multiplier:  req.Multiplier,
		ChangedAt:   *req.ChangedAt,
		Status:      &req.Status,
		Remark:      req.Remark,
		CreatedByID: &userID,
	}

	_, err = c.service.CreateVatHistory(ctx, &createdHistory, tx, parentSpan)

	// create history
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetVatParams{ID: req.ID}
	getVat, err := c.service.GetVatByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getVat}, paginationMeta, "Vat updated successfully", http.StatusOK, nil, nil)
}

// delete vat
func (c *VatController) DeleteVat(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-DeleteVat", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatParams{ID: req.ID}
	// GET vat by ID
	_, err := c.service.GetVatByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteVat(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Vat deleted successfully", http.StatusOK, nil, nil)
}

// restore vat
func (c *VatController) RestoreVat(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-RestoreVat", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetVatParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET vat by ID
	_, err := c.service.GetVatByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreVat(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Vat", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Vat restored successfully", http.StatusOK, nil, nil)
}

func (c *VatController) ExcelGetVats(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-ExcelGetVats", opentracing.ChildOf(apiSpan.Context()))
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

	vats, err := c.service.ExcelGetVats(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}

func (c *VatController) CsvGetVats(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetVats", opentracing.ChildOf(apiSpan.Context()))
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

	vats, err := c.service.CsvGetVats(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}

func (c *VatController) GetVatHistories(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-GetVatHistories", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("VatController-GetVatHistories: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	vats, total, err := c.service.GetVatHistories(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, response, http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, vats, paginationMeta, "Vat histories fetched successfully", http.StatusOK, nil, nil)
}

func (c *VatController) GetVatHistoryByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-GetVatHistoryByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetVatByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatHistoryParams{ID: &req.ID}
	vat, err := c.service.GetVatHistoryByID(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusNotFound, err.Error(), nil)
	}

	vatArray := []interface{}{vat}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, vatArray, paginationMeta, "Vat fetched successfully", http.StatusOK, nil, nil)
}

// update vat
func (c *VatController) UpdateVatHistory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-UpdateVat", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateVatHistoryRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewVatHistoryUpdateRequest().Validate(&req, ctx)
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

	vat := models.MixValue{
		ID:          req.VatID,
		GroupID:     utils.VatID,
		Remark:      req.Remark,
		Status:      req.Status,
		UpdatedByID: &userID,
		OptionsJSON: "{}",
	}

	tx := c.repo.BeginTransaction()

	isLatest := 1
	vatHistoryParams := &dtos.GetVatHistoryParams{ID: &req.ID, VatID: &req.VatID, IsLatest: &isLatest}
	vatLatestHistory, err := c.service.GetVatHistoryByID(ctx, vatHistoryParams, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusNotFound, err.Error(), nil)
	}

	if req.ChangedAt == nil {
		// current date
		changedAt := time.Now().Format("2006-01-02 15:04:05")
		req.ChangedAt = &changedAt
	}

	// if changed at < latest changed at, not update, create new history
	if vatLatestHistory != nil && vatLatestHistory.ChangedAt != nil && *req.ChangedAt >= *vatLatestHistory.ChangedAt {
		_, err := c.service.UpdateVat(ctx, &vat, tx, parentSpan)

		if err != nil {
			utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
			return utils.GetResponse(ctx, nil, nil, "Failed to update vat history", http.StatusInternalServerError, err.Error(), nil)
		}
	}

	updatedHistory := models.VatHistory{
		ID:          req.ID,
		Num:         &req.Num,
		Divider:     req.Divider,
		Multiplier:  req.Multiplier,
		ChangedAt:   *req.ChangedAt,
		Status:      &req.Status,
		Remark:      req.Remark,
		UpdatedByID: &userID,
	}

	_, err = c.service.UpdateVatHistory(ctx, &updatedHistory, tx, parentSpan)

	// create history
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update vat history", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetVatHistoryParams{ID: &req.ID}
	getVat, err := c.service.GetVatHistoryByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getVat}, paginationMeta, "Vat history updated successfully", http.StatusOK, nil, nil)
}

// delete vat history
func (c *VatController) DeleteVatHistory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-DeleteVatHistory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetVatHistoryParams{ID: &req.ID}
	// GET vat by ID
	_, err := c.service.GetVatHistoryByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteVatHistory(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete vat history", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Vat history deleted successfully", http.StatusOK, nil, nil)
}

// restore vat
func (c *VatController) RestoreVatHistory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-RestoreVatHistory", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteVatRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetVatHistoryParams{ID: &req.ID, IsDeleted: &isDeleted}
	// GET vat by ID
	_, err := c.service.GetVatHistoryByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Vat history not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreVatHistory(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore vat history", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Vat history restored successfully", http.StatusOK, nil, nil)
}

func (c *VatController) ExcelGetVatsHistory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("VatController-ExcelGetVatsHistory", opentracing.ChildOf(apiSpan.Context()))
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

	vats, err := c.service.ExcelGetVatsHistory(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}

func (c *VatController) CsvGetVatsHistory(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetVatsHistory", opentracing.ChildOf(apiSpan.Context()))
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

	vats, err := c.service.CsvGetVatsHistory(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(vats)
}
