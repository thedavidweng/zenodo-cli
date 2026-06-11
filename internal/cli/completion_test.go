package cli

import (
	"testing"
)

func TestCompletionCommandExists(t *testing.T) {
	if completionCmd.Use != "completion" {
		t.Errorf("Use = %q, want completion", completionCmd.Use)
	}
}

func TestCompletionHasSubcommands(t *testing.T) {
	expected := []string{"bash", "zsh", "fish", "powershell"}
	names := make(map[string]bool)
	for _, c := range completionCmd.Commands() {
		names[c.Name()] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("completion missing subcommand: %s", name)
		}
	}
}
