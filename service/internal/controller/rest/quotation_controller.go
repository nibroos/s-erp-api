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
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	quotationIDs := make([]uint, 0)
	for _, quotation := range quotations {
		quotationIDs = append(quotationIDs, uint(quotation.ID))
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, quotations, paginationMeta, "Master quotation fetched successfully", http.StatusOK, nil, nil)
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
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	params := &dtos.GetQuotationParams{ID: req.ID}
	quotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	createdQuotationIDs := make([]uint, 0)
	createdQuotationIDs = append(createdQuotationIDs, quotation.ID)

	// get created quoDts
	quoDts, err := c.service.GetQuoDtsByQuotationID(ctx, tx, createdQuotationIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	tx.Commit()

	quotation.QuoDts = quoDts

	quotationArray := []interface{}{quotation}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, quotationArray, paginationMeta, "Master quotation fetched successfully", http.StatusOK, nil, nil)
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

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	quotation := models.Quotation{
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		QuoNo:         req.QuoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		Status:        req.Status,
		IsApproved:    req.IsApproved,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		TotalQty:      req.TotalQty,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		CreatedByID:   &userID,
	}

	tx := c.repo.BeginTransaction()

	// create header quotation
	createdQuotation, tx, err := c.service.CreateQuotation(ctx, &quotation, tx, parentSpan)

	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create quotation", http.StatusInternalServerError)
	}

	// bulk create item quoDts ref ms items / product->boms
	quoDts, err := c.service.MapCreateQuoDts(ctx, req, createdQuotation, userID, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Quo Details", http.StatusInternalServerError)
	}

	tx, err = c.service.CreateQuoDts(ctx, quoDts, createdQuotation.ID, tx, parentSpan)

	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Quo Details", http.StatusInternalServerError)
	}

	createdQuotationIDs := make([]uint, 0)
	createdQuotationIDs = append(createdQuotationIDs, createdQuotation.ID)

	log.Println("createdQuotationIDs", createdQuotationIDs, "createdQuotation", createdQuotation.ID)

	// get created quoDts
	createdQuoDts, err := c.service.GetQuoDtsByQuotationID(ctx, tx, createdQuotationIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	log.Println("createdQuoDts", createdQuoDts)

	// bulk create boms
	quoDtBoms, err := c.service.MapCreateQuoDtBoms(ctx, req, createdQuoDts, userID, parentSpan)
	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Quo Detail BOMs", http.StatusInternalServerError)
	}

	// if quoDtBoms is not empty
	if len(quoDtBoms) > 0 {
		tx, err = c.service.CreateQuoDtBoms(ctx, quoDtBoms, tx, parentSpan)
		if err != nil {
			utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create Quo Detail BOMs", http.StatusInternalServerError)
		}
	}

	log.Println("Quotation created successfully quodtbom", quoDtBoms)

	tx.Commit()

	params := &dtos.GetQuotationParams{ID: createdQuotation.ID}
	getQuotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getQuotation}, paginationMeta, "Master quotation created successfully", http.StatusCreated, nil, nil)
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
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	branchID := claims["bid"].(*uint)

	quotation := models.Quotation{
		ID:            req.ID,
		CustomerID:    req.CustomerID,
		OrderTypeID:   req.OrderTypeID,
		CurrencyID:    req.CurrencyID,
		VatID:         req.VatID,
		PaymentID:     req.PaymentID,
		Pph23ID:       req.Pph23ID,
		QuoNo:         req.QuoNo,
		Title:         req.Title,
		Remark:        req.Remark,
		Status:        req.Status,
		IsApproved:    req.IsApproved,
		ExchangeRate:  req.ExchangeRate,
		VatPerc:       req.VatPerc,
		Pph23Perc:     req.Pph23Perc,
		TotalQty:      req.TotalQty,
		Subtotal:      req.Subtotal,
		TotalDiscount: req.TotalDiscount,
		TotalPph23:    req.TotalPph23,
		TotalVat:      req.TotalVat,
		GrandTotal:    req.GrandTotal,
		DueAt:         req.DueAt,
		ExpiredAt:     req.ExpiredAt,
		BranchID:      branchID,
		UpdatedByID:   &userID,
	}

	tx := c.repo.BeginTransaction()

	updatedQuotation, err := c.service.UpdateQuotation(ctx, &quotation, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update Master quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	refJSON := "{}"

	// Bulk/Create Update Batch QuoDts
	quoDts := make([]*models.QuoDt, 0)
	for _, quoDt := range req.QuoDts {
		quoDt := &models.QuoDt{
			ID:           quoDt.ID,
			QuotationID:  &updatedQuotation.ID,
			RefID:        quoDt.RefID,
			ItemID:       quoDt.ItemID,
			ItemUnitID:   quoDt.ItemUnitID,
			VatID:        quoDt.VatID,
			RefType:      quoDt.RefType,
			ItemType:     quoDt.ItemType,
			RefJSON:      &refJSON,
			Remark:       quoDt.Remark,
			VatPerc:      quoDt.VatPerc,
			VatPercAm:    quoDt.VatPercAm,
			QtySO:        quoDt.QtySO,
			Qty:          quoDt.Qty,
			PriceSell:    quoDt.PriceSell,
			PriceBuy:     quoDt.PriceBuy,
			SubtotalSell: quoDt.SubtotalSell,
			SubtotalBuy:  quoDt.SubtotalBuy,
			DiscAm:       quoDt.DiscAm,
			DiscPerc:     quoDt.DiscPerc,
			DiscPercNum:  quoDt.DiscPercNum,
			DiscPercAm:   quoDt.DiscPercAm,
			DiscFinal:    quoDt.DiscFinal,
			DiscType:     quoDt.DiscType,
			TotalAm:      quoDt.TotalAm,
			UpdatedByID:  &userID,
		}
		quoDts = append(quoDts, quoDt)
	}

	tx, err = c.service.BulkCreateUpdateQuoDts(ctx, quoDts, updatedQuotation.ID, tx, parentSpan)
	if err != nil {
		return utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to update Master quotation", http.StatusInternalServerError)
	}

	updatedQuotationIDs := make([]uint, 0)
	updatedQuotationIDs = append(updatedQuotationIDs, updatedQuotation.ID)

	// get updated quoDts
	updatedQuoDts, err := c.service.GetQuoDtsByQuotationID(ctx, tx, updatedQuotationIDs, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master quotation", http.StatusInternalServerError)
	}

	// Bulk/Create Update Batch QuoDtBoms
	err = c.service.BulkCreateUpdateQuoDtBoms(ctx, updatedQuoDts, updatedQuotation.ID, tx, parentSpan)
	if err != nil {
		return utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to update Master quotation", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetQuotationParams{ID: updatedQuotation.ID}
	getQuotation, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getQuotation}, paginationMeta, "Master quotation updated successfully", http.StatusOK, nil, nil)
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
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()

	params := &dtos.GetQuotationParams{ID: req.ID}
	// GET quotation by ID
	_, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	// DELETE quoDtBoms by Quotation ID
	err = c.service.DeleteQuoDtBomsByQuotationID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Master quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	// DELETE quoDts by Quotation ID
	err = c.service.DeleteQuoDtsByQuotationID(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Master quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	err = c.service.DeleteQuotation(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Master quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master quotation deleted successfully", http.StatusOK, nil, nil)
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
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetQuotationParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET quotation by ID
	_, err := c.service.GetQuotationByID(ctx, params, tx, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master quotation not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreQuotation(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Master quotation", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master quotation restored successfully", http.StatusOK, nil, nil)
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
