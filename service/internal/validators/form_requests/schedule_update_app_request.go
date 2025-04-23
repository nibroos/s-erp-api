package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// ScheduleUpdateAppRequest handles the validation for the RegisterRequest.
type ScheduleUpdateAppRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of ScheduleUpdateAppRequest.
func NewScheduleUpdateAppRequest() *ScheduleUpdateAppRequest {

	return &ScheduleUpdateAppRequest{}
}

// Validate validates the RegisterRequest.
func (r *ScheduleUpdateAppRequest) Validate(req *dtos.UpdateSalesOrderScheduleAppRequest, ctx *fiber.Ctx) (map[string][]string, bool) {

	rules := map[string][]string{
		"sales_order_id":       []string{"exists:sales_orders,id"},
		"tasks":                []string{"array"},
		"tasks.*.id":           []string{"required", "exists:schedule_tasks,id"},
		"tasks.*.assignee_id":  []string{"exists:users,id"},
		"tasks.*.title":        []string{},
		"tasks.*.remark":       []string{},
		"tasks.*.is_checked":   []string{},
		"attachments":          []string{"array"},
		"attachments.*.id":     []string{},
		"attachments.*.name":   []string{},
		"attachments.*.remark": []string{},
		"deleted_files":        []string{"array"},
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
