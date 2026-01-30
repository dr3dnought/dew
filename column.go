/*
TODO:

	Check if alias pointer is nil or not in As method
*/
package dew

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

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

type JSONBColumn[T any] struct {
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
		base = fmt.Sprintf("%s->'%s'", base, key)
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
		sql:  fmt.Sprintf("%s @> ?", c.baseSql()),
		args: []any{jsonbValue{val}},
	}
}

// ContainedBy checks if JSONB is contained by the given value (<@ operator)
func (c JSONBColumn[T]) ContainedBy(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s <@ ?", c.baseSql()),
		args: []any{jsonbValue{val}},
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
	return &simpleExpr{
		sql:  fmt.Sprintf("%s ??| ?", c.baseSql()),
		args: []any{pgStringArray(keys)},
	}
}

// HasAllKeys checks if JSONB has all of the given keys (?& operator)
func (c JSONBColumn[T]) HasAllKeys(keys ...string) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s ??& ?", c.baseSql()),
		args: []any{pgStringArray(keys)},
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
func (c JSONBColumn[T]) PathText(keys ...string) StringColumn {
	base := c.name
	if c.table != nil && *c.table != "" {
		base = fmt.Sprintf("%s.%s", *c.table, c.name)
	}

	// Build path: col->'key1'->'key2'->>'lastKey'
	for i, key := range keys {
		if i == len(keys)-1 {
			base = fmt.Sprintf("%s->>'%s'", base, key)
		} else {
			base = fmt.Sprintf("%s->'%s'", base, key)
		}
	}

	return StringColumn{name: base}
}

func (c JSONBColumn[T]) Eq(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s = ?", c.baseSql()),
		args: []any{jsonbValue{val}},
	}
}

func (c JSONBColumn[T]) NotEq(val T) Expression {
	return &simpleExpr{
		sql:  fmt.Sprintf("%s != ?", c.baseSql()),
		args: []any{jsonbValue{val}},
	}
}

func (c JSONBColumn[T]) IsNull() Expression {
	return colIsNull(c)
}

func (c JSONBColumn[T]) IsNotNull() Expression {
	return colIsNotNull(c)
}

// baseSql returns SQL without alias (for operators)
func (c JSONBColumn[T]) baseSql() string {
	base := c.name
	if c.table != nil && *c.table != "" {
		base = fmt.Sprintf("%s.%s", *c.table, c.name)
	}
	for _, key := range c.path {
		base = fmt.Sprintf("%s->'%s'", base, key)
	}
	return base
}

type jsonbValue struct {
	val any
}

func (j jsonbValue) Value() (driver.Value, error) {
	return json.Marshal(j.val)
}

// pgStringArray implements driver.Valuer for PostgreSQL text[] array
type pgStringArray []string

func (a pgStringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil
	}

	escaped := make([]string, len(a))
	for i, s := range a {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		escaped[i] = `"` + s + `"`
	}
	return "{" + strings.Join(escaped, ",") + "}", nil
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
