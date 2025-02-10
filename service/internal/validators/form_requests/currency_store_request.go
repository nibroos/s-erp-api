package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// CurrencyStoreRequest handles the validation for the RegisterRequest.
type CurrencyStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of CurrencyStoreRequest.
func NewCurrencyStoreRequest() *CurrencyStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &CurrencyStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *CurrencyStoreRequest) Validate(req *dtos.CreateCurrencyRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		// "name":        []string{"required", "unique:mix_values,name"},
		// "parent_id":   []string{"required"},
		"name":        []string{"required"},
		"description": []string{},
		"remarks":     []string{},
		"status":      []string{},
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
