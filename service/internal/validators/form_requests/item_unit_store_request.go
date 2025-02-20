package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// ItemUnitStoreRequest handles the validation for the RegisterRequest.
type ItemUnitStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of ItemUnitStoreRequest.
func NewItemUnitStoreRequest() *ItemUnitStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &ItemUnitStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *ItemUnitStoreRequest) Validate(req *dtos.CreateItemUnitRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"ms_item_id": []string{"required", "exists:ms_items,id"},
		"unit_id":    []string{"required", "exists:mix_values,id"},
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
