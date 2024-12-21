package fxjsons

import (
	"encoding/json"
	"fmt"
)

func GetAllJsonNodeFieldNames(node map[string]any) []string {
	keys := make([]string, 0, len(node))
	for k := range node {
		keys = append(keys, k)
	}
	return keys
}

func GetPropertyValueAsString(row map[string]any, key string) string {
	value, exists := row[key]
	if !exists || value == nil {
		return ""
	}
	result := fmt.Sprintf("%v", value)
	return result
}

func GetPropertyValueAsBool(row map[string]any, key string) bool {
	value, exists := row[key]
	if !exists {
		return false
	}
	result := fmt.Sprintf("%v", value)
	if result == "true" {
		return true
	}
	return false
}

func ToPrettyString(data []map[string]any) string {
	var result string
	for _, m := range data {
		jsonData, err := json.Marshal(m)
		if err != nil {
			return result
		}
		result += string(jsonData) + "\n"
	}
	return result
}
