package dew

import (
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

func (s *Selector[T]) One() (*T, error) {
	s.limitCount = 1
	results, err := s.All()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound
	}
	// Возвращаем указатель на элемент слайса
	return &results[0], nil
}

// All выполняет запрос и возвращает список записей
func (s *Selector[T]) All() ([]T, error) {
	selectClause := "*"
	if len(s.columns) > 0 {
		colSqls := make([]string, len(s.columns))
		for i, c := range s.columns {
			colSqls[i] = c.Sql()
		}
		selectClause = strings.Join(colSqls, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, s.tableName)

	// 2. Добавляем WHERE
	if len(s.wheres) > 0 {
		query += " WHERE " + strings.Join(s.wheres, " AND ")
	}

	// 3. Добавляем ORDER BY
	if len(s.orderBys) > 0 {
		var orderSqls []string
		for _, order := range s.orderBys {
			orderSqls = append(orderSqls, order.Sql())
			// Если бы у OrderBy были аргументы, их нужно добавить в args тут
		}
		query += " ORDER BY " + strings.Join(orderSqls, ", ")
	}

	// 4. Добавляем LIMIT и OFFSET
	if s.limitCount > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.limitCount)
	}
	if s.offsetCount > 0 {
		query += fmt.Sprintf(" OFFSET %d", s.offsetCount)
	}

	// 5. Выполняем запрос
	rows, err := s.db.Query(query, s.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 6. Сканирование результатов
	var results []T
	modelType := reflect.TypeOf(new(T)).Elem()

	for rows.Next() {
		newPtr := reflect.New(modelType)
		val := newPtr.Elem()
		var scanArgs []any

		if len(s.columns) > 0 {
			// СЛОЖНЫЙ ПУТЬ: Частичная выборка
			for _, colExpr := range s.columns {
				colName := colExpr.Sql()
				found := false
				// Ищем поле структуры по имени колонки
				for i := 0; i < val.NumField(); i++ {
					fieldInfo := modelType.Field(i)
					// Простое сравнение имен (можно улучшить, добавив чтение тегов `db`)
					if strings.EqualFold(fieldInfo.Name, colName) {
						scanArgs = append(scanArgs, val.Field(i).Addr().Interface())
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("dew: struct field for column '%s' not found", colName)
				}
			}
		} else {
			// ПРОСТОЙ ПУТЬ: Выборка всех полей по порядку
			numField := val.NumField()
			scanArgs = make([]any, numField)
			for i := 0; i < numField; i++ {
				scanArgs[i] = val.Field(i).Addr().Interface()
			}
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		results = append(results, *newPtr.Interface().(*T))
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
