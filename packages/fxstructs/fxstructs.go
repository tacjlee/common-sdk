package fxstructs

import (
	"fmt"
	"github.com/tacjlee/common-sdk/packages/fxstrings"
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
		fieldNames[i] = fxstrings.PascalToCamel(fieldName)
	}
	return fieldNames, nil
}

func CloneStruct(source, destination interface{}) error {
	srcValue := reflect.ValueOf(source)
	dstValue := reflect.ValueOf(destination)

	// Ensure both are pointers to structs
	if srcValue.Kind() != reflect.Ptr || dstValue.Kind() != reflect.Ptr {
		return fmt.Errorf("both source and destination must be pointers to structs")
	}

	srcElem := srcValue.Elem()
	dstElem := dstValue.Elem()

	// Ensure the underlying types are structs
	if srcElem.Kind() != reflect.Struct || dstElem.Kind() != reflect.Struct {
		return fmt.Errorf("both source and destination must point to structs")
	}

	// Iterate through the fields of the source struct
	for i := 0; i < srcElem.NumField(); i++ {
		srcField := srcElem.Field(i)
		dstField := dstElem.FieldByName(srcElem.Type().Field(i).Name)

		// Check if the field exists in the destination and is settable
		if dstField.IsValid() && dstField.CanSet() {
			dstField.Set(srcField)
		}
	}

	return nil
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
