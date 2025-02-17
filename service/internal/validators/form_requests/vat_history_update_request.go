package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// VatHistoryUpdateRequest handles the validation for the RegisterRequest.
type VatHistoryUpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of VatHistoryUpdateRequest.
func NewVatHistoryUpdateRequest() *VatHistoryUpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &VatHistoryUpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *VatHistoryUpdateRequest) Validate(req *dtos.UpdateVatHistoryRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"num":        []string{"float"},
		"remark":     []string{},
		"status":     []string{"float"},
		"divider":    []string{"float"},
		"multiplier": []string{"float"},
		"changed_at": []string{},
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
