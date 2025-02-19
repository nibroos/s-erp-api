package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// Pph23StoreRequest handles the validation for the RegisterRequest.
type Pph23StoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of Pph23StoreRequest.
func NewPph23StoreRequest() *Pph23StoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &Pph23StoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *Pph23StoreRequest) Validate(req *dtos.CreatePph23Request, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		"name":        []string{"required", "unique:mix_values,name"},
		"num":         []string{"numeric"},
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
