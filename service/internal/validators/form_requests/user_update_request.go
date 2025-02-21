package form_requests

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// UserdUpdateRequest handles the validation for the RegisterRequest.
type UserdUpdateRequest struct {
}

// NewRegisterdUpdateRequest creates a new instance of UserdUpdateRequest.
func NewUserdUpdateRequest() *UserdUpdateRequest {

	return &UserdUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *UserdUpdateRequest) Validate(req *dtos.UpdateUserRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		"name":     []string{"required", "min:3"},
		"username": []string{fmt.Sprintf("unique_ig:users,username,%d", req.ID)},
		"email":    []string{"required", "email", fmt.Sprintf("unique_ig:users,email,%d", req.ID)},
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
