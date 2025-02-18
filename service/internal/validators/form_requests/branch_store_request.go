package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// BranchStoreRequest handles the validation for the RegisterRequest.
type BranchStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of BranchStoreRequest.
func NewBranchStoreRequest() *BranchStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &BranchStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *BranchStoreRequest) Validate(req *dtos.CreateBranchRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		// "name":        []string{"required"},
		"name":        []string{"required"},
		"address":     []string{},
		"phone":       []string{},
		"email":       []string{},
		"website":     []string{},
		"logo":        []string{},
		"description": []string{},
		"remark":      []string{},
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
