package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ScheduleStoreRequest handles the validation for the RegisterRequest.
type ScheduleStoreRequest struct {
}

// NewRegisterStoreRequest creates a new instance of ScheduleStoreRequest.
func NewScheduleStoreRequest() *ScheduleStoreRequest {

	return &ScheduleStoreRequest{}
}

// Validate validates the RegisterRequest.
func (r *ScheduleStoreRequest) Validate(req *dtos.CreateScheduleNoRefRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"customer_id": []string{"exists:customers,id"},
		"assignee_id": []string{"exists:users,id"},
		"title":       []string{},
		"start_at":    []string{"date:yyyy-MM-dd"},
		"end_at":      []string{"date:yyyy-MM-dd"},
		"color":       []string{},
	}

	customFieldNames := map[string]string{}

	var requestBody map[string]interface{}
	if err := ctx.BodyParser(&requestBody); err != nil {
		return map[string][]string{"error": {"Invalid request body"}}, false
	}
	request := validators.NewRequest(rules, requestBody, customFieldNames)
	errors, valid := request.Validate()

	convertedErrors := make(map[string][]string)
	for key, value := range errors {
		if len(value) > 0 {
			convertedErrors[key] = value
		}
	}
	return convertedErrors, valid
}
