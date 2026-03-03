package commands_test

import (
	"bytes"
	"iter"
	"strings"
	"testing"

	"github.com/ossydotpy/veil/cmd/veil/commands"
	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/store"
)

// stubStore implements store.Store with minimal behavior needed by tests.
type stubStore struct {
	data map[string]string
}

func (s *stubStore) Save(vault, name, value string) error {
	if s.data == nil {
		s.data = make(map[string]string)
	}
	s.data[vault+"/"+name] = value
	return nil
}

func (s *stubStore) Get(vault, name string) (string, error) {
	if s.data == nil {
		return "", store.ErrNotFound
	}
	v, ok := s.data[vault+"/"+name]
	if !ok {
		return "", store.ErrNotFound
	}
	return v, nil
}

func (s *stubStore) Delete(vault, name string) error { return nil }
func (s *stubStore) List(vault string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {}
}
func (s *stubStore) ListVaults() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {}
}
func (s *stubStore) Search(pattern string) iter.Seq2[store.SecretRef, error] {
	return func(yield func(store.SecretRef, error) bool) {}
}
func (s *stubStore) Nuke() error  { return nil }
func (s *stubStore) Close() error { return nil }

func TestVersionCommand_Execute(t *testing.T) {
	cmd := commands.NewVersionCommand()

	// Test metadata
	if cmd.Name() != "version" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "version")
	}
	if cmd.NeedsDeps() {
		t.Error("NeedsDeps() = true, want false")
	}
	if cmd.NeedsMasterKey() {
		t.Error("NeedsMasterKey() = true, want false")
	}

	// Test execution
	var stdout bytes.Buffer
	deps := commands.Dependencies{
		Stdout: &stdout,
	}

	err := cmd.Execute(nil, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "veil version") {
		t.Errorf("Execute() output = %q, want to contain 'veil version'", output)
	}
}

func TestInitCommand_Metadata(t *testing.T) {
	cmd := commands.NewInitCommand()

	if cmd.Name() != "init" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "init")
	}
	if cmd.NeedsDeps() {
		t.Error("NeedsDeps() = true, want false")
	}
	if cmd.NeedsMasterKey() {
		t.Error("NeedsMasterKey() = true, want false")
	}
}

func TestInitCommand_Execute(t *testing.T) {
	cmd := commands.NewInitCommand()

	var stdout, stderr bytes.Buffer
	deps := commands.Dependencies{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Execute(nil, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()

	// Should contain the logo (Unicode art contains these characters)
	if !strings.Contains(output, "██") {
		t.Error("Execute() output should contain logo")
	}

	// Should contain a generated key (64 hex chars)
	if !strings.Contains(output, "MASTER_KEY") {
		t.Error("Execute() output should contain MASTER_KEY")
	}

	// Should contain safety warning
	if !strings.Contains(output, "SAVE THIS KEY") {
		t.Error("Execute() output should contain safety warning")
	}
}

func TestQuickCommand_Metadata(t *testing.T) {
	cmd := commands.NewQuickCommand()

	if cmd.Name() != "quick" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "quick")
	}
	if cmd.NeedsDeps() {
		t.Error("NeedsDeps() = true, want false")
	}
	if cmd.NeedsMasterKey() {
		t.Error("NeedsMasterKey() = true, want false")
	}
}

func TestQuickCommand_ExecuteDefault(t *testing.T) {
	cmd := commands.NewQuickCommand()

	var stdout bytes.Buffer
	deps := commands.Dependencies{
		Stdout: &stdout,
	}

	err := cmd.Execute(nil, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()
	// Default generates a password
	if !strings.Contains(output, "Generated password:") {
		t.Errorf("Execute() output = %q, want to contain 'Generated password:'", output)
	}
}

func TestQuickCommand_ExecuteWithLength(t *testing.T) {
	cmd := commands.NewQuickCommand()

	var stdout bytes.Buffer
	deps := commands.Dependencies{
		Stdout: &stdout,
	}

	err := cmd.Execute([]string{"--length", "8"}, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()
	// Should contain generated password output
	if !strings.Contains(output, "Generated password:") {
		t.Errorf("Execute() output = %q, want to contain 'Generated password:'", output)
	}
}

func TestQuickCommand_ExecuteMultiple(t *testing.T) {
	cmd := commands.NewQuickCommand()

	var stdout bytes.Buffer
	deps := commands.Dependencies{
		Stdout: &stdout,
	}

	err := cmd.Execute([]string{"--count", "3"}, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()
	// Should contain multiple numbered entries
	if !strings.Contains(output, "1.") || !strings.Contains(output, "2.") || !strings.Contains(output, "3.") {
		t.Errorf("Execute() output = %q, want to contain numbered entries 1., 2., 3.", output)
	}
}

// TestGetCommand_NoTrailingNewline ensures that the get command prints the secret
// exactly as stored without appending an extra newline character.
func TestGetCommand_NoTrailingNewline(t *testing.T) {
	cmd := commands.NewGetCommand()

	st := &stubStore{}
	engine, err := crypto.NewEngine(strings.Repeat("0", 64))
	if err != nil {
		t.Fatalf("failed to create crypto engine: %v", err)
	}
	app := app.New(st, engine)

	secret := "Line1\nLine2"
	if err := app.Set("vault", "key", secret); err != nil {
		t.Fatalf("app.Set error: %v", err)
	}

	var stdout bytes.Buffer
	deps := commands.Dependencies{
		App:    app,
		Stdout: &stdout,
	}

	err = cmd.Execute([]string{"vault", "key"}, deps)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	got := stdout.String()
	if got != secret {
		t.Errorf("unexpected output; got %q, want %q", got, secret)
	}
}
