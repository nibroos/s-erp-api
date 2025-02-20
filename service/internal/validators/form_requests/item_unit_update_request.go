package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// ItemUnitUpdateRequest handles the validation for the RegisterRequest.
type ItemUnitUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of ItemUnitUpdateRequest.
func NewItemUnitUpdateRequest() *ItemUnitUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &ItemUnitUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *ItemUnitUpdateRequest) Validate(req *dtos.UpdateItemUnitRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"id":         []string{"required"},
		"price_sell": []string{"float"},
		"price_buy":  []string{"float"},
		"conversion": []string{"float"},
		"status":     []string{"required"},
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
