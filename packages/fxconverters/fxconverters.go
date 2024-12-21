package fxconverters

import (
	"encoding/json"
	"fmt"
)

func BoolToAny[T any](value bool) (T, error) {
	var zeroValue T // Zero value of T
	// We are using the interface{} to hold any value and then perform a type assertion
	if value, ok := interface{}(value).(T); ok {
		zeroValue = value
		return zeroValue, nil
	} else {
		return zeroValue, fmt.Errorf("unable to convert value to type %T", zeroValue)
	}
}

func JsonToStruct[T any](jsonData string) (T, error) {
	var zeroValue T // Zero value of T
	err := json.Unmarshal([]byte(jsonData), &zeroValue)
	if err != nil {
		return zeroValue, err
	}
	return zeroValue, nil
}
