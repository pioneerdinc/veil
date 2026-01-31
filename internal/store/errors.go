package store

import "errors"

// Common errors for the store package
var (
	ErrDatabaseOpen = errors.New("failed to open database")

	ErrMigrationFailed = errors.New("database migration failed")

	ErrSaveFailed = errors.New("failed to save secret")

	ErrGetFailed = errors.New("failed to get secret")

	ErrDeleteFailed = errors.New("failed to delete secret")

	ErrListFailed = errors.New("failed to list secrets")

	ErrNukeFailed = errors.New("failed to nuke database")
)
