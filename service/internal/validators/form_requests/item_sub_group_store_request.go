package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// ItemSubGroupStoreRequest handles the validation for the RegisterRequest.
type ItemSubGroupStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of ItemSubGroupStoreRequest.
func NewItemSubGroupStoreRequest() *ItemSubGroupStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &ItemSubGroupStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *ItemSubGroupStoreRequest) Validate(req *dtos.CreateItemSubGroupRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		// "name":        []string{"required", "unique:mix_values,name"},
		"name":          []string{"required"},
		"item_group_id": []string{"required"},
		"description":   []string{},
		"remarks":       []string{},
		"status":        []string{},
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
