package dew

import (
	"testing"
)

func TestIntColumn_Sql(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		expected string
	}{
		{
			name:     "with table",
			column:   IntColumn{name: "id", table: stringPtr("table")},
			expected: "table.id",
		},
		{
			name:     "with alias",
			column:   IntColumn{name: "id", alias: stringPtr("alias")},
			expected: "alias",
		},
		{
			name:     "with table and alias (alias takes priority)",
			column:   IntColumn{name: "id", table: stringPtr("table"), alias: stringPtr("alias")},
			expected: "alias",
		},
		{
			name:     "with empty string table name",
			column:   IntColumn{name: "id", table: stringPtr("")},
			expected: "id",
		},
		{
			name:     "with alias and empty table (alias takes priority)",
			column:   IntColumn{name: "id", table: stringPtr(""), alias: stringPtr("alias")},
			expected: "alias",
		},
		{
			name:     "with empty alias",
			column:   IntColumn{name: "id", alias: stringPtr("")},
			expected: "",
		},
		{
			name:     "without table and alias",
			column:   IntColumn{name: "id"},
			expected: "id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.Sql()
			if got != tt.expected {
				t.Errorf("IntColumn.Sql() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIntColumn_TableName(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		expected string
	}{
		{
			name:     "with table",
			column:   IntColumn{name: "id", table: stringPtr("table")},
			expected: "table",
		},
		{
			name:     "with empty string table name",
			column:   IntColumn{name: "id", table: stringPtr("")},
			expected: "",
		},
		{
			name:     "without table",
			column:   IntColumn{name: "id"},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.column.TableName(); got != tt.expected {
				t.Errorf("IntColumn.TableName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIntColumn_Alias(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		expected *string
	}{
		{
			name:     "without alias",
			column:   IntColumn{name: "id"},
			expected: nil,
		},
		{
			name:     "with alias",
			column:   IntColumn{name: "id", alias: stringPtr("alias")},
			expected: stringPtr("alias"),
		},
		{
			name:     "with empty alias",
			column:   IntColumn{name: "id", alias: stringPtr("")},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.Alias()

			if tt.expected == nil {
				if got != nil {
					t.Errorf("IntColumn.Alias() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("IntColumn.Alias() = nil, want %v", *tt.expected)
				return
			}

			if *got != *tt.expected {
				t.Errorf("IntColumn.Alias() = %v, want %v", *got, *tt.expected)
			}
		})
	}
}

func TestIntColumn_As(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		alias    string
		expected IntColumn
	}{
		{
			name:     "with alias",
			column:   IntColumn{name: "id", alias: stringPtr("alias")},
			alias:    "new_alias",
			expected: IntColumn{name: "id", alias: stringPtr("new_alias")},
		},
		{
			name:     "with empty alias",
			column:   IntColumn{name: "id"},
			alias:    "",
			expected: IntColumn{name: "id", alias: stringPtr("")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.As(tt.alias)
			if got.name != tt.expected.name || got.table != tt.expected.table || *got.alias != *tt.expected.alias {
				t.Errorf("IntColumn.As() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIntColumn_Eq(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		value    int
		expected Expression
	}{
		{
			name:     "with value",
			column:   IntColumn{name: "id"},
			value:    1,
			expected: colEq(IntColumn{name: "id"}, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.Eq(tt.value)
			if got.Sql() != tt.expected.Sql() {
				t.Errorf("IntColumn.Eq() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIntColumn_NotEq(t *testing.T) {
	tests := []struct {
		name     string
		column   IntColumn
		value    int
		expected Expression
	}{
		{
			name:     "with value",
			column:   IntColumn{name: "id"},
			value:    1,
			expected: colNotEq(IntColumn{name: "id"}, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.NotEq(tt.value)
			if got.Sql() != tt.expected.Sql() {
				t.Errorf("IntColumn.NotEq() = %v, want %v", got, tt.expected)
			}
		})
	}
}
