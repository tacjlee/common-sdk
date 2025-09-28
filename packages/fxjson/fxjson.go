package fxjson

import (
	"encoding/json"
	"fmt"
	"log"
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

func ToPrettyString(data interface{}) string {
	// MarshalIndent indents the JSON output with two spaces
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}
	return string(prettyJSON)
}
