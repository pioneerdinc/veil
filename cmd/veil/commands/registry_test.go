package commands_test

import (
	"testing"

	"github.com/ossydotpy/veil/cmd/veil/commands"
)

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		wantErr bool
	}{
		{name: "version command exists", cmdName: "version", wantErr: false},
		{name: "init command exists", cmdName: "init", wantErr: false},
		{name: "quick command exists", cmdName: "quick", wantErr: false},
		{name: "set command exists", cmdName: "set", wantErr: false},
		{name: "get command exists", cmdName: "get", wantErr: false},
		{name: "delete command exists", cmdName: "delete", wantErr: false},
		{name: "list command exists", cmdName: "list", wantErr: false},
		{name: "vaults command exists", cmdName: "vaults", wantErr: false},
		{name: "search command exists", cmdName: "search", wantErr: false},
		{name: "export command exists", cmdName: "export", wantErr: false},
		{name: "generate command exists", cmdName: "generate", wantErr: false},
		{name: "reset command exists", cmdName: "reset", wantErr: false},
		{name: "unknown command", cmdName: "nonexistent", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := commands.Get(tt.cmdName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.cmdName, err, tt.wantErr)
				return
			}
			if !tt.wantErr && cmd == nil {
				t.Errorf("Get(%q) returned nil command", tt.cmdName)
			}
			if !tt.wantErr && cmd.Name() != tt.cmdName {
				t.Errorf("Get(%q) returned command with name %q", tt.cmdName, cmd.Name())
			}
		})
	}
}

func TestRegistry_All(t *testing.T) {
	all := commands.All()

	// Should have all 12 commands registered
	expectedCommands := []string{
		"version", "init", "quick", "set", "get", "delete",
		"list", "vaults", "search", "export", "generate", "reset",
	}

	if len(all) != len(expectedCommands) {
		t.Errorf("All() returned %d commands, want %d", len(all), len(expectedCommands))
	}

	for _, name := range expectedCommands {
		if _, ok := all[name]; !ok {
			t.Errorf("All() missing command %q", name)
		}
	}
}

func TestUsageError(t *testing.T) {
	err := &commands.UsageError{
		Command: "test",
		Usage:   "veil test <arg>",
	}

	expected := "Usage: veil test <arg>"
	if err.Error() != expected {
		t.Errorf("UsageError.Error() = %q, want %q", err.Error(), expected)
	}
}
