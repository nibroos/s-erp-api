package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// MsItemUpdateRequest handles the validation for the RegisterRequest.
type MsItemUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of MsItemUpdateRequest.
func NewMsItemUpdateRequest() *MsItemUpdateRequest {

	return &MsItemUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *MsItemUpdateRequest) Validate(req *dtos.UpdateMsItemRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                []string{"required"},
		"item_sub_group_id": []string{"required", "exists:mix_values,id"},
		"name":              []string{"required"},
		"specification":     []string{},
		"description":       []string{},
		"tpb_code":          []string{},
		"minimum_stock":     []string{"numeric"},
		"status":            []string{},
		"item_unit_id":      []string{"required"},
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
