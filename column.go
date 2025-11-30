package dew

import "fmt"

type Column interface {
	Expression
	ColumnName() string
	TableName() string
}

type IntColumn struct {
	name  string
	table string
}

func (c IntColumn) Sql() string        { return c.name }
func (c IntColumn) Args() []any        { return nil }
func (c IntColumn) ColumnName() string { return c.name }
func (c IntColumn) TableName() string  { return c.table }

func (c IntColumn) Eq(val int) Expression {
	return colEq(c, val)
}

func (c IntColumn) NotEq(val int) Expression {
	return colNotEq(c, val)
}

func (c IntColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c IntColumn) NotEqSub(subQuery Expression) Expression {
	return colNotEqSub(c, subQuery)
}

func (c IntColumn) Gt(val int) Expression {
	return colGt(c, val)
}

func (c IntColumn) Gte(val int) Expression {
	return colGte(c, val)
}

func (c IntColumn) Lt(val int) Expression {
	return colLt(c, val)
}

func (c IntColumn) Lte(val int) Expression {
	return colLte(c, val)
}

func (c IntColumn) In(vals ...int) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c IntColumn) InSub(subQuery Expression) Expression {
	return colInSub(c, subQuery)
}

func (c IntColumn) NotIn(vals ...int) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c IntColumn) NotInSub(subQuery Expression) Expression {
	return colNotInSub(c, subQuery)
}

func (c IntColumn) Between(min, max int) Expression {
	return colBetween(c, min, max)
}

func (c IntColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c IntColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

type StringColumn struct {
	name  string
	table string
}

func (c StringColumn) Sql() string        { return c.name }
func (c StringColumn) Args() []any        { return nil }
func (c StringColumn) ColumnName() string { return c.name }
func (c StringColumn) TableName() string  { return c.table }

func (c StringColumn) Eq(val string) Expression {
	return colEq(c, val)
}

func (c StringColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c StringColumn) NotEq(val string) Expression {
	return colNotEq(c, val)
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
	return colIn(c, convertToAny(vals...))
}

func (c StringColumn) InSub(subQuery Expression) Expression {
	return colInSub(c, subQuery)
}

func (c StringColumn) NotIn(vals ...string) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c StringColumn) NotInSub(subQuery Expression) Expression {
	return colNotInSub(c, subQuery)
}

func (c StringColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c StringColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

type BoolColumn struct {
	name  string
	table string
}

func (c BoolColumn) Sql() string        { return c.name }
func (c BoolColumn) Args() []any        { return nil }
func (c BoolColumn) ColumnName() string { return c.name }
func (c BoolColumn) TableName() string  { return c.table }

func (c BoolColumn) Eq(val bool) Expression {
	return colEq(c, val)
}

func (c BoolColumn) IsTrue() Expression {
	return colEq(c, true)
}

func (c BoolColumn) IsFalse() Expression {
	return colEq(c, false)
}

func (c BoolColumn) NotEq(val bool) Expression {
	return colNotEq(c, val)
}

func (c BoolColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c BoolColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

type FloatColumn struct {
	name  string
	table string
}

func (c FloatColumn) Sql() string        { return c.name }
func (c FloatColumn) Args() []any        { return nil }
func (c FloatColumn) ColumnName() string { return c.name }
func (c FloatColumn) TableName() string  { return c.table }

func (c FloatColumn) Eq(val float64) Expression {
	return colEq(c, val)
}

func (c FloatColumn) NotEq(val float64) Expression {
	return colNotEq(c, val)
}

func (c FloatColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c FloatColumn) NotEqSub(subQuery Expression) Expression {
	return colNotEqSub(c, subQuery)
}

func (c FloatColumn) Gt(val float64) Expression {
	return colGt(c, val)
}

func (c FloatColumn) Gte(val float64) Expression {
	return colGte(c, val)
}

func (c FloatColumn) Lt(val float64) Expression {
	return colLt(c, val)
}

func (c FloatColumn) Lte(val float64) Expression {
	return colLte(c, val)
}

func (c FloatColumn) In(vals ...float64) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c FloatColumn) InSub(subQuery Expression) Expression {
	return colInSub(c, subQuery)
}

func (c FloatColumn) NotIn(vals ...float64) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c FloatColumn) NotInSub(subQuery Expression) Expression {
	return colNotInSub(c, subQuery)
}

func (c FloatColumn) Between(min, max float64) Expression {
	return colBetween(c, min, max)
}

func (c FloatColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c FloatColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

func colEq(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", col.Sql()),
		args: []any{val},
	}
}

func colNotEq(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", col.Sql()),
		args: []any{val},
	}
}

func colEqSub(col Column, subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = (%s)", col.Sql(), subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func colNotEqSub(col Column, subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != (%s)", col.Sql(), subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func colGt(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s > ?", col.Sql()),
		args: []any{val},
	}
}

func colGte(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s >= ?", col.Sql()),
		args: []any{val},
	}
}

func colLt(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s < ?", col.Sql()),
		args: []any{val},
	}
}

func colLte(col Column, val any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s <= ?", col.Sql()),
		args: []any{val},
	}
}

func colIn(col Column, vals []any) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=0"}
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN %s", col.Sql(), buildPlaceholders(len(vals))),
		args: vals,
	}
}

func colInSub(col Column, subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IN (%s)", col.Sql(), subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func colNotIn(col Column, vals []any) Expression {
	if len(vals) == 0 {
		return &simpleExpr{sql: "1=1"}
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN %s", col.Sql(), buildPlaceholders(len(vals))),
		args: vals,
	}
}

func colNotInSub(col Column, subQuery Expression) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT IN (%s)", col.Sql(), subQuery.Sql()),
		args: subQuery.Args(),
	}
}

func colBetween(col Column, min, max any) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s BETWEEN ? AND ?", col.Sql()),
		args: []any{min, max},
	}
}

func colIsNull(col Column) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IS NULL", col.Sql()),
		args: nil,
	}
}

func colIsNotNull(col Column) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s IS NOT NULL", col.Sql()),
		args: nil,
	}
}

func convertToAny[T any](vals ...T) []any {
	result := make([]any, len(vals))
	for i, v := range vals {
		result[i] = v
	}
	return result
}
