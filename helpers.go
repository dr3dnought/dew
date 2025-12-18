package dew

import (
	"reflect"
	"strings"
)

func stringPtr(s string) *string {
	return &s
}

func getColumnName(field reflect.StructField) string {
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		if dbTag == "-" {
			return ""
		}

		if idx := strings.Index(dbTag, ","); idx != -1 {
			return dbTag[:idx]
		}
		if dbTag != "" {
			return strings.ToLower(strings.TrimSpace(dbTag))
		}
	}

	// Fallback to sql tag
	if sqlTag := field.Tag.Get("sql"); sqlTag != "" {
		if sqlTag == "-" {
			return ""
		}

		if idx := strings.Index(sqlTag, ","); idx != -1 {
			return sqlTag[:idx]
		}
		if sqlTag != "" {
			return strings.ToLower(strings.TrimSpace(sqlTag))
		}
	}

	// Fallback to snake_case
	return strings.ToLower(toSnakeCase(field.Name))
}
