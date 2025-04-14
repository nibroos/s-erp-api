package form_requests

import (
	"encoding/json"

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
		"po_buyer_no":       []string{},
		"status":            []string{"required"},
		"shipping_at":       []string{"date:yyyy-MM-dd"},
		"agree_at":          []string{"date:yyyy-MM-dd"},
		"is_vat":            []string{"numeric"},
		"is_pph23":          []string{"numeric"},
		"due_at":            []string{"date:yyyy-MM-dd"},
		"so_dts":            []string{"array"},
		"so_dts.*.item_id":  []string{"required", "exists:products,id"},
		"so_dts.*.ref_id":   []string{"required"},
		"so_dts.*.ref_type": []string{"required"},
		"so_dts.*.qty":      []string{"required", "numeric"},
	}

	customFieldNames := map[string]string{}

	var requestBody map[string]interface{}

	form := ctx.FormValue("data")

	if err := json.Unmarshal([]byte(form), &requestBody); err != nil {
		return map[string][]string{"error": {"Invalid JSON format"}}, false
	}

	// Convert empty strings to null in the map
	for key, value := range requestBody {
		if str, ok := value.(string); ok && str == "" {
			requestBody[key] = nil
		}
	}

	requestBody = make(map[string]interface{})
	// Parse the data string as JSON when present
	if err := json.Unmarshal([]byte(form), &requestBody); err != nil {
		return map[string][]string{"error": {"Invalid JSON format"}}, false
	}

	// Convert empty strings to null in the map
	for key, value := range requestBody {
		if str, ok := value.(string); ok && str == "" {
			requestBody[key] = nil
		}
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
