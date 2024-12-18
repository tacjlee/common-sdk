package fxoperators

import "reflect"

func If(condition bool, trueVal any, falseVal any) any {
	value := falseVal
	if condition {
		value = trueVal
	}
	result := reflect.ValueOf(value)
	return reflect.Indirect(result).Interface()
}
