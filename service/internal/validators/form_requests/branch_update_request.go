package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// BranchUpdateRequest handles the validation for the RegisterRequest.
type BranchUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of BranchUpdateRequest.
func NewBranchUpdateRequest() *BranchUpdateRequest {

	return &BranchUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *BranchUpdateRequest) Validate(req *dtos.UpdateBranchRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		// "name":        []string{"required"},
		"name":        []string{"required"},
		"owner_name":  []string{},
		"sign_name":   []string{},
		"address":     []string{},
		"phone":       []string{},
		"email":       []string{},
		"website":     []string{},
		"sign":        []string{},
		"logo":        []string{},
		"description": []string{},
		"remark":      []string{},
		"status":      []string{},
	}

	customFieldNames := map[string]string{}

	var requestBody map[string]interface{}

	if err := ctx.BodyParser(&requestBody); err != nil {
		form, err := ctx.MultipartForm()
		if err != nil {
			return map[string][]string{"error": {"Invalid request body or form-data"}}, false
		}
		requestBody = make(map[string]interface{})
		for key, values := range form.Value {
			if len(values) > 0 {
				requestBody[key] = values[0]
			}
		}
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
