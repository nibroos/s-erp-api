package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// MsItemStoreRequest handles the validation for the RegisterRequest.
type MsItemStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of MsItemStoreRequest.
func NewMsItemStoreRequest() *MsItemStoreRequest {

	return &MsItemStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *MsItemStoreRequest) Validate(req *dtos.CreateMsItemRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		"item_sub_group_id":  []string{"required", "exists:mix_values,id"},
		"name":               []string{"required"},
		"specification":      []string{},
		"tpb_code":           []string{},
		"price_sell":         []string{"numeric"},
		"price_buy":          []string{"numeric"},
		"minimum_stock":      []string{"numeric"},
		"status":             []string{},
		"item_unit_id":       []string{"required"},
		"units":              []string{"required", "array"},
		"units.*.unit_id":    []string{"required", "exists:ms_units,id"},
		"units.*.price_sell": []string{"required", "numeric"},
		"units.*.price_buy":  []string{"required", "numeric"},
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
