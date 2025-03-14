package form_requests

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

type WarehouseUpdateRequest struct {
}

func NewWarehouseUpdateRequest() *WarehouseUpdateRequest {
	return &WarehouseUpdateRequest{}
}

func (r *WarehouseUpdateRequest) Validate(req *dtos.UpdateWarehouseRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"name":        []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
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
