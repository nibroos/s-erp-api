package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// BranchItemUpdateRequest handles the validation for the RegisterRequest.
type BranchItemUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of BranchItemUpdateRequest.
func NewBranchItemUpdateRequest() *BranchItemUpdateRequest {
	return &BranchItemUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *BranchItemUpdateRequest) Validate(req *dtos.UpdateBranchItemRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":            []string{"required"},
		"name":          []string{},
		"specification": []string{},
		"description":   []string{},
		"tpb_code":      []string{},
		"price_sell":    []string{"numeric"},
		"price_buy":     []string{"numeric"},
		"minimum_stock": []string{"numeric"},
		"status":        []string{"required"},
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
