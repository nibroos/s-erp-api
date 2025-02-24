package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
)

var validate *validator.Validate
var db *sqlx.DB

func InitValidator(database *sqlx.DB) {
	db = database
}

// Request defines the validation rules and data
type Request struct {
	Rules            map[string][]string // Rules are now arrays of strings
	Data             map[string]interface{}
	CustomFieldNames map[string]string
}

// NewRequest creates a new Request instance
func NewRequest(rules map[string][]string, data map[string]interface{}, customFieldNames map[string]string) *Request {
	return &Request{
		Rules:            rules,
		Data:             data,
		CustomFieldNames: customFieldNames,
	}
}

// Validate validates the request data against the rules
func (r *Request) Validate() (map[string][]string, bool) {
	errors := make(map[string][]string)

	for field, rules := range r.Rules {
		value, exists := r.Data[field]
		if !exists {
			value = nil // Treat missing fields as nil
		}

		for _, rule := range rules {
			if strings.Contains(field, ".*.") {
				fieldParts := strings.Split(field, ".*.")
				arrayField := fieldParts[0]
				nestedField := fieldParts[1]

				arrayValue, exists := r.Data[arrayField]
				if !exists || !isArray(arrayValue) {
					continue
				}

				for i, item := range arrayValue.([]interface{}) {
					if itemMap, ok := item.(map[string]interface{}); ok {
						nestedValue, exists := itemMap[nestedField]
						if !exists {
							nestedValue = nil
						}

						// for _, rule := range rules {
						ruleParts := strings.Split(rule, ":")
						ruleName := ruleParts[0]
						var ruleParam string
						if len(ruleParts) > 1 {
							ruleParam = ruleParts[1]
						}

						customFieldName := strings.ReplaceAll(nestedField, "_", " ")
						if customName, ok := customFieldNames[nestedField]; ok {
							customFieldName = customName
						}
						if customName, ok := r.CustomFieldNames[nestedField]; ok {
							customFieldName = customName
						}

						switch ruleName {
						case "required":
							if isEmpty(nestedValue) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field is required", customFieldName))
							}
						case "email":
							if nestedValue == nil {
								continue
							}
							if !isEmail(nestedValue) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be a valid email address", customFieldName))
							}
						case "date":
							if !isDate(nestedValue.(string)) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be a valid date in the format %s", customFieldName, ruleParam))
							}
						case "min":
							min, err := strconv.Atoi(ruleParam)
							if err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], "Invalid min parameter")
								continue
							}
							if !isMin(nestedValue, min) {
								switch nestedValue.(type) {
								case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
									errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be at least %d", customFieldName, min))
								default:
									errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be at least %d characters long", customFieldName, min))
								}
							}
						// Add more rules here
						case "max":
							max, err := strconv.Atoi(ruleParam)
							if err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], "Invalid max parameter")
								continue
							}

							if !isMax(nestedValue, max) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("This field must be at most %d characters long", max))
							}
						case "numeric":
							if !isNumeric(nestedValue) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be a number", customFieldName))
							}

						case "float":
							if !isFloat(nestedValue.(string)) {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("The %s field must be a decimal number", customFieldName))
							}

						case "unique":
							if err := uniqueRule(customFieldName, rule, nestedValue); err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], err.Error())
							}
						case "unique_ig":
							if err := uniqueIgRule(customFieldName, rule, nestedValue); err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], err.Error())
							}
						case "array":
							if nestedValue == nil {
								continue
							}
							if err := arrayRule(customFieldName, nestedValue); err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], err.Error())
							}
						case "array_max":
							if err := arrayMaxRule(customFieldName, rule, nestedValue); err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], err.Error())
							}
						case "exists":
							if nestedValue == nil {
								continue
							}
							if err := isExistsRule(customFieldName, rule, nestedValue); err != nil {
								errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], err.Error())
							}
						default:
							errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)] = append(errors[fmt.Sprintf("%s.%d.%s", arrayField, i, nestedField)], fmt.Sprintf("Unknown validation rule: %s", ruleName))
						}
						// }
					}
				}
			} else {
				ruleParts := strings.Split(rule, ":")
				ruleName := ruleParts[0]
				var ruleParam string
				if len(ruleParts) > 1 {
					ruleParam = ruleParts[1]
				}

				customFieldName := strings.ReplaceAll(field, "_", " ")
				if customName, ok := customFieldNames[field]; ok {
					customFieldName = customName
				}
				if customName, ok := r.CustomFieldNames[field]; ok {
					customFieldName = customName
				}

				switch ruleName {
				case "required":
					if isEmpty(value) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field is required", customFieldName))
					}
				case "email":
					if value == nil {
						continue
					}
					if !isEmail(value) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field must be a valid email address", customFieldName))
					}
				case "date":
					if !isDate(value.(string)) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field must be a valid date in the format %s", customFieldName, ruleParam))
					}
				case "min":
					min, err := strconv.Atoi(ruleParam)
					if err != nil {
						errors[field] = append(errors[field], "Invalid min parameter")
						continue
					}
					if !isMin(value, min) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field must be at least %d characters long", customFieldName, min))
					}
				// Add more rules here
				case "max":
					max, err := strconv.Atoi(ruleParam)
					if err != nil {
						errors[field] = append(errors[field], "Invalid max parameter")
						continue
					}

					if !isMax(value, max) {
						errors[field] = append(errors[field], fmt.Sprintf("This field must be at most %d characters long", max))
					}
				case "numeric":
					if !isNumeric(value) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field must be a number", customFieldName))
					}

				case "float":
					if !isFloat(value.(string)) {
						errors[field] = append(errors[field], fmt.Sprintf("The %s field must be a decimal number", customFieldName))
					}

				case "unique":
					if err := uniqueRule(customFieldName, rule, value); err != nil {
						errors[field] = append(errors[field], err.Error())
					}
				case "unique_ig":
					if err := uniqueIgRule(customFieldName, rule, value); err != nil {
						errors[field] = append(errors[field], err.Error())
					}
				case "array":
					if value == nil {
						continue
					}
					if err := arrayRule(customFieldName, value); err != nil {
						errors[field] = append(errors[field], err.Error())
					}
				case "array_max":
					if err := arrayMaxRule(customFieldName, rule, value); err != nil {
						errors[field] = append(errors[field], err.Error())
					}
				case "exists":
					if value == nil {
						continue
					}
					if err := isExistsRule(customFieldName, rule, value); err != nil {
						errors[field] = append(errors[field], err.Error())
					}
				default:
					errors[field] = append(errors[field], fmt.Sprintf("Unknown validation rule: %s", ruleName))
				}
			}
		}
	}

	// Handle nested rules for arrays of objects
	// for field, rules := range r.Rules {
	// }

	return errors, len(errors) == 0
}

func isArray(x interface{}) bool {
	if x == nil {
		return false
	}
	kind := reflect.TypeOf(x).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

func isEmail(value interface{}) bool {
	email, ok := value.(string)
	if !ok {
		return false
	}
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailRegex).MatchString(email)
}

// func isDate(value interface{}, format string) bool {
// 	dateStr, ok := value.(string)
// 	if !ok {
// 		return false
// 	}
// 	_, err := time.Parse(format, dateStr)
// 	return err == nil
// }

func isMin(value interface{}, min int) bool {
	switch v := value.(type) {
	case string:
		return len(v) >= min
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Int() >= int64(min)
	case float32, float64:
		return reflect.ValueOf(v).Float() >= float64(min)
	default:
		return false
	}
}

func isMax(value interface{}, max int) bool {
	switch v := value.(type) {
	case string:
		return len(v) <= max
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Int() <= int64(max)
	case float32, float64:
		return reflect.ValueOf(v).Float() <= float64(max)
	default:
		return false
	}
}

func isNumeric(value interface{}) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(v, 64)
		return err == nil
	default:
		return false
	}
}

// uniqueValidator checks if a field value is unique in the database.
func uniqueValidator(fl validator.FieldLevel) bool {

	// utils.DD(map[string]interface{}{
	// 	"perPage": fl.Field().Interface(),
	// 	"fl":      fl,
	// })

	value := fl.Field().Interface()

	// Debugging: Dump the value
	// utils.DD(value)

	// If the value is empty or null, pass the validation
	if value == nil || reflect.ValueOf(value).IsZero() {
		return true
	}

	// Convert value to string for query
	valueStr, ok := value.(string)
	if !ok {
		return false
	}

	param := fl.Param()
	params := strings.Split(param, ",")
	if len(params) != 2 {
		return false
	}

	table := params[0]
	column := params[1]

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, column)
	err := db.Get(&count, query, valueStr)
	if err != nil {
		return false
	}

	return count == 0
}

// ValidateCreateUserRequest validates the CreateUserRequest struct.
func ValidateCreateUserRequest(req *dtos.CreateUserRequest) map[string]string {
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errors := make(map[string]string)
	for _, err := range validationErrors {
		errors[err.Field()] = err.Tag()
	}
	return errors
}

func ValidateUpdateUserRequest(req *dtos.UpdateUserRequest) map[string]string {
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errors := make(map[string]string)
	for _, err := range validationErrors {
		errors[err.Field()] = err.Tag()
	}
	return errors
}

// uniqueRule checks if a field value is unique in the database.
func uniqueRule(field string, rule string, value interface{}) error {
	valueStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type")
	}

	params := strings.Split(rule, ":")
	if len(params) != 2 {
		return fmt.Errorf("invalid rule format")
	}

	tableColumn := strings.Split(params[1], ",")
	if len(tableColumn) != 2 {
		return fmt.Errorf("invalid table and column format")
	}

	table := tableColumn[0]
	column := tableColumn[1]

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, column)
	err := db.Get(&count, query, valueStr)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("the %s has already been taken", field)
	}

	return nil
}

// uniqueIgRule checks if a field value is unique in the database, ignoring the current entity.
func uniqueIgRule(field string, rule string, value interface{}) error {
	// Check if value is uint or string
	var valueStr string
	switch v := value.(type) {
	case uint:
		valueStr = fmt.Sprintf("%d", v)
	case string:
		valueStr = v
	default:
		return fmt.Errorf("invalid value type")
	}

	params := strings.Split(rule, ":")
	if len(params) != 2 {
		return fmt.Errorf("invalid rule format")
	}

	tableColumn := strings.Split(params[1], ",")
	if len(tableColumn) != 3 {
		return fmt.Errorf("invalid table and column format")
	}

	table := tableColumn[0]
	column := tableColumn[1]
	currentID := tableColumn[2]

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1 AND id != $2", table, column)
	err := db.Get(&count, query, valueStr, currentID)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("the %s has already been taken", field)
	}

	return nil
}

func arrayRule(field string, value interface{}) error {
	_, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("the %s field must be an array", field)
	}

	return nil
}

func arrayMaxRule(field string, rule string, value interface{}) error {
	valueArr, ok := value.([]string)
	if !ok {
		return fmt.Errorf("invalid value typeb")
	}

	params := strings.Split(rule, ":")
	if len(params) != 2 {
		return fmt.Errorf("invalid rule format")
	}

	max, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("invalid max value")
	}

	if len(valueArr) > max {
		return fmt.Errorf("the %s field must have at most %d items", field, max)
	}

	return nil
}

func isExistsRule(field string, rule string, value interface{}) error {
	params := strings.Split(rule, ":")
	if len(params) != 2 {
		return fmt.Errorf("invalid rule format")
	}

	tableColumn := strings.Split(params[1], ",")
	if len(tableColumn) != 2 {
		return fmt.Errorf("invalid table and column format")
	}

	table := tableColumn[0]
	column := tableColumn[1]

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, column)
	err := db.Get(&count, query, value)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("the %s does not exist", field)
	}

	return nil
}

// isEmpty check a type is Zero
func isEmpty(x interface{}) bool {
	rt := reflect.TypeOf(x)
	if rt == nil {
		return true
	}
	rv := reflect.ValueOf(x)
	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return rv.Len() == 0
	}
	return reflect.DeepEqual(x, reflect.Zero(rt).Interface())
}

func requiredRule(field string, value interface{}) error {
	if value == nil {
		return fmt.Errorf("the %s field is required", field)
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("the %s field is required", field)
		}
	case int, int8, int16, int32, int64:
		if v == 0 {
			return fmt.Errorf("the %s field is required", field)
		}
	case uint, uint8, uint16, uint32, uint64:
		if v == 0 {
			return fmt.Errorf("the %s field is required", field)
		}
	case float32, float64:
		if v == 0 {
			return fmt.Errorf("the %s field is required", field)
		}
	case map[interface{}]interface{}:
		if len(v) == 0 {
			return fmt.Errorf("the %s field is required", field)
		}
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("the %s field is required", field)
		}
	default:
		return fmt.Errorf("unsupported type for field %s", field)
	}

	return nil
}

// TODO make a function to validate mix_values group, 2 params, group and value
