package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// VatStoreRequest handles the validation for the RegisterRequest.
type VatStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of VatStoreRequest.
func NewVatStoreRequest() *VatStoreRequest {

	return &VatStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *VatStoreRequest) Validate(req *dtos.CreateVatRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		"name":        []string{"required", "unique:mix_values,name"},
		"num":         []string{"numeric"},
		"description": []string{},
		"remarks":     []string{},
		"status":      []string{},
		"divider":     []string{"numeric"},
		"multiplier":  []string{"numeric"},
		"changed_at":  []string{},
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
