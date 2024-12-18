package fxoperators

import "reflect"

func If(condition bool, trueVal any, falseVal any) any {
	value := falseVal
	if condition {
		value = trueVal
	}
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return value
	} else if valueType.Kind() == reflect.Bool {
		return value.(bool)
	} else if valueType.Kind() == reflect.Float64 {
		return value.(float64)
	} else if valueType.Kind() == reflect.Float32 {
		return value.(float32)
	} else if valueType.Kind() == reflect.Int {
		return value.(int)
	} else if valueType.Kind() == reflect.Int8 {
		return value.(int8)
	} else if valueType.Kind() == reflect.Int16 {
		return value.(int16)
	} else if valueType.Kind() == reflect.Int32 {
		return value.(int32)
	} else if valueType.Kind() == reflect.Int64 {
		return value.(int64)
	} else if valueType.Kind() == reflect.Uint {
		return value.(uint)
	} else if valueType.Kind() == reflect.Uint8 {
		return value.(uint8)
	} else if valueType.Kind() == reflect.Uint16 {
		return value.(uint16)
	} else if valueType.Kind() == reflect.Uint32 {
		return value.(uint32)
	} else if valueType.Kind() == reflect.Uint64 {
		return value.(uint64)
	} else if valueType.Kind() == reflect.String {
		return value.(string)
	} else {
		return value
	}
}
