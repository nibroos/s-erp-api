package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type InvoiceDpStoreRequest struct {
}

func NewInvoiceDpStoreRequest() *InvoiceDpStoreRequest {
	return &InvoiceDpStoreRequest{}
}

func (r *InvoiceDpStoreRequest) Validate(req *dtos.CreateInvoiceDpRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"customer_id":  []string{"required", "exists:customers,id"},
		"currency_id":  []string{"required", "exists:mix_values,id"},
		"title":        []string{"required"},
		"invoice_date": []string{"required", "date:yyyy-MM-dd"},
		"due_date":     []string{"required", "date:yyyy-MM-dd"},
		// "dp_percentage":               []string{"required", "numeric"},
		"invoice_dp_dts":              []string{"required", "array"},
		"invoice_dp_dts.*.product_id": []string{"required"},
		"invoice_dp_dts.*.ref_id":     []string{"required"},
		"invoice_dp_dts.*.ref_type":   []string{"required"},
		"invoice_dp_dts.*.qty":        []string{"required", "numeric"},
		"invoice_dp_dts.*.price":      []string{"required", "numeric"},
	}

	customFieldNames := map[string]string{
		"invoice_dp_dts": "item sales order",
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
