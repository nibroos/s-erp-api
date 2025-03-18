package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// SalesOrderUpdateRequest handles the validation for the RegisterRequest.
type SalesOrderUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of SalesOrderUpdateRequest.
func NewSalesOrderUpdateRequest() *SalesOrderUpdateRequest {

	return &SalesOrderUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *SalesOrderUpdateRequest) Validate(req *dtos.UpdateSalesOrderRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                       []string{"required", "exists:products,id"},
		"unit_id":                  []string{"exists:mix_values,id"},
		"status":                   []string{},
		"expired_at":               []string{"date:yyyy-MM-dd"},
		"so_dts":                   []string{"array"},
		"so_dts.*.id":              []string{"exists:so_dts,id"},
		"so_dts.*.product_item_id": []string{"exists:products,id"},
		"so_dts.*.item_unit_id":    []string{"exists:item_units,id"},
		"so_dts.*.qty":             []string{"numeric"},
		"so_dts.*.gen_code":        []string{},
		"so_dts.*.remark":          []string{},
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
