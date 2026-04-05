package env

import (
	"testing"
)

func TestEscapeUnescape_RoundTrip(t *testing.T) {
	tests := []string{
		"simple",
		"with spaces",
		"with=equals",
		"with\"quotes",
		"with'ticks",
		"with\\backslash",
		"",
		"mixed \"quotes' and \\stuff",
		"special!@#$%^&*()",
		"unicode: 世界 🌍",
		"Uo9jwJW#HFH*fv4==", // base64 with double padding
		"single=pad=",       // single trailing =
		"triple===pad===",   // multiple trailing = (rare)
		"abc=def=ghi=jkl",   // multiple = in middle
	}

	for _, input := range tests {
		escaped := EscapeValue(input)
		unescaped := UnescapeValue(escaped)

		if unescaped != input {
			t.Errorf("RoundTrip failed for %q: escaped=%q, unescaped=%q", input, escaped, unescaped)
		}
	}
}

func TestUnescapeValue_DoubleQuotedWithEscapedQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"simple"`, "simple"},
		{`"with \"escaped\" quotes"`, `with "escaped" quotes`},
		{`"multiple \"escaped\" \"quotes\" here"`, `multiple "escaped" "quotes" here`},
		{`"just a \" in the middle"`, `just a " in the middle`},
		{`"\"leading escaped"`, `"leading escaped`},
		{`"trailing escaped\""`, `trailing escaped"`},
		// Note: UnescapeValue only handles \" -> ", not \\ -> \
		// So \\\" inside quotes becomes \\\" (backslashes unchanged, only \" -> ")
	}

	for _, tt := range tests {
		got := UnescapeValue(tt.input)
		if got != tt.expected {
			t.Errorf("UnescapeValue(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestUnescapeValue_SingleQuoted(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'simple'`, "simple"},
		{`'with "double" quotes'`, `with "double" quotes`}, // no escaping
		{`'with \backslash'`, `with \backslash`},           // literal backslash
		{`'with $variable'`, `with $variable`},             // no expansion
		{`'escaped \' quote'`, `escaped \' quote`},         // backslash is literal
	}

	for _, tt := range tests {
		got := UnescapeValue(tt.input)
		if got != tt.expected {
			t.Errorf("UnescapeValue(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParse_ComplexRealWorld(t *testing.T) {
	input := []byte(`# Database settings
DB_HOST=localhost
DB_PORT=5432

# API keys (quoted)
API_KEY="sk-abc123\"escaped\"xyz"

# Password with spaces (needs quoting)
DB_PASSWORD="my secret password"

# Empty value
EMPTY_KEY=

# Single quotes (literal)
SINGLE_QUOTED='no $expansion here'

# Unquoted special chars
UNQUOTED_SPECIAL=hello=world&foo

# Base64 secrets with padding
SECRET_1=Uo9jwJW#HFH*fv4==
SECRET_2=single=padding=

# Whitespace handling
  INDENTED_KEY=trimmed  
`)

	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	expected := map[string]string{
		"DB_HOST":          "localhost",
		"DB_PORT":          "5432",
		"API_KEY":          `sk-abc123"escaped"xyz`,
		"DB_PASSWORD":      "my secret password",
		"EMPTY_KEY":        "",
		"SINGLE_QUOTED":    "no $expansion here",
		"UNQUOTED_SPECIAL": "hello=world&foo",
		"SECRET_1":         "Uo9jwJW#HFH*fv4==",
		"SECRET_2":         "single=padding=",
		"INDENTED_KEY":     "trimmed",
	}

	if len(got) != len(expected) {
		t.Errorf("Got %d keys, want %d", len(got), len(expected))
	}

	for key, want := range expected {
		got, ok := got[key]
		if !ok {
			t.Errorf("Missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("Key %q = %q, want %q", key, got, want)
		}
	}
}
