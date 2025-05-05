package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type InvoiceAdjustmentStoreRequest struct {
}

func NewInvoiceAdjustmentStoreRequest() *InvoiceAdjustmentStoreRequest {
	return &InvoiceAdjustmentStoreRequest{}
}

func (r *InvoiceAdjustmentStoreRequest) Validate(req *dtos.CreateInvoiceAdjustmentRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"customer_id":             []string{"required", "exists:customers,id"},
		"currency_id":             []string{"required", "exists:mix_values,id"},
		"bank_id":                 []string{"exists:bank_informations,id"},
		"payment_date":            []string{"required", "date:yyyy-MM-dd"},
		"payment_amount":          []string{"required", "numeric"},
		"exchange_rate":           []string{"numeric"},
		"adjustment_dts":          []string{"required", "array"},
		"adjustment_dts.*.ref_id": []string{"required"},
	}

	customFieldNames := map[string]string{
		"adjustment_dts": "invoice adjustment items",
	}

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
