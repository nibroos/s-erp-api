package form_requests

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ClosingStockStoreRequest handles the validation for the RegisterRequest.
type ClosingStockStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of ClosingStockStoreRequest.
func NewClosingStockStoreRequest() *ClosingStockStoreRequest {

	return &ClosingStockStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *ClosingStockStoreRequest) Validate(req *dtos.FormClosingStockStoreRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"start_closing_at": []string{"date:yyyy-MM-dd", fmt.Sprintf("before_date:%s", req.EndClosingAt)},
		"end_closing_at":   []string{"required", "date:yyyy-MM-dd"},
		"password":         []string{"required"},
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
