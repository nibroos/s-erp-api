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

type BranchController struct {
	service *service.BranchService
	repo    *repository.BranchRepository
	tracer  opentracing.Tracer
}

func NewBranchController(service *service.BranchService, repo *repository.BranchRepository, tracer opentracing.Tracer) *BranchController {
	return &BranchController{service: service, repo: repo, tracer: tracer}
}

func (c *BranchController) GetBranches(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-GetBranches", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("BranchController-GetBranches: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	branches, total, err := c.service.GetBranches(ctx.Context(), filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, branches, paginationMeta, "Branch fetched successfully", http.StatusOK, nil, nil)
}

func (c *BranchController) CreateBranch(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-CreateBranch", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateBranchRequest

	// Parse form values and assign them to the struct fields
	req.ParentID = utils.ParseUintPointer(ctx.FormValue("parent_id"))
	req.CompanyProfileID = utils.ParseUintPointer(ctx.FormValue("company_profile_id"))
	req.OwnerName = utils.ParseStringPointer(ctx.FormValue("owner_name"))
	req.SignName = utils.ParseStringPointer(ctx.FormValue("sign_name"))
	req.Name = ctx.FormValue("name")
	req.Address = utils.ParseStringPointer(ctx.FormValue("address"))
	req.Phone = utils.ParseStringPointer(ctx.FormValue("phone"))
	req.Email = utils.ParseStringPointer(ctx.FormValue("email"))
	req.Website = utils.ParseStringPointer(ctx.FormValue("website"))
	req.Description = utils.ParseStringPointer(ctx.FormValue("description"))
	req.Remark = utils.ParseStringPointer(ctx.FormValue("remark"))
	req.Status = utils.ParseIntNullPointer(ctx.FormValue("status"))

	// Validate the request
	reqValidator, isValid := form_requests.NewBranchStoreRequest().Validate(&req, ctx)
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

	// Handle file upload
	logoSpan := opentracing.StartSpan("BranchController-CreateBranch-logoSpan", opentracing.ChildOf(parentSpan.Context()))
	file, err := ctx.FormFile("logo")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, logoSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.Logo = &filePath
	}

	signSpan := opentracing.StartSpan("BranchController-CreateBranch-signSpan", opentracing.ChildOf(parentSpan.Context()))
	fileSign, err := ctx.FormFile("sign")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, fileSign, userID, signSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.Logo = &filePath
	}

	branch := models.Branch{
		ParentID:         req.ParentID,
		CompanyProfileID: req.CompanyProfileID,
		OwnerName:        req.OwnerName,
		SignName:         req.SignName,
		Name:             req.Name,
		Address:          req.Address,
		Phone:            req.Phone,
		Email:            req.Email,
		Website:          req.Website,
		Logo:             req.Logo,
		Sign:             req.Sign,
		Description:      req.Description,
		Remark:           req.Remark,
		Status:           req.Status,
		OptionsJSON:      "{}",
		UpdatedByID:      userID,
	}

	tx := c.repo.BeginTransaction()

	createdBranch, err := c.service.CreateBranch(ctx.Context(), &branch, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to create branch", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetBranchParams{ID: createdBranch.ID}
	getBranch, err := c.service.GetBranchByID(ctx.Context(), params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getBranch}, paginationMeta, "Branch created successfully", http.StatusCreated, nil, nil)
}

func (c *BranchController) GetBranchByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-GetBranchByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()
	var req dtos.GetBranchByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetBranchParams{ID: req.ID}
	branch, err := c.service.GetBranchByID(ctx.Context(), params, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusNotFound, err.Error(), nil)
	}

	branchArray := []interface{}{branch}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, branchArray, paginationMeta, "Branch fetched successfully", http.StatusOK, nil, nil)
}

// update branch
func (c *BranchController) UpdateBranch(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-GetBranches", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateBranchRequest

	// Parse form values and assign them to the struct fields
	req.ID = *utils.ParseUintPointer(ctx.FormValue("id"))
	req.ParentID = utils.ParseUintPointer(ctx.FormValue("parent_id"))
	req.CompanyProfileID = utils.ParseUintPointer(ctx.FormValue("company_profile_id"))
	req.OwnerName = utils.ParseStringPointer(ctx.FormValue("owner_name"))
	req.SignName = utils.ParseStringPointer(ctx.FormValue("sign_name"))
	req.Name = ctx.FormValue("name")
	req.Address = utils.ParseStringPointer(ctx.FormValue("address"))
	req.Phone = utils.ParseStringPointer(ctx.FormValue("phone"))
	req.Email = utils.ParseStringPointer(ctx.FormValue("email"))
	req.Website = utils.ParseStringPointer(ctx.FormValue("website"))
	req.Description = utils.ParseStringPointer(ctx.FormValue("description"))
	req.Remark = utils.ParseStringPointer(ctx.FormValue("remark"))
	req.Status = utils.ParseIntNullPointer(ctx.FormValue("status"))

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	// Handle file upload
	logoSpan := opentracing.StartSpan("BranchController-CreateBranch-logoSpan", opentracing.ChildOf(parentSpan.Context()))
	file, err := ctx.FormFile("logo")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, logoSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.Logo = &filePath
	}

	signSpan := opentracing.StartSpan("BranchController-CreateBranch-signSpan", opentracing.ChildOf(parentSpan.Context()))
	fileSign, err := ctx.FormFile("sign")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, fileSign, userID, signSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.Logo = &filePath
	}

	// Validate the request
	reqValidator, isValid := form_requests.NewBranchUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	branch := models.Branch{
		ID:               req.ID,
		ParentID:         req.ParentID,
		CompanyProfileID: req.CompanyProfileID,
		OwnerName:        req.OwnerName,
		SignName:         req.SignName,
		Name:             req.Name,
		Address:          req.Address,
		Phone:            req.Phone,
		Email:            req.Email,
		Website:          req.Website,
		Logo:             req.Logo,
		Sign:             req.Sign,
		Description:      req.Description,
		Remark:           req.Remark,
		Status:           req.Status,
		OptionsJSON:      "{}",
		UpdatedByID:      userID,
	}

	tx := c.repo.BeginTransaction()
	updatedBranch, err := c.service.UpdateBranch(ctx.Context(), &branch, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))

		if err.Error() == "branch name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Branch already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Branch", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetBranchParams{ID: updatedBranch.ID}
	getBranch, err := c.service.GetBranchByID(ctx.Context(), params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getBranch}, paginationMeta, "Branch updated successfully", http.StatusOK, nil, nil)
}

// delete branch
func (c *BranchController) DeleteBranch(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-DeleteBranch", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteBranchRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetBranchParams{ID: req.ID}
	// GET branch by ID
	_, err := c.service.GetBranchByID(ctx.Context(), params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.DeleteBranch(ctx.Context(), params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Branch", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Branch deleted successfully", http.StatusOK, nil, nil)
}

// restore branch
func (c *BranchController) RestoreBranch(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("BranchController-RestoreBranch", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()
	var req dtos.DeleteBranchRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusBadRequest, "ID is required", nil)
	}

	isDeleted := 1
	params := &dtos.GetBranchParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET branch by ID
	_, err := c.service.GetBranchByID(ctx.Context(), params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Branch not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.RestoreBranch(ctx.Context(), params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Branch", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Branch restored successfully", http.StatusOK, nil, nil)
}
