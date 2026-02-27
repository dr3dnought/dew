package dew

import (
	"context"
	"database/sql"
)

// Querier is the common interface satisfied by both *DB and *Tx.
// It is unexported-method-sealed: only types in this package can implement it.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	getDialect() Dialect
}

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

func (db *DB) getDialect() Dialect {
	return db.dialect
}

// BeginTx starts a new transaction with the given options.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, dialect: db.dialect}, nil
}

// Begin starts a new transaction with default options.
func (db *DB) Begin() (*Tx, error) {
	return db.BeginTx(context.Background(), nil)
}

// Tx wraps *sql.Tx with dialect information, satisfying the Querier interface.
type Tx struct {
	*sql.Tx
	dialect Dialect
}

func (tx *Tx) getDialect() Dialect {
	return tx.dialect
}
