// testhelpers provides reusable test utilities for veil tests.
// These helpers are only imported in test files.
package testhelpers

import (
	"iter"
	"strings"

	"github.com/ossydotpy/veil/internal/store"
)

// SaveCall records a call to MemStore.Save for verification.
type SaveCall struct {
	Vault string
	Name  string
	Value string
}

// MemStore is a test double for store.Store with tracking capabilities.
// It stores data in memory and optionally tracks all Save() calls.
type MemStore struct {
	data      map[string]string
	SaveCalls []SaveCall // Records every Save() call for verification
	GetErr    error      // Configurable error for Get()
	SaveErr   error      // Configurable error for Save()
	ListErr   error      // Configurable error for List()
}

// NewMemStore creates a new MemStore with initialized data map.
func NewMemStore() *MemStore {
	return &MemStore{
		data:      make(map[string]string),
		SaveCalls: make([]SaveCall, 0),
	}
}

// Save stores the value and records the call.
func (s *MemStore) Save(vault, name, value string) error {
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.SaveCalls = append(s.SaveCalls, SaveCall{Vault: vault, Name: name, Value: value})
	s.data[vault+"/"+name] = value
	return nil
}

// Get retrieves a value from the store.
func (s *MemStore) Get(vault, name string) (string, error) {
	if s.GetErr != nil {
		return "", s.GetErr
	}
	val, ok := s.data[vault+"/"+name]
	if !ok {
		return "", store.ErrNotFound
	}
	return val, nil
}

// Delete removes a key from the store.
func (s *MemStore) Delete(vault, name string) error {
	delete(s.data, vault+"/"+name)
	return nil
}

// List returns all keys in a vault.
func (s *MemStore) List(vault string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if s.ListErr != nil {
			yield("", s.ListErr)
			return
		}
		for key := range s.data {
			if vaultName, name := splitKey(key); vaultName == vault {
				if !yield(name, nil) {
					return
				}
			}
		}
	}
}

// ListVaults returns all vault names (simplified - returns empty for now).
func (s *MemStore) ListVaults() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {}
}

// Search returns secrets matching a pattern (simplified - returns empty for now).
func (s *MemStore) Search(pattern string) iter.Seq2[store.SecretRef, error] {
	return func(yield func(store.SecretRef, error) bool) {}
}

// Nuke clears all data.
func (s *MemStore) Nuke() error {
	s.data = make(map[string]string)
	s.SaveCalls = make([]SaveCall, 0)
	return nil
}

// Close is a no-op for the in-memory store.
func (s *MemStore) Close() error { return nil }

// SaveCount returns the number of Save() calls made.
func (s *MemStore) SaveCount() int {
	return len(s.SaveCalls)
}

// HasKey checks if a key exists in a vault.
func (s *MemStore) HasKey(vault, name string) bool {
	_, err := s.Get(vault, name)
	return err == nil
}

// splitKey separates vault/name from the internal key format.
func splitKey(key string) (string, string) {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

// compile-time check that MemStore implements store.Store
var _ store.Store = (*MemStore)(nil)
