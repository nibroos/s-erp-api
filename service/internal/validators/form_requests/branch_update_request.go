package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// BranchUpdateRequest handles the validation for the RegisterRequest.
type BranchUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of BranchUpdateRequest.
func NewBranchUpdateRequest() *BranchUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &BranchUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *BranchUpdateRequest) Validate(req *dtos.UpdateBranchRequest, ctx context.Context) map[string]string {
	// utils.DD(req)
	rules := govalidator.MapData{
		// "name":        []string{"required"},
		"name":        []string{"required"},
		"owner_name":  []string{},
		"sign_name":   []string{},
		"address":     []string{},
		"phone":       []string{},
		"email":       []string{},
		"website":     []string{},
		"sign":        []string{},
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
