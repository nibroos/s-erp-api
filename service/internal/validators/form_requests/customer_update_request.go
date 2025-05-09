package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CustomerUpdateRequest handles the validation for the Request
type CustomerUpdateRequest struct{}

// NewRegisterUpdateRequest creates a new instance of CustomerUpdateRequest.
func NewCustomerUpdateRequest() *CustomerUpdateRequest {
	return &CustomerUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *CustomerUpdateRequest) Validate(req *dtos.UpdateCustomerRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":               []string{"required", "exists:customers,id"},
		"customer_type_id": []string{"required", "exists:mix_values,id"},
		"agent_id":         []string{"exists:customers,id"},
		"code":             []string{"required"},
		"name":             []string{"required", "min:3"},
		"shortname":        []string{"required"},
		"address":          []string{},
		"phone":            []string{},
		"email":            []string{"email"},
		"pic":              []string{},
		"status":           []string{},
	}

	customFieldNames := map[string]string{
		// "customer_type_id": "Customer Type",
		// Add more custom field names here
	}

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
