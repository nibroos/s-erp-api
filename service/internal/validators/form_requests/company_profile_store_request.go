package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CompanyProfileStoreRequest handles the validation for the RegisterRequest.
type CompanyProfileStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of CompanyProfileStoreRequest.
func NewCompanyProfileStoreRequest() *CompanyProfileStoreRequest {

	return &CompanyProfileStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *CompanyProfileStoreRequest) Validate(req *dtos.CreateCompanyProfileRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		// "name":        []string{"required"},
		"company_name":        []string{"required"},
		"company_address":     []string{},
		"company_phone":       []string{},
		"company_email":       []string{},
		"company_website":     []string{},
		"company_logo":        []string{},
		"company_description": []string{},
		"company_remark":      []string{},
		"company_status":      []string{},
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
