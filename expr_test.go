package dew

import "testing"

func TestReplacePlaceholders(t *testing.T) {
	dialects := map[string]Dialect{
		"SQLite":     SQLiteDialect{},
		"MySQL":      MySQLDialect{},
		"PostgreSQL": PostgreSQLDialect{},
		"MSSQL":      MSSQLDialect{},
	}

	tests := []struct {
		name      string
		sql       string
		argOffset int
		want      map[string]string
	}{
		{
			name:      "Empty string",
			sql:       "",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "",
				"MySQL":      "",
				"PostgreSQL": "",
				"MSSQL":      "",
			},
		},
		{
			name:      "No placeholders",
			sql:       "SELECT * FROM users",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT * FROM users",
				"MySQL":      "SELECT * FROM users",
				"PostgreSQL": "SELECT * FROM users",
				"MSSQL":      "SELECT * FROM users",
			},
		},
		{
			name:      "Single placeholder",
			sql:       "SELECT * FROM users WHERE id = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT * FROM users WHERE id = ?",
				"MySQL":      "SELECT * FROM users WHERE id = ?",
				"PostgreSQL": "SELECT * FROM users WHERE id = $1",
				"MSSQL":      "SELECT * FROM users WHERE id = @p1",
			},
		},
		{
			name:      "Multiple placeholders",
			sql:       "INSERT INTO users (name, age) VALUES (?, ?)",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "INSERT INTO users (name, age) VALUES (?, ?)",
				"MySQL":      "INSERT INTO users (name, age) VALUES (?, ?)",
				"PostgreSQL": "INSERT INTO users (name, age) VALUES ($1, $2)",
				"MSSQL":      "INSERT INTO users (name, age) VALUES (@p1, @p2)",
			},
		},
		{
			name:      "With offset",
			sql:       "UPDATE users SET name = ? WHERE id = ?",
			argOffset: 5,
			want: map[string]string{
				"SQLite":     "UPDATE users SET name = ? WHERE id = ?",
				"MySQL":      "UPDATE users SET name = ? WHERE id = ?",
				"PostgreSQL": "UPDATE users SET name = $6 WHERE id = $7",
				"MSSQL":      "UPDATE users SET name = @p6 WHERE id = @p7",
			},
		},
		{
			name:      "Ignore inside single quotes",
			sql:       "SELECT 'Is this a question?'",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'Is this a question?'",
				"MySQL":      "SELECT 'Is this a question?'",
				"PostgreSQL": "SELECT 'Is this a question?'",
				"MSSQL":      "SELECT 'Is this a question?'",
			},
		},
		{
			name:      "Mixed: outside and inside quotes",
			sql:       "SELECT 'Question?' FROM table WHERE id = ? AND text = 'Really?'",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'Question?' FROM table WHERE id = ? AND text = 'Really?'",
				"MySQL":      "SELECT 'Question?' FROM table WHERE id = ? AND text = 'Really?'",
				"PostgreSQL": "SELECT 'Question?' FROM table WHERE id = $1 AND text = 'Really?'",
				"MSSQL":      "SELECT 'Question?' FROM table WHERE id = @p1 AND text = 'Really?'",
			},
		},
		{
			name:      "SQL Standard Escaping (double single quote)",
			sql:       "SELECT 'It''s a match?' WHERE id = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'It''s a match?' WHERE id = ?",
				"MySQL":      "SELECT 'It''s a match?' WHERE id = ?",
				"PostgreSQL": "SELECT 'It''s a match?' WHERE id = $1",
				"MSSQL":      "SELECT 'It''s a match?' WHERE id = @p1",
			},
		},
		{
			name:      "Escaped single quote doubled",
			sql:       "SELECT 'O''Reilly' WHERE id = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'O''Reilly' WHERE id = ?",
				"MySQL":      "SELECT 'O''Reilly' WHERE id = ?",
				"PostgreSQL": "SELECT 'O''Reilly' WHERE id = $1",
				"MSSQL":      "SELECT 'O''Reilly' WHERE id = @p1",
			},
		},
		{
			name:      "Escaped single quote with backslash",
			sql:       "SELECT 'test\\'?' WHERE id = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'test\\'?' WHERE id = ?",
				"MySQL":      "SELECT 'test\\'?' WHERE id = ?",
				"PostgreSQL": "SELECT 'test\\'?' WHERE id = $1",
				"MSSQL":      "SELECT 'test\\'?' WHERE id = @p1",
			},
		},
		{
			name:      "UTF-8 Cyrillic check",
			sql:       "SELECT 'Привет?' WHERE name = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'Привет?' WHERE name = ?",
				"MySQL":      "SELECT 'Привет?' WHERE name = ?",
				"PostgreSQL": "SELECT 'Привет?' WHERE name = $1",
				"MSSQL":      "SELECT 'Привет?' WHERE name = @p1",
			},
		},
		{
			name:      "Multiline string",
			sql:       "SELECT *\nFROM table\nWHERE id = ?",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT *\nFROM table\nWHERE id = ?",
				"MySQL":      "SELECT *\nFROM table\nWHERE id = ?",
				"PostgreSQL": "SELECT *\nFROM table\nWHERE id = $1",
				"MSSQL":      "SELECT *\nFROM table\nWHERE id = @p1",
			},
		},
		{
			name:      "Many placeholders",
			sql:       "INSERT INTO t VALUES (?, ?, ?, ?, ?)",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "INSERT INTO t VALUES (?, ?, ?, ?, ?)",
				"MySQL":      "INSERT INTO t VALUES (?, ?, ?, ?, ?)",
				"PostgreSQL": "INSERT INTO t VALUES ($1, $2, $3, $4, $5)",
				"MSSQL":      "INSERT INTO t VALUES (@p1, @p2, @p3, @p4, @p5)",
			},
		},
		{
			name:      "Nested quotes",
			sql:       "SELECT 'test' WHERE id = ? AND name = 'test?'",
			argOffset: 0,
			want: map[string]string{
				"SQLite":     "SELECT 'test' WHERE id = ? AND name = 'test?'",
				"MySQL":      "SELECT 'test' WHERE id = ? AND name = 'test?'",
				"PostgreSQL": "SELECT 'test' WHERE id = $1 AND name = 'test?'",
				"MSSQL":      "SELECT 'test' WHERE id = @p1 AND name = 'test?'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for dialectName, dialect := range dialects {
				t.Run(dialectName, func(t *testing.T) {
					got := replacePlaceholders(tt.sql, dialect, tt.argOffset)
					want := tt.want[dialectName]
					if got != want {
						t.Errorf("replacePlaceholders() = %q, want %q", got, want)
					}
				})
			}
		})
	}
}

func TestReplacePlaceholders_NilDialect(t *testing.T) {
	tests := []string{
		"SELECT ?",
		"SELECT * FROM users WHERE id = ?",
		"INSERT INTO users (name, age) VALUES (?, ?)",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			got := replacePlaceholders(sql, nil, 0)
			if got != sql {
				t.Errorf("replacePlaceholders() with nil dialect = %q, want %q", got, sql)
			}
		})
	}
}

// --- Additional unit tests for expr.go ---

// helper test column implementing Column
type testColumn struct {
	sql   string
	name  string
	table string
	alias *string
}

func (c testColumn) Sql() string        { return c.sql }
func (c testColumn) Args() []any        { return nil }
func (c testColumn) ColumnName() string { return c.name }
func (c testColumn) TableName() string  { return c.table }
func (c testColumn) Alias() *string     { return c.alias }

func TestBuildPlaceholders(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{0, "()"},
		{1, "(?)"},
		{3, "(?, ?, ?)"},
	}
	for _, tt := range tests {
		if got := buildPlaceholders(tt.count); got != tt.want {
			t.Errorf("buildPlaceholders(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestOperatorString(t *testing.T) {
	if AND.String() != " AND " || OR.String() != " OR " {
		t.Errorf("operator String() mismatch: AND=%q OR=%q", AND.String(), OR.String())
	}
}

func TestCompoundExpr(t *testing.T) {
	a := simpleExpr{sql: "a = ?", args: []any{1}}
	b := simpleExpr{sql: "b = ?", args: []any{2}}

	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"And", And(a, b), "(a = ? AND b = ?)", []any{1, 2}},
		{"Or", Or(a, b), "(a = ? OR b = ?)", []any{1, 2}},
		{"Empty And", And(), "", nil},
		{"Empty Or", Or(), "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
			if len(tt.args) == 0 {
				if len(tt.expr.Args()) != 0 {
					t.Errorf("Args = %v, want []", tt.expr.Args())
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

func TestSortExpr(t *testing.T) {
	alias := "c_alias"
	colWithAlias := testColumn{sql: "col_sql", name: "c", alias: &alias}
	colWithAs := testColumn{sql: "col_sql AS x", name: "c"}
	colWithAsMixed := testColumn{sql: "COL_SQL as X", name: "c"}
	expr := simpleExpr{sql: "expr_sql"}

	tests := []struct {
		name string
		in   any
		desc bool
		want string
	}{
		{"column alias", colWithAlias, false, "c_alias ASC"},
		{"column with AS", colWithAs, false, "col_sql ASC"},
		{"column with AS mixed (Desc)", colWithAsMixed, true, "COL_SQL DESC"},
		{"expression", expr, true, "expr_sql DESC"},
		{"plain value", "price", false, "price ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Expression
			if tt.desc {
				e = Desc(tt.in)
			} else {
				e = Asc(tt.in)
			}
			if e.Sql() != tt.want {
				t.Errorf("Sql() = %q, want %q", e.Sql(), tt.want)
			}
			if args := e.Args(); args != nil {
				t.Errorf("Args() = %v, want nil", args)
			}
		})
	}
}

func TestAggColumns(t *testing.T) {
	base := simpleExpr{sql: "price"}
	withAlias := &aggColumn{fnType: SUM, col: "price", alias: "total_price"}

	col := testColumn{sql: "users.id", name: "id"}
	tests := []struct {
		name string
		expr Column
		sql  string
	}{
		{"Sum", Sum(base), "SUM(price)"},
		{"Sum with alias", withAlias, "SUM(price) AS total_price"},
		{"Count default", Count(), "COUNT(*)"},
		{"Count column", Count(col), "COUNT(users.id)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", tt.expr.Sql(), tt.sql)
			}
		})
	}

	if cn := withAlias.ColumnName(); cn != "total_price" {
		t.Errorf("ColumnName = %q, want %q", cn, "total_price")
	}
	if Count().Alias() != nil {
		t.Errorf("Count().Alias() = %v, want nil", Count().Alias())
	}
}

func TestAliasExpr(t *testing.T) {
	base := simpleExpr{sql: "a + b", args: []any{1, 2}}
	col := As(base, "sum_ab")

	tests := []struct {
		name string
		sql  string
		args []any
	}{
		{"alias", "a + b AS sum_ab", []any{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if col.Sql() != tt.sql {
				t.Errorf("Sql = %q, want %q", col.Sql(), tt.sql)
			}
			if len(col.Args()) != len(tt.args) {
				t.Errorf("Args len = %d, want %d", len(col.Args()), len(tt.args))
				return
			}
			for i, v := range tt.args {
				if col.Args()[i] != v {
					t.Errorf("Args[%d]=%v, want %v", i, col.Args()[i], v)
				}
			}
			if col.ColumnName() != "sum_ab" {
				t.Errorf("ColumnName = %q, want %q", col.ColumnName(), "sum_ab")
			}
			if col.TableName() != "" {
				t.Errorf("TableName = %q, want empty", col.TableName())
			}
			if col.Alias() == nil || *col.Alias() != "sum_ab" {
				t.Errorf("Alias() = %v, want sum_ab", col.Alias())
			}
		})
	}
}

func TestRaw(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
		sql  string
		args []any
	}{
		{"one arg", Raw("NOW() + ?", 5), "NOW() + ?", []any{5}},
		{"no args", Raw("NOW()"), "NOW()", nil},
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
