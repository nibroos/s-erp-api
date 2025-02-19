package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// VatStoreRequest handles the validation for the RegisterRequest.
type VatStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of VatStoreRequest.
func NewVatStoreRequest() *VatStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &VatStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *VatStoreRequest) Validate(req *dtos.CreateVatRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		"name":        []string{"required", "unique:mix_values,name"},
		"num":         []string{"float"},
		"description": []string{},
		"remarks":     []string{},
		"status":      []string{},
		"divider":     []string{"float"},
		"multiplier":  []string{"float"},
		"changed_at":  []string{},
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
