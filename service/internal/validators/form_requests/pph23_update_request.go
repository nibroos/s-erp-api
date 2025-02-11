package form_requests

import (
	"context"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// Pph23UpdateRequest handles the validation for the RegisterRequest.
type Pph23UpdateRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterUpdateRequest creates a new instance of Pph23UpdateRequest.
func NewPph23UpdateRequest() *Pph23UpdateRequest {
	v := govalidator.New(govalidator.Options{})
	return &Pph23UpdateRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *Pph23UpdateRequest) Validate(req *dtos.UpdatePph23Request, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		// "name":        []string{"required", fmt.Sprintf("unique_ig:mix_values,name,%d", req.ID)},
		"name":        []string{"required"},
		"num":         []string{"numeric"},
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
