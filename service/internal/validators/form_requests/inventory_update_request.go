package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// InventoryUpdateRequest handles the validation for the RegisterRequest.
type InventoryUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of InventoryUpdateRequest.
func NewInventoryUpdateRequest() *InventoryUpdateRequest {

	return &InventoryUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *InventoryUpdateRequest) Validate(req *dtos.FormInventoryRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                []string{"required", "exists:sales_orders,id"},
		"customer_id":       []string{"exists:customers,id"},
		"currency_id":       []string{"required", "exists:mix_values,id"},
		"warehouse_id":      []string{"required", "exists:mix_values,id"},
		"io_type_id":        []string{"required", "exists:mix_values,id"},
		"status":            []string{"required"},
		"do_at":             []string{"date:yyyy-MM-dd"},
		"ingoing_at":        []string{"date:yyyy-MM-dd"},
		"invoice_at":        []string{"date:yyyy-MM-dd"},
		"is_vat":            []string{"numeric"},
		"is_pph23":          []string{"numeric"},
		"so_dts":            []string{"array"},
		"so_dts.*.item_id":  []string{"required", "exists:products,id"},
		"so_dts.*.ref_type": []string{"required"},
		"so_dts.*.qty":      []string{"required", "numeric"},
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
