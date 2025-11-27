package dew

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type DB struct {
	*sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

type Selector[T any] struct {
	db        *DB
	tableName string
	columns   []Expression
	wheres    []string
	args      []any

	limitCount  int
	offsetCount int
	orderBys    []Expression
}

func From[T any](db *DB) *Selector[T] {
	var t T
	modelType := reflect.TypeOf(t)

	tableName := strings.ToLower(modelType.Name()) + "s"

	return &Selector[T]{
		db:        db,
		tableName: tableName,
	}
}

// *** SELECTOR *** ///

func (s *Selector[T]) Where(expr ...Expression) *Selector[T] {
	for _, exp := range expr {
		s.wheres = append(s.wheres, exp.Sql())
		s.args = append(s.args, exp.Args()...)
	}
	return s
}

func (s *Selector[T]) Select(columns ...Expression) *Selector[T] {
	s.columns = columns
	return s
}

func (s *Selector[T]) Limit(count int) *Selector[T] {
	s.limitCount = count
	return s
}

func (s *Selector[T]) Offset(count int) *Selector[T] {
	s.offsetCount = count
	return s
}

func (s *Selector[T]) OrderBy(expr ...Expression) *Selector[T] {
	s.orderBys = append(s.orderBys, expr...)
	return s
}

func (s *Selector[T]) ToSql() (string, []any) {
	return s.buildQuery(), s.args
}

// * SELECT EXECUTION * //

func (s *Selector[T]) One(ctxs ...context.Context) (*T, error) {
	ctx := getCtx(ctxs)

	s.limitCount = 1
	results, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound
	}
	// Возвращаем указатель на элемент слайса
	return &results[0], nil
}

func (s *Selector[T]) All(ctxs ...context.Context) ([]T, error) {
	ctx := getCtx(ctxs)

	// Шаг 1: Строим SQL
	query := s.buildQuery()

	// Шаг 2: Выполняем запрос
	rows, err := s.db.QueryContext(ctx, query, s.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Шаг 3: Подготовка рефлексии (Оптимизация)
	modelType := reflect.TypeOf(new(T)).Elem()
	targetIndices, err := resolveScanIndices(modelType, s.columns)
	if err != nil {
		return nil, err
	}

	// Шаг 4: Сканирование
	var results []T

	for rows.Next() {
		newItem := reflect.New(modelType)
		val := newItem.Elem()

		// Получаем список адресов (&field), куда драйвер запишет данные
		scanArgs := prepareScanArgs(val, targetIndices)

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		results = append(results, *newItem.Interface().(*T))
	}

	return results, nil
}

func (s *Selector[T]) First() (*T, error) {
	s.limitCount = 1

	results, err := s.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrNotFound
	}

	result := results[0]
	return &result, nil
}

// * Utils Functions * //

func (s *Selector[T]) buildQuery() string {
	// 1. SELECT clause
	selectClause := "*"
	if len(s.columns) > 0 {
		colSqls := make([]string, len(s.columns))
		for i, c := range s.columns {
			colSqls[i] = c.Sql()
		}
		selectClause = strings.Join(colSqls, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, s.tableName)

	// 2. WHERE clause
	if len(s.wheres) > 0 {
		query += " WHERE " + strings.Join(s.wheres, " AND ")
	}

	// 3. ORDER BY clause
	if len(s.orderBys) > 0 {
		var orderSqls []string
		for _, order := range s.orderBys {
			orderSqls = append(orderSqls, order.Sql())
		}
		query += " ORDER BY " + strings.Join(orderSqls, ", ")
	}

	// 4. LIMIT & OFFSET
	if s.limitCount > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.limitCount)
	}
	if s.offsetCount > 0 {
		query += fmt.Sprintf(" OFFSET %d", s.offsetCount)
	}

	return query
}

func resolveScanIndices(modelType reflect.Type, columns []Expression) ([]int, error) {
	// If there is no specific columns, return nil
	if len(columns) == 0 {
		return nil, nil
	}

	// Build map: "field_name_in_lower_case" -> index
	fieldMap := make(map[string]int, modelType.NumField())
	for i := 0; i < modelType.NumField(); i++ {
		fieldMap[strings.ToLower(modelType.Field(i).Name)] = i
	}

	indices := make([]int, len(columns))
	for i, col := range columns {
		colName := col.Sql()
		idx, ok := fieldMap[strings.ToLower(colName)]
		if !ok {
			return nil, fmt.Errorf("dew: struct field for column '%s' not found", colName)
		}
		indices[i] = idx
	}

	return indices, nil
}

func prepareScanArgs(val reflect.Value, indices []int) []any {
	if indices == nil {
		num := val.NumField()
		args := make([]any, num)
		for i := 0; i < num; i++ {
			args[i] = val.Field(i).Addr().Interface()
		}
		return args
	}

	args := make([]any, len(indices))
	for i, idx := range indices {
		args[i] = val.Field(idx).Addr().Interface()
	}
	return args
}

func getCtx(ctxs []context.Context) context.Context {
	if len(ctxs) > 0 {
		return ctxs[0]
	}
	return context.Background()
}
