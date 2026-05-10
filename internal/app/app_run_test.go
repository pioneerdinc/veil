package app

import (
	"errors"
	"testing"
)

func TestRunEnv_ReturnsAllSecrets(t *testing.T) {
	a, _, _ := setupTestApp(t)

	a.Set("myvault", "DB_HOST", "localhost")
	a.Set("myvault", "DB_PORT", "5432")

	secrets, err := a.RunEnv("myvault", nil, nil)
	if err != nil {
		t.Fatalf("RunEnv error: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("RunEnv returned %d secrets, want 2", len(secrets))
	}
	if secrets["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", secrets["DB_HOST"], "localhost")
	}
	if secrets["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT = %q, want %q", secrets["DB_PORT"], "5432")
	}
}

func TestRunEnv_NonExistentVault(t *testing.T) {
	a, _, _ := setupTestApp(t)

	_, err := a.RunEnv("nonexistent", nil, nil)
	if err == nil {
		t.Fatal("RunEnv should return error for non-existent vault")
	}
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("error should wrap ErrVaultNotFound, got: %v", err)
	}
}

func TestRunEnv_ListVaultsError(t *testing.T) {
	a, memStore, _ := setupTestApp(t)
	memStore.ListVaultsErr = errors.New("store unreachable")

	_, err := a.RunEnv("any-vault", nil, nil)
	if err == nil {
		t.Fatal("RunEnv should return error when ListVaults fails")
	}
	if errors.Is(err, ErrVaultNotFound) {
		t.Error("error should NOT be ErrVaultNotFound when store is unreachable")
	}
	if !errors.Is(err, memStore.ListVaultsErr) {
		t.Errorf("error should wrap the underlying store error, got: %v", err)
	}
}