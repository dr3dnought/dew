package dew

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type joinType string

const (
	InnerJoinType joinType = "INNER JOIN"
	LeftJoinType  joinType = "LEFT JOIN"
	RightJoinType joinType = "RIGHT JOIN"
)

type joinInfo struct {
	joinType  joinType
	tableName string
	onLeft    Column
	onRight   Column
}

type Selector[T any] struct {
	db        Querier
	tableName string
	columns   []Column
	whereExprs []Expression
	args       []any

	ctes []cteClause

	fromSubQuery Expression
	fromAlias    string

	distinctColumns []Column

	joins []joinInfo

	limitCount  int
	offsetCount int
	orderBys    []Expression

	groupBys []Expression
	havings  []Expression

	lockClause string
}

func From[T any](db Querier, schema Tabler) *Selector[T] {
	return &Selector[T]{
		db:        db,
		tableName: schema.TableName(),
	}
}

func FromSub[T any](db Querier, subQuery Expression, alias string) *Selector[T] {
	return &Selector[T]{
		db:           db,
		fromSubQuery: subQuery,
		fromAlias:    alias,
	}
}

// **** Implementations of the Expression interface **** //

func (s *Selector[T]) Sql() string {
	return s.buildRawQuery()
}

func (s *Selector[T]) Args() []any {
	return s.args
}

// *** SELECTOR *** ///

func (s *Selector[T]) Where(expr ...Expression) *Selector[T] {
	for _, exp := range expr {
		if exp.Sql() != "" {
			s.whereExprs = append(s.whereExprs, exp)
		}
	}
	return s
}

func (s *Selector[T]) With(ctes ...cteClause) *Selector[T] {
	s.ctes = append(s.ctes, ctes...)
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

func (s *Selector[T]) InnerJoin(schema Tabler, onLeft Column, onRight Column) *Selector[T] {
	s.joins = append(s.joins, joinInfo{
		joinType:  InnerJoinType,
		tableName: schema.TableName(),
		onLeft:    onLeft,
		onRight:   onRight,
	})
	return s
}

func (s *Selector[T]) LeftJoin(schema Tabler, onLeft Column, onRight Column) *Selector[T] {
	s.joins = append(s.joins, joinInfo{
		joinType:  LeftJoinType,
		tableName: schema.TableName(),
		onLeft:    onLeft,
		onRight:   onRight,
	})
	return s
}

func (s *Selector[T]) RightJoin(schema Tabler, onLeft Column, onRight Column) *Selector[T] {
	s.joins = append(s.joins, joinInfo{
		joinType:  RightJoinType,
		tableName: schema.TableName(),
		onLeft:    onLeft,
		onRight:   onRight,
	})
	return s
}

func (s *Selector[T]) ForUpdate() *Selector[T] {
	s.lockClause = "FOR UPDATE"
	return s
}

func (s *Selector[T]) ForShare() *Selector[T] {
	s.lockClause = "FOR SHARE"
	return s
}

func (s *Selector[T]) NoWait() *Selector[T] {
	s.lockClause += " NOWAIT"
	return s
}

func (s *Selector[T]) SkipLocked() *Selector[T] {
	s.lockClause += " SKIP LOCKED"
	return s
}

func (s *Selector[T]) ToSql() (string, []any, error) {
	return s.buildQuery(), s.args, nil
}

func (s *Selector[T]) Clone() *Selector[T] {
	clone := &Selector[T]{
		db:           s.db,
		tableName:    s.tableName,
		fromSubQuery: s.fromSubQuery,
		fromAlias:    s.fromAlias,
		limitCount:   s.limitCount,
		offsetCount:  s.offsetCount,
		lockClause:   s.lockClause,
	}

	if s.ctes != nil {
		clone.ctes = make([]cteClause, len(s.ctes))
		copy(clone.ctes, s.ctes)
	}

	if s.columns != nil {
		clone.columns = make([]Column, len(s.columns))
		copy(clone.columns, s.columns)
	}

	if s.whereExprs != nil {
		clone.whereExprs = make([]Expression, len(s.whereExprs))
		copy(clone.whereExprs, s.whereExprs)
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

	if s.joins != nil {
		clone.joins = make([]joinInfo, len(s.joins))
		copy(clone.joins, s.joins)
	}

	return clone
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

	return results[0], nil // results[0] уже указатель (*T)
}

func (s *Selector[T]) All(ctxs ...context.Context) ([]*T, error) {
	ctx := getCtx(ctxs)

	query := s.buildQuery()

	rows, err := s.db.QueryContext(ctx, query, s.args...)
	if err != nil {
		return nil, s.db.mapError(err)
	}
	defer rows.Close()

	results, err := scanAll[T](rows, s.columns)
	return results, s.db.mapError(err)
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

func (s *Selector[T]) Exists(ctxs ...context.Context) (bool, error) {
	ctx := getCtx(ctxs)

	clone := s.Clone()
	clone.columns = nil
	clone.limitCount = 1

	query := clone.buildQuery()
	query = fmt.Sprintf("SELECT 1 %s", query[strings.Index(query, " FROM "):])

	var exists int
	err := s.db.QueryRowContext(ctx, query, clone.args...).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, s.db.mapError(err)
	}

	return true, nil
}

func (s *Selector[T]) Count(ctxs ...context.Context) (int64, error) {
	ctx := getCtx(ctxs)

	clone := s.Clone()
	clone.columns = []Column{Count()}
	clone.limitCount = 0
	clone.offsetCount = 0
	clone.orderBys = nil

	query := clone.buildQuery()

	var count int64
	err := s.db.QueryRowContext(ctx, query, clone.args...).Scan(&count)

	if err != nil {
		return 0, s.db.mapError(err)
	}

	return count, nil
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
		return s.db.mapError(err)
	}
	defer rows.Close()

	if elem.Kind() == reflect.Slice {
		return s.db.mapError(s.scanIntoSlice(rows, elem))
	}

	return s.db.mapError(s.scanIntoOne(rows, firstDest))
}

func (s *Selector[T]) ScanWith(scanner func(*sql.Rows) (*T, error), ctxs ...context.Context) ([]*T, error) {
	ctx := getCtx(ctxs)

	query := s.buildQuery()

	rows, err := s.db.QueryContext(ctx, query, s.args...)
	if err != nil {
		return nil, s.db.mapError(err)
	}
	defer rows.Close()

	var results []*T

	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, s.db.mapError(err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, s.db.mapError(err)
	}

	return results, nil
}

func (s *Selector[T]) scanIntoSlice(rows *sql.Rows, sliceVal reflect.Value) error {
	elemType := sliceVal.Type().Elem()

	// Check if the element type implements RowScanner (via pointer receiver)
	elemPtrType := reflect.PointerTo(elemType)
	if elemType.Kind() == reflect.Struct && elemPtrType.Implements(reflect.TypeOf((*RowScanner)(nil)).Elem()) {
		for rows.Next() {
			newElemPtr := reflect.New(elemType)
			scanner := newElemPtr.Interface().(RowScanner)
			if err := scanner.ScanRow(rows); err != nil {
				return err
			}
			sliceVal.Set(reflect.Append(sliceVal, newElemPtr.Elem()))
		}
		return nil
	}

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

	// Check if dest implements RowScanner
	if scanner, ok := dest.(RowScanner); ok {
		return scanner.ScanRow(rows)
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

// buildRawQuery builds the query with ? placeholders (no dialect replacement).
// It collects all args and stores them in s.args.
func (s *Selector[T]) buildRawQuery() string {
	var finalArgs []any

	// CTE prefix
	var ctePrefix string
	if len(s.ctes) > 0 {
		var cteParts []string
		hasRecursive := false
		for _, cte := range s.ctes {
			if cte.recursive {
				hasRecursive = true
			}
			cteParts = append(cteParts, cte.name+" AS ("+cte.query.Sql()+")")
			finalArgs = append(finalArgs, cte.query.Args()...)
		}
		keyword := "WITH "
		if hasRecursive {
			keyword = "WITH RECURSIVE "
		}
		ctePrefix = keyword + strings.Join(cteParts, ", ") + " "
	}

	selectClause := "*"
	if len(s.columns) > 0 {
		colSqls := make([]string, len(s.columns))
		for i, c := range s.columns {
			sqlStr := c.Sql()
			if strings.Contains(strings.ToUpper(sqlStr), " AS ") {
				colSqls[i] = sqlStr
			} else if alias := c.Alias(); alias != nil {
				original := c.ColumnName()
				if table := c.TableName(); table != "" {
					original = fmt.Sprintf("%s.%s", table, original)
				}
				colSqls[i] = fmt.Sprintf("%s AS %s", original, *alias)
			} else {
				colSqls[i] = sqlStr
			}
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
				sqlStr := c.Sql()
				if strings.Contains(sqlStr, " AS ") {
					distinctColSqls[i] = sqlStr
				} else if alias := c.Alias(); alias != nil {
					original := c.ColumnName()
					if table := c.TableName(); table != "" {
						original = fmt.Sprintf("%s.%s", table, original)
					}
					distinctColSqls[i] = fmt.Sprintf("%s AS %s", original, *alias)
				} else {
					distinctColSqls[i] = sqlStr
				}
			}
			selectClause = strings.Join(distinctColSqls, ", ")
			distinctClause = "DISTINCT "
		}
	}

	var fromClause string
	if s.fromSubQuery != nil {
		fromClause = "(" + s.fromSubQuery.Sql() + ") AS " + s.fromAlias
		finalArgs = append(finalArgs, s.fromSubQuery.Args()...)
	} else {
		fromClause = s.tableName
	}

	query := ctePrefix + fmt.Sprintf("SELECT %s%s FROM %s", distinctClause, selectClause, fromClause)

	for _, join := range s.joins {
		query += fmt.Sprintf(" %s %s ON %s = %s",
			join.joinType,
			join.tableName,
			join.onLeft.Sql(),
			join.onRight.Sql(),
		)
	}

	if len(s.whereExprs) > 0 {
		var whereSqls []string
		for _, expr := range s.whereExprs {
			whereSqls = append(whereSqls, expr.Sql())
			finalArgs = append(finalArgs, expr.Args()...)
		}
		query += " WHERE " + strings.Join(whereSqls, " AND ")
	}

	if len(s.groupBys) > 0 {
		var groupSqls []string
		for _, expr := range s.groupBys {
			sqlStr := expr.Sql()
			if col, ok := expr.(Column); ok {
				if alias := col.Alias(); alias != nil {
					sqlStr = *alias
				} else {
					if idx := strings.Index(sqlStr, " AS "); idx != -1 {
						sqlStr = sqlStr[:idx]
					}
				}
			} else {
				if idx := strings.Index(sqlStr, " AS "); idx != -1 {
					sqlStr = sqlStr[:idx]
				}
			}
			groupSqls = append(groupSqls, sqlStr)
		}
		query += " GROUP BY " + strings.Join(groupSqls, ", ")
	}

	if len(s.havings) > 0 {
		var havingSqls []string
		for _, having := range s.havings {
			havingSqls = append(havingSqls, having.Sql())
			finalArgs = append(finalArgs, having.Args()...)
		}
		query += " HAVING " + strings.Join(havingSqls, " AND ")
	}

	if len(s.orderBys) > 0 {
		var orderSqls []string
		for _, order := range s.orderBys {
			orderSqls = append(orderSqls, order.Sql())
			finalArgs = append(finalArgs, order.Args()...)
		}
		query += " ORDER BY " + strings.Join(orderSqls, ", ")
	}

	if s.limitCount > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.limitCount)
	}
	if s.offsetCount > 0 {
		query += fmt.Sprintf(" OFFSET %d", s.offsetCount)
	}

	if s.lockClause != "" {
		query += " " + s.lockClause
	}

	s.args = finalArgs

	return query
}

// buildQuery builds the final query with dialect-specific placeholders.
func (s *Selector[T]) buildQuery() string {
	raw := s.buildRawQuery()
	return replacePlaceholders(raw, s.db.getDialect(), 0)
}

func buildFieldMap(typ reflect.Type) map[string]int {
	fieldMap := make(map[string]int, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		columnName := getColumnName(field)
		if columnName == "" {
			continue
		}
		key := strings.ToLower(columnName)
		fieldMap[key] = i
	}
	return fieldMap
}

// RowScanner allows a struct to define its own scanning logic instead of relying
// on reflection. Implement this interface on *T to use custom scanning.
//
//	func (u *User) ScanRow(rows *sql.Rows) error {
//	    return rows.Scan(&u.ID, &u.Name, &u.Email)
//	}
type RowScanner interface {
	ScanRow(rows *sql.Rows) error
}

// scanAll scans all rows into []*T. If *T implements RowScanner, it uses
// the custom scan method; otherwise falls back to reflection-based scanning.
func scanAll[T any](rows *sql.Rows, columns []Column) ([]*T, error) {
	// Check if *T implements RowScanner
	var zero T
	if _, ok := any(&zero).(RowScanner); ok {
		var results []*T
		for rows.Next() {
			item := new(T)
			if err := any(item).(RowScanner).ScanRow(rows); err != nil {
				return nil, err
			}
			results = append(results, item)
		}
		return results, nil
	}

	// Reflection-based scanning
	modelType := reflect.TypeOf(new(T)).Elem()
	targetIndices, err := resolveScanIndices(modelType, columns)
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
		if alias := col.Alias(); alias != nil {
			colName = *alias
		}
		key := strings.ToLower(colName)
		idx, ok := fieldMap[key]
		if !ok {
			return nil, fmt.Errorf("dew: struct field for column '%s' not found", colName)
		}
		indices[i] = idx
	}

	return indices, nil
}

func prepareScanArgs(val reflect.Value, indices []int) []any {
	if indices == nil {
		typ := val.Type()
		num := typ.NumField()
		args := make([]any, 0, num)
		for i := range num {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			if getColumnName(field) == "" {
				continue
			}
			args = append(args, val.Field(i).Addr().Interface())
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

func toSnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 5)

	var lastUpper bool

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && !lastUpper {
				b.WriteByte('_')
			}

			b.WriteRune(unicode.ToLower(r))
			lastUpper = true
		} else {
			b.WriteRune(r)
			lastUpper = false
		}
	}
	return b.String()
}
