package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// CurrencyUpdateRequest handles the validation for the RegisterRequest.
type CurrencyUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of CurrencyUpdateRequest.
func NewCurrencyUpdateRequest() *CurrencyUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &CurrencyUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *CurrencyUpdateRequest) Validate(req *dtos.UpdateCurrencyRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		// "name":        []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
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
