package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// BranchItemStoreRequest handles the validation for the RegisterRequest.
type BranchItemStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of BranchItemStoreRequest.
func NewBranchItemStoreRequest() *BranchItemStoreRequest {

	return &BranchItemStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *BranchItemStoreRequest) Validate(req *dtos.CreateBranchItemRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"product_id":    []string{"required", "exists:products,id"},
		"branch_id":     []string{"required", "exists:branches,id"},
		"name":          []string{},
		"specification": []string{},
		"description":   []string{},
		"tpb_code":      []string{},
		"price_sell":    []string{"numeric"},
		"price_buy":     []string{"numeric"},
		"minimum_stock": []string{"numeric"},
		"status":        []string{},
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
