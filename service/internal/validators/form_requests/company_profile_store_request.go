package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// CompanyProfileStoreRequest handles the validation for the RegisterRequest.
type CompanyProfileStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of CompanyProfileStoreRequest.
func NewCompanyProfileStoreRequest() *CompanyProfileStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &CompanyProfileStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *CompanyProfileStoreRequest) Validate(req *dtos.CreateCompanyProfileRequest, ctx context.Context) map[string]string {
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
