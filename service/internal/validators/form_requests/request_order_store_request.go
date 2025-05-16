package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type RequestOrderStoreRequest struct {
}

func NewRequestOrderStoreRequest() *RequestOrderStoreRequest {
	return &RequestOrderStoreRequest{}
}

func (r *RequestOrderStoreRequest) Validate(req *dtos.CreateRequestOrderRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"warehouse_id":                 []string{"required", "exists:mix_values,id"},
		"request_date":                 []string{"required", "date:yyyy-MM-dd"},
		"request_order_dts":            []string{"required", "array"},
		"request_order_dts.*.ref_type": []string{"required"},
		"request_order_dts.*.req_qty":  []string{"required", "numeric"},
	}

	customFieldNames := map[string]string{
		"request_order_dts": "item request order",
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
