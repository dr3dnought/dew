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

type Updater[T any] struct {
	db     Querier
	table  Tabler
	wheres []Expression
	sets   []setClause

	returningCols []Column
}

func Update[T any](db Querier, table Tabler) *Updater[T] {
	return &Updater[T]{
		db:    db,
		table: table,
	}
}

func (u *Updater[T]) Set(col Column, val any) *Updater[T] {
	u.sets = append(u.sets, setClause{
		column: col.ColumnName(),
		value:  val,
	})
	return u
}

func (u *Updater[T]) Where(expr ...Expression) *Updater[T] {
	u.wheres = append(u.wheres, expr...)
	return u
}

func (u *Updater[T]) Returning(cols ...Column) *Updater[T] {
	u.returningCols = append(u.returningCols, cols...)
	return u
}

func (u *Updater[T]) Exec(ctxs ...context.Context) error {
	ctx := getCtx(ctxs)

	query, args, err := u.buildUpdateQuery()
	if err != nil {
		return err
	}

	_, err = u.db.ExecContext(ctx, query, args...)
	return err
}

func (u *Updater[T]) Scan(ctx context.Context, dest ...any) error {
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

func (u *Updater[T]) RowsAffected(ctxs ...context.Context) (int64, error) {
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

func (u *Updater[T]) ToSql() (string, []any, error) {
	return u.buildUpdateQuery()
}

func (u *Updater[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
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

func (u *Updater[T]) Clone() *Updater[T] {
	return &Updater[T]{
		db:            u.db,
		table:         u.table,
		wheres:        append([]Expression{}, u.wheres...),
		sets:          append([]setClause{}, u.sets...),
		returningCols: append([]Column{}, u.returningCols...),
	}
}

func (u *Updater[T]) buildUpdateQuery() (string, []any, error) {
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
			sql = replacePlaceholders(sql, u.db.getDialect(), len(allArgs))
			setParts = append(setParts, fmt.Sprintf("%s = %s", set.column, sql))
			allArgs = append(allArgs, expr.Args()...)
		} else {
			placeholder := u.db.getDialect().Placeholder(len(allArgs))
			setParts = append(setParts, fmt.Sprintf("%s = %s", set.column, placeholder))
			allArgs = append(allArgs, set.value)
		}
	}

	builder.WriteString(strings.Join(setParts, ", "))

	var whereSqls []string
	for _, expr := range u.wheres {
		sql := expr.Sql()
		if sql != "" {
			sql = replacePlaceholders(sql, u.db.getDialect(), len(allArgs))
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
