package flags

import (
	"testing"
)

func TestParseRunFlags_IncludeAndExclude(t *testing.T) {
	opts, err := ParseRunFlags([]string{"--include", "DB_*", "--exclude", "DB_PASSWORD"})
	if err != nil {
		t.Fatalf("ParseRunFlags error: %v", err)
	}
	if len(opts.Include) != 1 || opts.Include[0] != "DB_*" {
		t.Errorf("Include = %v, want [DB_*]", opts.Include)
	}
	if len(opts.Exclude) != 1 || opts.Exclude[0] != "DB_PASSWORD" {
		t.Errorf("Exclude = %v, want [DB_PASSWORD]", opts.Exclude)
	}
}

func TestParseRunFlags_UnknownFlag(t *testing.T) {
	_, err := ParseRunFlags([]string{"--unknown"})
	if err == nil {
		t.Error("Expected error for unknown flag")
	}
}

func TestParseRunFlags_MissingValue(t *testing.T) {
	_, err := ParseRunFlags([]string{"--include"})
	if err == nil {
		t.Error("Expected error for --include without value")
	}
}