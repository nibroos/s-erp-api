package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// VatHistoryUpdateRequest handles the validation for the RegisterRequest.
type VatHistoryUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of VatHistoryUpdateRequest.
func NewVatHistoryUpdateRequest() *VatHistoryUpdateRequest {

	return &VatHistoryUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *VatHistoryUpdateRequest) Validate(req *dtos.UpdateVatHistoryRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"num":        []string{"numeric"},
		"remark":     []string{},
		"status":     []string{"numeric"},
		"divider":    []string{"numeric"},
		"multiplier": []string{"numeric"},
		"changed_at": []string{},
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
