package form_requests

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type AccountSettingUpdateRequest struct {
}

func NewAccountSettingUpdateRequest() *AccountSettingUpdateRequest {
	return &AccountSettingUpdateRequest{}
}

func (r *AccountSettingUpdateRequest) Validate(req *dtos.UpdateAccountSettingRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"name":         []string{"required", "min:3"},
		"username":     []string{"required", fmt.Sprintf("unique_ig:users,username,%d", req.ID)},
		"email":        []string{"required", "email", fmt.Sprintf("unique_ig:users,email,%d", req.ID)},
		"phone_number": []string{"required"},
		"address":      []string{"required"},
	}

	password := ctx.FormValue("password")
	if password != "" {
		rules["password"] = []string{"min:4"}

		passwordConfirmation := ctx.FormValue("password_confirmation")
		if passwordConfirmation == "" {
			return map[string][]string{
				"password_confirmation": {"Password confirmation is required"},
			}, false
		}
		if passwordConfirmation == "" || password != passwordConfirmation {
			return map[string][]string{
				"password_confirmation": {"Password confirmation does not match"},
			}, false
		}
	}

	customFieldNames := map[string]string{}

	requestBody := map[string]interface{}{
		"name":                  ctx.FormValue("name"),
		"username":              ctx.FormValue("username"),
		"email":                 ctx.FormValue("email"),
		"phone_number":          ctx.FormValue("phone_number"),
		"address":               ctx.FormValue("address"),
		"password":              ctx.FormValue("password"),
		"password_confirmation": ctx.FormValue("password_confirmation"),
	}

	request := validators.NewRequest(rules, requestBody, customFieldNames)
	errors, valid := request.Validate()

	convertedErrors := make(map[string][]string)
	for key, value := range errors {
		if len(value) > 0 {
			convertedErrors[key] = value
		}
	}
	return convertedErrors, valid
}
