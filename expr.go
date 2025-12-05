package dew

import (
	"fmt"
	"strings"
)

type Expression interface {
	Sql() string
	Args() []any
}

type simpleExpr struct {
	sql  string
	args []any
}

func (e simpleExpr) Sql() string { return e.sql }
func (e simpleExpr) Args() []any { return e.args }

var _ Expression = simpleExpr{}

type OPERATOR string

const (
	AND OPERATOR = " AND "
	OR  OPERATOR = " OR "
)

func (o OPERATOR) String() string {
	return string(o)
}

type compoundExpr struct {
	operator OPERATOR
	exprs    []Expression
}

func (c *compoundExpr) Sql() string {
	if len(c.exprs) == 0 {
		return ""
	}

	var parts []string
	for _, e := range c.exprs {
		parts = append(parts, e.Sql())
	}

	return "(" + strings.Join(parts, string(c.operator)) + ")"
}

func (c *compoundExpr) Args() []any {
	var args []any
	for _, e := range c.exprs {
		args = append(args, e.Args()...)
	}
	return args
}

// PUBLIC OPERATORS HELPERS

func And(exprs ...Expression) Expression {
	return &compoundExpr{
		operator: AND,
		exprs:    exprs,
	}
}

func Or(exprs ...Expression) Expression {
	return &compoundExpr{
		operator: OR,
		exprs:    exprs,
	}
}

type sortExpr struct {
	column string
	dir    string // ASC или DESC
}

func (s *sortExpr) Sql() string {
	return fmt.Sprintf("%s %s", s.column, s.dir)
}

func (s *sortExpr) Args() []any {
	return nil
}

func Asc(col any) Expression {
	var colSql string
	if colExpr, ok := col.(Column); ok {
		if alias := colExpr.Alias(); alias != nil {
			colSql = *alias
		} else {
			sqlStr := colExpr.Sql()
			if idx := strings.Index(sqlStr, " AS "); idx != -1 {
				colSql = sqlStr[:idx]
			} else {
				colSql = sqlStr
			}
		}
	} else if expr, ok := col.(Expression); ok {
		colSql = expr.Sql()
	} else {
		colSql = fmt.Sprintf("%v", col)
	}
	return &sortExpr{
		column: colSql,
		dir:    "ASC",
	}
}

func Desc(col any) Expression {
	var colSql string
	if colExpr, ok := col.(Column); ok {
		if alias := colExpr.Alias(); alias != nil {
			colSql = *alias
		} else {
			sqlStr := colExpr.Sql()
			if idx := strings.Index(sqlStr, " AS "); idx != -1 {
				colSql = sqlStr[:idx]
			} else {
				colSql = sqlStr
			}
		}
	} else if expr, ok := col.(Expression); ok {
		colSql = expr.Sql()
	} else {
		colSql = fmt.Sprintf("%v", col)
	}
	return &sortExpr{
		column: colSql,
		dir:    "DESC",
	}
}

// Agregate Funtions //

type AGREGATE_FUNCTION_TYPE string

const (
	SUM   AGREGATE_FUNCTION_TYPE = "SUM"
	AVG   AGREGATE_FUNCTION_TYPE = "AVG"
	MAX   AGREGATE_FUNCTION_TYPE = "MAX"
	MIN   AGREGATE_FUNCTION_TYPE = "MIN"
	COUNT AGREGATE_FUNCTION_TYPE = "COUNT"
)

type aggColumn struct {
	fnType AGREGATE_FUNCTION_TYPE
	col    string
	alias  string
}

func (a *aggColumn) Sql() string {
	sql := fmt.Sprintf("%s(%s)", a.fnType, a.col)
	if a.alias != "" {
		return fmt.Sprintf("%s AS %s", sql, a.alias)
	}
	return sql
}

func (a *aggColumn) Args() []any { return nil }

func (a *aggColumn) ColumnName() string {
	if a.alias != "" {
		return a.alias
	}
	return strings.ToLower(string(a.fnType))
}

func (a *aggColumn) TableName() string { return "" }

func (a *aggColumn) Alias() *string {
	if a.alias != "" {
		return &a.alias
	}
	return nil
}

func Sum(col Expression) Column {
	return &aggColumn{fnType: SUM, col: col.Sql()}
}

func Avg(col Expression) Column {
	return &aggColumn{fnType: AVG, col: col.Sql()}
}

func Max(col Expression) Column {
	return &aggColumn{fnType: MAX, col: col.Sql()}
}

func Min(col Expression) Column {
	return &aggColumn{fnType: MIN, col: col.Sql()}
}

func Count(columns ...Column) Column {
	col := "*"
	if len(columns) > 0 {
		col = columns[0].Sql()
	}
	return &aggColumn{fnType: COUNT, col: col}
}

// * Alias Expr * //

type aliasExpr struct {
	expr  Expression
	alias string
}

func (a *aliasExpr) Sql() string { return fmt.Sprintf("%s AS %s", a.expr.Sql(), a.alias) }

func (a *aliasExpr) Args() []any { return a.expr.Args() }

func (a *aliasExpr) ColumnName() string { return a.alias }

func (a *aliasExpr) TableName() string { return "" }

func (a *aliasExpr) Alias() *string { return &a.alias }

func As(exp Expression, alias string) Column { return &aliasExpr{expr: exp, alias: alias} }

// * Raw Expr * //

type rawExpr struct {
	sql  string
	args []any
}

func (e *rawExpr) Sql() string { return e.sql }
func (e *rawExpr) Args() []any { return e.args }

func Raw(sql string, args ...any) Expression {
	return &rawExpr{
		sql:  sql,
		args: args,
	}
}

// * Helpers * //

func buildPlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return "(" + strings.Join(placeholders, ", ") + ")"
}

func replacePlaceholders(sql string, dialect Dialect, argOffset int) string {
	if dialect == nil {
		return sql
	}

	if strings.IndexByte(sql, '?') == -1 {
		return sql
	}

	inQuote := false

	var b strings.Builder
	b.Grow(len(sql) + 16)

	for i := 0; i < len(sql); i++ {
		c := sql[i]

		isEscaped := i > 0 && sql[i-1] == '\\'

		if c == '\'' && !isEscaped {
			inQuote = !inQuote
		}

		if c == '?' && !inQuote {
			b.WriteString(dialect.Placeholder(argOffset))
			argOffset++
		} else {
			b.WriteByte(c)
		}
	}

	return b.String()
}
