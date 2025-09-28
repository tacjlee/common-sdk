package fxstruct

import (
	"fmt"
	"github.com/tacjlee/common-sdk/packages/fxstring"
	"reflect"
)

func GetAllStructFieldNames(inputStruct any) ([]string, error) {
	// Ensure the input is a struct or a pointer to a struct
	v := reflect.ValueOf(inputStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or a pointer to a struct")
	}

	// Get the type of the struct
	t := v.Type()
	fieldNames := make([]string, t.NumField())

	// Iterate over all fields
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldNames[i] = fxstring.PascalToCamel(fieldName)
	}
	return fieldNames, nil
}

func CloneTo[T any](input any) (T, error) {
	var cloned T
	// Use reflection to copy the fields from the input to the cloned value
	srcValue := reflect.ValueOf(input)
	dstField := reflect.ValueOf(&cloned)
	dstElem := dstField.Elem()
	// Iterate through the fields of the source struct
	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		dstField := dstElem.FieldByName(srcValue.Type().Field(i).Name)
		// Check if the field exists in the destination and is settable
		if dstField.IsValid() && dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
	return cloned, nil
}

func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Struct {
		return false
	}
	// Iterate over all fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// Check if the field is a struct itself
		if field.Kind() == reflect.Struct {
			if IsEmpty(field.Interface()) {
				continue
			}
			return false
		}
		// Check if the field is a zero value
		if !field.IsValid() || reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			continue
		}
		return false
	}
	return true
}

func AppendToList[T any](existingList []T, newList []T) []T {
	// Append newList to existingList and return the updated list
	return append(existingList, newList...)
}
