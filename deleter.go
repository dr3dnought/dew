package dew

import (
	"context"
	"fmt"
	"strings"
)

type Deleter[T any] struct {
	db     Querier
	table  Tabler
	wheres []Expression

	returningCols []Column
}

func Delete[T any](db Querier, table Tabler) *Deleter[T] {
	return &Deleter[T]{
		db:    db,
		table: table,
	}
}

func (d *Deleter[T]) Where(expr ...Expression) *Deleter[T] {
	d.wheres = append(d.wheres, expr...)
	return d
}

func (d *Deleter[T]) Returning(cols ...Column) *Deleter[T] {
	d.returningCols = append(d.returningCols, cols...)
	return d
}

func (d *Deleter[T]) Exec(ctxs ...context.Context) error {
	ctx := getCtx(ctxs)

	query, args, err := d.buildDeleteQuery()
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, query, args...)
	return d.db.mapError(err)
}

func (d *Deleter[T]) Scan(ctx context.Context, dest ...any) error {
	if len(d.returningCols) == 0 {
		return fmt.Errorf("dew: Scan requires .Returning(...)")
	}

	query, args, err := d.buildDeleteQuery()
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return d.db.mapError(d.db.QueryRowContext(ctx, query, args...).Scan(dest...))
}

func (d *Deleter[T]) RowsAffected(ctxs ...context.Context) (int64, error) {
	ctx := getCtx(ctxs)

	query, args, err := d.buildDeleteQuery()
	if err != nil {
		return 0, err
	}

	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, d.db.mapError(err)
	}

	return res.RowsAffected()
}

func (d *Deleter[T]) ToSql() (string, []any, error) {
	return d.buildDeleteQuery()
}

func (d *Deleter[T]) Clone() *Deleter[T] {
	return &Deleter[T]{
		db:            d.db,
		table:         d.table,
		wheres:        append([]Expression{}, d.wheres...),
		returningCols: append([]Column{}, d.returningCols...),
	}
}

func (d *Deleter[T]) buildDeleteQuery() (string, []any, error) {
	if len(d.wheres) == 0 {
		return "", nil, fmt.Errorf("dew: UNSAFE DELETE! You must provide a Where clause. Use Where(dew.Raw(\"1=1\")) to force delete all")
	}

	var builder strings.Builder
	builder.Grow(128) // 128 is a good default size for most queries

	builder.WriteString("DELETE FROM ")
	builder.WriteString(d.table.TableName())

	var whereSqls []string
	var allArgs []any
	for _, expr := range d.wheres {
		sql := expr.Sql()
		if sql != "" {
			sql = replacePlaceholders(sql, d.db.getDialect(), len(allArgs))
			whereSqls = append(whereSqls, sql)
			allArgs = append(allArgs, expr.Args()...)
		}
	}

	if len(whereSqls) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(whereSqls, " AND "))
	}

	if len(d.returningCols) > 0 {
		cols := make([]string, len(d.returningCols))
		for i, c := range d.returningCols {
			cols[i] = c.ColumnName()
		}
		builder.WriteString(" RETURNING ")
		builder.WriteString(strings.Join(cols, ", "))
	}

	return builder.String(), allArgs, nil
}
