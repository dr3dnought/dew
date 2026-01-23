package dew

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type setClause struct {
	column string
	value  any
}

type Updator[T any] struct {
	db     *DB
	table  Tabler
	wheres []Expression
	sets   []setClause

	returningCols []Column
}

func Update[T any](db *DB, table Tabler) *Updator[T] {
	return &Updator[T]{
		db:    db,
		table: table,
	}
}

func (u *Updator[T]) Set(col Column, val any) *Updator[T] {
	u.sets = append(u.sets, setClause{
		column: col.ColumnName(),
		value:  val,
	})
	return u
}

func (u *Updator[T]) Where(expr ...Expression) *Updator[T] {
	u.wheres = append(u.wheres, expr...)
	return u
}

func (u *Updator[T]) Returning(cols ...Column) *Updator[T] {
	u.returningCols = append(u.returningCols, cols...)
	return u
}

func (u *Updator[T]) Exec(ctxs ...context.Context) error {
	ctx := getCtx(ctxs)

	query, args, err := u.buildUpdateQuery()
	if err != nil {
		return err
	}

	_, err = u.db.ExecContext(ctx, query, args...)
	return err
}

func (u *Updator[T]) Scan(ctx context.Context, dest ...any) error {
	if len(u.returningCols) == 0 {
		return fmt.Errorf("dew: Scan requires .Returning(...)")
	}

	query, args, err := u.buildUpdateQuery()
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return u.db.QueryRowContext(ctx, query, args...).Scan(dest...)
}

func (u *Updator[T]) RowsAffected(ctxs ...context.Context) (int64, error) {
	ctx := getCtx(ctxs)

	query, args, err := u.buildUpdateQuery()
	if err != nil {
		return 0, err
	}

	res, err := u.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (u *Updator[T]) ToSql() (string, []any, error) {
	return u.buildUpdateQuery()
}

func (u *Updator[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
	if len(u.returningCols) == 0 {
		return nil, fmt.Errorf("dew: ScanWith requires Returning() to be called")
	}

	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	query, args, err := u.buildUpdateQuery()
	if err != nil {
		return nil, err
	}

	rows, err := u.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*T

	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (u *Updator[T]) Clone() *Updator[T] {
	return &Updator[T]{
		db:            u.db,
		table:         u.table,
		wheres:        append([]Expression{}, u.wheres...),
		sets:          append([]setClause{}, u.sets...),
		returningCols: append([]Column{}, u.returningCols...),
	}
}

func (u *Updator[T]) buildUpdateQuery() (string, []any, error) {
	if len(u.sets) == 0 {
		return "", nil, fmt.Errorf("dew: UPDATE requires at least one Set()")
	}

	if len(u.wheres) == 0 {
		return "", nil, fmt.Errorf("dew: UNSAFE UPDATE! You must provide a Where clause. Use Where(dew.Raw(\"1=1\")) to force update all")
	}

	var builder strings.Builder
	builder.Grow(128)

	builder.WriteString("UPDATE ")
	builder.WriteString(u.table.TableName())
	builder.WriteString(" SET ")

	var allArgs []any
	var setParts []string

	for _, set := range u.sets {
		if expr, ok := set.value.(Expression); ok {
			sql := expr.Sql()
			sql = replacePlaceholders(sql, u.db.dialect, len(allArgs))
			setParts = append(setParts, fmt.Sprintf("%s = %s", set.column, sql))
			allArgs = append(allArgs, expr.Args()...)
		} else {
			placeholder := u.db.dialect.Placeholder(len(allArgs))
			setParts = append(setParts, fmt.Sprintf("%s = %s", set.column, placeholder))
			allArgs = append(allArgs, set.value)
		}
	}

	builder.WriteString(strings.Join(setParts, ", "))

	var whereSqls []string
	for _, expr := range u.wheres {
		sql := expr.Sql()
		if sql != "" {
			sql = replacePlaceholders(sql, u.db.dialect, len(allArgs))
			whereSqls = append(whereSqls, sql)
			allArgs = append(allArgs, expr.Args()...)
		}
	}

	if len(whereSqls) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(whereSqls, " AND "))
	}

	if len(u.returningCols) > 0 {
		cols := make([]string, len(u.returningCols))
		for i, c := range u.returningCols {
			cols[i] = c.ColumnName()
		}
		builder.WriteString(" RETURNING ")
		builder.WriteString(strings.Join(cols, ", "))
	}

	return builder.String(), allArgs, nil
}
