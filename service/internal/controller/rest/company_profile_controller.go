package rest

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/nibroos/s-erp-api/service/internal/validators/form_requests"
	"github.com/opentracing/opentracing-go"
)

type CompanyProfileController struct {
	service *service.CompanyProfileService
	tracer  opentracing.Tracer
}

func NewCompanyProfileController(service *service.CompanyProfileService, tracer opentracing.Tracer) *CompanyProfileController {
	return &CompanyProfileController{service: service, tracer: tracer}
}

func (c *CompanyProfileController) GetCompanyProfiles(ctx *fiber.Ctx) error {
	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	companyProfiles, total, err := c.service.GetCompanyProfiles(ctx.Context(), filters)
	if err != nil {
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, companyProfiles, paginationMeta, "Company profile fetched successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) CreateCompanyProfile(ctx *fiber.Ctx) error {
	var req dtos.CreateCompanyProfileRequest

	// Use the utility function to parse the request body
	if err := utils.BodyParserWithNull(ctx, &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": err.Error(), "message": "Invalid request", "status": http.StatusBadRequest})
	}

	// Validate the request
	reqValidator := form_requests.NewCompanyProfileStoreRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	companyProfile := models.CompanyProfile{
		ParentID:           req.ParentID,
		IsPrimary:          *req.IsPrimary,
		CompanyOwnerName:   req.CompanyOwnerName,
		CompanySignName:    req.CompanySignName,
		CompanyName:        req.CompanyName,
		CompanyAddress:     req.CompanyAddress,
		CompanyPhone:       req.CompanyPhone,
		CompanyEmail:       req.CompanyEmail,
		CompanyWebsite:     req.CompanyWebsite,
		CompanyLogo:        req.CompanyLogo,
		CompanyDescription: req.CompanyDescription,
		CompanyRemark:      req.CompanyRemark,
		CompanyStatus:      req.CompanyStatus,
		CompanyOptionsJSON: "{}",
		UpdatedByID:        userID,
	}

	createdCompanyProfile, err := c.service.CreateCompanyProfile(ctx.Context(), &companyProfile)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: createdCompanyProfile.ID}
	getCompanyProfile, err := c.service.GetCompanyProfileByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCompanyProfile}, paginationMeta, "Company profile created successfully", http.StatusCreated, nil, nil)
}
func (c *CompanyProfileController) GetCompanyProfileByID(ctx *fiber.Ctx) error {
	var req dtos.GetCompanyProfileByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: req.ID}
	companyProfile, err := c.service.GetCompanyProfileByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	companyProfileArray := []interface{}{companyProfile}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, companyProfileArray, paginationMeta, "Company profile fetched successfully", http.StatusOK, nil, nil)
}

// update companyProfile
func (c *CompanyProfileController) UpdateCompanyProfile(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetCompanyProfiles", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateCompanyProfileRequest

	// Parse form values and assign them to the struct fields
	req.ID = *utils.ParseUintPointer(ctx.FormValue("id"))
	req.ParentID = utils.ParseUintPointer(ctx.FormValue("parent_id"))
	req.IsPrimary = utils.ParseIntPointer(ctx.FormValue("is_primary"))
	req.CompanyOwnerName = utils.ParseStringPointer(ctx.FormValue("company_owner_name"))
	req.CompanySignName = utils.ParseStringPointer(ctx.FormValue("company_sign_name"))
	req.CompanyName = ctx.FormValue("company_name")
	req.CompanyAddress = utils.ParseStringPointer(ctx.FormValue("company_address"))
	req.CompanyPhone = utils.ParseStringPointer(ctx.FormValue("company_phone"))
	req.CompanyEmail = utils.ParseStringPointer(ctx.FormValue("company_email"))
	req.CompanyWebsite = utils.ParseStringPointer(ctx.FormValue("company_website"))
	req.CompanyDescription = utils.ParseStringPointer(ctx.FormValue("company_description"))
	req.CompanyRemark = utils.ParseStringPointer(ctx.FormValue("company_remark"))
	req.CompanyStatus = *utils.ParseIntPointer(ctx.FormValue("company_status"))

	// log.Println("UpdateCompanyProfileRequest2", req.ID, *req.CompanyRemark, *req.CompanyAddress)
	// log.Println(ctx.FormFile("company_name"))

	// Extract user ID from JWT
	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"errors": err.Error(), "message": "Unauthorized", "status": fiber.StatusUnauthorized})
	}
	userID := uint(claims["user_id"].(float64))

	// Handle file upload
	file, err := ctx.FormFile("company_logo")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, parentSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.CompanyLogo = &filePath
	}

	// Validate the request
	reqValidator := form_requests.NewCompanyProfileUpdateRequest().Validate(&req, ctx.Context())
	if reqValidator != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	companyProfile := models.CompanyProfile{
		ID:                 req.ID,
		ParentID:           req.ParentID,
		IsPrimary:          *req.IsPrimary,
		CompanyOwnerName:   req.CompanyOwnerName,
		CompanySignName:    req.CompanySignName,
		CompanyName:        req.CompanyName,
		CompanyAddress:     req.CompanyAddress,
		CompanyPhone:       req.CompanyPhone,
		CompanyEmail:       req.CompanyEmail,
		CompanyWebsite:     req.CompanyWebsite,
		CompanyLogo:        req.CompanyLogo,
		CompanyDescription: req.CompanyDescription,
		CompanyRemark:      req.CompanyRemark,
		CompanyStatus:      req.CompanyStatus,
		CompanyOptionsJSON: "{}",
		UpdatedByID:        userID,
	}

	updatedCompanyProfile, err := c.service.UpdateCompanyProfile(ctx.Context(), &companyProfile)
	if err != nil {
		if err.Error() == "companyProfile name already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Company profile already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update Company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: updatedCompanyProfile.ID}
	getCompanyProfile, err := c.service.GetCompanyProfileByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCompanyProfile}, paginationMeta, "Company profile updated successfully", http.StatusOK, nil, nil)
}

// delete companyProfile
func (c *CompanyProfileController) DeleteCompanyProfile(ctx *fiber.Ctx) error {
	var req dtos.DeleteCompanyProfileRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: req.ID}
	// GET companyProfile by ID
	_, err := c.service.GetCompanyProfileByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.DeleteCompanyProfile(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to delete Company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Company profile deleted successfully", http.StatusOK, nil, nil)
}

// restore companyProfile
func (c *CompanyProfileController) RestoreCompanyProfile(ctx *fiber.Ctx) error {
	var req dtos.DeleteCompanyProfileRequest

	if err := ctx.BodyParser(&req); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	isDeleted := 1
	params := &dtos.GetCompanyProfileParams{ID: req.ID, IsDeleted: &isDeleted}
	// GET companyProfile by ID
	_, err := c.service.GetCompanyProfileByID(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	err = c.service.RestoreCompanyProfile(ctx.Context(), params)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to restore Company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.GetResponse(ctx, nil, nil, "Company profile restored successfully", http.StatusOK, nil, nil)
}
