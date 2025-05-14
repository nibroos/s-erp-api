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

type QuotationController struct {
	service *service.QuotationService
	repo    *repository.QuotationRepository
	tracer  opentracing.Tracer
}

func NewQuotationController(service *service.QuotationService, repo *repository.QuotationRepository, tracer opentracing.Tracer) *QuotationController {
	return &QuotationController{service: service, repo: repo, tracer: tracer}
}

func (c *QuotationController) GetQuotations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-GetQuotations", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("QuotationController-GetQuotations: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quotations, total, err := c.service.GetQuotations(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	quotationIDs := make([]uint, 0)
	for _, quotation := range quotations {
		quotationIDs = append(quotationIDs, uint(quotation.ID))
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quotations, paginationMeta, "quotation fetched successfully", http.StatusOK, nil, nil)
}

func (c *QuotationController) GetQuotationByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-GetQuotationByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetQuotationByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetQuotationParams{ID: req.ID}
	quotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	createdQuotationIDs := make([]uint, 0)
	createdQuotationIDs = append(createdQuotationIDs, quotation.ID)

	// get quoDts
	quoDts, err := c.service.GetQuoDtsByQuotationIDs(ctx, tx, createdQuotationIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	filters := ctx.Locals("filters").(map[string]string)
	// get quoDtBoms
	quoDtBoms, err := c.service.GetQuoDtsBomByQuotations(ctx, filters, createdQuotationIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	quoDts = c.service.MapFilterQuoDtBomsToQuoDts(ctx, quoDtBoms, quoDts, parentSpan)

	quotation.QuoDts = quoDts

	quotationArray := []interface{}{quotation}

	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, quotationArray, paginationMeta, "quotation fetched successfully", http.StatusOK, nil, nil)
}

func (c *QuotationController) CreateQuotation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-CreateQuotation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateQuotationRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewQuotationStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create quotation", reqValidator)
	}

	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()

	// create header quotation
	createdQuotation, tx, err := c.service.CreateQuotation(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create quotation", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetQuotationParams{ID: createdQuotation.ID}
	getQuotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getQuotation}, paginationMeta, "quotation created successfully", http.StatusCreated, nil, nil)
}

// update quotation
func (c *QuotationController) UpdateQuotation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-UpdateQuotation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateQuotationRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewQuotationUpdateRequest().Validate(&req, ctx)
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
	tx, err := c.service.LockQuotationTable(ctx, tx, req, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	updatedQuotation, err := c.service.UpdateQuotation(ctx, req, userID, branchID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetQuotationParams{ID: updatedQuotation.ID}
	getQuotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getQuotation}, paginationMeta, "quotation updated successfully", http.StatusOK, nil, nil)
}

// delete quotation
func (c *QuotationController) DeleteQuotation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-DeleteQuotation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteQuotationRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()

	params := &dtos.GetQuotationParams{ID: req.ID}
	// GET quotation by ID
	_, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	// DELETE quoDtBoms by Quotation ID
	err = c.service.DeleteQuoDtBomsByQuotationID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	// DELETE quoDts by Quotation ID
	err = c.service.DeleteQuoDtsByQuotationID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	err = c.service.DeleteQuotation(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "quotation deleted successfully", http.StatusOK, nil, nil)
}

// restore quotation
func (c *QuotationController) RestoreQuotation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-RestoreQuotation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteQuotationRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetQuotationParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET quotation by ID
	_, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreQuotation(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "quotation restored successfully", http.StatusOK, nil, nil)
}

func (c *QuotationController) ExcelGetQuotations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-ExcelGetQuotations", opentracing.ChildOf(apiSpan.Context()))
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

	quotations, err := c.service.ExcelGetQuotations(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(quotations)
}

func (c *QuotationController) CsvGetQuotations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetQuotations", opentracing.ChildOf(apiSpan.Context()))
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

	quotations, err := c.service.CsvGetQuotations(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(quotations)
}

func (c *QuotationController) GetWidgetQuotations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("QuotationController-GetWidgetQuotations", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("QuotationController-GetWidgetQuotations: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	quotations, total, err := c.service.GetWidgetQuotations(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch quotation", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quotations, paginationMeta, "quotation fetched successfully", http.StatusOK, nil, nil)
}

func (c *QuotationController) PdfGetQuotations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-PdfGetQuotations", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.FormQuotationRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims := utils.GetClaims(ctx, parentSpan)
	userID := uint(claims["user_id"].(float64))
	branchID := utils.GetDefaultBranchID(ctx)

	tx := c.repo.BeginTransaction()
	link, err := c.service.PdfGetQuotations(ctx, req, userID, branchID, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	link = utils.MapStringToURL(link)

	return utils.GetResponse(ctx, map[string]string{"link": *link}, nil, "PDF generated successfully", http.StatusOK, nil, nil)
}
