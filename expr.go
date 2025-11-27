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

type IntColumn string

func (c IntColumn) Sql() string { return string(c) }
func (c IntColumn) Args() []any { return nil }

func (c IntColumn) Eq(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{val},
	}
}

func (c IntColumn) Gt(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s > ?", c),
		args: []any{val},
	}
}

func (c IntColumn) Lt(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s < ?", c),
		args: []any{val},
	}
}

func (c IntColumn) In(vals ...int) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=0"} // Всегда ложь для пустого списка
	}
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN %s", c, buildPlaceholders(len(vals))),
		args: args,
	}
}

type StringColumn string

func (c StringColumn) Sql() string { return string(c) }
func (c StringColumn) Args() []any { return nil }

func (c StringColumn) Eq(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{val},
	}
}

func (c StringColumn) Like(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s LIKE ?", c),
		args: []any{val},
	}
}

func (c StringColumn) In(vals ...string) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=0"}
	}
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN %s", c, buildPlaceholders(len(vals))),
		args: args,
	}
}

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
	return &sortExpr{
		column: fmt.Sprintf("%v", col),
		dir:    "ASC",
	}
}

func Desc(col any) Expression {
	return &sortExpr{
		column: fmt.Sprintf("%v", col),
		dir:    "DESC",
	}
}

// Agregate Funtions //

type AGREGATE_FUNCTION_TYPE string

const (
	SUM AGREGATE_FUNCTION_TYPE = "SUM"
	AVG AGREGATE_FUNCTION_TYPE = "AVG"
	MAX AGREGATE_FUNCTION_TYPE = "MAX"
	MIN AGREGATE_FUNCTION_TYPE = "MIN"
)

type aggExpr struct {
	fnType AGREGATE_FUNCTION_TYPE
	col    string
}

func (a *aggExpr) Sql() string { return fmt.Sprintf("%s(%s)", a.fnType, a.col) }

func (a *aggExpr) Args() []any { return nil }

func Sum(col Expression) Expression { return &aggExpr{fnType: SUM, col: col.Sql()} }

func Avg(col Expression) Expression { return &aggExpr{fnType: AVG, col: col.Sql()} }

func Max(col Expression) Expression { return &aggExpr{fnType: MAX, col: col.Sql()} }

func Min(col Expression) Expression { return &aggExpr{fnType: MIN, col: col.Sql()} }

// * Alias Expr * //

type aliasExpr struct {
	expr  Expression
	alias string
}

func (a *aliasExpr) Sql() string { return fmt.Sprintf("%s AS %s", a.expr.Sql(), a.alias) }

func (a *aliasExpr) Args() []any { return a.expr.Args() }

func As(exp Expression, alias string) Expression { return &aliasExpr{expr: exp, alias: alias} }

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
