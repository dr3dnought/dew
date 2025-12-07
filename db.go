package dew

import "database/sql"

type DB struct {
	*sql.DB
	dialect Dialect
}

func Open(driverName, dataSourceName string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db, dialect: dialect}, nil
}
