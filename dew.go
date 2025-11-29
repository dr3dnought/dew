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
	columns   []Column
	wheres    []string
	args      []any

	distinctColumns []Column

	limitCount  int
	offsetCount int
	orderBys    []Expression

	groupBys []Expression
	havings  []Expression
}

func From[T any](db *DB, schema Tabler) *Selector[T] {
	return &Selector[T]{
		db:        db,
		tableName: schema.TableName(),
	}
}

// **** Implementations of the Expression interface **** //

func (s *Selector[T]) Sql() string {
	return "(" + s.buildQuery() + ")"
}

func (s *Selector[T]) Args() []any {
	return s.args
}

// *** SELECTOR *** ///

func (s *Selector[T]) Where(expr ...Expression) *Selector[T] {
	for _, exp := range expr {
		s.wheres = append(s.wheres, exp.Sql())
		s.args = append(s.args, exp.Args()...)
	}
	return s
}

func (s *Selector[T]) Select(columns ...Column) *Selector[T] {
	s.columns = columns
	return s
}

func (s *Selector[T]) Distinct(columns ...Column) *Selector[T] {
	s.distinctColumns = columns
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

func (s *Selector[T]) GroupBy(columns ...Expression) *Selector[T] {
	s.groupBys = append(s.groupBys, columns...)
	return s
}

func (s *Selector[T]) Having(columns ...Expression) *Selector[T] {
	s.havings = append(s.havings, columns...)
	return s
}

func (s *Selector[T]) ToSql() (string, []any) {
	return s.buildQuery(), s.args
}

func (s *Selector[T]) Clone() *Selector[T] {
	clone := &Selector[T]{
		db:          s.db,
		tableName:   s.tableName,
		limitCount:  s.limitCount,
		offsetCount: s.offsetCount,
	}

	if s.columns != nil {
		clone.columns = make([]Column, len(s.columns))
		copy(clone.columns, s.columns)
	}

	if s.wheres != nil {
		clone.wheres = make([]string, len(s.wheres))
		copy(clone.wheres, s.wheres)
	}

	if s.args != nil {
		clone.args = make([]any, len(s.args))
		copy(clone.args, s.args)
	}

	if s.distinctColumns != nil {
		clone.distinctColumns = make([]Column, len(s.distinctColumns))
		copy(clone.distinctColumns, s.distinctColumns)
	}

	if s.orderBys != nil {
		clone.orderBys = make([]Expression, len(s.orderBys))
		copy(clone.orderBys, s.orderBys)
	}

	if s.groupBys != nil {
		clone.groupBys = make([]Expression, len(s.groupBys))
		copy(clone.groupBys, s.groupBys)
	}

	if s.havings != nil {
		clone.havings = make([]Expression, len(s.havings))
		copy(clone.havings, s.havings)
	}

	return clone
}

// * SELECT EXECUTION * //

// One выполняет запрос с LIMIT 1 и возвращает первую строку как указатель на модель T.
// Возвращает ErrNotFound если строк нет.
// Пример: user, err := dew.From(db, UserSchema).Where(...).One() вернет *User
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

	return results[0], nil // results[0] уже указатель (*T)
}

func (s *Selector[T]) All(ctxs ...context.Context) ([]*T, error) {
	ctx := getCtx(ctxs)

	query := s.buildQuery()

	rows, err := s.db.QueryContext(ctx, query, s.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modelType := reflect.TypeOf(new(T)).Elem()
	targetIndices, err := resolveScanIndices(modelType, s.columns)
	if err != nil {
		return nil, err
	}

	var results []*T

	for rows.Next() {
		newItem := reflect.New(modelType)
		val := newItem.Elem()

		scanArgs := prepareScanArgs(val, targetIndices)

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		results = append(results, newItem.Interface().(*T))
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

	return results[0], nil
}

func (s *Selector[T]) Scan(dest ...any) error {
	return s.ScanCtx(context.Background(), dest...)
}

func (s *Selector[T]) ScanCtx(ctx context.Context, dest ...any) error {
	if len(dest) == 0 {
		return fmt.Errorf("dew: Scan expects at least one destination")
	}

	firstDest := dest[0]
	val := reflect.ValueOf(firstDest)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("dew: Scan expects a non-nil pointer")
	}

	elem := val.Elem()

	if elem.Kind() == reflect.Slice {
		s.limitCount = 0
	}

	query := s.buildQuery()

	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := s.db.QueryContext(ctx, query, s.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if elem.Kind() == reflect.Slice {
		return s.scanIntoSlice(rows, elem)
	}

	return s.scanIntoOne(rows, firstDest)
}

func (s *Selector[T]) scanIntoSlice(rows *sql.Rows, sliceVal reflect.Value) error {
	elemType := sliceVal.Type().Elem()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	numColumns := len(columns)

	var targetIndices []int
	isStruct := elemType.Kind() == reflect.Struct

	if isStruct {
		fieldMap := buildFieldMap(elemType)

		targetIndices = make([]int, numColumns)
		for i, colName := range columns {
			idx, ok := fieldMap[strings.ToLower(colName)]
			if !ok {
				return fmt.Errorf("dew: struct field for column '%s' not found", colName)
			}
			targetIndices[i] = idx
		}
	} else {
		if numColumns != 1 {
			return fmt.Errorf("dew: cannot scan %d columns into %s", numColumns, elemType.String())
		}
	}

	for rows.Next() {
		newElemPtr := reflect.New(elemType)

		if isStruct {
			elemVal := newElemPtr.Elem()
			scanArgs := prepareScanArgs(elemVal, targetIndices)

			if err := rows.Scan(scanArgs...); err != nil {
				return err
			}
		} else {
			if err := rows.Scan(newElemPtr.Interface()); err != nil {
				return err
			}
		}

		sliceVal.Set(reflect.Append(sliceVal, newElemPtr.Elem()))
	}
	return nil
}

func (s *Selector[T]) scanIntoOne(rows *sql.Rows, dest any) error {
	if !rows.Next() {
		return ErrNotFound
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	val := reflect.ValueOf(dest).Elem()

	if val.Kind() == reflect.Struct {
		modelType := val.Type()
		fieldMap := buildFieldMap(modelType)

		scanArgs := make([]any, len(columns))
		for i, colName := range columns {
			idx, ok := fieldMap[strings.ToLower(colName)]
			if !ok {
				return fmt.Errorf("dew: struct field for column '%s' not found", colName)
			}
			scanArgs[i] = val.Field(idx).Addr().Interface()
		}

		return rows.Scan(scanArgs...)
	}

	if len(columns) != 1 {
		return fmt.Errorf("dew: cannot scan %d columns into %s", len(columns), val.Type().String())
	}

	return rows.Scan(dest)
}

// * Utils Functions * //

func (s *Selector[T]) buildQuery() string {
	selectClause := "*"
	if len(s.columns) > 0 {
		colSqls := make([]string, len(s.columns))
		for i, c := range s.columns {
			colSqls[i] = c.Sql()
		}
		selectClause = strings.Join(colSqls, ", ")
	}

	distinctClause := ""
	if s.distinctColumns != nil {
		if len(s.distinctColumns) == 0 {
			distinctClause = "DISTINCT "
		} else {
			distinctColSqls := make([]string, len(s.distinctColumns))
			for i, c := range s.distinctColumns {
				distinctColSqls[i] = c.Sql()
			}
			selectClause = strings.Join(distinctColSqls, ", ")
			distinctClause = "DISTINCT "
		}
	}

	query := fmt.Sprintf("SELECT %s%s FROM %s", distinctClause, selectClause, s.tableName)

	if len(s.wheres) > 0 {
		query += " WHERE " + strings.Join(s.wheres, " AND ")
	}

	if len(s.orderBys) > 0 {
		var orderSqls []string
		for _, order := range s.orderBys {
			orderSqls = append(orderSqls, order.Sql())
		}
		query += " ORDER BY " + strings.Join(orderSqls, ", ")
	}

	if len(s.groupBys) > 0 {
		var groupSqls []string
		for _, order := range s.groupBys {
			groupSqls = append(groupSqls, order.Sql())
		}
		query += " GROUP BY " + strings.Join(groupSqls, ", ")
	}

	if len(s.havings) > 0 {
		var havingSqls []string
		for _, having := range s.havings {
			havingSqls = append(havingSqls, having.Sql())
		}
		query += " HAVING " + strings.Join(havingSqls, " AND ")
	}

	if s.limitCount > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.limitCount)
	}
	if s.offsetCount > 0 {
		query += fmt.Sprintf(" OFFSET %d", s.offsetCount)
	}

	return query
}

func buildFieldMap(typ reflect.Type) map[string]int {
	fieldMap := make(map[string]int, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fieldMap[strings.ToLower(typ.Field(i).Name)] = i
	}
	return fieldMap
}

func resolveScanIndices(modelType reflect.Type, columns []Column) ([]int, error) {
	// If there is no specific columns, return nil
	if len(columns) == 0 {
		return nil, nil
	}

	// Build map: "field_name_in_lower_case" -> index
	fieldMap := buildFieldMap(modelType)

	indices := make([]int, len(columns))
	for i, col := range columns {
		colName := col.ColumnName()
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
