package fxconverter

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func ToIntDefault(s string, defaultVal int) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func ToLongDefault(s string, defaultVal int64) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func ToBoolDefault(s string, defaultVal bool) bool {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func ToFloatDefault(s string, defaultVal float64) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func ToStringDefault(s string, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
