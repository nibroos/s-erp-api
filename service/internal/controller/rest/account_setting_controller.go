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

type AccountSettingController struct {
	service *service.AccountSettingService
	repo    *repository.AccountSettingRepository
	tracer  opentracing.Tracer
}

func NewAccountSettingController(service *service.AccountSettingService, repo *repository.AccountSettingRepository, tracer opentracing.Tracer) *AccountSettingController {
	return &AccountSettingController{service: service, repo: repo, tracer: tracer}
}

func (c *AccountSettingController) GetAccountSettingUser(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("AccountSettingController-GetAccountSettingUser", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))

	params := &dtos.GetUserByIDParams{ID: userID}
	user, err := c.service.GetAccountSettingUser(ctx, params, parentSpan)
	if err != nil {
		response := utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError)
		utils.LogResponse(apiSpan, response)
		return utils.GetResponse(ctx, nil, nil, "User not found", http.StatusNotFound, err.Error(), nil)
	}

	userArray := []interface{}{user}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, userArray, paginationMeta, "User fetched successfully", http.StatusOK, nil, nil)
}

func (c *AccountSettingController) UpdateAccountSetting(ctx *fiber.Ctx) error {
	apiSpan := utils.StartSpanFromController(ctx, c.tracer, ctx.Path())
	parentSpan := opentracing.StartSpan("AccountSettingController-UpdateAccountSetting", opentracing.ChildOf(apiSpan.Context()))
	defer func() {
		if utils.FilterOtel(ctx) {
			defer apiSpan.Finish()
			defer parentSpan.Finish()
		}
	}()

	var req dtos.UpdateAccountSettingRequest

	req.Name = ctx.FormValue("name")
	req.Username = utils.ParseStringPointer(ctx.FormValue("username"))
	req.Email = ctx.FormValue("email")
	req.Address = utils.ParseStringPointer(ctx.FormValue("address"))
	req.PhoneNumber = utils.ParseStringPointer(ctx.FormValue("phone_number"))
	req.Password = utils.ParseStringPointer(ctx.FormValue("password"))
	req.PasswordConfirmation = utils.ParseStringPointer(ctx.FormValue("password_confirmation"))

	claims, err := middleware.GetAuthUser(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	userID := uint(claims["user_id"].(float64))
	req.ID = userID

	imageSpan := opentracing.StartSpan("AccountSettingController-UpdateAccountSetting-ImageUpload", opentracing.ChildOf(parentSpan.Context()))
	file, err := ctx.FormFile("profile_image")
	if err != nil {
		if err.Error() != "there is no uploaded file associated with the given key" {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"errors":  err.Error(),
				"message": "Failed to process file upload",
				"status":  http.StatusInternalServerError,
			})
		}
	} else {
		filePath, err := utils.HandleFileUpload(ctx, file, userID, imageSpan)
		if err != nil {
			utils.LogErrors(parentSpan, err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"errors":  err.Error(),
				"message": "Failed to upload file",
				"status":  http.StatusInternalServerError,
			})
		}
		req.ProfileImageURL = &filePath
	}

	reqValidator, isValid := form_requests.NewAccountSettingUpdateRequest().Validate(&req, ctx)
	if !isValid {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": reqValidator, "message": "Validation failed", "status": http.StatusBadRequest})
	}

	_, err = c.service.GetAccountSettingUser(ctx, &dtos.GetUserByIDParams{ID: userID}, parentSpan)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "User not found", http.StatusNotFound, err.Error(), nil)
	}

	user := models.User{
		ID:              req.ID,
		Name:            req.Name,
		Username:        req.Username,
		Email:           req.Email,
		Address:         req.Address,
		PhoneNumber:     req.PhoneNumber,
		ProfileImageURL: req.ProfileImageURL,
		UpdatedByID:     userID,
	}

	fieldsToUpdate := []string{"name", "username", "email", "address", "phone_number", "updated_by_id"}

	if req.ProfileImageURL != nil {
		fieldsToUpdate = append(fieldsToUpdate, "profile_image_url")
	}

	if req.Password != nil && *req.Password != "" {
		user.Password = *req.Password
		fieldsToUpdate = append(fieldsToUpdate, "password")
	}

	tx := c.repo.BeginTransaction()

	updatedUser, err := c.service.UpdateAccountSetting(ctx, tx, &user, fieldsToUpdate, parentSpan)
	if err != nil {
		tx.Rollback()
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		if err.Error() == "username already exists" {
			return ctx.Status(http.StatusConflict).JSON(fiber.Map{"errors": err.Error(), "message": "Username already exists", "status": http.StatusConflict})
		}
		return utils.GetResponse(ctx, nil, nil, "Failed to update account setting", http.StatusInternalServerError, err.Error(), nil)
	}

	tx.Commit()

	params := &dtos.GetUserByIDParams{ID: updatedUser.ID}
	getUser, err := c.service.GetAccountSettingUser(ctx, params, parentSpan)
	if err != nil {
		utils.LogResponse(apiSpan, utils.WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
		return utils.GetResponse(ctx, nil, nil, "User not found", http.StatusNotFound, err.Error(), nil)
	}

	filters := make(map[string]string)
	paginationMeta := utils.CreatePaginationMeta(filters, 1)

	return utils.GetResponse(ctx, getUser, paginationMeta, "Account setting updated successfully", http.StatusOK, nil, nil)
}
