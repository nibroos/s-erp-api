package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// BranchStoreRequest handles the validation for the RegisterRequest.
type BranchStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of BranchStoreRequest.
func NewBranchStoreRequest() *BranchStoreRequest {

	return &BranchStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *BranchStoreRequest) Validate(req *dtos.CreateBranchRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		// "name":        []string{"required"},
		"name":        []string{"required"},
		"address":     []string{},
		"phone":       []string{},
		"email":       []string{},
		"website":     []string{},
		"logo":        []string{},
		"description": []string{},
		"remark":      []string{},
		"status":      []string{},
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
