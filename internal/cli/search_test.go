package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

func TestSearchCommandExists(t *testing.T) {
	if searchCmd.Name() != "search" {
		t.Errorf("Name = %q, want search", searchCmd.Name())
	}
	if searchCmd.RunE == nil {
		t.Error("search should have RunE")
	}
}

// --- Integration tests using FakeZenodo ---

func TestSearchCommandWithQuery(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	_, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Quantum Computing Basics", Description: "An introduction to quantum computing",
		PublicationDate: "2026-01-01", ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err = client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Machine Learning 101", Description: "Intro to ML",
		PublicationDate: "2026-01-01", ResourceType: zenodo.ResourceType{Type: "publication"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, searchCmd, []string{"quantum"}, nil, nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(out, "Quantum Computing Basics") {
		t.Errorf("expected 'Quantum Computing Basics' in output: %s", out)
	}
	if strings.Contains(out, "Machine Learning 101") {
		t.Errorf("should not contain unrelated record: %s", out)
	}
}

func TestSearchCommandWithQueryJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	_, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Climate Change Report", Description: "Annual climate data",
		PublicationDate: "2026-01-01", ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, searchCmd, []string{"climate"}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("search --json: %v", err)
	}
	if !strings.Contains(out, "Climate Change Report") {
		t.Errorf("expected title in JSON output: %s", out)
	}
}

func TestSearchCommandNoResults(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, searchCmd, []string{"nonexistent-query-xyz"}, nil, nil)
	if err != nil {
		t.Fatalf("search no results: %v", err)
	}
	if !strings.Contains(out, "Total: 0") {
		t.Errorf("expected 'Total: 0' in output: %s", out)
	}
}
