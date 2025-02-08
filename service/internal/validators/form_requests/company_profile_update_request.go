package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// CompanyProfileUpdateRequest handles the validation for the RegisterRequest.
type CompanyProfileUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of CompanyProfileUpdateRequest.
func NewCompanyProfileUpdateRequest() *CompanyProfileUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &CompanyProfileUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *CompanyProfileUpdateRequest) Validate(req *dtos.UpdateCompanyProfileRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
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

	opts := govalidator.Options{
		Data:  req,
		Rules: rules,
	}

	v := govalidator.New(opts)
	mappedErrors := v.ValidateStruct()

	if len(mappedErrors) == 0 {
		return nil
	}

	errors := make(map[string]string)
	for field, err := range mappedErrors {
		errors[field] = err[0]
	}
	return errors
}
