package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestApiCommandExists(t *testing.T) {
	if apiCmd.Use != "api" {
		t.Errorf("Use = %q, want api", apiCmd.Use)
	}
}

func TestApiHasSubcommands(t *testing.T) {
	expected := []string{"get", "post"}
	names := make(map[string]bool)
	for _, c := range apiCmd.Commands() {
		names[c.Name()] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("api missing subcommand: %s", name)
		}
	}
}

func TestApiPostHasDataFlag(t *testing.T) {
	var postCmd *cobra.Command
	for _, c := range apiCmd.Commands() {
		if c.Name() == "post" {
			postCmd = c
			break
		}
	}
	if postCmd == nil {
		t.Fatal("post command not found")
	}
	if postCmd.Flags().Lookup("data") == nil {
		t.Error("post should have --data flag")
	}
}
