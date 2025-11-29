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

func (c IntColumn) NotEq(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", c),
		args: []any{val},
	}
}

func (c IntColumn) EqSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c IntColumn) NotEqSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c IntColumn) Gt(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s > ?", c),
		args: []any{val},
	}
}

func (c IntColumn) Gte(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s >= ?", c),
		args: []any{val},
	}
}

func (c IntColumn) Lt(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s < ?", c),
		args: []any{val},
	}
}

func (c IntColumn) Lte(val int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s <= ?", c),
		args: []any{val},
	}
}

func (c IntColumn) In(vals ...int) Expression {
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

func (c IntColumn) InSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c IntColumn) NotIn(vals ...int) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=1"}
	}
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN %s", c, buildPlaceholders(len(vals))),
		args: args,
	}
}

func (c IntColumn) NotInSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c IntColumn) Between(min, max int) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s BETWEEN ? AND ?", c),
		args: []any{min, max},
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

func (c StringColumn) EqSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c StringColumn) NotEq(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", c),
		args: []any{val},
	}
}

func (c StringColumn) Like(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s LIKE ?", c),
		args: []any{val},
	}
}

func (c StringColumn) NotLike(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT LIKE ?", c),
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

func (c StringColumn) InSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c StringColumn) NotIn(vals ...string) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=1"}
	}
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN %s", c, buildPlaceholders(len(vals))),
		args: args,
	}
}

func (c StringColumn) NotInSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
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

func (c BoolColumn) NotEq(val bool) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", c),
		args: []any{val},
	}
}

type FloatColumn string

func (c FloatColumn) Sql() string        { return string(c) }
func (c FloatColumn) Args() []any        { return nil }
func (c FloatColumn) ColumnName() string { return string(c) }

func (c FloatColumn) Eq(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) NotEq(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) EqSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c FloatColumn) NotEqSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c FloatColumn) Gt(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s > ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) Gte(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s >= ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) Lt(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s < ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) Lte(val float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s <= ?", c),
		args: []any{val},
	}
}

func (c FloatColumn) In(vals ...float64) Expression {
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

func (c FloatColumn) InSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c FloatColumn) NotIn(vals ...float64) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=1"}
	}
	args := make([]any, len(vals))
	for i, v := range vals {
		args[i] = v
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN %s", c, buildPlaceholders(len(vals))),
		args: args,
	}
}

func (c FloatColumn) NotInSub(subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN (%s)", c, subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func (c FloatColumn) Between(min, max float64) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s BETWEEN ? AND ?", c),
		args: []any{min, max},
	}
}
