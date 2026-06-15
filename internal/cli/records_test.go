package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
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
		t.Error("show should validate args")
	}
}

// --- helper to find a subcommand of recordsCmd ---
func recordsSubcmd(name string) *cobra.Command {
	for _, c := range recordsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// --- Integration tests using FakeZenodo ---

func TestRecordsListCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	for _, title := range []string{"Alpha Record", "Beta Record"} {
		_, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
			Title: title, Description: "desc", PublicationDate: "2026-01-01",
			ResourceType: zenodo.ResourceType{Type: "dataset"},
		})
		if err != nil {
			t.Fatalf("seed %s: %v", title, err)
		}
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("list"), nil, nil, nil)
	if err != nil {
		t.Fatalf("records list: %v", err)
	}
	if !strings.Contains(out, "Alpha Record") {
		t.Errorf("output missing 'Alpha Record': %s", out)
	}
	if !strings.Contains(out, "Beta Record") {
		t.Errorf("output missing 'Beta Record': %s", out)
	}
}

func TestRecordsListCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	_, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "JSON Test", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("list"), nil, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records list --json: %v", err)
	}
	if !strings.Contains(out, "JSON Test") {
		t.Errorf("JSON output missing 'JSON Test': %s", out)
	}
}

func TestRecordsCreateWithTitle(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, nil, map[string]string{
		"title":       "My Dataset",
		"description": "A test",
	})
	if err != nil {
		t.Fatalf("records create: %v", err)
	}
	if !strings.Contains(out, "Created draft") {
		t.Errorf("expected 'Created draft' in output: %s", out)
	}
}

func TestRecordsCreateWithMetadataFile(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "meta.json")
	metaJSON := map[string]any{
		"metadata": map[string]any{
			"title":            "Metadata File Record",
			"description":      "From file",
			"publication_date": "2026-06-01",
			"resource_type":    map[string]any{"type": "publication"},
		},
	}
	data, _ := json.Marshal(metaJSON)
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, nil, map[string]string{
		"metadata": metaPath,
	})
	if err != nil {
		t.Fatalf("records create --metadata: %v", err)
	}
	if !strings.Contains(out, "Created draft") {
		t.Errorf("expected 'Created draft' in output: %s", out)
	}
}

func TestRecordsCreateMissingTitle(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	// resetFlags in runCmd clears any leftover title from previous tests.
	_, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, nil, nil)
	if err == nil {
		t.Error("expected error when creating without title")
	}
}

func TestRecordsCreateDryRun(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, map[string]bool{"dry-run": true}, map[string]string{
		"title": "Dry Run",
	})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if !strings.Contains(out, "Would create") {
		t.Errorf("expected 'Would create' in dry-run output: %s", out)
	}
}

func TestRecordsShowCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Show Me", Description: "visible", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("show"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("records show: %v", err)
	}
	if !strings.Contains(out, "Show Me") {
		t.Errorf("expected 'Show Me' in output: %s", out)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in output: %s", out)
	}
}

func TestRecordsShowCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Show JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("show"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records show --json: %v", err)
	}
	if !strings.Contains(out, "Show JSON") {
		t.Errorf("expected 'Show JSON' in output: %s", out)
	}
}

func TestRecordsShowPublishedRecord(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Published", Description: "pub", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("show"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("records show published: %v", err)
	}
	if !strings.Contains(out, "Published") {
		t.Errorf("expected 'Published' in output: %s", out)
	}
}

func TestRecordsShowNotFound(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, recordsSubcmd("show"), []string{"999999"}, nil, nil)
	if err == nil {
		t.Error("expected error for non-existent record")
	}
}

func TestRecordsDeleteCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "To Delete", Description: "bye", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("delete"), []string{rec.ID}, map[string]bool{"confirm": true}, nil)
	if err != nil {
		t.Fatalf("records delete --confirm: %v", err)
	}
	if !strings.Contains(out, "Deleted draft") {
		t.Errorf("expected 'Deleted draft' in output: %s", out)
	}
}

func TestRecordsDeleteDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Delete", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("delete"), []string{rec.ID}, map[string]bool{"dry-run": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("dry run delete: %v", err)
	}
	if !strings.Contains(out, "Would delete") {
		t.Errorf("expected 'Would delete' in output: %s", out)
	}
}

func TestRecordsDeleteJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Delete JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("delete"), []string{rec.ID}, map[string]bool{"json": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("records delete --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestRecordsPublishCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "To Publish", Description: "pub", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("publish"), []string{rec.ID}, map[string]bool{"confirm": true}, nil)
	if err != nil {
		t.Fatalf("records publish: %v", err)
	}
	if !strings.Contains(out, "Published") {
		t.Errorf("expected 'Published' in output: %s", out)
	}
}

func TestRecordsPublishDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Publish", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("publish"), []string{rec.ID}, map[string]bool{"dry-run": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("dry run publish: %v", err)
	}
	if !strings.Contains(out, "Would publish") {
		t.Errorf("expected 'Would publish' in output: %s", out)
	}
}

func TestRecordsPublishJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Publish JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("publish"), []string{rec.ID}, map[string]bool{"json": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("records publish --json: %v", err)
	}
	if !strings.Contains(out, "Publish JSON") {
		t.Errorf("expected title in JSON output: %s", out)
	}
}

func TestRecordsNewVersionCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Versioned", Description: "versions", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("new-version"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("records new-version: %v", err)
	}
	if !strings.Contains(out, "Created new version") {
		t.Errorf("expected 'Created new version' in output: %s", out)
	}
}

func TestRecordsNewVersionDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Version", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("new-version"), []string{rec.ID}, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("dry run new-version: %v", err)
	}
	if !strings.Contains(out, "Would create new version") {
		t.Errorf("expected 'Would create new version' in output: %s", out)
	}
}

func TestRecordsNewVersionJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Version JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("new-version"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records new-version --json: %v", err)
	}
	if !strings.Contains(out, "Version JSON") {
		t.Errorf("expected title in JSON output: %s", out)
	}
}

func TestRecordsVersionsCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Versioned List", Description: "list versions", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("versions"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("records versions: %v", err)
	}
	if !strings.Contains(out, "Versioned List") {
		t.Errorf("expected 'Versioned List' in output: %s", out)
	}
}

func TestRecordsVersionsCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Version JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("versions"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records versions --json: %v", err)
	}
	if !strings.Contains(out, "Version JSON") {
		t.Errorf("expected title in JSON output: %s", out)
	}
}

func TestRecordsReserveDOICommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "DOI Record", Description: "doi", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("reserve-doi"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("records reserve-doi: %v", err)
	}
	if !strings.Contains(out, "Reserved DOI") {
		t.Errorf("expected 'Reserved DOI' in output: %s", out)
	}
}

func TestRecordsReserveDOIJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "DOI JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("reserve-doi"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records reserve-doi --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestRecordsReserveDOIDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "DOI Dry", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("reserve-doi"), []string{rec.ID}, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("dry run reserve-doi: %v", err)
	}
	if !strings.Contains(out, "Would reserve DOI") {
		t.Errorf("expected 'Would reserve DOI' in output: %s", out)
	}
}

func TestRecordsSubmitCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Submit Record", Description: "submit", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("submit"), []string{rec.ID}, map[string]bool{"confirm": true}, map[string]string{
		"community": "my-community",
	})
	if err != nil {
		t.Fatalf("records submit: %v", err)
	}
	if !strings.Contains(out, "Submitted") {
		t.Errorf("expected 'Submitted' in output: %s", out)
	}
}

func TestRecordsSubmitJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Submit JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("submit"), []string{rec.ID}, map[string]bool{"json": true, "confirm": true}, map[string]string{
		"community": "test-community",
	})
	if err != nil {
		t.Fatalf("records submit --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestRecordsSubmitMissingCommunity(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "No Community", Description: "no", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("submit"), []string{rec.ID}, map[string]bool{"confirm": true}, nil)
	if err == nil {
		t.Error("expected error when --community is missing")
	}
}

func TestRecordsSubmitDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Submit", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, recordsSubcmd("submit"), []string{rec.ID}, map[string]bool{"dry-run": true, "confirm": true}, map[string]string{
		"community": "test-community",
	})
	if err != nil {
		t.Fatalf("dry run submit: %v", err)
	}
	if !strings.Contains(out, "Would submit") {
		t.Errorf("expected 'Would submit' in output: %s", out)
	}
}

func TestRecordsRequestsCommand(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("requests"), nil, nil, nil)
	if err != nil {
		t.Fatalf("records requests: %v", err)
	}
	if !strings.Contains(out, "Total:") {
		t.Errorf("expected 'Total:' in output: %s", out)
	}
}

func TestRecordsRequestsJSON(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("requests"), nil, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("records requests --json: %v", err)
	}
	if !strings.Contains(out, "total") {
		t.Errorf("expected 'total' in JSON output: %s", out)
	}
}

func TestRecordsRequestsQuery(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("requests"), []string{"test-query"}, nil, nil)
	if err != nil {
		t.Fatalf("records requests with query: %v", err)
	}
	if !strings.Contains(out, "Total:") {
		t.Errorf("expected 'Total:' in output: %s", out)
	}
}

func TestRecordsCreateJSON(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, map[string]bool{"json": true}, map[string]string{
		"title":       "JSON Create",
		"description": "json",
	})
	if err != nil {
		t.Fatalf("records create --json: %v", err)
	}
	if !strings.Contains(out, "JSON Create") {
		t.Errorf("JSON output missing title: %s", out)
	}
}

func TestRecordsCreateBadMetadataFile(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(metaPath, []byte("not-json{{{"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, nil, map[string]string{
		"metadata": metaPath,
	})
	if err == nil {
		t.Error("expected error for bad metadata JSON")
	}
}

func TestRecordsCreateMissingMetadataFile(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, recordsSubcmd("create"), nil, nil, map[string]string{
		"metadata": "/nonexistent/meta.json",
	})
	if err == nil {
		t.Error("expected error for missing metadata file")
	}
}

func TestRecordsDeleteWithoutConfirm(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "No Confirm", Description: "nc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("delete"), []string{rec.ID}, nil, nil)
	if err == nil {
		t.Error("expected error when --confirm is missing")
	}
}

func TestRecordsPublishWithoutConfirm(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "No Confirm Pub", Description: "nc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("publish"), []string{rec.ID}, nil, nil)
	if err == nil {
		t.Error("expected error when --confirm is missing")
	}
}

func TestRecordsReadOnlyBlocksDelete(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Read Only", Description: "ro", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("delete"), []string{rec.ID}, map[string]bool{"confirm": true, "read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestRecordsReadOnlyBlocksPublish(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Read Only Pub", Description: "ro", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("publish"), []string{rec.ID}, map[string]bool{"confirm": true, "read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestRecordsReadOnlyBlocksNewVersion(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Read Only NV", Description: "ro", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, recordsSubcmd("new-version"), []string{rec.ID}, map[string]bool{"read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}
