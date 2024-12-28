package fxstrings

import (
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func IsEmpty(value any) bool {
	if value == nil {
		return true
	}
	result := fmt.Sprintf("%v", value)
	if result == "" {
		return true
	} else {
		return false
	}
}

func PascalToCamel(pascal string) string {
	if pascal == "" {
		return pascal
	}
	// Convert the first letter to lowercase
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func ToSnakeCase(input string) string {
	// Regular expression to match uppercase letters or delimiters
	reg := regexp.MustCompile(`[A-Z][a-z]*|[a-z]+|\d+`)
	// Find all words in the string
	words := reg.FindAllString(input, -1)
	// Join words with underscores and convert to lowercase
	return strings.ToLower(strings.Join(words, "_"))
}

func ToJsonCase(snake string) string {
	// Split the snake_case string by underscores
	words := strings.Split(snake, "_")
	// The first word should be in lowercase (no change needed)
	for i := 1; i < len(words); i++ {
		// Capitalize the first letter of each subsequent word
		words[i] = strings.Title(words[i])
	}
	// Join the words to form the camelCase string
	return strings.Join(words, "")
}

func ToString(input interface{}) string {
	if input == nil {
		return ""
	}
	result := fmt.Sprintf("%v", input)
	return result
}

func StringToInt(s string, defaultValue int) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return value
}

func StringToInt64(s string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func StringToDouble(s string, defaultValue float64) float64 {
	// Convert string to float64 (double)
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func StringToUuid(s string, defaultValue uuid.UUID) uuid.UUID {
	convertedUUID, err := uuid.Parse(s)
	if err != nil {
		return defaultValue
	}
	return convertedUUID
}

func StringToIsoDateTime(s string, defaultValue time.Time) time.Time {
	// time.RFC3339 is the Go standard for parsing and formatting ISO 8601 dates (like 2024-11-30T13:45:00Z).
	datetime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return defaultValue
	}
	return datetime
}
