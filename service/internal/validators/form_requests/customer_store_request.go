package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CustomerStoreRequest handles the validation for the Request
type CustomerStoreRequest struct{}

// NewRegisterStoreRequest creates a new instance of CustomerStoreRequest.
func NewCustomerStoreRequest() *CustomerStoreRequest {
	return &CustomerStoreRequest{}
}

// Validate validates the Request
func (r *CustomerStoreRequest) Validate(req *dtos.CreateCustomerRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"customer_type_id": []string{"required", "numeric", "exists:mix_values,id"},
		"agent_id":         []string{"exists:customers,id"},
		"code":             []string{},
		"name":             []string{"required", "min:3"},
		"shortname":        []string{"required", "unique:customers,shortname"},
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
