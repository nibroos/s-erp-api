package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// BranchItemStoreRequest handles the validation for the RegisterRequest.
type BranchItemStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of BranchItemStoreRequest.
func NewBranchItemStoreRequest() *BranchItemStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &BranchItemStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *BranchItemStoreRequest) Validate(req *dtos.CreateBranchItemRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"ms_item_id":    []string{"required", "exists:ms_items,id"},
		"branch_id":     []string{"required", "exists:branches,id"},
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
