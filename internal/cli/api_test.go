package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

// --- helper to find a subcommand of apiCmd ---
func apiSubcmd(name string) *cobra.Command {
	for _, c := range apiCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// --- Integration tests using FakeZenodo ---

func TestApiGetCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "API Get Test", Description: "api get", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, apiSubcmd("get"), []string{"/api/records/" + rec.ID + "/draft"}, nil, nil)
	if err != nil {
		t.Fatalf("api get: %v", err)
	}
	if !strings.Contains(out, "API Get Test") {
		t.Errorf("expected 'API Get Test' in output: %s", out)
	}
}

func TestApiGetCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "API Get JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, apiSubcmd("get"), []string{"/api/records/" + rec.ID + "/draft"}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("api get --json: %v", err)
	}
	if !strings.Contains(out, "API Get JSON") {
		t.Errorf("expected title in JSON output: %s", out)
	}
}

func TestApiGetWithoutLeadingSlash(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "No Slash", Description: "no slash", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, apiSubcmd("get"), []string{"api/records/" + rec.ID + "/draft"}, nil, nil)
	if err != nil {
		t.Fatalf("api get no slash: %v", err)
	}
	if !strings.Contains(out, "No Slash") {
		t.Errorf("expected 'No Slash' in output: %s", out)
	}
}

func TestApiPostCommand(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, apiSubcmd("post"), []string{"/api/records"}, map[string]bool{"confirm": true}, map[string]string{
		"data": `{"metadata":{"title":"API Created","description":"via api post","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`,
	})
	if err != nil {
		t.Fatalf("api post: %v", err)
	}
	if !strings.Contains(out, "API Created") {
		t.Errorf("expected 'API Created' in output: %s", out)
	}
}

func TestApiPostDryRun(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	out, err := runCmd(t, cfgPath, apiSubcmd("post"), []string{"/api/records"}, map[string]bool{"dry-run": true, "confirm": true}, map[string]string{
		"data": `{"metadata":{"title":"Dry Post"}}`,
	})
	if err != nil {
		t.Fatalf("api post dry run: %v", err)
	}
	if !strings.Contains(out, "Would POST") {
		t.Errorf("expected 'Would POST' in output: %s", out)
	}
}

func TestApiPostInvalidJSON(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, apiSubcmd("post"), []string{"/api/records"}, map[string]bool{"confirm": true}, map[string]string{
		"data": "not-valid-json{{{",
	})
	if err == nil {
		t.Error("expected error for invalid JSON data")
	}
}

func TestApiPutCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Original Title", Description: "original", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, apiSubcmd("put"), []string{"/api/records/" + rec.ID + "/draft"}, map[string]bool{"confirm": true}, map[string]string{
		"data": `{"metadata":{"title":"Updated via PUT","description":"updated","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`,
	})
	if err != nil {
		t.Fatalf("api put: %v", err)
	}
	if !strings.Contains(out, "Updated via PUT") {
		t.Errorf("expected 'Updated via PUT' in output: %s", out)
	}
}

func TestApiPutDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Put", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, apiSubcmd("put"), []string{"/api/records/" + rec.ID + "/draft"}, map[string]bool{"dry-run": true, "confirm": true}, map[string]string{
		"data": `{"metadata":{"title":"Dry Put"}}`,
	})
	if err != nil {
		t.Fatalf("api put dry run: %v", err)
	}
	if !strings.Contains(out, "Would PUT") {
		t.Errorf("expected 'Would PUT' in output: %s", out)
	}
}

func TestApiPutInvalidJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Bad Put", Description: "bad", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, apiSubcmd("put"), []string{"/api/records/" + rec.ID + "/draft"}, map[string]bool{"confirm": true}, map[string]string{
		"data": "invalid-json",
	})
	if err == nil {
		t.Error("expected error for invalid JSON data")
	}
}

func TestApiGetNotFound(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, apiSubcmd("get"), []string{"/api/records/999999"}, nil, nil)
	if err == nil {
		t.Error("expected error for non-existent record")
	}
}

func TestApiReadOnlyBlocksPost(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, apiSubcmd("post"), []string{"/api/records"}, map[string]bool{"confirm": true, "read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestApiReadOnlyBlocksPut(t *testing.T) {
	_, cfgPath := setupFakeZenodoTest(t)

	_, err := runCmd(t, cfgPath, apiSubcmd("put"), []string{"/api/records/1/draft"}, map[string]bool{"confirm": true, "read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestPrintJSON(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	app := &AppContext{}
	r := newRenderer(app, cmd)
	meta := metaInput(app, "test")
	ctx := &CmdContext{App: app, Cmd: cmd, R: r, Meta: meta}

	err := printJSON(ctx, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("printJSON: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, `"key"`) || !strings.Contains(output, `"value"`) {
		t.Errorf("output should contain key/value JSON, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}
}
