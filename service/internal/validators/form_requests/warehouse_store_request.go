package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type WarehouseStoreRequest struct {
}

func NewWarehouseStoreRequest() *WarehouseStoreRequest {
	return &WarehouseStoreRequest{}
}

func (r *WarehouseStoreRequest) Validate(req *dtos.CreateWarehouseRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"name":        []string{"required", "unique:mix_values,name"},
		"description": []string{},
		"remarks":     []string{},
		"status":      []string{},
		"code":        []string{},
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
