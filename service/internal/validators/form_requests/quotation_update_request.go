package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// QuotationUpdateRequest handles the validation for the RegisterRequest.
type QuotationUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of QuotationUpdateRequest.
func NewQuotationUpdateRequest() *QuotationUpdateRequest {

	return &QuotationUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *QuotationUpdateRequest) Validate(req *dtos.UpdateQuotationRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                        []string{"required", "exists:quotations,id"},
		"customer_id":               []string{"required", "exists:customers,id"},
		"order_type_id":             []string{"required", "exists:mix_values,id"},
		"currency_id":               []string{"required", "exists:mix_values,id"},
		"status":                    []string{},
		"expired_at":                []string{"date:yyyy-MM-dd"},
		"is_vat":                    []string{"numeric"},
		"is_pph23":                  []string{"numeric"},
		"quo_dts":                   []string{"array"},
		"quo_dts.*.id":              []string{"exists:quo_dts,id"},
		"quo_dts.*.product_item_id": []string{"exists:products,id"},
		"quo_dts.*.item_unit_id":    []string{"exists:item_units,id"},
		"quo_dts.*.qty":             []string{"numeric"},
		"quo_dts.*.gen_code":        []string{},
		"quo_dts.*.remark":          []string{},
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
