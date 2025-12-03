package dew

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Insertor[T any] struct {
	db      *DB
	table   Tabler
	columns []Column
	values  [][]any
	models  []*T
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

func (i *Insertor[T]) Model(model *T) *Insertor[T] {
	i.models = append(i.models, model)
	return i
}

func (i *Insertor[T]) Models(models ...*T) *Insertor[T] {
	i.models = append(i.models, models...)
	return i
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

		// TODO: read struct tags

		columnName := toSnakeCase(field.Name)
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

		allArgs = append(allArgs, rowArgs...)

		placeholders := make([]string, len(columns))
		for j := range placeholders {
			placeholders[j] = "?"
		}
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

		allArgs = append(allArgs, row...)

		placeholders := make([]string, len(row))
		for j := range placeholders {
			placeholders[j] = "?"
		}
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

	if hasModels {
		if len(i.columns) > 0 {
			return "", nil, fmt.Errorf("dew: do not use Columns() with Models(), columns are inferred from struct fields")
		}
		return i.buildFromModels()
	}

	if hasValues {
		if len(i.columns) == 0 {
			return "", nil, fmt.Errorf("dew: Columns() are required when using Values()")
		}
		return i.buildFromValues()
	}

	return "", nil, fmt.Errorf("dew: no data to insert (call Models or Values)")
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
