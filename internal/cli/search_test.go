package cli

import (
	"testing"
)

func TestSearchCommandExists(t *testing.T) {
	if searchCmd.Name() != "search" {
		t.Errorf("Name = %q, want search", searchCmd.Name())
	}
	if searchCmd.RunE == nil {
		t.Error("search should have RunE")
	}
}
