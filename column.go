/*
TODO:

	Check if alias pointer is nil or not in As method
*/
package dew

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// JSONB is the constraint for JSONBColumn type parameters.
// Types used with JSONBColumn must implement both sql.Scanner (for reading)
// and driver.Valuer (for writing).
type JSONB interface {
	sql.Scanner
	driver.Valuer
}

type Column interface {
	Expression
	ColumnName() string
	TableName() string
	Alias() *string
}

type IntColumn struct {
	name  string
	table *string
	alias *string
}

func (c IntColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c IntColumn) Args() []any        { return nil }
func (c IntColumn) ColumnName() string { return c.name }
func (c IntColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c IntColumn) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c IntColumn) As(alias string) IntColumn {
	return IntColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

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

type Int64Column struct {
	name  string
	table *string
	alias *string
}

func (c Int64Column) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c Int64Column) Args() []any        { return nil }
func (c Int64Column) ColumnName() string { return c.name }
func (c Int64Column) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c Int64Column) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c Int64Column) As(alias string) Int64Column {
	return Int64Column{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c Int64Column) Eq(val int64) Expression    { return colEq(c, val) }
func (c Int64Column) NotEq(val int64) Expression { return colNotEq(c, val) }
func (c Int64Column) Gt(val int64) Expression    { return colGt(c, val) }
func (c Int64Column) Gte(val int64) Expression   { return colGte(c, val) }
func (c Int64Column) Lt(val int64) Expression    { return colLt(c, val) }
func (c Int64Column) Lte(val int64) Expression   { return colLte(c, val) }

func (c Int64Column) EqSub(subQuery Expression) Expression    { return colEqSub(c, subQuery) }
func (c Int64Column) NotEqSub(subQuery Expression) Expression { return colNotEqSub(c, subQuery) }

func (c Int64Column) In(vals ...int64) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c Int64Column) InSub(subQuery Expression) Expression { return colInSub(c, subQuery) }

func (c Int64Column) NotIn(vals ...int64) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c Int64Column) NotInSub(subQuery Expression) Expression { return colNotInSub(c, subQuery) }

func (c Int64Column) Between(min, max int64) Expression { return colBetween(c, min, max) }
func (c Int64Column) IsNull() Expression                { return colIsNull(c) }
func (c Int64Column) IsNotNull() Expression             { return colIsNotNull(c) }

type StringColumn struct {
	name  string
	table *string
	alias *string
}

func (c StringColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c StringColumn) Args() []any        { return nil }
func (c StringColumn) ColumnName() string { return c.name }
func (c StringColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c StringColumn) Alias() *string { return c.alias }

func (c StringColumn) As(alias string) StringColumn {
	return StringColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c StringColumn) Eq(val string) Expression {
	return colEq(c, val)
}

func (c StringColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c StringColumn) NotEqSub(subQuery Expression) Expression {
	return colNotEqSub(c, subQuery)
}

func (c StringColumn) NotEq(val string) Expression {
	return colNotEq(c, val)
}

func (c StringColumn) Like(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s LIKE ?", c.Sql()),
		args: []any{val},
	}
}

func (c StringColumn) NotLike(val string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s NOT LIKE ?", c.Sql()),
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
	table *string
	alias *string
}

func (c BoolColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c BoolColumn) Args() []any        { return nil }
func (c BoolColumn) ColumnName() string { return c.name }
func (c BoolColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c BoolColumn) Alias() *string { return c.alias }

func (c BoolColumn) As(alias string) BoolColumn {
	return BoolColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

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
	table *string
	alias *string
}

func (c FloatColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c FloatColumn) Args() []any        { return nil }
func (c FloatColumn) ColumnName() string { return c.name }
func (c FloatColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c FloatColumn) Alias() *string { return c.alias }

func (c FloatColumn) As(alias string) FloatColumn {
	return FloatColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

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

type TimeColumn struct {
	name  string
	table *string
	alias *string
}

func (c TimeColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c TimeColumn) Args() []any        { return nil }
func (c TimeColumn) ColumnName() string { return c.name }
func (c TimeColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c TimeColumn) Alias() *string { return c.alias }

func (c TimeColumn) As(alias string) TimeColumn {
	return TimeColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c TimeColumn) Eq(val time.Time) Expression {
	return colEq(c, val)
}

func (c TimeColumn) NotEq(val time.Time) Expression {
	return colNotEq(c, val)
}

func (c TimeColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c TimeColumn) NotEqSub(subQuery Expression) Expression {
	return colNotEqSub(c, subQuery)
}

func (c TimeColumn) Gt(val time.Time) Expression {
	return colGt(c, val)
}

func (c TimeColumn) Gte(val time.Time) Expression {
	return colGte(c, val)
}

func (c TimeColumn) Lt(val time.Time) Expression {
	return colLt(c, val)
}

func (c TimeColumn) Lte(val time.Time) Expression {
	return colLte(c, val)
}

func (c TimeColumn) In(vals ...time.Time) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c TimeColumn) InSub(subQuery Expression) Expression {
	return colInSub(c, subQuery)
}

func (c TimeColumn) NotIn(vals ...time.Time) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c TimeColumn) NotInSub(subQuery Expression) Expression {
	return colNotInSub(c, subQuery)
}

func (c TimeColumn) Between(min, max time.Time) Expression {
	return colBetween(c, min, max)
}

func (c TimeColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c TimeColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

type UUIDColumn struct {
	name  string
	table *string
	alias *string
}

func (c UUIDColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c UUIDColumn) Args() []any        { return nil }
func (c UUIDColumn) ColumnName() string { return c.name }
func (c UUIDColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c UUIDColumn) Alias() *string { return c.alias }

func (c UUIDColumn) As(alias string) UUIDColumn {
	return UUIDColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c UUIDColumn) Eq(val string) Expression {
	return colEq(c, val)
}

func (c UUIDColumn) NotEq(val string) Expression {
	return colNotEq(c, val)
}

func (c UUIDColumn) EqSub(subQuery Expression) Expression {
	return colEqSub(c, subQuery)
}

func (c UUIDColumn) NotEqSub(subQuery Expression) Expression {
	return colNotEqSub(c, subQuery)
}

func (c UUIDColumn) In(vals ...string) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c UUIDColumn) InSub(subQuery Expression) Expression {
	return colInSub(c, subQuery)
}

func (c UUIDColumn) NotIn(vals ...string) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c UUIDColumn) NotInSub(subQuery Expression) Expression {
	return colNotInSub(c, subQuery)
}

func (c UUIDColumn) IsNull() Expression {
	return colIsNull(c)
}

func (c UUIDColumn) IsNotNull() Expression {
	return colIsNotNull(c)
}

type DecimalColumn struct {
	name  string
	table *string
	alias *string
}

func (c DecimalColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c DecimalColumn) Args() []any        { return nil }
func (c DecimalColumn) ColumnName() string { return c.name }
func (c DecimalColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c DecimalColumn) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c DecimalColumn) As(alias string) DecimalColumn {
	return DecimalColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c DecimalColumn) Eq(val string) Expression       { return colEq(c, val) }
func (c DecimalColumn) NotEq(val string) Expression    { return colNotEq(c, val) }
func (c DecimalColumn) Gt(val string) Expression       { return colGt(c, val) }
func (c DecimalColumn) Gte(val string) Expression      { return colGte(c, val) }
func (c DecimalColumn) Lt(val string) Expression       { return colLt(c, val) }
func (c DecimalColumn) Lte(val string) Expression      { return colLte(c, val) }

func (c DecimalColumn) EqSub(subQuery Expression) Expression    { return colEqSub(c, subQuery) }
func (c DecimalColumn) NotEqSub(subQuery Expression) Expression { return colNotEqSub(c, subQuery) }

func (c DecimalColumn) In(vals ...string) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c DecimalColumn) InSub(subQuery Expression) Expression { return colInSub(c, subQuery) }

func (c DecimalColumn) NotIn(vals ...string) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c DecimalColumn) NotInSub(subQuery Expression) Expression { return colNotInSub(c, subQuery) }

func (c DecimalColumn) Between(min, max string) Expression { return colBetween(c, min, max) }
func (c DecimalColumn) IsNull() Expression                 { return colIsNull(c) }
func (c DecimalColumn) IsNotNull() Expression              { return colIsNotNull(c) }

type BytesColumn struct {
	name  string
	table *string
	alias *string
}

func (c BytesColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c BytesColumn) Args() []any        { return nil }
func (c BytesColumn) ColumnName() string { return c.name }
func (c BytesColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c BytesColumn) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c BytesColumn) As(alias string) BytesColumn {
	return BytesColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c BytesColumn) Eq(val []byte) Expression    { return colEq(c, val) }
func (c BytesColumn) NotEq(val []byte) Expression { return colNotEq(c, val) }
func (c BytesColumn) IsNull() Expression           { return colIsNull(c) }
func (c BytesColumn) IsNotNull() Expression        { return colIsNotNull(c) }

type Float32Column struct {
	name  string
	table *string
	alias *string
}

func (c Float32Column) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c Float32Column) Args() []any        { return nil }
func (c Float32Column) ColumnName() string { return c.name }
func (c Float32Column) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c Float32Column) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c Float32Column) As(alias string) Float32Column {
	return Float32Column{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c Float32Column) Eq(val float32) Expression    { return colEq(c, val) }
func (c Float32Column) NotEq(val float32) Expression { return colNotEq(c, val) }
func (c Float32Column) Gt(val float32) Expression    { return colGt(c, val) }
func (c Float32Column) Gte(val float32) Expression   { return colGte(c, val) }
func (c Float32Column) Lt(val float32) Expression    { return colLt(c, val) }
func (c Float32Column) Lte(val float32) Expression   { return colLte(c, val) }

func (c Float32Column) EqSub(subQuery Expression) Expression    { return colEqSub(c, subQuery) }
func (c Float32Column) NotEqSub(subQuery Expression) Expression { return colNotEqSub(c, subQuery) }

func (c Float32Column) In(vals ...float32) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c Float32Column) InSub(subQuery Expression) Expression { return colInSub(c, subQuery) }

func (c Float32Column) NotIn(vals ...float32) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c Float32Column) NotInSub(subQuery Expression) Expression { return colNotInSub(c, subQuery) }

func (c Float32Column) Between(min, max float32) Expression { return colBetween(c, min, max) }
func (c Float32Column) IsNull() Expression                  { return colIsNull(c) }
func (c Float32Column) IsNotNull() Expression               { return colIsNotNull(c) }

type AnyColumn struct {
	name  string
	table *string
	alias *string
}

func (c AnyColumn) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c AnyColumn) Args() []any        { return nil }
func (c AnyColumn) ColumnName() string { return c.name }
func (c AnyColumn) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c AnyColumn) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c AnyColumn) As(alias string) AnyColumn {
	return AnyColumn{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c AnyColumn) Eq(val any) Expression    { return colEq(c, val) }
func (c AnyColumn) NotEq(val any) Expression { return colNotEq(c, val) }
func (c AnyColumn) Gt(val any) Expression    { return colGt(c, val) }
func (c AnyColumn) Gte(val any) Expression   { return colGte(c, val) }
func (c AnyColumn) Lt(val any) Expression    { return colLt(c, val) }
func (c AnyColumn) Lte(val any) Expression   { return colLte(c, val) }

func (c AnyColumn) EqSub(subQuery Expression) Expression    { return colEqSub(c, subQuery) }
func (c AnyColumn) NotEqSub(subQuery Expression) Expression { return colNotEqSub(c, subQuery) }

func (c AnyColumn) In(vals ...any) Expression    { return colIn(c, vals) }
func (c AnyColumn) InSub(subQuery Expression) Expression { return colInSub(c, subQuery) }

func (c AnyColumn) NotIn(vals ...any) Expression    { return colNotIn(c, vals) }
func (c AnyColumn) NotInSub(subQuery Expression) Expression { return colNotInSub(c, subQuery) }

func (c AnyColumn) Between(min, max any) Expression { return colBetween(c, min, max) }
func (c AnyColumn) IsNull() Expression              { return colIsNull(c) }
func (c AnyColumn) IsNotNull() Expression           { return colIsNotNull(c) }

type EnumColumn[T ~string] struct {
	name  string
	table *string
	alias *string
}

func (c EnumColumn[T]) Sql() string {
	if c.alias != nil {
		return *c.alias
	}
	if c.table != nil && *c.table != "" {
		return fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	return c.name
}

func (c EnumColumn[T]) Args() []any        { return nil }
func (c EnumColumn[T]) ColumnName() string { return c.name }
func (c EnumColumn[T]) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c EnumColumn[T]) Alias() *string {
	if c.alias != nil && *c.alias != "" {
		return c.alias
	}
	return nil
}

func (c EnumColumn[T]) As(alias string) EnumColumn[T] {
	return EnumColumn[T]{
		name:  c.name,
		table: c.table,
		alias: &alias,
	}
}

func (c EnumColumn[T]) Eq(val T) Expression    { return colEq(c, val) }
func (c EnumColumn[T]) NotEq(val T) Expression { return colNotEq(c, val) }

func (c EnumColumn[T]) In(vals ...T) Expression {
	return colIn(c, convertToAny(vals...))
}

func (c EnumColumn[T]) NotIn(vals ...T) Expression {
	return colNotIn(c, convertToAny(vals...))
}

func (c EnumColumn[T]) IsNull() Expression    { return colIsNull(c) }
func (c EnumColumn[T]) IsNotNull() Expression { return colIsNotNull(c) }

type JSONBColumn[T JSONB] struct {
	name  string
	table *string
	alias *string
	path  []string // for nested path access
}

func (c JSONBColumn[T]) Sql() string {
	if c.alias != nil {
		return *c.alias
	}

	base := c.name
	if c.table != nil && *c.table != "" {
		base = fmt.Sprintf("%s.%s", *c.table, c.name)
	}

	for _, key := range c.path {
		base = fmt.Sprintf("%s->'%s'", base, escapePathKey(key))
	}

	return base
}

func (c JSONBColumn[T]) Args() []any        { return nil }
func (c JSONBColumn[T]) ColumnName() string { return c.name }
func (c JSONBColumn[T]) TableName() string {
	if c.table != nil {
		return *c.table
	}
	return ""
}

func (c JSONBColumn[T]) Alias() *string { return c.alias }

func (c JSONBColumn[T]) As(alias string) JSONBColumn[T] {
	return JSONBColumn[T]{
		name:  c.name,
		table: c.table,
		alias: &alias,
		path:  c.path,
	}
}

// Contains checks if JSONB contains the given value (@> operator)
func (c JSONBColumn[T]) Contains(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s @> ?::jsonb", c.baseSql()),
		args: []any{val},
	}
}

// ContainedBy checks if JSONB is contained by the given value (<@ operator)
func (c JSONBColumn[T]) ContainedBy(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s <@ ?::jsonb", c.baseSql()),
		args: []any{val},
	}
}

// HasKey checks if JSONB has the given key (? operator)
func (c JSONBColumn[T]) HasKey(key string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s ?? ?", c.baseSql()), // ?? escapes ? for placeholder replacement
		args: []any{key},
	}
}

// HasAnyKey checks if JSONB has any of the given keys (?| operator)
func (c JSONBColumn[T]) HasAnyKey(keys ...string) Expression {
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		placeholders[i] = "?"
		args[i] = k
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s ??| ARRAY[%s]", c.baseSql(), strings.Join(placeholders, ", ")),
		args: args,
	}
}

// HasAllKeys checks if JSONB has all of the given keys (?& operator)
func (c JSONBColumn[T]) HasAllKeys(keys ...string) Expression {
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		placeholders[i] = "?"
		args[i] = k
	}
	return &simpleExpr{
		sql:  fmt.Sprintf("%s ??& ARRAY[%s]", c.baseSql(), strings.Join(placeholders, ", ")),
		args: args,
	}
}

// Path returns a new JSONBColumn for nested path access (-> operator)
func (c JSONBColumn[T]) Path(keys ...string) JSONBColumn[T] {
	newPath := make([]string, len(c.path)+len(keys))
	copy(newPath, c.path)
	copy(newPath[len(c.path):], keys)
	return JSONBColumn[T]{
		name:  c.name,
		table: c.table,
		path:  newPath,
	}
}

// PathText returns a StringColumn for text extraction (->> operator)
func (c JSONBColumn[T]) PathText(keys ...string) Column {
	base := c.baseSql()

	for i, key := range keys {
		escaped := escapePathKey(key)
		if i == len(keys)-1 {
			base = fmt.Sprintf("%s->>'%s'", base, escaped)
		} else {
			base = fmt.Sprintf("%s->'%s'", base, escaped)
		}
	}

	return rawColumn{sql: base}
}

func (c JSONBColumn[T]) Eq(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?::jsonb", c.baseSql()),
		args: []any{val},
	}
}

func (c JSONBColumn[T]) NotEq(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?::jsonb", c.baseSql()),
		args: []any{val},
	}
}

func (c JSONBColumn[T]) IsNull() Expression {
	return colIsNull(c)
}

func (c JSONBColumn[T]) IsNotNull() Expression {
	return colIsNotNull(c)
}

// rawColumn wraps a raw SQL expression as a Column (e.g. for JSONB path expressions).
type rawColumn struct {
	sql   string
	alias *string
}

func (c rawColumn) Sql() string {
	if c.alias != nil {
		return fmt.Sprintf("%s AS %s", c.sql, *c.alias)
	}
	return c.sql
}
func (c rawColumn) Args() []any        { return nil }
func (c rawColumn) ColumnName() string { return c.sql }
func (c rawColumn) TableName() string  { return "" }
func (c rawColumn) Alias() *string     { return c.alias }

// escapePathKey escapes single quotes in JSONB path keys by doubling them (PostgreSQL standard).
func escapePathKey(key string) string {
	return strings.ReplaceAll(key, "'", "''")
}

// baseSql returns SQL without alias (for operators)
func (c JSONBColumn[T]) baseSql() string {
	base := c.name
	if c.table != nil && *c.table != "" {
		base = fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	for _, key := range c.path {
		base = fmt.Sprintf("%s->'%s'", base, escapePathKey(key))
	}
	return base
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
