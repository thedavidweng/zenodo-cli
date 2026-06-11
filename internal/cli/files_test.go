package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestFilesCommandExists(t *testing.T) {
	if filesCmd.Use != "files" {
		t.Errorf("Use = %q, want files", filesCmd.Use)
	}
}

func TestFilesHasSubcommands(t *testing.T) {
	expected := []string{"upload", "list", "download"}
	names := make(map[string]bool)
	for _, c := range filesCmd.Commands() {
		names[c.Name()] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("files missing subcommand: %s", name)
		}
	}
}

func TestFilesUploadTakesIDAndFiles(t *testing.T) {
	var uploadCmd *cobra.Command
	for _, c := range filesCmd.Commands() {
		if c.Name() == "upload" {
			uploadCmd = c
			break
		}
	}
	if uploadCmd == nil {
		t.Fatal("upload command not found")
	}
	if uploadCmd.RunE == nil {
		t.Error("upload should have RunE")
	}
}

func TestFilesDownloadHasDestFlag(t *testing.T) {
	var downloadCmd *cobra.Command
	for _, c := range filesCmd.Commands() {
		if c.Name() == "download" {
			downloadCmd = c
			break
		}
	}
	if downloadCmd == nil {
		t.Fatal("download command not found")
	}
	if downloadCmd.Flags().Lookup("dest") == nil {
		t.Error("download should have --dest flag")
	}
}
