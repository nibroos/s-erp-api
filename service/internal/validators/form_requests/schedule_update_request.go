package form_requests

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ScheduleUpdateRequest handles the validation for the RegisterRequest.
type ScheduleUpdateRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of ScheduleUpdateRequest.
func NewScheduleUpdateRequest() *ScheduleUpdateRequest {

	return &ScheduleUpdateRequest{}
}

// Validate validates the RegisterRequest.
func (r *ScheduleUpdateRequest) Validate(req *dtos.UpdateSalesOrderScheduleRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		// "id":             []string{"required", "exists:schedules,id"},
		"assignee_id":    []string{"exists:users,id"},
		"customer_id":    []string{"exists:customers,id"},
		"sales_order_id": []string{},
		"uuid":           []string{},
		"steps_id":       []string{},
		"title":          []string{},
		"remark":         []string{},
		"status":         []string{},
		// "start_at":       []string{"date_format:Y-m-d"},
		// "end_at":         []string{"date_format:Y-m-d"},
		"color": []string{},

		// "steps":               []string{"array"},
		// "steps.*.id":          []string{"required", "exists:schedule_tasks,id"},
		// "steps.*.assignee_id": []string{"exists:users,id"},
		// "steps.*.parent_id":   []string{},
		// "steps.*.entity_id":   []string{},
		// "steps.*.entity_type": []string{},
		// "steps.*.uuid":        []string{},
		// "steps.*.parent_uuid": []string{},
		// "steps.*.title":       []string{"required", "string"},
		// "steps.*.remark":      []string{},
		// "steps.*.order_item":  []string{},
		// "steps.*.color":       []string{},
		// "steps.*.is_checked":  []string{},
		// "steps.*.start_at":    []string{"date_format:Y-m-d H:i:s"},
		// "steps.*.end_at":      []string{"date_format:Y-m-d H:i:s"},
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
