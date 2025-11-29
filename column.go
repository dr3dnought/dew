package dew

import "fmt"

type Column interface {
	Expression
	ColumnName() string
}

type IntColumn string

func (c IntColumn) Sql() string        { return string(c) }
func (c IntColumn) Args() []any        { return nil }
func (c IntColumn) ColumnName() string { return string(c) }

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

func (c StringColumn) Sql() string        { return string(c) }
func (c StringColumn) Args() []any        { return nil }
func (c StringColumn) ColumnName() string { return string(c) }

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

type BoolColumn string

func (c BoolColumn) Sql() string        { return string(c) }
func (c BoolColumn) Args() []any        { return nil }
func (c BoolColumn) ColumnName() string { return string(c) }

func (c BoolColumn) Eq(val bool) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{val},
	}
}

func (c BoolColumn) IsTrue() Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{true},
	}
}

func (c BoolColumn) IsFalse() Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{false},
	}
}
