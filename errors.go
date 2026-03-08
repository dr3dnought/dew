package dew

import "errors"

var (
	ErrNotFound         = errors.New("dew: record not found")
	ErrUniqueViolation  = errors.New("dew: unique constraint violation")
	ErrForeignKey       = errors.New("dew: foreign key violation")
	ErrCheckViolation   = errors.New("dew: check constraint violation")
	ErrNotNull          = errors.New("dew: not null violation")
)

// ErrorMapper transforms a database driver error into a dew sentinel error.
// Return the original error unchanged if no mapping applies.
type ErrorMapper func(err error) error
