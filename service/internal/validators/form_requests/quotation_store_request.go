package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// QuotationStoreRequest handles the validation for the RegisterRequest.
type QuotationStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of QuotationStoreRequest.
func NewQuotationStoreRequest() *QuotationStoreRequest {

	return &QuotationStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *QuotationStoreRequest) Validate(req *dtos.CreateQuotationRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		// "item_sub_group_id":         []string{"required", "exists:mix_values,id"},
		// "unit_id":                   []string{"exists:mix_values,id"},
		// "item_unit_id":              []string{"required"},
		// "status":                    []string{},
		// "expired_at":                []string{"date:yyyy-MM-dd"},
		// "quo_dts":                   []string{"array"},
		// "quo_dts.*.product_item_id": []string{"exists:products,id"},
		// "quo_dts.*.item_unit_id":    []string{"exists:item_units,id"},
		// "quo_dts.*.qty":             []string{"numeric"},
		// "quo_dts.*.gen_code":        []string{},
		// "quo_dts.*.remark":          []string{},
		"customer_id":          []string{"required", "exists:customers,id"},
		"order_type_id":        []string{"required", "exists:mix_values,id"},
		"currency_id":          []string{"required", "exists:mix_values,id"},
		"quo_no":               []string{"required"},
		"title":                []string{"required"},
		"status":               []string{"required"},
		"expired_at":           []string{"date:yyyy-MM-dd"},
		"quo_dts":              []string{"array"},
		"quo_dts.*.product_id": []string{"required", "exists:products,id"},
		"quo_dts.*.item_id":    []string{"required", "exists:products,id"},
		"quo_dts.*.ref_id":     []string{"required"},
		"quo_dts.*.ref_type":   []string{"required"},
		"quo_dts.*.qty":        []string{"required", "numeric"},
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
