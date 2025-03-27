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

type ProductController struct {
	service *service.ProductService
	repo    *repository.ProductRepository
	tracer  opentracing.Tracer
}

func NewProductController(service *service.ProductService, repo *repository.ProductRepository, tracer opentracing.Tracer) *ProductController {
	return &ProductController{service: service, repo: repo, tracer: tracer}
}

func (c *ProductController) GetProducts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-GetProducts", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)

	if !ok {
		apiSpan.LogKV("response_body", string("ProductController-GetProducts: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	products, total, err := c.service.GetProducts(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	productIDs := make([]uint, 0)
	for _, product := range products {
		productIDs = append(productIDs, uint(product.ID))
	}

	// get all boms
	boms, err := c.service.GetBomsByProductIDs(ctx, filters, productIDs, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	// map bom to product
	for i, product := range products {
		productBoms := make([]dtos.ProductBomListDTO, 0)
		for _, bom := range boms {
			if bom.ProductID == uint(product.ID) {
				productBoms = append(productBoms, bom)
			}
		}
		products[i].Boms = productBoms
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, products, paginationMeta, "Master product fetched successfully", http.StatusOK, nil, nil)
}

func (c *ProductController) CreateProduct(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-CreateProduct", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateProductRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewProductStoreRequest().Validate(&req, ctx)
	if !isValid {
		return utils.ErrValidResponse(ctx, apiSpan, "Failed to create product", reqValidator)
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	prodType := "single"
	if len(req.Boms) > 0 {
		prodType = "product"
	}

	isVat := 0
	if req.VatID != nil {
		isVat = 1
	}

	isPph23 := 0
	if req.Pph23ID != nil {
		isPph23 = 1
	}

	product := models.Product{
		ItemSubGroupID: req.ItemSubGroupID,
		ItemUnitID:     &req.ItemUnitID,
		VatID:          req.VatID,
		Pph23ID:        req.Pph23ID,
		Code:           req.Code,
		FactoryCode:    req.FactoryCode,
		Name:           req.Name,
		ProdType:       &prodType,
		Sku:            req.Sku,
		Barcode:        req.Barcode,
		Specification:  req.Specification,
		Description:    req.Description,
		TpbCode:        req.TpbCode,
		MinimumStock:   req.MinimumStock,
		IsAllBranch:    req.IsAllBranch,
		IsVat:          &isVat,
		IsPph23:        &isPph23,
		Remark:         req.Remark,
		Status:         req.Status,
		ExpiredAt:      req.ExpiredAt,
		CreatedByID:    &userID,
	}

	tx := c.repo.BeginTransaction()
	createdProduct, err := c.service.CreateProduct(ctx, &product, tx, parentSpan)

	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create product", http.StatusInternalServerError)
	}

	// bulk create item units
	itemUnits := make([]*models.ItemUnit, 0)
	for _, unit := range req.Units {
		itemUnit := &models.ItemUnit{
			// MsItemID:    createdMsItem.ID,
			ProductID:   createdProduct.ID,
			UnitID:      unit.UnitID,
			Conversion:  &unit.Conversion,
			PriceSell:   &unit.PriceSell,
			PriceBuy:    &unit.PriceBuy,
			Status:      1,
			CreatedByID: &userID,
		}
		itemUnits = append(itemUnits, itemUnit)
	}

	err = c.service.CreateItemUnits(ctx, itemUnits, createdProduct.ID, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	// get selected item unit id by unit id
	paramsItemUnit := &dtos.GetProductItemUnitParams{ProductID: createdProduct.ID, UnitID: req.ItemUnitID}
	selectedItemUnit, err := c.service.GetItemUnitIDBySelectedItemID(ctx, tx, paramsItemUnit, parentSpan)
	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	// update product with selected item unit id
	product.ItemUnitID = &selectedItemUnit.ID
	_, err = c.service.UpdateProduct(ctx, &product, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "Failed to create master items", http.StatusInternalServerError, err.Error(), nil)
	}

	// bulk create item boms
	boms := make([]*models.Bom, 0)
	for _, bom := range req.Boms {
		bom := &models.Bom{
			ProductID:     &createdProduct.ID,
			ProductItemID: &bom.ProductItemID,
			ItemUnitID:    &bom.ItemUnitID,
			Qty:           &bom.Qty,
			Remark:        bom.Remark,
			CreatedByID:   &userID,
		}
		boms = append(boms, bom)
	}

	err = c.service.CreateBoms(ctx, boms, createdProduct.ID, tx, parentSpan)

	if err != nil {
		utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to create BOM", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetProductParams{ID: createdProduct.ID}
	getProduct, err := c.service.GetProductByID(ctx, params, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	createdBoms, err := c.service.GetBomsByProductID(ctx, createdProduct.ID, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	getProduct.Boms = createdBoms

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getProduct}, paginationMeta, "Master product created successfully", http.StatusCreated, nil, nil)
}
func (c *ProductController) GetProductByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-GetProductByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetProductByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetProductParams{ID: req.ID}
	product, err := c.service.GetProductByID(ctx, params, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	boms, err := c.service.GetBomsByProductID(ctx, req.ID, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	product.Boms = boms

	productArray := []interface{}{product}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, productArray, paginationMeta, "Master product fetched successfully", http.StatusOK, nil, nil)
}

// update product
func (c *ProductController) UpdateProduct(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-UpdateProduct", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateProductRequest

	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewProductUpdateRequest().Validate(&req, ctx)
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

	prodType := "single"
	if len(req.Boms) > 0 {
		prodType = "product"
	}

	isVat := 0
	if req.VatID != nil {
		isVat = 1
	}

	isPph23 := 0
	if req.Pph23ID != nil {
		isPph23 = 1
	}

	product := models.Product{
		ID:             req.ID,
		ItemSubGroupID: req.ItemSubGroupID,
		ItemUnitID:     &req.ItemUnitID,
		VatID:          req.VatID,
		Pph23ID:        req.Pph23ID,
		Code:           req.Code,
		FactoryCode:    req.FactoryCode,
		Name:           req.Name,
		Sku:            req.Sku,
		ProdType:       &prodType,
		Barcode:        req.Barcode,
		Specification:  req.Specification,
		Description:    req.Description,
		TpbCode:        req.TpbCode,
		MinimumStock:   req.MinimumStock,
		IsAllBranch:    req.IsAllBranch,
		IsVat:          &isVat,
		IsPph23:        &isPph23,
		Remark:         req.Remark,
		Status:         req.Status,
		ExpiredAt:      req.ExpiredAt,
		UpdatedByID:    &userID,
	}

	tx := c.repo.BeginTransaction()

	updatedProduct, err := c.service.UpdateProduct(ctx, &product, tx, parentSpan)

	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "product name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Master product already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Master product", http.StatusInternalServerError, err.Error(), nil)
	}

	// Bulk/Create Update Batch Boms
	boms := make([]*models.Bom, 0)
	for _, bom := range req.Boms {
		bom := &models.Bom{
			ID:            bom.ID,
			ProductID:     &updatedProduct.ID,
			ProductItemID: &bom.ProductItemID,
			ItemUnitID:    &bom.ItemUnitID,
			Qty:           &bom.Qty,
			Remark:        bom.Remark,
			UpdatedByID:   &userID,
		}
		boms = append(boms, bom)
	}

	err = c.service.BulkCreateUpdateBoms(ctx, boms, updatedProduct.ID, tx, parentSpan)
	if err != nil {
		return utils.ErrTrxResponse(ctx, tx, apiSpan, err, "Failed to update Master product", http.StatusInternalServerError)
	}

	tx.Commit()

	params := &dtos.GetProductParams{ID: updatedProduct.ID}
	getProduct, err := c.service.GetProductByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusNotFound, err.Error(), nil)
	}

	createdBoms, err := c.service.GetBomsByProductID(ctx, getProduct.ID, parentSpan)
	if err != nil {
		utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch Master product", http.StatusInternalServerError)
	}

	getProduct.Boms = createdBoms

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getProduct}, paginationMeta, "Master product updated successfully", http.StatusOK, nil, nil)
}

// delete product
func (c *ProductController) DeleteProduct(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-DeleteProduct", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteProductRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetProductParams{ID: req.ID}
	// GET product by ID
	_, err := c.service.GetProductByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusNotFound, err.Error(), nil)
	}

	// Transaction handling
	tx := c.repo.BeginTransaction()
	err = c.service.DeleteProduct(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Master product", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master product deleted successfully", http.StatusOK, nil, nil)
}

// restore product
func (c *ProductController) RestoreProduct(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-RestoreProduct", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteProductRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusBadRequest, "ID is required", nil)
	}

	tx := c.repo.BeginTransaction()

	isDeleted := 1
	params := &dtos.GetProductParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET product by ID
	_, err := c.service.GetProductByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Master product not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreProduct(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Master product", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Master product restored successfully", http.StatusOK, nil, nil)
}

func (c *ProductController) ExcelGetProducts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("ProductController-ExcelGetProducts", opentracing.ChildOf(apiSpan.Context()))
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

	products, err := c.service.ExcelGetProducts(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(products)
}

func (c *ProductController) CsvGetProducts(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CustomerTypeController-CsvGetProducts", opentracing.ChildOf(apiSpan.Context()))
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

	products, err := c.service.CsvGetProducts(ctx, filters, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return ctx.Send(products)
}
