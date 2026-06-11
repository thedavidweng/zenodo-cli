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

func TestSearchHasPageAndSizeFlags(t *testing.T) {
	for _, flag := range []string{"page", "size"} {
		if searchCmd.Flags().Lookup(flag) == nil {
			t.Errorf("search missing --%s flag", flag)
		}
	}
}
