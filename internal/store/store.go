package store

import (
	"errors"
	"iter"
)

var (
	ErrNotFound = errors.New("secret not found")
)

type Store interface {
	Save(vault, name, value string) error
	Get(vault, name string) (string, error)
	Delete(vault, name string) error
	List(vault string) iter.Seq2[string, error]
	ListVaults() iter.Seq2[string, error]
	Nuke() error
	Close() error
}
