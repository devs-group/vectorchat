package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Load populates the fields of the configuration struct pointed to by `cfg`
// based on environment variables.
//
// It uses struct field tags to determine the mapping:
//   - `env:"ENV_VAR_NAME"`: Specifies the environment variable name.
//   - `envDefault:"value"`: Provides a default value if the environment variable is not set.
//   - `envRequired:"true"`: Marks the field as required. If the environment variable
//     is not set and no default is provided, Load will return an error.
//   - `envSeparator:","`: Specifies the separator for slice types (default is ",").
//
// Supported field types: string, int, int64, bool, float64, time.Duration,
// []string, []int, []int64, []bool, []float64.
//
// The `cfg` argument must be a pointer to a struct.
func Load(cfg any) error {
	// 1. Validate input: Must be a non-nil pointer to a struct.
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("config: Load expects a non-nil pointer to a struct, got %T", cfg)
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("config: Load expects a pointer to a struct, got pointer to %v", elem.Kind())
	}

	typ := elem.Type()

	// 2. Iterate over struct fields.
	for i := range elem.NumField() {
		field := elem.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields.
		if !field.CanSet() {
			continue
		}

		// 3. Get tags.
		envVar := fieldType.Tag.Get("env")
		if envVar == "" {
			// Skip fields without 'env' tag.
			continue
		}

		defaultValue := fieldType.Tag.Get("envDefault")
		required := fieldType.Tag.Get("envRequired") == "true"
		separator := fieldType.Tag.Get("envSeparator")
		if separator == "" {
			separator = "," // Default separator
		}

		// 4. Get value from environment or default.
		value, exists := os.LookupEnv(envVar)
		if !exists {
			if defaultValue != "" {
				value = defaultValue
				exists = true // Treat default value as existing for parsing logic
			} else if required {
				return fmt.Errorf("config: required environment variable %q for field %q is not set and no default value provided", envVar, fieldType.Name)
			} else {
				// Not required, no env var, no default -> leave field as zero value.
				continue
			}
		}

		// 5. Set field value based on type.
		if err := setField(field, fieldType.Name, envVar, value, separator); err != nil {
			return err
		}
	}

	return nil
}

// setField attempts to parse the string value and set it to the reflected field.
func setField(field reflect.Value, fieldName, envVar, value, separator string) error {
	fieldType := field.Type()

	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Check for time.Duration, which has an underlying type of int64
		if fieldType == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return formatError(fieldName, envVar, value, "parse as time.Duration", err)
			}
			field.SetInt(int64(duration))
		} else {
			// Standard integers
			intValue, err := strconv.ParseInt(value, 0, fieldType.Bits())
			if err != nil {
				return formatError(fieldName, envVar, value, fmt.Sprintf("parse as %s", fieldType.Kind()), err)
			}
			field.SetInt(intValue)
		}
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return formatError(fieldName, envVar, value, "parse as bool", err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, fieldType.Bits())
		if err != nil {
			return formatError(fieldName, envVar, value, fmt.Sprintf("parse as %s", fieldType.Kind()), err)
		}
		field.SetFloat(floatValue)
	case reflect.Slice:
		return setSliceField(field, fieldName, envVar, value, separator)
	default:
		return fmt.Errorf("config: unsupported type %q for field %q (env %q)", fieldType.Kind(), fieldName, envVar)
	}
	return nil
}

// setSliceField handles parsing and setting slice types.
func setSliceField(field reflect.Value, fieldName, envVar, value, separator string) error {
	sliceKind := field.Type().Elem().Kind()
	parts := strings.Split(value, separator)

	// Handle empty string input for slices (results in a slice with one empty element)
	// Treat empty string as an empty slice unless the separator is also empty
	// (which is unlikely but technically possible).
	if len(parts) == 1 && parts[0] == "" && separator != "" {
		parts = []string{}
	}

	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part) // Trim whitespace from each part
		elem := slice.Index(i)
		switch sliceKind {
		case reflect.String:
			elem.SetString(part)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(part, 0, elem.Type().Bits())
			if err != nil {
				return formatError(fieldName, envVar, part, fmt.Sprintf("parse slice element as %s", sliceKind), err)
			}
			elem.SetInt(intValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(part)
			if err != nil {
				return formatError(fieldName, envVar, part, "parse slice element as bool", err)
			}
			elem.SetBool(boolValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(part, elem.Type().Bits())
			if err != nil {
				return formatError(fieldName, envVar, part, fmt.Sprintf("parse slice element as %s", sliceKind), err)
			}
			elem.SetFloat(floatValue)
		default:
			return fmt.Errorf("config: unsupported slice element type %q for field %q (env %q)", sliceKind, fieldName, envVar)
		}
	}

	field.Set(slice)
	return nil
}

// formatError creates a standard error message for parsing issues.
func formatError(fieldName, envVar, value, action string, parseErr error) error {
	return fmt.Errorf("config: field %q (env %q): failed to %s value %q: %w",
		fieldName, envVar, action, value, parseErr)
}
