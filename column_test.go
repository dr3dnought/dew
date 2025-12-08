package dew

import (
	"testing"
	"time"
)

// helper column for low-level col* tests
type helperColumn struct {
	sql string
}

func (c helperColumn) Sql() string        { return c.sql }
func (c helperColumn) Args() []any        { return nil }
func (c helperColumn) ColumnName() string { return c.sql }
func (c helperColumn) TableName() string  { return "" }
func (c helperColumn) Alias() *string     { return nil }

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

func TestIntColumn_In(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"values", col.In(1, 2, 3), "id IN (?, ?, ?)", []any{1, 2, 3}},
		{"empty", col.In(), "1=0", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestIntColumn_NotIn(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"values", col.NotIn(4, 5), "id NOT IN (?, ?)", []any{4, 5}},
		{"empty", col.NotIn(), "1=1", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestIntColumn_Between(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"between", col.Between(10, 20), "id BETWEEN ? AND ?", []any{10, 20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestIntColumn_IsNull(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
	}{
		{"is null", col.IsNull(), "id IS NULL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.expr.Args() != nil {
				t.Errorf("Args = %v, want nil", tt.expr.Args())
			}
		})
	}
}

func TestIntColumn_IsNotNull(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
	}{
		{"is not null", col.IsNotNull(), "id IS NOT NULL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.expr.Args() != nil {
				t.Errorf("Args = %v, want nil", tt.expr.Args())
			}
		})
	}
}

func TestIntColumn_Compare(t *testing.T) {
	col := IntColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		arg  any
	}{
		{"Gt", col.Gt(5), "id > ?", 5},
		{"Gte", col.Gte(5), "id >= ?", 5},
		{"Lt", col.Lt(5), "id < ?", 5},
		{"Lte", col.Lte(5), "id <= ?", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != 1 || tt.expr.Args()[0] != tt.arg {
				t.Errorf("Args = %v, want [%v]", tt.expr.Args(), tt.arg)
			}
		})
	}
}

func TestIntColumn_SubQueries(t *testing.T) {
	col := IntColumn{name: "id"}
	sub := Raw("SELECT uid FROM users WHERE active = ?", true)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"EqSub", col.EqSub(sub), "id = (SELECT uid FROM users WHERE active = ?)", []any{true}},
		{"NotEqSub", col.NotEqSub(sub), "id != (SELECT uid FROM users WHERE active = ?)", []any{true}},
		{"InSub", col.InSub(sub), "id IN (SELECT uid FROM users WHERE active = ?)", []any{true}},
		{"NotInSub", col.NotInSub(sub), "id NOT IN (SELECT uid FROM users WHERE active = ?)", []any{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

// StringColumn tests

func TestStringColumn_Sql(t *testing.T) {
	tests := []struct {
		name     string
		column   StringColumn
		expected string
	}{
		{"no table/alias", StringColumn{name: "name"}, "name"},
		{"with table", StringColumn{name: "name", table: stringPtr("users")}, "users.name"},
		{"with alias", StringColumn{name: "name", alias: stringPtr("n")}, "n"},
		{"table and alias", StringColumn{name: "name", table: stringPtr("users"), alias: stringPtr("n")}, "n"},
		{"empty table", StringColumn{name: "name", table: stringPtr("")}, "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.column.Sql(); got != tt.expected {
				t.Errorf("StringColumn.Sql() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestStringColumn_Eq_NotEq(t *testing.T) {
	col := StringColumn{name: "name"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		arg  any
	}{
		{"Eq", col.Eq("alice"), "name = ?", "alice"},
		{"NotEq", col.NotEq("alice"), "name != ?", "alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != 1 || tt.expr.Args()[0] != tt.arg {
				t.Errorf("Args = %v, want [%v]", tt.expr.Args(), tt.arg)
			}
		})
	}
}

func TestStringColumn_Like(t *testing.T) {
	col := StringColumn{name: "name"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"like", col.Like("%a%"), "name LIKE ?", []any{"%a%"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestStringColumn_NotLike(t *testing.T) {
	col := StringColumn{name: "name"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"not like", col.NotLike("a%"), "name NOT LIKE ?", []any{"a%"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestStringColumn_In(t *testing.T) {
	col := StringColumn{name: "name"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"single", col.In("a"), "name IN (?)", []any{"a"}},
		{"values", col.In("a", "b"), "name IN (?, ?)", []any{"a", "b"}},
		{"empty", col.In(), "1=0", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestStringColumn_NotIn_InSub(t *testing.T) {
	col := StringColumn{name: "name"}
	sub := Raw("SELECT name FROM users WHERE active = ?", true)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"NotIn values", col.NotIn("x", "y"), "name NOT IN (?, ?)", []any{"x", "y"}},
		{"NotIn single", col.NotIn("x"), "name NOT IN (?)", []any{"x"}},
		{"NotIn empty", col.NotIn(), "1=1", nil},
		{"InSub", col.InSub(sub), "name IN (SELECT name FROM users WHERE active = ?)", []any{true}},
		{"NotInSub", col.NotInSub(sub), "name NOT IN (SELECT name FROM users WHERE active = ?)", []any{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestStringColumn_IsNull(t *testing.T) {
	col := StringColumn{name: "name"}
	expr := col.IsNull()
	if expr.Sql() != "name IS NULL" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "name IS NULL")
	}
	if expr.Args() != nil {
		t.Errorf("Args = %v, want nil", expr.Args())
	}
}

func TestStringColumn_IsNotNull(t *testing.T) {
	col := StringColumn{name: "name"}
	expr := col.IsNotNull()
	if expr.Sql() != "name IS NOT NULL" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "name IS NOT NULL")
	}
	if expr.Args() != nil {
		t.Errorf("Args = %v, want nil", expr.Args())
	}
}

func TestStringColumn_As_Alias_Table(t *testing.T) {
	base := StringColumn{name: "name", table: stringPtr("users")}
	tests := []struct {
		name      string
		col       StringColumn
		wantSql   string
		wantTable string
		wantAlias *string
	}{
		{"no alias", base, "users.name", "users", nil},
		{"with alias", base.As("n"), "n", "users", stringPtr("n")},
		{"empty alias", base.As(""), "", "users", stringPtr("")},
		{"no table", StringColumn{name: "name"}, "name", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.col.Sql() != tt.wantSql {
				t.Errorf("Sql = %q, want %q", tt.col.Sql(), tt.wantSql)
			}
			if tt.col.TableName() != tt.wantTable {
				t.Errorf("TableName = %q, want %q", tt.col.TableName(), tt.wantTable)
			}
			if tt.wantAlias == nil {
				if tt.col.Alias() != nil {
					t.Errorf("Alias = %v, want nil", tt.col.Alias())
				}
			} else if tt.col.Alias() == nil || *tt.col.Alias() != *tt.wantAlias {
				t.Errorf("Alias = %v, want %v", tt.col.Alias(), *tt.wantAlias)
			}
		})
	}
}

func TestStringColumn_SubQueries(t *testing.T) {
	col := StringColumn{name: "name"}
	sub := Raw("SELECT name FROM users WHERE id = ?", 1)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"EqSub", col.EqSub(sub), "name = (SELECT name FROM users WHERE id = ?)", []any{1}},
		{"NotEqSub", col.NotEqSub(sub), "name != (SELECT name FROM users WHERE id = ?)", []any{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

// BoolColumn tests

func TestBoolColumn_IsTrue(t *testing.T) {
	col := BoolColumn{name: "active"}
	expr := col.IsTrue()
	if expr.Sql() != "active = ?" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "active = ?")
	}
	if len(expr.Args()) != 1 || expr.Args()[0] != true {
		t.Errorf("Args = %v, want [true]", expr.Args())
	}
}

func TestBoolColumn_IsFalse(t *testing.T) {
	col := BoolColumn{name: "active"}
	expr := col.IsFalse()
	if expr.Sql() != "active = ?" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "active = ?")
	}
	if len(expr.Args()) != 1 || expr.Args()[0] != false {
		t.Errorf("Args = %v, want [false]", expr.Args())
	}
}

func TestBoolColumn_Eq(t *testing.T) {
	col := BoolColumn{name: "active"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		arg  any
	}{
		{"Eq true", col.Eq(true), "active = ?", true},
		{"Eq false", col.Eq(false), "active = ?", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != 1 || tt.expr.Args()[0] != tt.arg {
				t.Errorf("Args = %v, want [%v]", tt.expr.Args(), tt.arg)
			}
		})
	}
}

func TestBoolColumn_NotEq(t *testing.T) {
	col := BoolColumn{name: "active"}
	expr := col.NotEq(true)
	if expr.Sql() != "active != ?" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "active != ?")
	}
	if len(expr.Args()) != 1 || expr.Args()[0] != true {
		t.Errorf("Args = %v, want [true]", expr.Args())
	}
}

func TestBoolColumn_IsNull(t *testing.T) {
	col := BoolColumn{name: "active"}
	expr := col.IsNull()
	if expr.Sql() != "active IS NULL" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "active IS NULL")
	}
	if expr.Args() != nil {
		t.Errorf("Args = %v, want nil", expr.Args())
	}
}

func TestBoolColumn_IsNotNull(t *testing.T) {
	col := BoolColumn{name: "active"}
	expr := col.IsNotNull()
	if expr.Sql() != "active IS NOT NULL" {
		t.Errorf("Sql = %q, want %q", expr.Sql(), "active IS NOT NULL")
	}
	if expr.Args() != nil {
		t.Errorf("Args = %v, want nil", expr.Args())
	}
}

func TestBoolColumn_Sql(t *testing.T) {
	tests := []struct {
		name     string
		column   BoolColumn
		expected string
	}{
		{"with table", BoolColumn{name: "active", table: stringPtr("users")}, "users.active"},
		{"with alias", BoolColumn{name: "active", alias: stringPtr("a")}, "a"},
		{"table and alias", BoolColumn{name: "active", table: stringPtr("users"), alias: stringPtr("a")}, "a"},
		{"empty table", BoolColumn{name: "active", table: stringPtr("")}, "active"},
		{"no table/alias", BoolColumn{name: "active"}, "active"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.column.Sql() != tt.expected {
				t.Errorf("Sql = %q, want %q", tt.column.Sql(), tt.expected)
			}
		})
	}
}

func TestBoolColumn_TableName(t *testing.T) {
	tests := []struct {
		name     string
		column   BoolColumn
		expected string
	}{
		{"with table", BoolColumn{name: "active", table: stringPtr("users")}, "users"},
		{"empty table", BoolColumn{name: "active", table: stringPtr("")}, ""},
		{"nil table", BoolColumn{name: "active"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.column.TableName() != tt.expected {
				t.Errorf("TableName = %q, want %q", tt.column.TableName(), tt.expected)
			}
		})
	}
}

func TestBoolColumn_Alias_As(t *testing.T) {
	col := BoolColumn{name: "active", table: stringPtr("users")}

	tests := []struct {
		name     string
		column   BoolColumn
		expected *string
	}{
		{"no alias", col, nil},
		{"with alias", col.As("a"), stringPtr("a")},
		{"empty alias", col.As(""), stringPtr("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.Alias()
			if tt.expected == nil {
				if got != nil {
					t.Errorf("Alias = %v, want nil", got)
				}
				return
			}
			if got == nil || *got != *tt.expected {
				t.Errorf("Alias = %v, want %v", got, *tt.expected)
			}
		})
	}

	asCol := col.As("b")
	if asCol.name != "active" || asCol.table == nil || *asCol.table != "users" || asCol.alias == nil || *asCol.alias != "b" {
		t.Errorf("As() not preserving fields: %+v", asCol)
	}
}

// FloatColumn tests

func TestFloatColumn_Between(t *testing.T) {
	col := FloatColumn{name: "price"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"Between", col.Between(1.5, 3.5), "price BETWEEN ? AND ?", []any{1.5, 3.5}},
		{"Gt", col.Gt(10.0), "price > ?", []any{10.0}},
		{"Gte", col.Gte(10.0), "price >= ?", []any{10.0}},
		{"Lt", col.Lt(10.0), "price < ?", []any{10.0}},
		{"Lte", col.Lte(10.0), "price <= ?", []any{10.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestFloatColumn_Eq_NotEq(t *testing.T) {
	col := FloatColumn{name: "price"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		arg  any
	}{
		{"Eq", col.Eq(9.9), "price = ?", 9.9},
		{"NotEq", col.NotEq(9.9), "price != ?", 9.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != 1 || tt.expr.Args()[0] != tt.arg {
				t.Errorf("Args = %v, want [%v]", tt.expr.Args(), tt.arg)
			}
		})
	}
}

func TestFloatColumn_In_NotIn(t *testing.T) {
	col := FloatColumn{name: "price"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"In one", col.In(1.1), "price IN (?)", []any{1.1}},
		{"In empty", col.In(), "1=0", nil},
		{"NotIn one", col.NotIn(2.2), "price NOT IN (?)", []any{2.2}},
		{"NotIn empty", col.NotIn(), "1=1", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestFloatColumn_SubQueries(t *testing.T) {
	col := FloatColumn{name: "price"}
	sub := Raw("SELECT p FROM prices WHERE c = ?", "usd")
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"EqSub", col.EqSub(sub), "price = (SELECT p FROM prices WHERE c = ?)", []any{"usd"}},
		{"NotEqSub", col.NotEqSub(sub), "price != (SELECT p FROM prices WHERE c = ?)", []any{"usd"}},
		{"InSub", col.InSub(sub), "price IN (SELECT p FROM prices WHERE c = ?)", []any{"usd"}},
		{"NotInSub", col.NotInSub(sub), "price NOT IN (SELECT p FROM prices WHERE c = ?)", []any{"usd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestFloatColumn_As_Alias_Table(t *testing.T) {
	base := FloatColumn{name: "price", table: stringPtr("products")}
	tests := []struct {
		name      string
		col       FloatColumn
		wantSql   string
		wantTable string
		wantAlias *string
	}{
		{"no alias", base, "products.price", "products", nil},
		{"with alias", base.As("p"), "p", "products", stringPtr("p")},
		{"empty alias", base.As(""), "", "products", stringPtr("")},
		{"no table", FloatColumn{name: "price"}, "price", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.col.Sql() != tt.wantSql {
				t.Errorf("Sql = %q, want %q", tt.col.Sql(), tt.wantSql)
			}
			if tt.col.TableName() != tt.wantTable {
				t.Errorf("TableName = %q, want %q", tt.col.TableName(), tt.wantTable)
			}
			if tt.wantAlias == nil {
				if tt.col.Alias() != nil {
					t.Errorf("Alias = %v, want nil", tt.col.Alias())
				}
			} else if tt.col.Alias() == nil || *tt.col.Alias() != *tt.wantAlias {
				t.Errorf("Alias = %v, want %v", tt.col.Alias(), *tt.wantAlias)
			}
		})
	}
}

// TimeColumn tests

func TestTimeColumn_Between(t *testing.T) {
	col := TimeColumn{name: "created_at"}
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"Between", col.Between(t1, t2), "created_at BETWEEN ? AND ?", []any{t1, t2}},
		{"Gt", col.Gt(t1), "created_at > ?", []any{t1}},
		{"Gte", col.Gte(t1), "created_at >= ?", []any{t1}},
		{"Lt", col.Lt(t2), "created_at < ?", []any{t2}},
		{"Lte", col.Lte(t2), "created_at <= ?", []any{t2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestTimeColumn_Eq_NotEq(t *testing.T) {
	col := TimeColumn{name: "created_at"}
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		expr Expression
		sql  string
		arg  any
	}{
		{"Eq", col.Eq(t1), "created_at = ?", t1},
		{"NotEq", col.NotEq(t1), "created_at != ?", t1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != 1 || tt.expr.Args()[0] != tt.arg {
				t.Errorf("Args = %v, want [%v]", tt.expr.Args(), tt.arg)
			}
		})
	}
}

func TestTimeColumn_In_NotIn(t *testing.T) {
	col := TimeColumn{name: "created_at"}
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"In one", col.In(t1), "created_at IN (?)", []any{t1}},
		{"In two", col.In(t1, t2), "created_at IN (?, ?)", []any{t1, t2}},
		{"In empty", col.In(), "1=0", nil},
		{"NotIn one", col.NotIn(t1), "created_at NOT IN (?)", []any{t1}},
		{"NotIn empty", col.NotIn(), "1=1", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestTimeColumn_SubQueries(t *testing.T) {
	col := TimeColumn{name: "created_at"}
	sub := Raw("SELECT created_at FROM users WHERE active = ?", true)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"EqSub", col.EqSub(sub), "created_at = (SELECT created_at FROM users WHERE active = ?)", []any{true}},
		{"NotEqSub", col.NotEqSub(sub), "created_at != (SELECT created_at FROM users WHERE active = ?)", []any{true}},
		{"InSub", col.InSub(sub), "created_at IN (SELECT created_at FROM users WHERE active = ?)", []any{true}},
		{"NotInSub", col.NotInSub(sub), "created_at NOT IN (SELECT created_at FROM users WHERE active = ?)", []any{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestTimeColumn_As_Alias_Table(t *testing.T) {
	base := TimeColumn{name: "created_at", table: stringPtr("users")}
	tests := []struct {
		name      string
		col       TimeColumn
		wantSql   string
		wantTable string
		wantAlias *string
	}{
		{"no alias", base, "users.created_at", "users", nil},
		{"with alias", base.As("created"), "created", "users", stringPtr("created")},
		{"empty alias", base.As(""), "", "users", stringPtr("")},
		{"no table", TimeColumn{name: "created_at"}, "created_at", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.col.Sql() != tt.wantSql {
				t.Errorf("Sql = %q, want %q", tt.col.Sql(), tt.wantSql)
			}
			if tt.col.TableName() != tt.wantTable {
				t.Errorf("TableName = %q, want %q", tt.col.TableName(), tt.wantTable)
			}
			if tt.wantAlias == nil {
				if tt.col.Alias() != nil {
					t.Errorf("Alias = %v, want nil", tt.col.Alias())
				}
			} else if tt.col.Alias() == nil || *tt.col.Alias() != *tt.wantAlias {
				t.Errorf("Alias = %v, want %v", tt.col.Alias(), *tt.wantAlias)
			}
		})
	}
}

// UUIDColumn tests

func TestUUIDColumn_Eq(t *testing.T) {
	col := UUIDColumn{name: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"Eq", col.Eq("abc-123"), "id = ?", []any{"abc-123"}},
		{"NotEq", col.NotEq("abc-123"), "id != ?", []any{"abc-123"}},
		{"EqSub", col.EqSub(Raw("SELECT uid FROM t WHERE a = ?", 1)), "id = (SELECT uid FROM t WHERE a = ?)", []any{1}},
		{"NotEqSub", col.NotEqSub(Raw("SELECT uid FROM t WHERE a = ?", 1)), "id != (SELECT uid FROM t WHERE a = ?)", []any{1}},
		{"InSub", col.InSub(Raw("SELECT uid FROM t WHERE a = ?", 1)), "id IN (SELECT uid FROM t WHERE a = ?)", []any{1}},
		{"NotInSub", col.NotInSub(Raw("SELECT uid FROM t WHERE a = ?", 1)), "id NOT IN (SELECT uid FROM t WHERE a = ?)", []any{1}},
		{"In single", col.In("a"), "id IN (?)", []any{"a"}},
		{"In", col.In("a", "b"), "id IN (?, ?)", []any{"a", "b"}},
		{"NotIn single", col.NotIn("a"), "id NOT IN (?)", []any{"a"}},
		{"NotIn", col.NotIn("a", "b"), "id NOT IN (?, ?)", []any{"a", "b"}},
		{"In empty", col.In(), "1=0", nil},
		{"NotIn empty", col.NotIn(), "1=1", nil},
		{"IsNull", col.IsNull(), "id IS NULL", nil},
		{"IsNotNull", col.IsNotNull(), "id IS NOT NULL", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestUUIDColumn_As_Alias_Table(t *testing.T) {
	base := UUIDColumn{name: "id", table: stringPtr("users")}
	tests := []struct {
		name      string
		col       UUIDColumn
		wantSql   string
		wantTable string
		wantAlias *string
	}{
		{"no alias", base, "users.id", "users", nil},
		{"with alias", base.As("uid"), "uid", "users", stringPtr("uid")},
		{"empty alias", base.As(""), "", "users", stringPtr("")},
		{"no table", UUIDColumn{name: "id"}, "id", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.col.Sql() != tt.wantSql {
				t.Errorf("Sql = %q, want %q", tt.col.Sql(), tt.wantSql)
			}
			if tt.col.TableName() != tt.wantTable {
				t.Errorf("TableName = %q, want %q", tt.col.TableName(), tt.wantTable)
			}
			if tt.wantAlias == nil {
				if tt.col.Alias() != nil {
					t.Errorf("Alias = %v, want nil", tt.col.Alias())
				}
			} else if tt.col.Alias() == nil || *tt.col.Alias() != *tt.wantAlias {
				t.Errorf("Alias = %v, want %v", tt.col.Alias(), *tt.wantAlias)
			}
		})
	}
}

// Low-level col* helper tests

func TestColEq_NotEq(t *testing.T) {
	col := helperColumn{sql: "users.id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"Eq", colEq(col, 1), "users.id = ?", []any{1}},
		{"NotEq", colNotEq(col, 2), "users.id != ?", []any{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestColCompare(t *testing.T) {
	col := helperColumn{sql: "price"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"Gt", colGt(col, 10), "price > ?", []any{10}},
		{"Gte", colGte(col, 10), "price >= ?", []any{10}},
		{"Lt", colLt(col, 5), "price < ?", []any{5}},
		{"Lte", colLte(col, 5), "price <= ?", []any{5}},
		{"Between", colBetween(col, 1, 2), "price BETWEEN ? AND ?", []any{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestColIn_NotIn(t *testing.T) {
	col := helperColumn{sql: "id"}
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"In values", colIn(col, []any{1, 2}), "id IN (?, ?)", []any{1, 2}},
		{"In one", colIn(col, []any{1}), "id IN (?)", []any{1}},
		{"In empty", colIn(col, []any{}), "1=0", nil},
		{"NotIn values", colNotIn(col, []any{3, 4}), "id NOT IN (?, ?)", []any{3, 4}},
		{"NotIn one", colNotIn(col, []any{3}), "id NOT IN (?)", []any{3}},
		{"NotIn empty", colNotIn(col, []any{}), "1=1", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if tt.args == nil {
				if tt.expr.Args() != nil {
					t.Errorf("Args = %v, want nil", tt.expr.Args())
				}
				return
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}

func TestColNulls(t *testing.T) {
	col := helperColumn{sql: "id"}
	if colIsNull(col).Sql() != "id IS NULL" || colIsNull(col).Args() != nil {
		t.Errorf("colIsNull mismatch")
	}
	if colIsNotNull(col).Sql() != "id IS NOT NULL" || colIsNotNull(col).Args() != nil {
		t.Errorf("colIsNotNull mismatch")
	}
}

func TestColSubQueries(t *testing.T) {
	col := helperColumn{sql: "id"}
	sub := Raw("SELECT id FROM t WHERE a = ?", 10)
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"EqSub", colEqSub(col, sub), "id = (SELECT id FROM t WHERE a = ?)", []any{10}},
		{"NotEqSub", colNotEqSub(col, sub), "id != (SELECT id FROM t WHERE a = ?)", []any{10}},
		{"InSub", colInSub(col, sub), "id IN (SELECT id FROM t WHERE a = ?)", []any{10}},
		{"NotInSub", colNotInSub(col, sub), "id NOT IN (SELECT id FROM t WHERE a = ?)", []any{10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.expr.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(tt.expr.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if tt.expr.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, tt.expr.Args()[i], v)
				}
			}
		})
	}
}
