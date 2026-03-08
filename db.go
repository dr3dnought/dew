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
	mapError(err error) error
}

type DB struct {
	*sql.DB
	dialect     Dialect
	errorMapper ErrorMapper
}

func Open(driverName, dataSourceName string, dialect Dialect, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	d := &DB{DB: db, dialect: dialect}
	for _, opt := range opts {
		opt(d)
	}
	return d, nil
}

// NewDB wraps an existing *sql.DB with a dialect and options.
func NewDB(db *sql.DB, dialect Dialect, opts ...DBOption) *DB {
	d := &DB{DB: db, dialect: dialect}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// DBOption configures a DB instance.
type DBOption func(*DB)

// WithErrorMapper sets a custom error mapper for the DB.
func WithErrorMapper(mapper ErrorMapper) DBOption {
	return func(db *DB) {
		db.errorMapper = mapper
	}
}

func (db *DB) getDialect() Dialect {
	return db.dialect
}

func (db *DB) mapError(err error) error {
	if err == nil || db.errorMapper == nil {
		return err
	}
	return db.errorMapper(err)
}

// BeginTx starts a new transaction with the given options.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, dialect: db.dialect, errorMapper: db.errorMapper}, nil
}

// Begin starts a new transaction with default options.
func (db *DB) Begin() (*Tx, error) {
	return db.BeginTx(context.Background(), nil)
}

// Tx wraps *sql.Tx with dialect information, satisfying the Querier interface.
type Tx struct {
	*sql.Tx
	dialect     Dialect
	errorMapper ErrorMapper
}

func (tx *Tx) getDialect() Dialect {
	return tx.dialect
}

func (tx *Tx) mapError(err error) error {
	if err == nil || tx.errorMapper == nil {
		return err
	}
	return tx.errorMapper(err)
}
