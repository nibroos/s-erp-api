package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type SalesInvoiceUpdateRequest struct {
}

func NewSalesInvoiceUpdateRequest() *SalesInvoiceUpdateRequest {
	return &SalesInvoiceUpdateRequest{}
}

func (r *SalesInvoiceUpdateRequest) Validate(req *dtos.UpdateSalesInvoiceRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"id":                             []string{"required", "exists:sales_invoices,id"},
		"customer_id":                    []string{"required", "exists:customers,id"},
		"currency_id":                    []string{"required", "exists:mix_values,id"},
		"title":                          []string{"required"},
		"invoice_date":                   []string{"required", "date:yyyy-MM-dd"},
		"due_date":                       []string{"required", "date:yyyy-MM-dd"},
		"sales_invoice_dts":              []string{"required", "array"},
		"sales_invoice_dts.*.product_id": []string{"required"},
		"sales_invoice_dts.*.ref_id":     []string{"required"},
		"sales_invoice_dts.*.ref_type":   []string{"required"},
		"sales_invoice_dts.*.qty":        []string{"required", "numeric"},
		"sales_invoice_dts.*.price":      []string{"required", "numeric"},
	}

	customFieldNames := map[string]string{
		"sales_invoice_dts": "item sales order",
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
