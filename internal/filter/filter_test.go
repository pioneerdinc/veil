package filter

import (
	"testing"
)

func TestMatchPattern_WildcardCombinations(t *testing.T) {
	tests := []struct {
		key     string
		pattern string
		want    bool
	}{
		// Exact matches
		{"DB_HOST", "DB_HOST", true},
		{"DB_HOST", "DB_PORT", false},

		// Prefix wildcards: PATTERN* matches keys starting with PATTERN
		{"DB_HOST", "DB_*", true},
		{"DB_PORT", "DB_*", true},
		{"DB_PASSWORD", "DB_*", true},
		{"DB", "DB_*", false},           // Does NOT match (no underscore)
		{"DATABASE_URL", "DB_*", false}, // Does NOT match (no underscore after DB)
		{"MY_DB_HOST", "DB_*", false},

		// Suffix wildcards: *PATTERN matches keys ending with PATTERN
		{"API_KEY", "*_KEY", true},
		{"SECRET_KEY", "*_KEY", true},
		{"MY_API_KEY", "*_KEY", true}, // Matches (ends with _KEY)
		{"KEY", "*_KEY", false},       // Does NOT match (no underscore)
		{"API_KEY_SECRET", "*_KEY", false},

		// Both wildcards: middle wildcards NOT supported
		{"DB_PASSWORD", "DB_*_PASSWORD", false},
		{"DB_PASSWORD", "*PASSWORD", true}, // Suffix match works

		// No wildcards - partial match should fail
		{"DATABASE_URL", "DATA", false},
		{"DATABASE_URL", "URL", false},

		// Empty pattern edge case
		{"KEY", "", false},
		{"", "", true},
	}

	for _, tt := range tests {
		got := MatchPattern(tt.key, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.key, tt.pattern, got, tt.want)
		}
	}
}

func TestFilterSecrets_IncludeAndExcludeConflict(t *testing.T) {
	secrets := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_PASSWORD": "secret123",
		"API_KEY":     "key123",
		"API_SECRET":  "secret456",
	}

	tests := []struct {
		name    string
		include []string
		exclude []string
		want    map[string]string
	}{
		{
			name:    "include DB_*, exclude DB_PASSWORD",
			include: []string{"DB_*"},
			exclude: []string{"DB_PASSWORD"},
			want: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
			},
		},
		{
			name:    "include DB_* and API_*, exclude *_PASSWORD",
			include: []string{"DB_*", "API_*"},
			exclude: []string{"*_PASSWORD"},
			want: map[string]string{
				"DB_HOST":    "localhost",
				"DB_PORT":    "5432",
				"API_KEY":    "key123",
				"API_SECRET": "secret456",
			},
		},
		{
			name:    "include DB_*, exclude DB_* (everything excluded)",
			include: []string{"DB_*"},
			exclude: []string{"DB_*"},
			want:    map[string]string{},
		},
		{
			name:    "include all with *, exclude specific",
			include: []string{}, // empty = include all
			exclude: []string{"DB_PASSWORD", "API_SECRET"},
			want: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"API_KEY": "key123",
			},
		},
		{
			name:    "include exact, exclude prefix (exclude ignored)",
			include: []string{"DB_PASSWORD"},
			exclude: []string{"DB_*"},
			want:    map[string]string{}, // exclude wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterSecrets(secrets, tt.include, tt.exclude)

			if len(got) != len(tt.want) {
				t.Errorf("Got %d keys, want %d", len(got), len(tt.want))
			}

			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("Missing key %q", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("Key %q = %q, want %q", key, gotVal, wantVal)
				}
			}

			// Check for unexpected keys
			for key := range got {
				if _, ok := tt.want[key]; !ok {
					t.Errorf("Unexpected key %q in result", key)
				}
			}
		})
	}
}
