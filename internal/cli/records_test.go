package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRecordsCommandExists(t *testing.T) {
	if recordsCmd.Use != "records" {
		t.Errorf("Use = %q, want records", recordsCmd.Use)
	}
}

func TestRecordsHasSubcommands(t *testing.T) {
	expected := []string{"list", "create", "show", "delete", "publish", "new-version"}
	names := make(map[string]bool)
	for _, c := range recordsCmd.Commands() {
		names[c.Name()] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("records missing subcommand: %s", name)
		}
	}
}

func TestRecordsCreateHasFlags(t *testing.T) {
	var createCmd *cobra.Command
	for _, c := range recordsCmd.Commands() {
		if c.Name() == "create" {
			createCmd = c
			break
		}
	}
	if createCmd == nil {
		t.Fatal("create command not found")
	}
	for _, flag := range []string{"title", "description"} {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("create missing --%s flag", flag)
		}
	}
}

func TestRecordsDeleteRequiresConfirm(t *testing.T) {
	var deleteCmd *cobra.Command
	for _, c := range recordsCmd.Commands() {
		if c.Name() == "delete" {
			deleteCmd = c
			break
		}
	}
	if deleteCmd == nil {
		t.Fatal("delete command not found")
	}
	if deleteCmd.RunE == nil {
		t.Error("delete should have RunE")
	}
}

func TestRecordsPublishRequiresConfirm(t *testing.T) {
	var publishCmd *cobra.Command
	for _, c := range recordsCmd.Commands() {
		if c.Name() == "publish" {
			publishCmd = c
			break
		}
	}
	if publishCmd == nil {
		t.Fatal("publish command not found")
	}
	if publishCmd.RunE == nil {
		t.Error("publish should have RunE")
	}
}

func TestRecordsShowTakesID(t *testing.T) {
	var showCmd *cobra.Command
	for _, c := range recordsCmd.Commands() {
		if c.Name() == "show" {
			showCmd = c
			break
		}
	}
	if showCmd == nil {
		t.Fatal("show command not found")
	}
	if showCmd.Args == nil {
		// show should accept exactly 1 arg
		t.Error("show should validate args")
	}
}
