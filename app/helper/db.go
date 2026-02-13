package helper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/webcore-go/webcore/port"
)

func MarshalDbMap(v any) (port.DbMap, error) {
	result := make(port.DbMap)
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input harus struct, dapat: %s", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		tag := field.Tag.Get("db")

		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		fieldName := parts[0]

		// Cek opsi omitempty
		isOmitEmpty := false
		for _, opt := range parts[1:] {
			if opt == "omitempty" {
				isOmitEmpty = true
			}
		}

		// Jika field adalah pointer, ambil nilai elemennya
		actualVal := fieldVal
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				if !isOmitEmpty {
					result[fieldName] = nil
				}
				continue
			}
			actualVal = fieldVal.Elem() // Ambil nilai di balik pointer
		}

		if isOmitEmpty && actualVal.IsZero() {
			continue
		}

		result[fieldName] = actualVal.Interface()
	}
	return result, nil
}

func UnmarshalDbMap(data port.DbMap, out any) error {
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("out harus berupa pointer ke struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")

		if tag == "" || tag == "-" {
			continue
		}

		dbFieldName := strings.Split(tag, ",")[0]
		mapVal, exists := data[dbFieldName]
		if !exists || mapVal == nil {
			continue
		}

		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		inputVal := reflect.ValueOf(mapVal)

		// Jika struct field adalah pointer, siapkan memorinya
		if structField.Kind() == reflect.Ptr {
			// Buat instance baru sesuai tipe yang ditunjuk pointer (misal string)
			ptrValue := reflect.New(structField.Type().Elem())

			// Set nilainya ke instance baru tersebut
			if inputVal.Type().AssignableTo(ptrValue.Elem().Type()) {
				ptrValue.Elem().Set(inputVal)
				structField.Set(ptrValue) // Masukkan pointer ke field struct
			}
		} else {
			// Normal field (bukan pointer)
			if inputVal.Type().AssignableTo(structField.Type()) {
				structField.Set(inputVal)
			}
		}
	}
	return nil
}
