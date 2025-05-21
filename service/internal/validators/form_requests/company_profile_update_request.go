package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// CompanyProfileUpdateRequest handles the validation for the RegisterRequest.
type CompanyProfileUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of CompanyProfileUpdateRequest.
func NewCompanyProfileUpdateRequest() *CompanyProfileUpdateRequest {

	return &CompanyProfileUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *CompanyProfileUpdateRequest) Validate(req *dtos.UpdateCompanyProfileRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	// utils.DD(req)
	rules := map[string][]string{
		// "name":        []string{"required"},
		"company_name":       []string{"required"},
		"company_owner_name": []string{},
		"company_sign_name":  []string{},
		"company_address":    []string{},
		"company_phone":      []string{},
		// "company_email":          []string{"required"},
		// "company_email_password": []string{"required"},
		"company_website":     []string{},
		"company_sign":        []string{},
		"company_logo":        []string{},
		"company_description": []string{},
		"company_remark":      []string{},
		"company_status":      []string{},
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
