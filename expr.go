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
	// В простой сортировке аргументов нет
	return nil
}

// Asc возвращает Expression для сортировки по возрастанию
func Asc(col any) Expression {
	return &sortExpr{
		column: fmt.Sprintf("%v", col),
		dir:    "ASC",
	}
}

// Desc возвращает Expression для сортировки по убыванию
func Desc(col any) Expression {
	return &sortExpr{
		column: fmt.Sprintf("%v", col),
		dir:    "DESC",
	}
}

func buildPlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return "(" + strings.Join(placeholders, ", ") + ")"
}
