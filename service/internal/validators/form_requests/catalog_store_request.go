package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CatalogStoreRequest handles the validation for the RegisterRequest.
type CatalogStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of CatalogStoreRequest.
func NewCatalogStoreRequest() *CatalogStoreRequest {

	return &CatalogStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *CatalogStoreRequest) Validate(req *dtos.CreateCatalogRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"unit_id":             []string{"exists:mix_values,id"},
		"collection_id":       []string{"exists:mix_values,id"},
		"branch_id":           []string{"exists:branches,id"},
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
