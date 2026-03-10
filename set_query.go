package dew

import (
	"context"
	"fmt"
	"strings"
)

type setOp string

const (
	setOpUnion     setOp = "UNION"
	setOpUnionAll  setOp = "UNION ALL"
	setOpIntersect setOp = "INTERSECT"
	setOpExcept    setOp = "EXCEPT"
)

type setPart struct {
	op    setOp
	query Expression
}

type SetQuery[T any] struct {
	db    Querier
	first Expression
	parts []setPart

	orderBys    []Expression
	limitCount  int
	offsetCount int
}

func newSetQuery[T any](db Querier, op setOp, left, right Expression) *SetQuery[T] {
	return &SetQuery[T]{
		db:    db,
		first: left,
		parts: []setPart{{op: op, query: right}},
	}
}

func Union[T any](db Querier, left, right Expression) *SetQuery[T] {
	return newSetQuery[T](db, setOpUnion, left, right)
}

func UnionAll[T any](db Querier, left, right Expression) *SetQuery[T] {
	return newSetQuery[T](db, setOpUnionAll, left, right)
}

func Intersect[T any](db Querier, left, right Expression) *SetQuery[T] {
	return newSetQuery[T](db, setOpIntersect, left, right)
}

func Except[T any](db Querier, left, right Expression) *SetQuery[T] {
	return newSetQuery[T](db, setOpExcept, left, right)
}

func (sq *SetQuery[T]) Union(query Expression) *SetQuery[T] {
	sq.parts = append(sq.parts, setPart{op: setOpUnion, query: query})
	return sq
}

func (sq *SetQuery[T]) UnionAll(query Expression) *SetQuery[T] {
	sq.parts = append(sq.parts, setPart{op: setOpUnionAll, query: query})
	return sq
}

func (sq *SetQuery[T]) Intersect(query Expression) *SetQuery[T] {
	sq.parts = append(sq.parts, setPart{op: setOpIntersect, query: query})
	return sq
}

func (sq *SetQuery[T]) Except(query Expression) *SetQuery[T] {
	sq.parts = append(sq.parts, setPart{op: setOpExcept, query: query})
	return sq
}

func (sq *SetQuery[T]) OrderBy(exprs ...Expression) *SetQuery[T] {
	sq.orderBys = append(sq.orderBys, exprs...)
	return sq
}

func (sq *SetQuery[T]) Limit(count int) *SetQuery[T] {
	sq.limitCount = count
	return sq
}

func (sq *SetQuery[T]) Offset(count int) *SetQuery[T] {
	sq.offsetCount = count
	return sq
}

// Expression interface — allows SetQuery to be used as CTE body or subquery FROM.

func (sq *SetQuery[T]) Sql() string {
	sql, _ := sq.buildRawSetQuery()
	return sql
}

func (sq *SetQuery[T]) Args() []any {
	_, args := sq.buildRawSetQuery()
	return args
}

func (sq *SetQuery[T]) ToSql() (string, []any, error) {
	raw, args := sq.buildRawSetQuery()
	return replacePlaceholders(raw, sq.db.getDialect(), 0), args, nil
}

func (sq *SetQuery[T]) All(ctxs ...context.Context) ([]*T, error) {
	ctx := getCtx(ctxs)
	query, args, _ := sq.ToSql()

	rows, err := sq.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, sq.db.mapError(err)
	}
	defer rows.Close()

	results, err := scanAll[T](rows, nil)
	if err != nil {
		return nil, sq.db.mapError(err)
	}
	return results, nil
}

// buildRawSetQuery builds the set query with ? placeholders (no dialect replacement).
func (sq *SetQuery[T]) buildRawSetQuery() (string, []any) {
	var finalArgs []any

	// Build first query — Sql() must be called before Args() to populate args
	var b strings.Builder
	firstSql := sq.first.Sql()
	b.WriteString("(")
	b.WriteString(firstSql)
	b.WriteString(")")
	finalArgs = append(finalArgs, sq.first.Args()...)

	// Build subsequent parts
	for _, part := range sq.parts {
		partSql := part.query.Sql()
		b.WriteString(" ")
		b.WriteString(string(part.op))
		b.WriteString(" (")
		b.WriteString(partSql)
		b.WriteString(")")
		finalArgs = append(finalArgs, part.query.Args()...)
	}

	// ORDER BY
	if len(sq.orderBys) > 0 {
		var orderSqls []string
		for _, order := range sq.orderBys {
			sqlStr := order.Sql()
			orderSqls = append(orderSqls, sqlStr)
			finalArgs = append(finalArgs, order.Args()...)
		}
		b.WriteString(" ORDER BY ")
		b.WriteString(strings.Join(orderSqls, ", "))
	}

	if sq.limitCount > 0 {
		b.WriteString(fmt.Sprintf(" LIMIT %d", sq.limitCount))
	}
	if sq.offsetCount > 0 {
		b.WriteString(fmt.Sprintf(" OFFSET %d", sq.offsetCount))
	}

	return b.String(), finalArgs
}
