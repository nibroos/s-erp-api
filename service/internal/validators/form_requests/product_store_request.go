package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ProductStoreRequest handles the validation for the RegisterRequest.
type ProductStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of ProductStoreRequest.
func NewProductStoreRequest() *ProductStoreRequest {

	return &ProductStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *ProductStoreRequest) Validate(req *dtos.CreateProductRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"item_sub_group_id":      []string{"required", "exists:mix_values,id"},
		"unit_id":                []string{"exists:mix_values,id"},
		"code":                   []string{},
		"factory_code":           []string{},
		"name":                   []string{"required"},
		"sku":                    []string{},
		"barcode":                []string{},
		"specification":          []string{},
		"description":            []string{},
		"tpb_code":               []string{},
		"remark":                 []string{},
		"price_sell":             []string{"numeric"},
		"price_buy":              []string{"numeric"},
		"margin":                 []string{"numeric"},
		"minimum_stock":          []string{"numeric"},
		"status":                 []string{},
		"item_unit_id":           []string{"required"},
		"expired_at":             []string{"date:yyyy-MM-dd"},
		"units":                  []string{"required", "array"},
		"units.*.unit_id":        []string{"required", "exists:mix_values,id"},
		"units.*.conversion":     []string{"required", "numeric"},
		"units.*.price_sell":     []string{"required", "numeric"},
		"units.*.price_buy":      []string{"required", "numeric"},
		"boms":                   []string{"array"},
		"boms.*.product_item_id": []string{"exists:products,id"},
		"boms.*.item_unit_id":    []string{"exists:item_units,id"},
		"boms.*.qty":             []string{"numeric"},
		"boms.*.remark":          []string{},
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
