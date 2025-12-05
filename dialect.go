package dew

import "strconv"

type Dialect interface {
	Placeholder(index int) string
}

type SQLiteDialect struct{}

func (d SQLiteDialect) Placeholder(index int) string {
	return "?"
}

type MySQLDialect struct{}

func (d MySQLDialect) Placeholder(index int) string {
	return "?"
}

type PostgreSQLDialect struct{}

func (d PostgreSQLDialect) Placeholder(index int) string {
	return "$" + strconv.Itoa(index+1)
}

type MSSQLDialect struct{}

func (d MSSQLDialect) Placeholder(index int) string {
	return "@p" + strconv.Itoa(index+1)
}
