package form_requests

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ItemGroupUpdateRequest handles the validation for the RegisterRequest.
type ItemGroupUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of ItemGroupUpdateRequest.
func NewItemGroupUpdateRequest() *ItemGroupUpdateRequest {

	return &ItemGroupUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *ItemGroupUpdateRequest) Validate(req *dtos.UpdateItemGroupRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"name":        []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
		"description": []string{},
		"remarks":     []string{},
		"status":      []string{},
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
