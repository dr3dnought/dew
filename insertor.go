package dew

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type ConflictActionType string

const (
	ConflictActionTypeNothing ConflictActionType = "NOTHING"
	ConflictActionTypeUpdate  ConflictActionType = "UPDATE"
)

type Insertor[T any] struct {
	db      *DB
	table   Tabler
	columns []Column
	values  [][]any
	models  []*T

	returningCols []Column
}

func Insert[T any](db *DB, table Tabler) *Insertor[T] {
	return &Insertor[T]{
		db:    db,
		table: table,
	}
}

func (i *Insertor[T]) Columns(cols ...Column) *Insertor[T] {
	i.columns = append(i.columns, cols...)
	return i
}

func (i *Insertor[T]) Values(vals ...any) *Insertor[T] {
	i.values = append(i.values, vals)
	return i
}

func (i *Insertor[T]) Models(models ...*T) *Insertor[T] {
	i.models = append(i.models, models...)
	return i
}

func (i *Insertor[T]) Returning(cols ...Column) *Insertor[T] {
	i.returningCols = append(i.returningCols, cols...)
	return i
}

func (i *Insertor[T]) OnConflict(cols ...Column) *ConfilctInsertor[T] {
	conflictInsertor := newConfilctInsertor(i)
	for _, c := range cols {
		conflictInsertor.conflictTargets = append(conflictInsertor.conflictTargets, c.ColumnName())
	}
	return conflictInsertor
}

func (i *Insertor[T]) buildFromModels() (string, []any, error) {
	if len(i.models) == 0 {
		return "", nil, fmt.Errorf("dew: no models to insert")
	}

	modelType := reflect.TypeOf(i.models[0]).Elem()
	if modelType.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("dew: expected struct, got %s", modelType.Kind())
	}

	var columns []string
	var allArgs []any

	var validFieldIndexes []int

	numFields := modelType.NumField()

	for j := range numFields {
		field := modelType.Field(j)

		if !field.IsExported() {
			continue
		}

		columnName := getColumnName(field)
		if columnName == "" {
			continue // field is ignored (tag "-")
		}

		columns = append(columns, columnName)
		validFieldIndexes = append(validFieldIndexes, j)
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("dew: no exportable fields in model")
	}

	valuePlaceholders := make([]string, len(i.models))
	for idx, model := range i.models {
		modelVal := reflect.ValueOf(model).Elem()

		var rowArgs []any

		for _, fieldIdx := range validFieldIndexes {
			val := modelVal.Field(fieldIdx).Interface()
			rowArgs = append(rowArgs, val)
		}

		placeholders := make([]string, len(columns))
		start := len(allArgs)
		for j := range placeholders {
			placeholders[j] = i.db.dialect.Placeholder(start + j)
		}
		allArgs = append(allArgs, rowArgs...)
		valuePlaceholders[idx] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		i.table.TableName(),
		strings.Join(columns, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	return query, allArgs, nil
}

func (i *Insertor[T]) buildFromValues() (string, []any, error) {
	if len(i.columns) == 0 {
		return "", nil, fmt.Errorf("dew: no columns specified")
	}

	if len(i.values) == 0 {
		return "", nil, fmt.Errorf("dew: no values specified")
	}

	columnNames := make([]string, len(i.columns))
	for idx, col := range i.columns {
		columnNames[idx] = col.ColumnName()
	}

	valuePlaceholders := make([]string, len(i.values))
	var allArgs []any

	for idx, row := range i.values {
		if len(row) != len(i.columns) {
			return "", nil, fmt.Errorf("dew: row %d has %d values but %d columns", idx, len(row), len(i.columns))
		}

		placeholders := make([]string, len(row))
		start := len(allArgs)
		for j := range placeholders {
			placeholders[j] = i.db.dialect.Placeholder(start + j)
		}
		allArgs = append(allArgs, row...)
		valuePlaceholders[idx] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		i.table.TableName(),
		strings.Join(columnNames, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	return query, allArgs, nil
}

func (i *Insertor[T]) buildInsertQuery() (string, []any, error) {
	hasModels := len(i.models) > 0
	hasValues := len(i.values) > 0

	if hasModels && hasValues {
		return "", nil, fmt.Errorf("dew: cannot use both Models() and Values() in the same insert statement")
	}

	var query string
	var args []any
	var err error

	if hasModels {
		if len(i.columns) > 0 {
			return "", nil, fmt.Errorf("dew: do not use Columns() with Models(), columns are inferred from struct fields")
		}
		query, args, err = i.buildFromModels()
	} else if hasValues {
		if len(i.columns) == 0 {
			return "", nil, fmt.Errorf("dew: Columns() are required when using Values()")
		}
		query, args, err = i.buildFromValues()
	} else {
		return "", nil, fmt.Errorf("dew: no data to insert (call Models or Values)")
	}

	if err != nil {
		return "", nil, err
	}

	if len(i.returningCols) > 0 {
		query += " RETURNING "
		for _, col := range i.returningCols {
			query += col.Sql() + ", "
		}
		query = query[:len(query)-2]
	}

	return query, args, nil
}

func (i *Insertor[T]) Exec(ctxs ...context.Context) error {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return err
	}

	_, err = i.db.ExecContext(ctx, query, args...)
	return err
}

func (i *Insertor[T]) ToSql() (string, []any, error) {
	return i.buildInsertQuery()
}

func (i *Insertor[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
	if len(i.returningCols) == 0 {
		return nil, fmt.Errorf("dew: ScanWith requires Returning() to be called")
	}

	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return nil, err
	}

	rows, err := i.db.QueryContext(ctx, query, args...)
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

type ConfilctInsertor[T any] struct {
	*Insertor[T]

	conflictTargets []string
	conflictAction  ConflictActionType
	conflictSets    map[string]any
}

func newConfilctInsertor[T any](insertor *Insertor[T]) *ConfilctInsertor[T] {
	return &ConfilctInsertor[T]{
		Insertor:     insertor,
		conflictSets: make(map[string]any),
	}
}

func (i *ConfilctInsertor[T]) DoNothing() *ConfilctInsertor[T] {
	i.conflictAction = "NOTHING"
	return i
}

func (i *ConfilctInsertor[T]) SetUpdate(col Column, val any) *ConfilctInsertor[T] {
	if i.conflictSets == nil {
		i.conflictSets = make(map[string]any)
	}
	i.conflictAction = "UPDATE"
	i.conflictSets[col.ColumnName()] = val
	return i
}

func (i *ConfilctInsertor[T]) buildInsertQuery() (string, []any, error) {
	var query string
	var args []any
	var err error

	hasModels := len(i.models) > 0
	hasValues := len(i.values) > 0

	if hasModels && hasValues {
		return "", nil, fmt.Errorf("dew: cannot use both Models() and Values() in the same insert statement")
	}

	if hasModels {
		if len(i.columns) > 0 {
			return "", nil, fmt.Errorf("dew: do not use Columns() with Models(), columns are inferred from struct fields")
		}
		query, args, err = i.buildFromModels()
	} else if hasValues {
		if len(i.columns) == 0 {
			return "", nil, fmt.Errorf("dew: Columns() are required when using Values()")
		}
		query, args, err = i.buildFromValues()
	} else {
		return "", nil, fmt.Errorf("dew: no data to insert (call Models or Values)")
	}

	if err != nil {
		return "", nil, err
	}

	if len(i.conflictTargets) > 0 {
		query += " ON CONFLICT (" + strings.Join(i.conflictTargets, ", ") + ")"

		switch i.conflictAction {
		case ConflictActionTypeNothing:
			query += " DO NOTHING"
		case ConflictActionTypeUpdate:
			if len(i.conflictSets) == 0 {
				return "", nil, fmt.Errorf("dew: OnConflict DoUpdate requires at least one SetUpdate()")
			}

			var setParts []string
			var updateArgs []any

			for col, val := range i.conflictSets {
				if expr, ok := val.(Expression); ok {
					setParts = append(setParts, fmt.Sprintf("%s = %s", col, expr.Sql()))
					updateArgs = append(updateArgs, expr.Args()...)
				} else {
					setParts = append(setParts, fmt.Sprintf("%s = ?", col))
					updateArgs = append(updateArgs, val)
				}
			}

			query += " DO UPDATE SET " + strings.Join(setParts, ", ")

			args = append(args, updateArgs...)
		}
	}

	if len(i.returningCols) > 0 {
		query += " RETURNING "
		for _, col := range i.returningCols {
			query += col.Sql() + ", "
		}
		query = query[:len(query)-2]
	}

	return query, args, nil
}

func (i *ConfilctInsertor[T]) Exec(ctxs ...context.Context) error {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return err
	}

	_, err = i.db.ExecContext(ctx, query, args...)
	return err
}

func (i *ConfilctInsertor[T]) ToSql() (string, []any, error) {
	return i.buildInsertQuery()
}

func (i *ConfilctInsertor[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
	if len(i.returningCols) == 0 {
		return nil, fmt.Errorf("dew: ScanWith requires Returning() to be called")
	}

	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return nil, err
	}

	rows, err := i.db.QueryContext(ctx, query, args...)
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
