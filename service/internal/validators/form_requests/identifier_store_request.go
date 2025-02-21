package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// IdentifierStoreRequest handles the validation for the RegisterRequest.
type IdentifierStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of IdentifierStoreRequest.
func NewIdentifierStoreRequest() *IdentifierStoreRequest {

	return &IdentifierStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *IdentifierStoreRequest) Validate(req *dtos.CreateIdentifierRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"type_identifier_id": []string{"required", "exists:mix_values,id"},
		"user_id":            []string{"required", "exists:users,id"},
		"ref_num":            []string{"required"},
		"status":             []string{},
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
