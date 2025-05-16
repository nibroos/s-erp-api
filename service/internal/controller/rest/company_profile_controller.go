package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

type CompanyProfileController struct {
	service *service.CompanyProfileService
	repo    *repository.CompanyProfileRepository
	tracer  opentracing.Tracer
}

func NewCompanyProfileController(service *service.CompanyProfileService, repo *repository.CompanyProfileRepository, tracer opentracing.Tracer) *CompanyProfileController {
	return &CompanyProfileController{service: service, repo: repo, tracer: tracer}
}

func (c *CompanyProfileController) GetCompanyProfiles(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetCompanyProfiles", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("CompanyProfileController-GetCompanyProfiles: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	companyProfiles, total, err := c.service.GetCompanyProfiles(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, companyProfiles, paginationMeta, "Company profile fetched successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) CreateCompanyProfile(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-CreateCompanyProfile", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.CreateCompanyProfileRequest

	req.ParentID = utils.ParseUintPointer(ctx.FormValue("parent_id"))
	req.IsPrimary = utils.ParseIntNullPointer(ctx.FormValue("is_primary"))
	req.CompanyOwnerName = utils.ParseStringPointer(ctx.FormValue("company_owner_name"))
	req.CompanySignName = utils.ParseStringPointer(ctx.FormValue("company_sign_name"))
	req.CompanyName = ctx.FormValue("company_name")
	req.CompanyCity = utils.ParseStringPointer(ctx.FormValue("company_city"))
	req.CompanyProvince = utils.ParseStringPointer(ctx.FormValue("company_province"))
	req.CompanyDistrict = utils.ParseStringPointer(ctx.FormValue("company_district"))
	req.CompanyPostalCode = utils.ParseStringPointer(ctx.FormValue("company_postal_code"))
	req.CompanyAddress = utils.ParseStringPointer(ctx.FormValue("company_address"))
	req.CompanyPhone = utils.ParseStringPointer(ctx.FormValue("company_phone"))
	req.CompanyEmail = utils.ParseStringPointer(ctx.FormValue("company_email"))
	req.CompanyEmailPassword = utils.ParseStringPointer(ctx.FormValue("company_email_password"))
	req.CompanyWebsite = utils.ParseStringPointer(ctx.FormValue("company_website"))
	req.CompanyDescription = utils.ParseStringPointer(ctx.FormValue("company_description"))
	req.CompanyRemark = utils.ParseStringPointer(ctx.FormValue("company_remark"))
	req.CompanyStatus = utils.ParseIntNullPointer(ctx.FormValue("company_status"))

	reqValidator, isValid := form_requests.NewCompanyProfileStoreRequest().Validate(&req, ctx)
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

	logoSpan := opentracing.StartSpan("CompanyProfileController-CreateCompanyProfile-logoSpan", opentracing.ChildOf(parentSpan.Context()))
	file, err := ctx.FormFile("company_logo")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, logoSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.CompanyLogo = &filePath
	}

	signSpan := opentracing.StartSpan("CompanyProfileController-CreateCompanyProfile-signSpan", opentracing.ChildOf(parentSpan.Context()))
	fileSign, err := ctx.FormFile("company_sign")
	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, fileSign, userID, signSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.CompanySign = &filePath
	}

	companyProfile := models.CompanyProfile{
		ParentID:             req.ParentID,
		IsPrimary:            req.IsPrimary,
		CompanyOwnerName:     req.CompanyOwnerName,
		CompanySignName:      req.CompanySignName,
		CompanyName:          req.CompanyName,
		CompanyCity:          req.CompanyCity,
		CompanyProvince:      req.CompanyProvince,
		CompanyDistrict:      req.CompanyDistrict,
		CompanyPostalCode:    req.CompanyPostalCode,
		CompanyAddress:       req.CompanyAddress,
		CompanyPhone:         req.CompanyPhone,
		CompanyEmail:         req.CompanyEmail,
		CompanyEmailPassword: req.CompanyEmailPassword,
		CompanyWebsite:       req.CompanyWebsite,
		CompanyLogo:          req.CompanyLogo,
		CompanySign:          req.CompanySign,
		CompanyDescription:   req.CompanyDescription,
		CompanyRemark:        req.CompanyRemark,
		CompanyStatus:        req.CompanyStatus,
		CompanyOptionsJSON:   "{}",
		CreatedByID:          userID,
		UpdatedByID:          userID,
	}

	var bankInformations []*models.BankInformation
	bankInfosJSON := ctx.FormValue("bank_informations")
	if bankInfosJSON != "" {
		var bankInfosData []map[string]interface{}
		if err := json.Unmarshal([]byte(bankInfosJSON), &bankInfosData); err != nil {
			utils.LogErrors(parentSpan, err)
			return utils.GetResponse(ctx, nil, nil, "Invalid bank information format", http.StatusBadRequest, err.Error(), nil)
		}

		for _, bankData := range bankInfosData {
			name, _ := bankData["name"].(string)
			accountNumber, _ := bankData["account_number"].(string)
			accountName, _ := bankData["account_name"].(string)

			var description *string
			if descVal, ok := bankData["description"].(string); ok && descVal != "" {
				description = &descVal
			}

			bankInfo := &models.BankInformation{
				Name:          &name,
				AccountNumber: &accountNumber,
				AccountName:   &accountName,
				Description:   description,
				CreatedByID:   &userID,
				UpdatedByID:   &userID,
			}
			bankInformations = append(bankInformations, bankInfo)
		}
	}

	tx := c.repo.BeginTransaction()

	createdCompanyProfile, err := c.service.CreateCompanyProfile(ctx, &companyProfile, bankInformations, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to create company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetCompanyProfileParams{ID: createdCompanyProfile.ID}
	getCompanyProfile, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCompanyProfile}, paginationMeta, "Company profile created successfully", http.StatusCreated, nil, nil)
}

func (c *CompanyProfileController) GetCompanyProfileByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetCompanyProfileByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()
	var req dtos.GetCompanyProfileByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: req.ID}
	companyProfile, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	companyProfileArray := []interface{}{companyProfile}

	filters := ctx.Locals("filters").(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, companyProfileArray, paginationMeta, "Company profile fetched successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) UpdateCompanyProfile(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-UpdateCompanyProfile", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateCompanyProfileRequest
	req.ID = *utils.ParseUintPointer(ctx.FormValue("id"))
	req.ParentID = utils.ParseUintPointer(ctx.FormValue("parent_id"))
	req.IsPrimary = utils.ParseIntNullPointer(ctx.FormValue("is_primary"))
	req.CompanyOwnerName = utils.ParseStringPointer(ctx.FormValue("company_owner_name"))
	req.CompanySignName = utils.ParseStringPointer(ctx.FormValue("company_sign_name"))
	req.CompanyName = ctx.FormValue("company_name")
	req.CompanyCity = utils.ParseStringPointer(ctx.FormValue("company_city"))
	req.CompanyProvince = utils.ParseStringPointer(ctx.FormValue("company_province"))
	req.CompanyDistrict = utils.ParseStringPointer(ctx.FormValue("company_district"))
	req.CompanyPostalCode = utils.ParseStringPointer(ctx.FormValue("company_postal_code"))
	req.CompanyAddress = utils.ParseStringPointer(ctx.FormValue("company_address"))
	req.CompanyPhone = utils.ParseStringPointer(ctx.FormValue("company_phone"))
	req.CompanyEmail = utils.ParseStringPointer(ctx.FormValue("company_email"))
	req.CompanyEmailPassword = utils.ParseStringPointer(ctx.FormValue("company_email_password"))
	req.CompanyWebsite = utils.ParseStringPointer(ctx.FormValue("company_website"))
	req.CompanyDescription = utils.ParseStringPointer(ctx.FormValue("company_description"))
	req.CompanyRemark = utils.ParseStringPointer(ctx.FormValue("company_remark"))
	req.CompanyStatus = utils.ParseIntNullPointer(ctx.FormValue("company_status"))

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	reqValidator, isValid := form_requests.NewCompanyProfileUpdateRequest().Validate(&req, ctx)
	if !isValid {
		utils.LogResponse(apiSpan, reqValidator)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	params := &dtos.GetCompanyProfileParams{ID: req.ID}
	existingCompanyProfile, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	logoSpan := opentracing.StartSpan("CompanyProfileController-UpdateCompanyProfile-logoSpan", opentracing.ChildOf(parentSpan.Context()))
	deleteLogo := ctx.FormValue("company_logo_deleted") == "1"
	file, err := ctx.FormFile("company_logo")

	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, logoSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.CompanyLogo = &filePath
	} else if deleteLogo {
		req.CompanyLogo = nil
	} else if existingCompanyProfile.CompanyLogo != nil {
		req.CompanyLogo = existingCompanyProfile.CompanyLogo
	}

	signSpan := opentracing.StartSpan("CompanyProfileController-UpdateCompanyProfile-signSpan", opentracing.ChildOf(parentSpan.Context()))
	deleteSign := ctx.FormValue("company_sign_deleted") == "1"
	fileSign, err := ctx.FormFile("company_sign")

	if err == nil {
		filePath, err := utils.HandleFileUpload(ctx, fileSign, userID, signSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"errors": err.Error(), "message": "Failed to upload file", "status": http.StatusInternalServerError})
		}
		req.CompanySign = &filePath
	} else if deleteSign {
		req.CompanySign = nil
	} else if existingCompanyProfile.CompanySign != nil {
		req.CompanySign = existingCompanyProfile.CompanySign
	}

	companyProfile := models.CompanyProfile{
		ID:                   req.ID,
		ParentID:             req.ParentID,
		IsPrimary:            req.IsPrimary,
		CompanyOwnerName:     req.CompanyOwnerName,
		CompanySignName:      req.CompanySignName,
		CompanyName:          req.CompanyName,
		CompanyCity:          req.CompanyCity,
		CompanyProvince:      req.CompanyProvince,
		CompanyDistrict:      req.CompanyDistrict,
		CompanyPostalCode:    req.CompanyPostalCode,
		CompanyAddress:       req.CompanyAddress,
		CompanyPhone:         req.CompanyPhone,
		CompanyEmail:         req.CompanyEmail,
		CompanyEmailPassword: req.CompanyEmailPassword,
		CompanyWebsite:       req.CompanyWebsite,
		CompanyLogo:          req.CompanyLogo,
		CompanySign:          req.CompanySign,
		CompanyDescription:   req.CompanyDescription,
		CompanyRemark:        req.CompanyRemark,
		CompanyStatus:        req.CompanyStatus,
		CompanyOptionsJSON:   "{}",
		UpdatedByID:          userID,
	}

	var bankInformations []*models.BankInformation
	bankInfosJSON := ctx.FormValue("bank_informations")
	if bankInfosJSON != "" {
		var bankInfosData []map[string]interface{}
		if err := json.Unmarshal([]byte(bankInfosJSON), &bankInfosData); err != nil {
			utils.LogErrors(parentSpan, err)
			return utils.GetResponse(ctx, nil, nil, "Invalid bank information format", http.StatusBadRequest, err.Error(), nil)
		}

		for _, bankData := range bankInfosData {
			var bankID uint
			if id, ok := bankData["id"].(float64); ok {
				bankID = uint(id)
			}
			name, _ := bankData["name"].(string)
			accountNumber, _ := bankData["account_number"].(string)
			accountName, _ := bankData["account_name"].(string)
			var description *string
			if descVal, ok := bankData["description"].(string); ok && descVal != "" {
				description = &descVal
			}

			bankInfo := &models.BankInformation{
				ID:                bankID,
				CommpanyProfileID: &req.ID,
				Name:              &name,
				AccountNumber:     &accountNumber,
				AccountName:       &accountName,
				Description:       description,
				UpdatedByID:       &userID,
			}

			if bankID == 0 {
				bankInfo.CreatedByID = &userID
			}
			bankInformations = append(bankInformations, bankInfo)
		}
	}

	tx := c.repo.BeginTransaction()
	_, err = c.service.UpdateCompanyProfile(ctx, &companyProfile, bankInformations, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update company profile", http.StatusInternalServerError, err.Error(), nil)
	}
	tx.Commit()

	getCompanyProfile, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, []interface{}{getCompanyProfile}, paginationMeta, "Company profile updated successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) DeleteCompanyProfile(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-DeleteCompanyProfile", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.DeleteCompanyProfileRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	params := &dtos.GetCompanyProfileParams{ID: req.ID}
	_, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.DeleteCompanyProfile(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Company profile deleted successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) RestoreCompanyProfile(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-RestoreCompanyProfile", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		// If no error, delete span
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.GetCompanyProfileByIDRequest

	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == 0 {
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusBadRequest, "ID is required", nil)
	}

	isDeleted := 1
	params := &dtos.GetCompanyProfileParams{ID: req.ID, IsDeleted: &isDeleted}
	_, err := c.service.GetCompanyProfileByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.RestoreCompanyProfile(ctx, params, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore company profile", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Company profile restored successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) GetBankInformations(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetBankInformations", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters, ok := ctx.Locals("filters").(map[string]string)
	if !ok {
		apiSpan.LogKV("response_body", string("CompanyProfileController-GetBankInformations: Invalid filters"))
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, "Invalid filters", http.StatusBadRequest), http.StatusBadRequest)
	}

	bankInformations, total, err := c.service.GetBankInformations(ctx, filters, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.SendResponse(ctx, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError), http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, bankInformations, paginationMeta, "Bank information fetched successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) GetBankInformationByID(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetBankInformationByID", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, "Invalid ID", http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid ID", http.StatusBadRequest, err.Error(), nil)
	}

	params := &dtos.GetBankInformationParams{ID: uint(id)}
	bankInformation, err := c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	return utils.GetResponse(ctx, []interface{}{bankInformation}, nil, "Bank information fetched successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) CreateBankInformation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-CreateBankInformation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.BankInformationRequest
	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	companyProfileID, err := strconv.ParseUint(ctx.FormValue("commpany_profile_id"), 10, 64)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, "Invalid company profile ID", http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid company profile ID", http.StatusBadRequest, err.Error(), nil)
	}

	companyParams := &dtos.GetCompanyProfileParams{ID: uint(companyProfileID)}
	_, err = c.service.GetCompanyProfileByID(ctx, companyParams, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, "Company profile not found", http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Company profile not found", http.StatusNotFound, err.Error(), nil)
	}

	companyProfileIDUint := uint(companyProfileID)
	bankInformation := models.BankInformation{
		CommpanyProfileID: &companyProfileIDUint,
		Name:              &req.BankName,
		AccountNumber:     &req.AccountNumber,
		AccountName:       &req.AccountName,
		Description:       req.Description,
		CreatedByID:       &userID,
		UpdatedByID:       &userID,
	}

	tx := c.repo.BeginTransaction()

	createdBankInfo, err := c.service.CreateBankInformation(ctx, &bankInformation, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to create bank information", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetBankInformationParams{ID: createdBankInfo.ID}
	getBankInfo, err := c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	return utils.GetResponse(ctx, []interface{}{getBankInfo}, nil, "Bank information created successfully", http.StatusCreated, nil, nil)
}

func (c *CompanyProfileController) UpdateBankInformation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-UpdateBankInformation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.BankInformationRequest
	if err := ctx.BodyParser(&req); err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid request", http.StatusBadRequest, err.Error(), nil)
	}

	if req.ID == nil {
		return utils.GetResponse(ctx, nil, nil, "Bank information ID is required", http.StatusBadRequest, "ID is required", nil)
	}

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		utils.LogErrors(parentSpan, err)
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	params := &dtos.GetBankInformationParams{ID: *req.ID}
	existingBankInfo, err := c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	companyProfileID := existingBankInfo.CommpanyProfileID

	bankInformation := models.BankInformation{
		ID:                *req.ID,
		Name:              &req.BankName,
		AccountNumber:     &req.AccountNumber,
		AccountName:       &req.AccountName,
		Description:       req.Description,
		CommpanyProfileID: &companyProfileID,
		UpdatedByID:       &userID,
	}

	tx := c.repo.BeginTransaction()

	_, err = c.service.UpdateBankInformation(ctx, &bankInformation, tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to update bank information", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	getBankInfo, err := c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	return utils.GetResponse(ctx, []interface{}{getBankInfo}, nil, "Bank information updated successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) DeleteBankInformation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-DeleteBankInformation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, "Invalid ID", http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid ID", http.StatusBadRequest, err.Error(), nil)
	}

	params := &dtos.GetBankInformationParams{ID: uint(id)}
	_, err = c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.DeleteBankInformation(ctx, uint(id), tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to delete bank information", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Bank information deleted successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) RestoreBankInformation(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-RestoreBankInformation", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, "Invalid ID", http.StatusBadRequest))
		return utils.GetResponse(ctx, nil, nil, "Invalid ID", http.StatusBadRequest, err.Error(), nil)
	}

	// Check if bank information exists (in deleted state)
	isDeleted := 1
	params := &dtos.GetBankInformationParams{ID: uint(id), IsDeleted: &isDeleted}
	_, err = c.service.GetBankInformationByID(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusNotFound))
		return utils.GetResponse(ctx, nil, nil, "Bank information not found", http.StatusNotFound, err.Error(), nil)
	}

	tx := c.repo.BeginTransaction()

	err = c.service.RestoreBankInformation(ctx, uint(id), tx, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "Failed to restore bank information", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	return utils.GetResponse(ctx, nil, nil, "Bank information restored successfully", http.StatusOK, nil, nil)
}

func (c *CompanyProfileController) GetBankInformationsWithCompany(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("CompanyProfileController-GetBankInformationsWithCompany", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	filters := make(map[string]string)

	if len(ctx.Body()) > 0 {
		var req map[string]interface{}
		if err := ctx.BodyParser(&req); err != nil {
			return utils.GetResponse(ctx, nil, nil, "Invalid request format", http.StatusBadRequest, err.Error(), nil)
		}

		for key, value := range req {
			if value != nil {
				switch v := value.(type) {
				case string:
					filters[key] = v
				case float64:
					filters[key] = fmt.Sprintf("%v", v)
				case int:
					filters[key] = fmt.Sprintf("%d", v)
				default:
					filters[key] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	if _, ok := filters["per_page"]; !ok {
		filters["per_page"] = "10"
	}
	if _, ok := filters["page"]; !ok {
		filters["page"] = "1"
	}
	if _, ok := filters["order_column"]; !ok {
		filters["order_column"] = "id"
	}
	if _, ok := filters["order_direction"]; !ok {
		filters["order_direction"] = "desc"
	}

	ctx.Locals("filters", filters)

	bankInformations, total, err := c.service.GetBankInformationsWithCompany(ctx, filters, parentSpan)
	if err != nil {
		return utils.ErrGetReponse(ctx, apiSpan, err, "Failed to fetch bank informations", http.StatusInternalServerError)
	}

	paginationMeta := utils.CreatePaginationMeta(filters, total)

	return utils.GetResponse(ctx, bankInformations, paginationMeta, "Bank informations fetched successfully", http.StatusOK, nil, nil)
}
