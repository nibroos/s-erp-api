package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type InvoiceMaintenanceStoreRequest struct {
}

func NewInvoiceMaintenanceStoreRequest() *InvoiceMaintenanceStoreRequest {
	return &InvoiceMaintenanceStoreRequest{}
}

func (r *InvoiceMaintenanceStoreRequest) Validate(req *dtos.CreateInvoiceMaintenanceRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"customer_id":                          []string{"required", "exists:customers,id"},
		"currency_id":                          []string{"required", "exists:mix_values,id"},
		"invoice_date":                         []string{"required", "date:yyyy-MM-dd"},
		"invoice_maintenance_dts":              []string{"required", "array"},
		"invoice_maintenance_dts.*.product_id": []string{"required"},
		"invoice_maintenance_dts.*.ref_id":     []string{"required"},
		"invoice_maintenance_dts.*.ref_type":   []string{"required"},
		"invoice_maintenance_dts.*.qty":        []string{"required", "numeric"},
		"invoice_maintenance_dts.*.price":      []string{"required", "numeric"},
	}

	customFieldNames := map[string]string{
		"invoice_maintenance_dts": "item sales order",
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
