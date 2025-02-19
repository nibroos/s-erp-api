package form_requests

import (
	"context"
	"fmt"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// OrderTypeUpdateRequest handles the validation for the RegisterRequest.
type OrderTypeUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of OrderTypeUpdateRequest.
func NewOrderTypeUpdateRequest() *OrderTypeUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &OrderTypeUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *OrderTypeUpdateRequest) Validate(req *dtos.UpdateOrderTypeRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"name":        []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
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
