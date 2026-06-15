package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
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

// --- helper to find a subcommand of filesCmd ---
func filesSubcmd(name string) *cobra.Command {
	for _, c := range filesCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// --- Integration tests using FakeZenodo ---

func TestFilesUploadCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Upload Target", Description: "upload test", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "data.csv")
	if err := os.WriteFile(tmpFile, []byte("a,b\n1,2\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("upload"), []string{rec.ID, tmpFile}, nil, nil)
	if err != nil {
		t.Fatalf("files upload: %v", err)
	}
	if !strings.Contains(out, "Uploaded") {
		t.Errorf("expected 'Uploaded' in output: %s", out)
	}
}

func TestFilesUploadMultipleFiles(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Multi Upload", Description: "multi", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	filePaths := make([]string, 0, 2)
	for _, name := range []string{"a.csv", "b.csv"} {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		filePaths = append(filePaths, p)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("upload"), append([]string{rec.ID}, filePaths...), nil, nil)
	if err != nil {
		t.Fatalf("files upload multiple: %v", err)
	}
	if !strings.Contains(out, "Uploaded") {
		t.Errorf("expected 'Uploaded' in output: %s", out)
	}
}

func TestFilesUploadDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Upload", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(tmpFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("upload"), []string{rec.ID, tmpFile}, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("dry run upload: %v", err)
	}
	if !strings.Contains(out, "Would upload") {
		t.Errorf("expected 'Would upload' in output: %s", out)
	}
}

func TestFilesUploadJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "JSON Upload", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "data.json")
	if err := os.WriteFile(tmpFile, []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("upload"), []string{rec.ID, tmpFile}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("files upload --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestFilesListCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "List Files", Description: "list", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "myfile.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("list"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("files list: %v", err)
	}
	if !strings.Contains(out, "myfile.txt") {
		t.Errorf("expected 'myfile.txt' in output: %s", out)
	}
}

func TestFilesListCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "List JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "data.bin")
	if err := os.WriteFile(tmpFile, []byte("binary-data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("list"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("files list --json: %v", err)
	}
	if !strings.Contains(out, "data.bin") {
		t.Errorf("expected 'data.bin' in JSON output: %s", out)
	}
}

func TestFilesListPublishedFiles(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Published Files", Description: "pub", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "pub.txt")
	if err := os.WriteFile(tmpFile, []byte("published"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("list"), []string{rec.ID}, nil, nil)
	if err != nil {
		t.Fatalf("files list published: %v", err)
	}
	if !strings.Contains(out, "pub.txt") {
		t.Errorf("expected 'pub.txt' in output: %s", out)
	}
}

func TestFilesDownloadCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Download Me", Description: "download", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	uploadFile := filepath.Join(tmpDir, "upload.txt")
	content := []byte("download content")
	if err := os.WriteFile(uploadFile, content, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, uploadFile); err != nil {
		t.Fatalf("upload: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	destDir := t.TempDir()
	cmd := filesSubcmd("download")
	out, err := runCmd(t, cfgPath, cmd, []string{rec.ID}, nil, map[string]string{"dest": destDir})
	if err != nil {
		t.Fatalf("files download: %v", err)
	}
	if !strings.Contains(out, "Downloaded") {
		t.Errorf("expected 'Downloaded' in output: %s", out)
	}

	downloaded := filepath.Join(destDir, "upload.txt")
	data, err := os.ReadFile(downloaded)
	if err != nil {
		t.Fatalf("read downloaded: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content = %q, want %q", data, content)
	}
}

func TestFilesDownloadDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Download", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("download"), []string{rec.ID}, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("dry run download: %v", err)
	}
	if !strings.Contains(out, "Would download") {
		t.Errorf("expected 'Would download' in output: %s", out)
	}
}

func TestFilesDownloadJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "JSON Download", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "f.txt")
	if err := os.WriteFile(tmpFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}
	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	destDir := t.TempDir()
	out, err := runCmd(t, cfgPath, filesSubcmd("download"), []string{rec.ID}, map[string]bool{"json": true}, map[string]string{"dest": destDir})
	if err != nil {
		t.Fatalf("files download --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestFilesDeleteCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Delete File", Description: "delete", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "todelete.txt")
	if err := os.WriteFile(tmpFile, []byte("delete me"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("delete"), []string{rec.ID, "todelete.txt"}, map[string]bool{"confirm": true}, nil)
	if err != nil {
		t.Fatalf("files delete: %v", err)
	}
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected 'Deleted' in output: %s", out)
	}
}

func TestFilesDeleteDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Delete File", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("delete"), []string{rec.ID, "file.txt"}, map[string]bool{"dry-run": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("dry run delete: %v", err)
	}
	if !strings.Contains(out, "Would delete") {
		t.Errorf("expected 'Would delete' in output: %s", out)
	}
}

func TestFilesDeleteJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Delete File JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "del.txt")
	if err := os.WriteFile(tmpFile, []byte("del"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("delete"), []string{rec.ID, "del.txt"}, map[string]bool{"json": true, "confirm": true}, nil)
	if err != nil {
		t.Fatalf("files delete --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}

func TestFilesInfoCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "File Info", Description: "info", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "info.txt")
	if err := os.WriteFile(tmpFile, []byte("info content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("info"), []string{rec.ID, "info.txt"}, nil, nil)
	if err != nil {
		t.Fatalf("files info: %v", err)
	}
	if !strings.Contains(out, "info.txt") {
		t.Errorf("expected 'info.txt' in output: %s", out)
	}
}

func TestFilesInfoCommandJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "File Info JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "info.json")
	if err := os.WriteFile(tmpFile, []byte("json data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("info"), []string{rec.ID, "info.json"}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("files info --json: %v", err)
	}
	if !strings.Contains(out, "info.json") {
		t.Errorf("expected 'info.json' in JSON output: %s", out)
	}
}

func TestFilesImportDryRun(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Dry Import", Description: "dry", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("import"), []string{rec.ID}, map[string]bool{"dry-run": true}, nil)
	if err != nil {
		t.Fatalf("dry run import: %v", err)
	}
	if !strings.Contains(out, "Would import") {
		t.Errorf("expected 'Would import' in output: %s", out)
	}
}

func TestFilesImportCommand(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Import Target", Description: "import", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	// FakeZenodo doesn't implement /draft/actions/files-import.
	_, _ = runCmd(t, cfgPath, filesSubcmd("import"), []string{rec.ID}, nil, nil)
}

func TestFilesUploadMissingFile(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Missing File", Description: "missing", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, filesSubcmd("upload"), []string{rec.ID, "/nonexistent/file.txt"}, nil, nil)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFilesReadOnlyBlocksUpload(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "RO Upload", Description: "ro", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "x.txt")
	if err := os.WriteFile(tmpFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err = runCmd(t, cfgPath, filesSubcmd("upload"), []string{rec.ID, tmpFile}, map[string]bool{"read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestFilesReadOnlyBlocksDelete(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "RO Delete", Description: "ro", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, filesSubcmd("delete"), []string{rec.ID, "file.txt"}, map[string]bool{"confirm": true, "read-only": true}, nil)
	if err == nil {
		t.Error("expected error in read-only mode")
	}
}

func TestFilesDeleteWithoutConfirm(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "No Confirm Delete", Description: "nc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err = runCmd(t, cfgPath, filesSubcmd("delete"), []string{rec.ID, "file.txt"}, nil, nil)
	if err == nil {
		t.Error("expected error when --confirm is missing")
	}
}

func TestFilesDownloadLatest(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Latest Test", Description: "latest", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "v1.txt")
	if err := os.WriteFile(tmpFile, []byte("version 1"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), rec.ID, tmpFile); err != nil {
		t.Fatalf("upload: %v", err)
	}

	_, err = client.PublishDraft(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	newVer, err := client.NewVersion(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("new version: %v", err)
	}

	tmpFile2 := filepath.Join(tmpDir, "v2.txt")
	if err := os.WriteFile(tmpFile2, []byte("version 2"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(context.Background(), newVer.ID, tmpFile2); err != nil {
		t.Fatalf("upload v2: %v", err)
	}

	_, err = client.PublishDraft(context.Background(), newVer.ID)
	if err != nil {
		t.Fatalf("publish v2: %v", err)
	}

	destDir := t.TempDir()
	cmd := filesSubcmd("download")
	out, err := runCmd(t, cfgPath, cmd, []string{rec.ID}, nil, map[string]string{"dest": destDir, "latest": "true"})
	if err != nil {
		t.Fatalf("files download --latest: %v", err)
	}
	if !strings.Contains(out, "Downloaded") {
		t.Errorf("expected 'Downloaded' in output: %s", out)
	}
}

func TestFilesImportJSON(t *testing.T) {
	fz, cfgPath := setupFakeZenodoTest(t)
	client := newTestClient(fz)

	rec, err := client.CreateRecord(context.Background(), zenodo.RecordMetadata{
		Title: "Import JSON", Description: "json", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	out, err := runCmd(t, cfgPath, filesSubcmd("import"), []string{rec.ID}, map[string]bool{"json": true}, nil)
	if err != nil {
		t.Fatalf("files import --json: %v", err)
	}
	if !strings.Contains(out, rec.ID) {
		t.Errorf("expected record ID in JSON output: %s", out)
	}
}
