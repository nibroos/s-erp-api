package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// BranchItemUpdateRequest handles the validation for the RegisterRequest.
type BranchItemUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of BranchItemUpdateRequest.
func NewBranchItemUpdateRequest() *BranchItemUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &BranchItemUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *BranchItemUpdateRequest) Validate(req *dtos.UpdateBranchItemRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"id":            []string{"required"},
		"name":          []string{},
		"specification": []string{},
		"description":   []string{},
		"tpb_code":      []string{},
		"price_sell":    []string{"float"},
		"price_buy":     []string{"float"},
		"minimum_stock": []string{"float"},
		"status":        []string{"required"},
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
