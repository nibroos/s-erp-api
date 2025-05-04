package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type PurchaseOrderUpdateRequest struct {
}

func NewPurchaseOrderUpdateRequest() *PurchaseOrderUpdateRequest {
	return &PurchaseOrderUpdateRequest{}
}

func (r *PurchaseOrderUpdateRequest) Validate(req *dtos.UpdatePurchaseOrderRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                  []string{"required", "exists:purchase_orders,id"},
		"purchase_type_id":    []string{"required", "exists:mix_values,id"},
		"currency_id":         []string{"required", "exists:mix_values,id"},
		"po_no":               []string{},
		"po_date":             []string{"required", "date:yyyy-MM-dd"},
		"delivery_date":       []string{"date:yyyy-MM-dd"},
		"status":              []string{"required"},
		"po_dts":              []string{"array"},
		"po_dts.*.product_id": []string{"required", "exists:products,id"},
		"po_dts.*.qty":        []string{"numeric"},
		"po_dts.*.gen_code":   []string{},
		"po_dts.*.remark":     []string{},
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
