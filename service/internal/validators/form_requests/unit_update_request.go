package form_requests

import (
	"context"
	"fmt"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// UnitUpdateRequest handles the validation for the RegisterRequest.
type UnitUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of UnitUpdateRequest.
func NewUnitUpdateRequest() *UnitUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &UnitUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *UnitUpdateRequest) Validate(req *dtos.UpdateUnitRequest, ctx context.Context) map[string]string {
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
