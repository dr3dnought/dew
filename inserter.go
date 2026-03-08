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

type Inserter[T any] struct {
	db      Querier
	table   Tabler
	columns []Column
	values  [][]any
	models  []*T

	returningCols []Column
	batchSize     int
}

func Insert[T any](db Querier, table Tabler) *Inserter[T] {
	return &Inserter[T]{
		db:    db,
		table: table,
	}
}

func (i *Inserter[T]) Columns(cols ...Column) *Inserter[T] {
	i.columns = append(i.columns, cols...)
	return i
}

func (i *Inserter[T]) Values(vals ...any) *Inserter[T] {
	i.values = append(i.values, vals)
	return i
}

func (i *Inserter[T]) Models(models ...*T) *Inserter[T] {
	i.models = append(i.models, models...)
	return i
}

func (i *Inserter[T]) Returning(cols ...Column) *Inserter[T] {
	i.returningCols = append(i.returningCols, cols...)
	return i
}

func (i *Inserter[T]) Batch(size int) *Inserter[T] {
	i.batchSize = size
	return i
}

func (i *Inserter[T]) OnConflict(cols ...Column) *ConflictInserter[T] {
	conflictInserter := newConflictInserter(i)
	for _, c := range cols {
		conflictInserter.conflictTargets = append(conflictInserter.conflictTargets, c.ColumnName())
	}
	return conflictInserter
}

func (i *Inserter[T]) buildFromModels() (string, []any, error) {
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
			placeholders[j] = i.db.getDialect().Placeholder(start + j)
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

func (i *Inserter[T]) buildFromValues() (string, []any, error) {
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
			placeholders[j] = i.db.getDialect().Placeholder(start + j)
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

func (i *Inserter[T]) buildInsertQuery() (string, []any, error) {
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

func (i *Inserter[T]) Exec(ctxs ...context.Context) error {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	if i.batchSize > 0 {
		return i.execBatch(ctx)
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return err
	}

	_, err = i.db.ExecContext(ctx, query, args...)
	return i.db.mapError(err)
}

func (i *Inserter[T]) execBatch(ctx context.Context) error {
	if len(i.models) > 0 {
		for start := 0; start < len(i.models); start += i.batchSize {
			end := min(start+i.batchSize, len(i.models))
			chunk := &Inserter[T]{
				db:     i.db,
				table:  i.table,
				models: i.models[start:end],
			}
			query, args, err := chunk.buildInsertQuery()
			if err != nil {
				return err
			}
			if _, err := i.db.ExecContext(ctx, query, args...); err != nil {
				return i.db.mapError(err)
			}
		}
		return nil
	}

	if len(i.values) > 0 {
		for start := 0; start < len(i.values); start += i.batchSize {
			end := min(start+i.batchSize, len(i.values))
			chunk := &Inserter[T]{
				db:      i.db,
				table:   i.table,
				columns: i.columns,
				values:  i.values[start:end],
			}
			query, args, err := chunk.buildInsertQuery()
			if err != nil {
				return err
			}
			if _, err := i.db.ExecContext(ctx, query, args...); err != nil {
				return i.db.mapError(err)
			}
		}
		return nil
	}

	return fmt.Errorf("dew: no data to insert (call Models or Values)")
}

func (i *Inserter[T]) ToSql() (string, []any, error) {
	return i.buildInsertQuery()
}

// BatchQueries returns the SQL and args for each batch chunk.
// If Batch() was not called, returns a single-element slice.
func (i *Inserter[T]) BatchQueries() ([]string, [][]any, error) {
	if i.batchSize <= 0 {
		q, a, err := i.buildInsertQuery()
		if err != nil {
			return nil, nil, err
		}
		return []string{q}, [][]any{a}, nil
	}

	if len(i.models) > 0 {
		var queries []string
		var allArgs [][]any
		for start := 0; start < len(i.models); start += i.batchSize {
			end := min(start+i.batchSize, len(i.models))
			chunk := &Inserter[T]{
				db:    i.db,
				table: i.table,
				models: i.models[start:end],
			}
			q, a, err := chunk.buildInsertQuery()
			if err != nil {
				return nil, nil, err
			}
			queries = append(queries, q)
			allArgs = append(allArgs, a)
		}
		return queries, allArgs, nil
	}

	if len(i.values) > 0 {
		var queries []string
		var allArgs [][]any
		for start := 0; start < len(i.values); start += i.batchSize {
			end := min(start+i.batchSize, len(i.values))
			chunk := &Inserter[T]{
				db:      i.db,
				table:   i.table,
				columns: i.columns,
				values:  i.values[start:end],
			}
			q, a, err := chunk.buildInsertQuery()
			if err != nil {
				return nil, nil, err
			}
			queries = append(queries, q)
			allArgs = append(allArgs, a)
		}
		return queries, allArgs, nil
	}

	return nil, nil, fmt.Errorf("dew: no data to insert (call Models or Values)")
}

func (i *Inserter[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
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
		return nil, i.db.mapError(err)
	}
	defer rows.Close()

	var results []*T

	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, i.db.mapError(err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, i.db.mapError(err)
	}

	return results, nil
}

type ConflictInserter[T any] struct {
	*Inserter[T]

	conflictTargets []string
	conflictAction  ConflictActionType
	conflictSets    map[string]any
}

func newConflictInserter[T any](insertor *Inserter[T]) *ConflictInserter[T] {
	return &ConflictInserter[T]{
		Inserter:     insertor,
		conflictSets: make(map[string]any),
	}
}

func (i *ConflictInserter[T]) Batch(size int) *ConflictInserter[T] {
	i.batchSize = size
	return i
}

func (i *ConflictInserter[T]) DoNothing() *ConflictInserter[T] {
	i.conflictAction = "NOTHING"
	return i
}

func (i *ConflictInserter[T]) SetUpdate(col Column, val any) *ConflictInserter[T] {
	if i.conflictSets == nil {
		i.conflictSets = make(map[string]any)
	}
	i.conflictAction = "UPDATE"
	i.conflictSets[col.ColumnName()] = val
	return i
}

func (i *ConflictInserter[T]) buildInsertQuery() (string, []any, error) {
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
					placeholderIndex := len(args) + len(updateArgs)
					setParts = append(setParts, fmt.Sprintf("%s = %s", col, i.db.getDialect().Placeholder(placeholderIndex)))
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

func (i *ConflictInserter[T]) Exec(ctxs ...context.Context) error {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	if i.batchSize > 0 {
		return i.execBatch(ctx)
	}

	query, args, err := i.buildInsertQuery()
	if err != nil {
		return err
	}

	_, err = i.db.ExecContext(ctx, query, args...)
	return i.db.mapError(err)
}

func (i *ConflictInserter[T]) execBatch(ctx context.Context) error {
	if len(i.models) > 0 {
		for start := 0; start < len(i.models); start += i.batchSize {
			end := min(start+i.batchSize, len(i.models))
			chunk := &ConflictInserter[T]{
				Inserter: &Inserter[T]{
					db:    i.db,
					table: i.table,
					models: i.models[start:end],
				},
				conflictTargets: i.conflictTargets,
				conflictAction:  i.conflictAction,
				conflictSets:    i.conflictSets,
			}
			query, args, err := chunk.buildInsertQuery()
			if err != nil {
				return err
			}
			if _, err := i.db.ExecContext(ctx, query, args...); err != nil {
				return i.db.mapError(err)
			}
		}
		return nil
	}

	if len(i.values) > 0 {
		for start := 0; start < len(i.values); start += i.batchSize {
			end := min(start+i.batchSize, len(i.values))
			chunk := &ConflictInserter[T]{
				Inserter: &Inserter[T]{
					db:      i.db,
					table:   i.table,
					columns: i.columns,
					values:  i.values[start:end],
				},
				conflictTargets: i.conflictTargets,
				conflictAction:  i.conflictAction,
				conflictSets:    i.conflictSets,
			}
			query, args, err := chunk.buildInsertQuery()
			if err != nil {
				return err
			}
			if _, err := i.db.ExecContext(ctx, query, args...); err != nil {
				return i.db.mapError(err)
			}
		}
		return nil
	}

	return fmt.Errorf("dew: no data to insert (call Models or Values)")
}

func (i *ConflictInserter[T]) ToSql() (string, []any, error) {
	return i.buildInsertQuery()
}

func (i *ConflictInserter[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
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
		return nil, i.db.mapError(err)
	}
	defer rows.Close()

	var results []*T

	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, i.db.mapError(err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, i.db.mapError(err)
	}

	return results, nil
}
