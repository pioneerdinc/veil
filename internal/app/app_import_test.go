package app

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/importer"
	"github.com/ossydotpy/veil/internal/store"
	"github.com/ossydotpy/veil/internal/testhelpers"
)

func setupTestApp(t *testing.T) (*App, *testhelpers.MemStore, *crypto.Engine) {
	t.Helper()
	// Use a fixed test key
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	engine, err := crypto.NewEngine(keyHex)
	if err != nil {
		t.Fatalf("Failed to create crypto engine: %v", err)
	}
	memStore := testhelpers.NewMemStore()
	return New(memStore, engine), memStore, engine
}

func TestImport_Categorization_NewUpdatedSkipped(t *testing.T) {
	app, ts, _ := setupTestApp(t)

	app.Set("test-vault", "EXISTING_SAME", "same-value")
	app.Set("test-vault", "EXISTING_DIFFERENT", "old-value")
	ts.SaveCalls = nil // Reset tracking

	imported := map[string]string{
		"NEW_KEY_1":          "value1",
		"NEW_KEY_2":          "value2",
		"EXISTING_SAME":      "same-value", // Should be skipped (same value)
		"EXISTING_DIFFERENT": "new-value",  // Should be updated (different value, but no force)
	}

	// Mock the importer by directly calling app.Import with the parsed data
	preview, err := app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	// Verify categorization
	if len(preview.NewKeys) != 2 {
		t.Errorf("NewKeys = %v (count=%d), want 2", preview.NewKeys, len(preview.NewKeys))
	}
	if !slices.Contains(preview.NewKeys, "NEW_KEY_1") || !slices.Contains(preview.NewKeys, "NEW_KEY_2") {
		t.Errorf("NewKeys missing expected keys: %v", preview.NewKeys)
	}

	// Without force, different values should be skipped
	if len(preview.UpdatedKeys) != 0 {
		t.Errorf("UpdatedKeys = %v (count=%d), want 0 (no force)", preview.UpdatedKeys, len(preview.UpdatedKeys))
	}

	if len(preview.SkippedKeys) != 2 {
		t.Errorf("SkippedKeys = %v (count=%d), want 2", preview.SkippedKeys, len(preview.SkippedKeys))
	}

	// Sorted order check
	if !slices.IsSorted(preview.NewKeys) {
		t.Errorf("NewKeys not sorted: %v", preview.NewKeys)
	}
}

func TestImport_ForceOverwritesDifferentValues(t *testing.T) {
	app, ts, _ := setupTestApp(t)

	// Pre-populate
	app.Set("test-vault", "EXISTING_DIFFERENT", "old-value")
	ts.SaveCalls = nil

	imported := map[string]string{
		"EXISTING_DIFFERENT": "new-value",
	}

	// Without force - should skip
	preview, err := app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     true,
		Force:      false,
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	if len(preview.SkippedKeys) != 1 {
		t.Errorf("Without force: SkippedKeys = %d, want 1", len(preview.SkippedKeys))
	}
	if len(preview.UpdatedKeys) != 0 {
		t.Errorf("Without force: UpdatedKeys = %d, want 0", len(preview.UpdatedKeys))
	}

	// With force - should update
	preview, err = app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     true,
		Force:      true,
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	if len(preview.UpdatedKeys) != 1 {
		t.Errorf("With force: UpdatedKeys = %d, want 1", len(preview.UpdatedKeys))
	}
	if len(preview.SkippedKeys) != 0 {
		t.Errorf("With force: SkippedKeys = %d, want 0", len(preview.SkippedKeys))
	}

	// Actually perform the import with force and verify store is updated
	_, err = app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     false,
		Force:      true,
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	// Verify the value was updated in store
	updatedVal, err := app.Get("test-vault", "EXISTING_DIFFERENT")
	if err != nil {
		t.Fatalf("Failed to get updated value: %v", err)
	}
	if updatedVal != "new-value" {
		t.Errorf("Updated value = %q, want %q", updatedVal, "new-value")
	}
}

func TestImport_DryRun_NoChanges(t *testing.T) {
	app, ts, _ := setupTestApp(t)

	// Pre-populate
	app.Set("test-vault", "EXISTING", "value")
	originalCalls := len(ts.SaveCalls)

	imported := map[string]string{
		"NEW_KEY":  "new-value",
		"EXISTING": "different-value",
	}

	// Dry run - should NOT call Save
	_, err := app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     true,
		Force:      true, // Even with force, dry run shouldn't save
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	if len(ts.SaveCalls) != originalCalls {
		t.Errorf("Dry run made %d Save calls, expected %d", len(ts.SaveCalls), originalCalls)
	}

	// Verify existing value unchanged
	val, _ := app.Get("test-vault", "EXISTING")
	if val != "value" {
		t.Errorf("Dry run modified existing value to %q", val)
	}

	// Verify new key not added
	_, err = app.Get("test-vault", "NEW_KEY")
	if err != store.ErrNotFound {
		t.Errorf("Dry run created NEW_KEY, should not exist")
	}
}

func TestImport_WithFilters_AppliesCorrectly(t *testing.T) {
	app, ts, _ := setupTestApp(t)
	ts.SaveCalls = nil

	imported := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_PASSWORD": "secret",
		"API_KEY":     "key123",
		"API_SECRET":  "secret456",
		"OTHER":       "value",
	}

	// Import with include filter for DB_* and API_*, exclude *_PASSWORD
	_, err := app.Import("test-vault", importer.ImportOptions{
		SourcePath: createTempEnvFile(t, imported),
		Format:     "env",
		DryRun:     false,
		Include:    []string{"DB_*", "API_*"},
		Exclude:    []string{"*_PASSWORD", "*_SECRET"},
	})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}

	// Should have imported DB_HOST, DB_PORT (from DB_*) and API_KEY (from API_*)
	// Excluded: DB_PASSWORD (*_PASSWORD), API_SECRET (*_SECRET)
	expectedKeys := map[string]string{
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
		"API_KEY": "key123",
	}

	for key, expectedVal := range expectedKeys {
		val, err := app.Get("test-vault", key)
		if err != nil {
			t.Errorf("Expected key %q not found", key)
			continue
		}
		if val != expectedVal {
			t.Errorf("Key %q = %q, want %q", key, val, expectedVal)
		}
	}

	// These should NOT exist (filtered out by exclude patterns or not matching include)
	shouldNotExist := []string{"DB_PASSWORD", "API_SECRET", "OTHER"}
	for _, key := range shouldNotExist {
		_, err := app.Get("test-vault", key)
		if err != store.ErrNotFound {
			t.Errorf("Key %q should not exist but was found", key)
		}
	}
}

// Helper function to create a temp env file for testing
func createTempEnvFile(t *testing.T, data map[string]string) string {
	t.Helper()

	file, err := os.CreateTemp("", "test-*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	for key, value := range data {
		fmt.Fprintf(file, "%s=%s\n", key, value)
	}

	return file.Name()
}
