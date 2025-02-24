package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// UserStoreRequest handles the validation for the RegisterRequest.
type UserStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of UserStoreRequest.
func NewUserStoreRequest() *UserStoreRequest {

	return &UserStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *UserStoreRequest) Validate(req *dtos.CreateUserRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		"name":     []string{"required", "min:3"},
		"username": []string{"unique:users,username"},
		"email":    []string{"required", "email", "unique:users,email"},
		"password": []string{"required", "min:4"},
		"role_ids": []string{"required"},
	}

	customFieldNames := map[string]string{}

	var requestBody map[string]interface{}
	if err := ctx.BodyParser(&requestBody); err != nil {
		return map[string][]string{"error": {"Invalid request body"}}, false
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
