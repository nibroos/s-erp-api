package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jLog "github.com/opentracing/opentracing-go/log"
)

// ErrorHandler middleware
func ErrorHandler(ctx *fiber.Ctx, err error) error {
	// Default to 500 Internal Server Error
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	message := "Internal server error"

	if err == sql.ErrNoRows {
		// code = http.StatusNotFound
		message = "No result found"
	} else if e, ok := err.(*fiber.Error); ok {
		// Use Fiber's default error message
		code = e.Code
		message = e.Message
	}

	// Capture the stack trace
	_, file, line, _ := runtime.Caller(1)
	stackTrace := fmt.Sprintf("%s:%d", file, line)

	// Log the error and stack trace
	log.Printf("[ERROR] %v\nStack Trace: %s\n", err, stackTrace)

	// Return a JSON response with the error
	return ctx.Status(code).JSON(fiber.Map{
		"status":  code,
		"message": message,
		"errors":  err.Error(),
		// "stack":   stackTrace, // Optionally include stack trace
	})
	// return fiber.NewError(code, message)
}
func ConvertRequestToFilters() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Get("Content-Type") == "application/json; charset=utf-8" {
			ctx.Set("Content-Type", "application/json")
		}

		log.Println("ConvertRequestToFilters-Content-Type:", ctx.Get("Content-Type"))

		// Check if the content type is JSON
		// application/json or application/json; charset=utf-8
		if strings.Contains(ctx.Get("Content-Type"), "application/json") {
			// if ctx.Get("Content-Type") == "application/json" {
			var requestBody map[string]interface{}
			if err := ctx.BodyParser(&requestBody); err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Failed to parse request body",
					"status":  "error",
					"err":     err.Error(),
				})
			}

			filters := make(map[string]string)
			for key, value := range requestBody {
				switch v := value.(type) {
				case string:
					if v == "" {
						requestBody[key] = nil
					} else {
						filters[key] = v
					}
				case int:
					filters[key] = strconv.Itoa(v)
				case float64:
					filters[key] = strconv.FormatFloat(v, 'f', -1, 64)
				// case array
				case []interface{}:
					var arr []string
					for _, val := range v {
						switch valType := val.(type) {
						case string:
							arr = append(arr, valType)
						case int:
							arr = append(arr, strconv.Itoa(valType))
						case float64:
							arr = append(arr, strconv.FormatFloat(valType, 'f', -1, 64))
						}
					}
					filters[key] = strings.Join(arr, ",")
				// case object
				case map[string]interface{}:
					jsonValue, err := json.Marshal(v)
					if err != nil {
						log.Printf("Failed to marshal object for key %s: %v", key, err)
					} else {
						filters[key] = string(jsonValue)
					}
				default:
					log.Printf("Unsupported type for key %s: %T", key, v)
				}
			}

			// Marshal the modified body back to JSON
			modifiedBody, err := json.Marshal(requestBody)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request body"})
			}

			// Replace the request body with the modified body
			ctx.Request().SetBody(modifiedBody)

			ctx.Locals("filters", filters)
		}

		// Check if the content type is multipart/form-data
		if ctx.Get("Content-Type") == "multipart/form-data" {
			form, err := ctx.MultipartForm()
			if err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Failed to parse multipart form",
					"status":  "error",
					"err":     err.Error(),
				})
			}

			filters := make(map[string]string)
			for key, values := range form.Value {
				if len(values) > 0 {
					filters[key] = values[0]
				}
			}

			// Convert form values to JSON
			requestBody := make(map[string]interface{})
			for key, values := range form.Value {
				if len(values) > 0 {
					requestBody[key] = values[0]
				}
			}

			// Marshal the modified body back to JSON
			modifiedBody, err := json.Marshal(requestBody)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request body"})
			}

			// Replace the request body with the modified body
			ctx.Request().SetBody(modifiedBody)

			ctx.Locals("filters", filters)
		}
		// else {

		return ctx.Next()
	}
}

func ConvertEmptyStringsToNull() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Parse the request body into a map
		var body map[string]interface{}

		if strings.Contains(ctx.Get("Content-Type"), "application/json") {
			ctx.Set("Content-Type", "application/json")
		}

		// log.Println("ctxBody", ctx.Body())

		log.Println("ConvertEmptyStringsToNull-Content-Type:", ctx.Get("Content-Type"))

		// if ctx.Get("Content-Type") == "application/json" {
		if strings.Contains(ctx.Get("Content-Type"), "application/json") {
			if err := json.Unmarshal(ctx.Body(), &body); err != nil {
				log.Println("convertEmptyStringsToNull-Error:", err)
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body, application/json"})
			}

			// Convert empty strings to null
			for key, value := range body {
				if str, ok := value.(string); ok && str == "" {
					body[key] = nil
				}
			}

			// Marshal the modified body back to JSON
			modifiedBody, err := json.Marshal(body)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request body, application/json"})
			}

			// Replace the request body with the modified body
			ctx.Request().SetBody(modifiedBody)
		} else if ctx.Get("Content-Type") == "multipart/form-data" {
			form, err := ctx.MultipartForm()
			if err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse multipart form"})
			}

			// Convert form values to JSON
			body = make(map[string]interface{})
			for key, values := range form.Value {
				if len(values) > 0 {
					if values[0] == "" {
						body[key] = nil
					} else {
						body[key] = values[0]
					}
				}
			}

			// Marshal the modified body back to JSON
			modifiedBody, err := json.Marshal(body)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request body, multipart/form-data"})
			}

			// Replace the request body with the modified body
			ctx.Request().SetBody(modifiedBody)
		}

		return ctx.Next()
	}
}

// PermissionMiddleware checks if the user has the required permission
func PermissionMiddleware(requiredPermission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !utils.HasPermission(ctx, requiredPermission) {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
		}
		return ctx.Next()
	}
}

func ConvertToClientTimezone() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientTimezone := c.Get("X-Client-Timezone", "Asia/Jakarta")
		location, err := time.LoadLocation(clientTimezone)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid timezone"})
		}

		// Convert start_at and end_at if they exist in the request body
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			log.Println("convertToClientTimezone-Error:", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if startAt, ok := body["start_at"].(string); ok {
			if parsedTime, err := time.ParseInLocation("2006-01-02 15:04", startAt, location); err == nil {
				body["start_at"] = parsedTime
			}
		}

		if endAt, ok := body["end_at"].(string); ok {
			if parsedTime, err := time.ParseInLocation("2006-01-02 15:04", endAt, location); err == nil {
				body["end_at"] = parsedTime
			}
		}

		// Replace the request body with the modified one
		modifiedBody, err := json.Marshal(body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process request body"})
		}
		c.Request().SetBody(modifiedBody)

		return c.Next()
	}
}

func JaegerTracingMiddleware(tracer opentracing.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Call the next handler
		err := c.Next()

		if err != nil || c.Response().StatusCode() >= 400 {
			span := tracer.StartSpan(c.Path())
			defer span.Finish()

			// Add custom tag to indicate middleware-based tracing
			span.SetTag("type", "middleware")

			// Set standard HTTP tags
			ext.HTTPMethod.Set(span, c.Method())
			ext.HTTPUrl.Set(span, c.Path())

			// // Capture the request body
			// var bodyBytes []byte
			// if c.Body() != nil {
			// 	bodyBytes = c.Body()
			// 	span.LogKV("request_body", string(bodyBytes)) // Log the request body
			// }

			// Pass the context with the span to the next handler
			ctx := opentracing.ContextWithSpan(c.Context(), span)
			c.SetUserContext(ctx)

			// Check for errors or 500 status code
			ext.Error.Set(span, true)

			// Log the error message
			if err != nil {
				span.LogFields(
					jLog.String("event", "error"),
					jLog.String("message", err.Error()),
				)
			}

			// Log the response body if available (e.g., error response)
			if c.Response().Body() != nil {
				responseBody := c.Response().Body()
				span.LogKV("response_body", string(responseBody))
				span.LogKV("request_body", string(c.Body()))
			}

			// Set the HTTP status code
			ext.HTTPStatusCode.Set(span, uint16(c.Response().StatusCode()))
		}

		return err
	}
}

func SafeHandler(handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic in handler: %v\n", err)
				c.Status(500).JSON(fiber.Map{
					"error":   true,
					"message": "Internal server error",
				})
			}
		}()
		return handler(c)
	}
}

func SafetyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Printf("Panic recovered: %v\nStack trace: %s\n", r, string(stack))

				// Return error response to client
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "Internal server error",
				})
			}
		}()

		return c.Next()
	}
}

func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVER] %v\n", r)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal Server Error",
				})
			}
		}()
		return c.Next()
	}
}
