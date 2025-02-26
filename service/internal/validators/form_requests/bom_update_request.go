package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CatalogUpdateRequest handles the validation for the RegisterRequest.
type CatalogUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of CatalogUpdateRequest.
func NewCatalogUpdateRequest() *CatalogUpdateRequest {

	return &CatalogUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *CatalogUpdateRequest) Validate(req *dtos.UpdateCatalogRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                  []string{"required", "exists:catalogs,id"},
		"unit_id":             []string{"exists:mix_values,id"},
		"collection_id":       []string{"exists:mix_values,id"},
		"code":                []string{},
		"factory_code":        []string{},
		"name":                []string{"required"},
		"sku":                 []string{},
		"barcode":             []string{},
		"specification":       []string{},
		"description":         []string{},
		"remark":              []string{},
		"price_sell":          []string{"numeric"},
		"price_buy":           []string{"numeric"},
		"margin":              []string{"numeric"},
		"status":              []string{},
		"expired_at":          []string{},
		"boms":                []string{"array"},
		"boms.*.ms_item_id":   []string{"exists:ms_items,id"},
		"boms.*.item_unit_id": []string{"exists:item_units,id"},
		"boms.*.qty":          []string{"numeric"},
		"boms.*.remark":       []string{},
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
