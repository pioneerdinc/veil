package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ossydotpy/veil/cmd/veil/commands"
)

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
