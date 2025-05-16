package form_requests

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/validators"
)

// SendSolutionEmailRequest handles the validation for the RegisterRequest.
type SendSolutionEmailRequest struct {
}

// NewRegisterUpdateRequest creates a new instance of SendSolutionEmailRequest.
func NewSendSolutionEmailRequest() *SendSolutionEmailRequest {

	return &SendSolutionEmailRequest{}
}

// Validate validates the RegisterRequest.
func (r *SendSolutionEmailRequest) Validate(req *dtos.FormTicketRequest, ctx *fiber.Ctx) (map[string][]string, bool) {
	rules := map[string][]string{
		"email": []string{"required"},
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
