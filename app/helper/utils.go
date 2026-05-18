package helper

import (
	"fmt"
	"reflect"
	"strings"
)

// Helper function to convert data type to its pointer
func StringPtr(s string) *string    { return &s }
func IntPtr(i int) *int             { return &i }
func Int8Ptr(i int8) *int8          { return &i }
func Int16Ptr(i int16) *int16       { return &i }
func Int32Ptr(i int32) *int32       { return &i }
func Int64Ptr(i int64) *int64       { return &i }
func UIntPtr(i uint) *uint          { return &i }
func UInt8Ptr(i uint8) *uint8       { return &i }
func UInt16Ptr(i uint16) *uint16    { return &i }
func UInt32Ptr(i uint32) *uint32    { return &i }
func UInt64Ptr(i uint64) *uint64    { return &i }
func FloatPtr(f float32) *float32   { return &f }
func Float32Ptr(f float32) *float32 { return &f }
func Float64Ptr(f float64) *float64 { return &f }
func BoolPtr(b bool) *bool          { return &b }

// ToJSON converts an interface to JSON string
func ToJSON(v any) (string, error) {
	data, err := JSONMarshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ToIndentJSON(v any) (string, error) {
	data, err := JSONMarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ToLogJSON(v any) string {
	data, err := JSONMarshalIndent(v, "", "  ")
	if err != nil {
		return "<error marshal to json>"
	}
	return string(data)
}

// FromJSON converts a JSON string to an interface
func FromJSON(jsonStr string, v any) error {
	return JSONUnmarshal([]byte(jsonStr), v)
}

// StructToMap converts a struct to a map
func StructToMap(obj any) (map[string]any, error) {
	result := make(map[string]any)

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		result[fieldName] = v.Field(i).Interface()
	}

	return result, nil
}

// MapToStruct converts a map to a struct
func MapToStruct(data map[string]any, result any) error {
	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := fieldType.Name
		if jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		if value, exists := data[fieldName]; exists {
			if reflect.TypeOf(value).AssignableTo(field.Type()) {
				field.Set(reflect.ValueOf(value))
			}
		}
	}

	return nil
}

// Contains checks if a slice contains a value
func Contains(slice any, item any) bool {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < s.Len(); i++ {
		if reflect.DeepEqual(s.Index(i).Interface(), item) {
			return true
		}
	}

	return false
}

// RemoveDuplicates removes duplicate values from a slice
func RemoveDuplicates(slice any) any {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		return slice
	}

	result := make([]any, 0)
	seen := make(map[any]bool)

	for i := 0; i < s.Len(); i++ {
		item := s.Index(i).Interface()
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
