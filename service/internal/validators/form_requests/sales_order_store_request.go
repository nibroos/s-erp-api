package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// SalesOrderStoreRequest handles the validation for the RegisterRequest.
type SalesOrderStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of SalesOrderStoreRequest.
func NewSalesOrderStoreRequest() *SalesOrderStoreRequest {

	return &SalesOrderStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *SalesOrderStoreRequest) Validate(req *dtos.CreateSalesOrderRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"customer_id":       []string{"required", "exists:customers,id"},
		"order_type_id":     []string{"required", "exists:mix_values,id"},
		"currency_id":       []string{"required", "exists:mix_values,id"},
		"po_buyer_no":       []string{"required"},
		"status":            []string{"required"},
		"shipping_at":       []string{"date:yyyy-MM-dd"},
		"agree_at":          []string{"date:yyyy-MM-dd"},
		"due_at":            []string{"date:yyyy-MM-dd"},
		"expired_at":        []string{"date:yyyy-MM-dd"},
		"so_dts":            []string{"array"},
		"so_dts.*.item_id":  []string{"required", "exists:products,id"},
		"so_dts.*.ref_id":   []string{"required"},
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
