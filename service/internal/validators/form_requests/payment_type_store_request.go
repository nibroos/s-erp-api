package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// PaymentTypeStoreRequest handles the validation for the RegisterRequest.
type PaymentTypeStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of PaymentTypeStoreRequest.
func NewPaymentTypeStoreRequest() *PaymentTypeStoreRequest {

	return &PaymentTypeStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *PaymentTypeStoreRequest) Validate(req *dtos.CreatePaymentTypeRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		"name":        []string{"required", "unique:mix_values,name"},
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
