package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jLog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Meta struct {
	Total       int  `json:"total"`
	PerPage     int  `json:"per_page"`
	CurrentPage int  `json:"current_page"`
	LastPage    int  `json:"last_page"`
	NextPageUrl *int `json:"next_page_url"`
}

type Response struct {
	Data     interface{} `json:"data"`
	Meta     *Meta       `json:"meta,omitempty"`
	Message  string      `json:"message"`
	Status   int16       `json:"status"`
	Errors   interface{} `json:"errors"`
	Optional interface{} `json:"optional,omitempty"`
}

// Nullable is a generic type that can handle null values for different data types.
type Nullable[T any] struct {
	Value *T
}

// ContextKey is a type for context keys used in this package
type ContextKey string

const (
	// ResponseWriterKey is the context key for the http.ResponseWriter
	ResponseWriterKey ContextKey = "ResponseWriter"
)

// JSONError formats and returns an error response
func JSONError(ctx *fiber.Ctx, status int, err error) error {
	return ctx.Status(status).JSON(fiber.Map{
		"error": err.Error(),
	})
}

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string, span opentracing.Span) (string, error) {
	// Create a child span for the controller
	childSpan := opentracing.StartSpan("HashPassword", opentracing.ChildOf(span.Context()))
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		LogErrors(childSpan, err)
		return "", err
	}

	return string(bytes), err
}

func WrapResponse(data interface{}, pagination *Meta, message string, status int16, errors ...interface{}) Response {
	meta := Meta{}
	if pagination != nil {
		meta = *pagination
	}

	return Response{
		Data:    data,
		Meta:    &meta,
		Message: message,
		Status:  status,
		Errors:  errors,
	}
}

func GetResponse(ctx *fiber.Ctx, data interface{}, pagination *Meta, message string, status int16, errors interface{}, options interface{}) error {
	meta := Meta{}
	if pagination != nil {
		meta = *pagination
	}

	response := Response{
		Data:     data,
		Meta:     &meta,
		Message:  message,
		Status:   status,
		Errors:   errors,
		Optional: options,
	}

	return SendResponse(ctx, response, int(status))
}

func SendResponse(ctx *fiber.Ctx, response Response, statusCode int) error {
	return ctx.Status(statusCode).JSON(response)
}

func AtoiDefault(str string, def int) int {
	value, err := strconv.Atoi(str)
	if err != nil {
		return def
	}
	return value
}

// ConvertStructToMap function converts a struct to a map
func ConvertStructToMap(filters interface{}) map[string]string {
	result := make(map[string]string)

	v := reflect.ValueOf(filters)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		key := typeOfS.Field(i).Tag.Get("json")
		value := v.Field(i).Interface()

		// Convert value to string
		switch v := value.(type) {
		case int:
			result[key] = strconv.Itoa(v)
		case string:
			result[key] = v
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}

	return result
}

// GenerateIndexName creates a standardized index name based on table and columns
func GenerateIndexName(table string, columns ...string) string {
	return fmt.Sprintf("idx_%s_%s", table, strings.Join(columns, "_"))
}

func DefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func DefaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// // DD takes multiple values, creates a JSON response, and stops execution.
// func DD(c *fiber.Ctx, values ...interface{}) error {
// 	// Create a map to hold the values
// 	response := fiber.Map{}

// 	// Dynamically add the passed values to the response
// 	for i, value := range values {
// 		// The key will be "value_0", "value_1", etc.
// 		key := fmt.Sprintf("value_%d", i)
// 		response[key] = value
// 	}

// 	// Return a JSON response with status 200 and stop further execution
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  "debug",
// 		"message": "Debugging Output",
// 		"data":    response,
// 	})
// }

// ErrorWithLocation returns an error message with file and line number information.
func ErrorWithLocation(err error) string {
	// Retrieve the program counter, file, and line number
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("error: %v at %s:%d", err, file, line)
}

// GetStringOrDefault retrieves a string value from a variable or map-based on the provided default value.
// It accepts both direct values and map-based values.
// If the value is empty or the key does not exist, it returns the provided default value.
func GetStringOrDefault(value interface{}, defaultValue string) string {
	// Check if value is a string directly
	if str, ok := value.(string); ok {
		if str != "" {
			return str
		}
		return defaultValue
	}

	// Check if value is a map
	if reflect.TypeOf(value).Kind() == reflect.Map {
		// Ensure the value is a map of strings to interfaces
		if m, ok := value.(map[string]interface{}); ok {
			// Try to retrieve value from map and check if it is a string
			if v, exists := m["order_column"]; exists {
				if str, ok := v.(string); ok && str != "" {
					return str
				}
			}
		}
	}

	return defaultValue
}

// GetIntOrDefault retrieves an int value from a variable or map-based on the provided default value.
// It accepts both direct int values and string values (which are converted to int).
// If the value is empty, invalid, or the key does not exist, it returns the provided default value.
func GetIntOrDefault(value interface{}, defaultValue int) int {
	// Check if value is an int directly
	if intValue, ok := value.(int); ok {
		return intValue
	}

	// Check if value is a string and convert it to int
	if str, ok := value.(string); ok {
		if intValue, err := strconv.Atoi(str); err == nil {
			return intValue
		}
	}

	// Check if value is a map
	if reflect.TypeOf(value).Kind() == reflect.Map {
		// Ensure the value is a map of strings to interfaces
		if m, ok := value.(map[string]interface{}); ok {
			// Try to retrieve value from map and check if it is a string or int
			for _, v := range m {
				if intValue, ok := v.(int); ok {
					return intValue
				}
				if str, ok := v.(string); ok {
					if intValue, err := strconv.Atoi(str); err == nil {
						return intValue
					}
				}
			}
		}
	}

	return defaultValue
}

// DD is a helper function to dump the value of a variable, stop the process, and optionally send a response to the client.
func DD(value interface{}) {
	// Print the value to the console
	fmt.Printf("%+v\n", value)

	// Convert the value to JSON
	jsonValue, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal value:", err)
		os.Exit(1)
	}

	// Print the JSON value to the console
	fmt.Println(string(jsonValue))

	// Stop the process
	os.Exit(0)
}

func CreatePaginationMeta(filters map[string]string, total int) *Meta {
	currentPage := GetIntOrDefault(filters["page"], 1)
	perPage := GetIntOrDefault(filters["per_page"], 10)
	lastPage := (total + perPage - 1) / perPage
	// nextPageUrl = if currentPage - lastPage nextpage url null else 1
	var nextPageUrl *int

	if currentPage < lastPage {
		nextPage := currentPage + 1
		nextPageUrl = &nextPage
	}

	return &Meta{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: currentPage,
		LastPage:    lastPage,
		NextPageUrl: nextPageUrl,
	}
}

func ExecuteSeeders(db *sql.DB, seedFiles []string) error {
	for _, file := range seedFiles {
		err := executeSQLFile(db, file)
		if err != nil {
			return fmt.Errorf("error executing %s: %v", file, err)
		}
		fmt.Printf("Executed %s successfully\n", file)
	}
	return nil
}

func executeSQLFile(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening SQL file %s: %v", filePath, err)
	}
	defer file.Close()

	sqlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading SQL file %s: %v", filePath, err)
	}

	sqlContent := string(sqlBytes)
	fmt.Printf("Executing SQL file %s:\n%s\n", filePath, sqlContent)

	_, err = db.Exec(sqlContent)
	if err != nil {
		return fmt.Errorf("error executing SQL file %s: %v", filePath, err)
	}
	return nil
}

func BodyParserWithNull(ctx *fiber.Ctx, out interface{}) error {
	// Parse the request body into a map
	var body map[string]interface{}
	if ctx.Get("Content-Type") == "application/json" {
		if err := json.Unmarshal(ctx.Body(), &body); err != nil {
			return err
		}

		// Convert empty strings to null in the map
		for key, value := range body {
			if str, ok := value.(string); ok && str == "" {
				body[key] = nil
			}
		}
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
	}

	// Marshal the modified body back to JSON
	modifiedBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Unmarshal the modified body into the provided struct
	if err := json.Unmarshal(modifiedBody, out); err != nil {
		return err
	}

	// Convert empty strings to null in the struct fields
	convertEmptyStringsToNull(out)

	return nil
}

// convertEmptyStringsToNull uses reflection to convert empty strings to null in struct fields.
func convertEmptyStringsToNull(out interface{}) {
	v := reflect.ValueOf(out).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

// StringPointerToString converts a string pointer to a string, returning an empty string if the pointer is nil.
func StringPointerToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if reflect.ValueOf(value).IsZero() {
		n.Value = nil
	} else {
		n.Value = &value
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.Value == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(*n.Value)
}

// Scan implements the sql.Scanner interface.
func (n *Nullable[T]) Scan(value interface{}) error {
	if value == nil {
		n.Value = nil
		return nil
	}
	val, ok := value.(T)
	if !ok {
		return errors.New("type assertion failed")
	}
	n.Value = &val
	return nil
}

// GetStringOrDefaultFromArray retrieves a string value from a variable or map-based on the provided default value.
// It accepts both direct values and map-based values.
// If the value is not in the allowed values array, it returns the provided default value.
func GetStringOrDefaultFromArray(value interface{}, allowedValues []string, defaultValue string, key ...string) string {
	// Determine the key to use
	mapKey := "order_column"
	if len(key) > 0 {
		mapKey = key[0]
	}

	// Check if value is a string directly
	if str, ok := value.(string); ok {
		for _, allowedValue := range allowedValues {
			if str == allowedValue {
				return str
			}
		}
		return defaultValue
	}

	// Check if value is a map
	if reflect.TypeOf(value).Kind() == reflect.Map {
		// Ensure the value is a map of strings to interfaces
		if m, ok := value.(map[string]interface{}); ok {
			// Try to retrieve value from map and check if it is a string
			if v, exists := m[mapKey]; exists {
				if str, ok := v.(string); ok {
					for _, allowedValue := range allowedValues {
						if str == allowedValue {
							return str
						}
					}
					return defaultValue
				}
			}
		}
	}

	return defaultValue
}

func Ptr(s string) *string {
	return &s
}

// HasPermission checks if the user has the required permission
func HasPermission(ctx *fiber.Ctx, requiredPermission string) bool {
	userClaims, ok := ctx.Locals("user").(jwt.MapClaims)
	if !ok {
		return false
	}

	permissions, ok := userClaims["permissions"].([]interface{})
	if !ok {
		return false
	}

	for _, permission := range permissions {
		if perm, ok := permission.(string); ok && perm == requiredPermission {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if a substring is present in a string, ignoring case.
func ContainsIgnoreCase(str, substr string) bool {
	str = strings.ToLower(str)
	substr = strings.ToLower(substr)
	return strings.Contains(str, substr)
}

// GetPtrVal safely dereferences a string pointer
func GetPtrVal(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// GetFloatPtrVal safely dereferences a float pointer
func GetFloatPtrVal(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

func JaegerMiddleware(c *fiber.Ctx, tracer opentracing.Tracer) opentracing.SpanContext {
	httpHeaders := make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpHeaders.Add(string(key), string(value))
	})
	parentSpanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpHeaders))
	return parentSpanCtx
}

func StartSpanFromController(ctx *fiber.Ctx, tracer opentracing.Tracer, funcDesc string) opentracing.Span {
	parentSpan := opentracing.StartSpan(funcDesc)

	// Add custom tag to indicate per-service/per-layer tracing
	parentSpan.SetTag("type", "service")

	// requestBody := ctx.Body()
	// parentSpan.LogKV("request_body", string(requestBody))
	// headers
	httpHeaders := make(http.Header)
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		httpHeaders.Add(string(key), string(value))
	})
	parentSpan.LogKV("request_headers", httpHeaders)

	return parentSpan
}

func FilterOtel(ctx *fiber.Ctx) bool {
	// filter response status code >= 400
	isError := ctx.Response().StatusCode() >= 400

	return isError
}

func LogErrors(span opentracing.Span, err error) {
	defer span.Finish()
	span.LogFields(
		jLog.String("event", "error"),
		jLog.String("message", err.Error()),
	)
	ext.Error.Set(span, true)
}

func LogResponse(span opentracing.Span, responseBody interface{}) {
	responseBodyJSON, _ := json.Marshal(responseBody)
	span.LogKV("response_body", string(responseBodyJSON))
}

func GetBodyPayloadValue(ctx *fiber.Ctx, key string) string {
	body := ctx.Body()
	var payload map[string]interface{}
	json.Unmarshal(body, &payload)
	return payload[key].(string)
}

func HandleFileUpload(ctx *fiber.Ctx, file *multipart.FileHeader, userID uint, span opentracing.Span) (string, error) {
	childSpan := opentracing.StartSpan("HandleFileUpload", opentracing.ChildOf(span.Context()))
	// Define the directory to save the uploaded files
	uploadDir := "./public/uploads"

	// Ensure the directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		LogErrors(childSpan, err)
		return "", err
	}

	// timestamp yyyyMMdd_HHmm
	timestamp := time.Now().Format("20060102_1504")
	uuid := uuid.New().String()[:5]
	fileName := fmt.Sprintf("%s-%s-%s-%s", timestamp, strconv.FormatUint(uint64(userID), 10), uuid, file.Filename)

	// Save the file with a unique name
	filePath := fmt.Sprintf("%s/%s", uploadDir, fileName)
	if err := ctx.SaveFile(file, filePath); err != nil {
		LogErrors(childSpan, err)
		return "", err
	}

	// Return the file path
	return filePath, nil
}

// ParseUintPointer parses a string to a uint pointer
func ParseUintPointer(value string) *uint {
	if value == "" {
		return nil
	}
	parsedValue, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return nil
	}
	uintValue := uint(parsedValue)
	return &uintValue
}

// ParseIntPointer parses a string to an int pointer
func ParseIntPointer(value string) *int {
	if value == "" {
		return nil
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &parsedValue
}

// ParseStringPointer parses a string to a string pointer
func ParseStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ParseInt parses a string to an int
func ParseInt(value string) int {
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsedValue
}

// ParseStringPointer parses a string to a string pointer
func ParseIntNullPointer(value interface{}) *int {
	if value == nil {
		return nil
	}

	// if value string
	if str, ok := value.(string); ok {
		parsedValue, err := strconv.Atoi(str)
		if err != nil {
			return nil
		}
		return &parsedValue
	}

	// if value int
	if intValue, ok := value.(int); ok {
		return &intValue
	}

	return nil
}

func RemoveDotAtStart(s string) string {
	if strings.HasPrefix(s, ".") {
		return s[1:]
	}
	return s
}

func AddHostURLToImageURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", os.Getenv("APP_HOST"), RemoveDotAtStart(imageURL))
}

func IsAdmin(ctx *fiber.Ctx) bool {
	userClaims, ok := ctx.Locals("user").(jwt.MapClaims)
	if !ok {
		return false
	}

	roles, ok := userClaims["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == "superadmin" {
			return true
		}
	}
	return false
}

func ErrTrxResponse(ctx *fiber.Ctx, tx *gorm.DB, apiSpan opentracing.Span, err error, message string, status int16) error {
	tx.Rollback()
	LogResponse(apiSpan, WrapResponse(nil, nil, err.Error(), http.StatusInternalServerError))
	return GetResponse(ctx, nil, nil, message, status, err.Error(), nil)
}

func ErrGetReponse(ctx *fiber.Ctx, apiSpan opentracing.Span, err error, message string, status int16) error {
	response := WrapResponse(nil, nil, err.Error(), status)
	LogResponse(apiSpan, response)
	return SendResponse(ctx, response, int(status))
}

func ErrValidResponse(ctx *fiber.Ctx, apiSpan opentracing.Span, message string, errors map[string][]string) error {
	LogResponse(apiSpan, errors)
	return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": errors, "message": "Validation failed", "status": http.StatusBadRequest})
}

func GetClaims(ctx *fiber.Ctx, parentSpan opentracing.Span) jwt.MapClaims {
	// Extract user ID from JWT
	claims, err := auth.GetAuthUser(ctx)
	if err != nil {
		LogErrors(parentSpan, err)
		GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
		return nil
	}

	return claims
}

func JoinUintsToString(ints []uint, sep string) string {
	var strInts []string
	for _, i := range ints {
		strInts = append(strInts, strconv.Itoa(int(i)))
	}
	return strings.Join(strInts, sep)
}

func JoinUintPtrsToString(ints []*uint, sep string) string {
	var strInts []string
	for _, i := range ints {
		strInts = append(strInts, strconv.Itoa(int(*i)))
	}
	return strings.Join(strInts, sep)
}

func SplitStringArrayOfInts(str []string) ([]int, error) {
	var intIDs []int
	for _, id := range str {
		intID, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		intIDs = append(intIDs, intID)
	}

	return intIDs, nil
}

func GetDefaultBranchID(ctx *fiber.Ctx) uint {
	claims := GetClaims(ctx, nil)
	branchID := claims["bid"]

	if branchID != nil {
		return uint(branchID.(float64))
	}

	defaultBranchID := os.Getenv("DEFAULT_BRANCH_ID")
	if defaultBranchID != "" {
		branchID, _ = strconv.ParseFloat(defaultBranchID, 64)
	}

	return uint(branchID.(float64))
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
