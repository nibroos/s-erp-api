package form_requests

import (
	"encoding/json"

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

	form := ctx.FormValue("data")

	if err := json.Unmarshal([]byte(form), &requestBody); err != nil {
		return map[string][]string{"error": {"Invalid JSON format"}}, false
	}

	// Convert empty strings to null in the map
	for key, value := range requestBody {
		if str, ok := value.(string); ok && str == "" {
			requestBody[key] = nil
		}
	}

	requestBody = make(map[string]interface{})
	// Parse the data string as JSON when present
	if err := json.Unmarshal([]byte(form), &requestBody); err != nil {
		return map[string][]string{"error": {"Invalid JSON format"}}, false
	}

	// Convert empty strings to null in the map
	for key, value := range requestBody {
		if str, ok := value.(string); ok && str == "" {
			requestBody[key] = nil
		}
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
