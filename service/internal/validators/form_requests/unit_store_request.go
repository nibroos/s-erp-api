package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// UnitStoreRequest handles the validation for the RegisterRequest.
type UnitStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of UnitStoreRequest.
func NewUnitStoreRequest() *UnitStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &UnitStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *UnitStoreRequest) Validate(req *dtos.CreateUnitRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		"name":        []string{"required", "unique:mix_values,name"},
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
