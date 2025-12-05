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

// --- BENCHMARKS ---

func BenchmarkReplacePlaceholders_Simple(b *testing.B) {
	dialects := []struct {
		name    string
		dialect Dialect
	}{
		{"SQLite", SQLiteDialect{}},
		{"PostgreSQL", PostgreSQLDialect{}},
		{"MSSQL", MSSQLDialect{}},
	}

	sql := "SELECT * FROM users WHERE id = ? AND status = ?"

	for _, d := range dialects {
		b.Run(d.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				replacePlaceholders(sql, d.dialect, 0)
			}
		})
	}
}

func BenchmarkReplacePlaceholders_NoOp(b *testing.B) {
	dialect := PostgreSQLDialect{}
	sql := "SELECT * FROM users WHERE id = 1 AND status = 'active'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacePlaceholders(sql, dialect, 0)
	}
}

func BenchmarkReplacePlaceholders_Complex(b *testing.B) {
	dialect := PostgreSQLDialect{}
	sql := "INSERT INTO t VALUES "
	for range 50 {
		sql += "(?, ?, ?),"
	}

	for b.Loop() {
		replacePlaceholders(sql, dialect, 0)
	}
}

func BenchmarkReplacePlaceholders_WithQuotes(b *testing.B) {
	dialect := PostgreSQLDialect{}
	sql := "SELECT 'test?' FROM table WHERE id = ? AND name = 'test?' AND value = ?"

	for b.Loop() {
		replacePlaceholders(sql, dialect, 0)
	}
}
