package dew

import "database/sql"

type DB struct {
	*sql.DB
	dialect Dialect
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	var dialect Dialect
	switch driverName {
	case "postgres", "pgx":
		dialect = PostgreSQLDialect{}
	case "sqlserver", "mssql":
		dialect = MSSQLDialect{}
	case "mysql":
		dialect = MySQLDialect{}
	default:
		dialect = SQLiteDialect{}
	}

	return &DB{db, dialect}, nil
}

func OpenWithDialect(driverName, dataSourceName string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db, dialect}, nil
}
