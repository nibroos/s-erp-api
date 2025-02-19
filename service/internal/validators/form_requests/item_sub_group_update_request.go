package form_requests

import (
	"context"
	"fmt"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// ItemSubGroupUpdateRequest handles the validation for the RegisterRequest.
type ItemSubGroupUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of ItemSubGroupUpdateRequest.
func NewItemSubGroupUpdateRequest() *ItemSubGroupUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &ItemSubGroupUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *ItemSubGroupUpdateRequest) Validate(req *dtos.UpdateItemSubGroupRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"name":          []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
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
