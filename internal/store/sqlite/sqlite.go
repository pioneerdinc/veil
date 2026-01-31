package sqlite

import (
	"database/sql"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossydotpy/veil/internal/store"
	_ "modernc.org/sqlite"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(dbPath string) (store.Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("could not create storage directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOpen, err)
	}

	if err := os.Chmod(dbPath, 0600); err != nil && !os.IsNotExist(err) {
	}

	s := &SqliteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SqliteStore) migrate() error {
	query := `
CREATE TABLE IF NOT EXISTS secrets (
vault TEXT NOT NULL,
name TEXT NOT NULL,
value TEXT NOT NULL,
PRIMARY KEY (vault, name)
);`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%w: %v", store.ErrMigrationFailed, err)
	}
	return nil
}

func (s *SqliteStore) Save(vault, name, value string) error {
	query := `INSERT OR REPLACE INTO secrets (vault, name, value) VALUES (?, ?, ?);`
	_, err := s.db.Exec(query, vault, name, value)
	if err != nil {
		return fmt.Errorf("%w: %v", store.ErrSaveFailed, err)
	}
	return nil
}

func (s *SqliteStore) Get(vault, name string) (string, error) {
	var value string
	query := `SELECT value FROM secrets WHERE vault = ? AND name = ?;`
	err := s.db.QueryRow(query, vault, name).Scan(&value)
	if err == sql.ErrNoRows {
		return "", store.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", store.ErrGetFailed, err)
	}
	return value, nil
}

func (s *SqliteStore) Delete(vault, name string) error {
	query := `DELETE FROM secrets WHERE vault = ? AND name = ?;`
	_, err := s.db.Exec(query, vault, name)
	if err != nil {
		return fmt.Errorf("%w: %v", store.ErrDeleteFailed, err)
	}
	return nil
}

func (s *SqliteStore) List(vault string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		query := `SELECT name FROM secrets WHERE vault = ? ORDER BY name ASC;`
		rows, err := s.db.Query(query, vault)
		if err != nil {
			yield("", fmt.Errorf("%w: %v", store.ErrListFailed, err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				if !yield("", err) {
					return
				}
				continue
			}
			if !yield(name, nil) {
				return
			}
		}
	}
}

func (s *SqliteStore) ListVaults() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		query := `SELECT DISTINCT vault FROM secrets ORDER BY vault ASC;`
		rows, err := s.db.Query(query)
		if err != nil {
			yield("", fmt.Errorf("%w: %v", store.ErrListFailed, err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var vault string
			if err := rows.Scan(&vault); err != nil {
				if !yield("", err) {
					return
				}
				continue
			}
			if !yield(vault, nil) {
				return
			}
		}
	}
}

func (s *SqliteStore) Search(pattern string) iter.Seq2[store.SecretRef, error] {
	return func(yield func(store.SecretRef, error) bool) {
		sqlPattern := convertPattern(pattern)
		query := `SELECT vault, name FROM secrets WHERE LOWER(name) LIKE LOWER(?) ORDER BY vault ASC, name ASC;`
		rows, err := s.db.Query(query, sqlPattern)
		if err != nil {
			yield(store.SecretRef{}, fmt.Errorf("failed to search secrets: %w", err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var ref store.SecretRef
			if err := rows.Scan(&ref.Vault, &ref.Name); err != nil {
				if !yield(store.SecretRef{}, err) {
					return
				}
				continue
			}
			if !yield(ref, nil) {
				return
			}
		}
	}
}

func convertPattern(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "*", "%")
	pattern = strings.ReplaceAll(pattern, "?", "_")
	return pattern
}

func (s *SqliteStore) Nuke() error {
	query := `DELETE FROM secrets;`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("%w: %v", store.ErrNukeFailed, err)
	}
	_, _ = s.db.Exec("VACUUM;")
	return nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}
