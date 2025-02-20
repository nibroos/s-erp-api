package form_requests

import (
	"context"
	"log"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/thedevsaddam/govalidator"
)

// CustomerStoreRequest handles the validation for the RegisterRequest.
type CustomerStoreRequest struct {
	Validator *govalidator.Validator
}

// NewRegisterStoreRequest creates a new instance of CustomerStoreRequest.
func NewCustomerStoreRequest() *CustomerStoreRequest {
	v := govalidator.New(govalidator.Options{})
	return &CustomerStoreRequest{Validator: v}
}

// Validate validates the RegisterRequest.
func (r *CustomerStoreRequest) Validate(req *dtos.CreateCustomerRequest, ctx context.Context) map[string]string {
	rules := govalidator.MapData{
		"customer_type_id": []string{"required", "numeric", "exists:mix_values,id"},
		"agent_id":         []string{"exists:customers,id"},
		"code":             []string{},
		"name":             []string{"required"},
		"address":          []string{},
		"phone":            []string{},
		"email":            []string{"email"},
		"pic":              []string{},
		"status":           []string{"required"},
	}

	log.Println("rules", rules)

	opts := govalidator.Options{
		Data:  req,
		Rules: rules,
	}

	log.Println("opts", opts)

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
