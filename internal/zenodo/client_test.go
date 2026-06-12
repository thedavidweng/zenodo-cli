package zenodo_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/thedavidweng/zenodo-cli/internal/testutil"
	"github.com/thedavidweng/zenodo-cli/internal/zenodo"
)

const testToken = "test-token-abc123"

func newClientAndServer(t *testing.T) (*testutil.FakeZenodo, *zenodo.Client) {
	t.Helper()
	fz := testutil.NewFakeZenodo(testToken)
	client := zenodo.NewClient(fz.URL(), testToken)
	t.Cleanup(func() { fz.Close() })
	return fz, client
}

// --- NewClient ---

func TestNewClient(t *testing.T) {
	c := zenodo.NewClient("https://zenodo.org/api", "mytoken")
	if c.BaseURL != "https://zenodo.org/api" {
		t.Fatalf("BaseURL = %q, want %q", c.BaseURL, "https://zenodo.org/api")
	}
	if c.Token != "mytoken" {
		t.Fatalf("Token = %q, want %q", c.Token, "mytoken")
	}
	if c.HTTPClient == nil {
		t.Fatal("HTTPClient should not be nil")
	}
	if c.Retries != 3 {
		t.Fatalf("Retries = %d, want 3", c.Retries)
	}
}

// --- CreateRecord ---

func TestCreateRecord(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "Test Record",
		Description:     "A test description",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
		Creators: []zenodo.Creator{{PersonOrOrg: &zenodo.PersonOrOrg{
			Type:       "personal",
			FamilyName: "Smith",
			GivenName:  "Alice",
		}}},
	}

	rec, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	if rec.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if rec.Metadata.Title != "Test Record" {
		t.Fatalf("Title = %q, want %q", rec.Metadata.Title, "Test Record")
	}
	if rec.Status != "draft" {
		t.Fatalf("Status = %q, want %q", rec.Status, "draft")
	}
}

// --- GetRecord ---

func TestGetRecord(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	// Create and publish a record via the fake server directly
	meta := zenodo.RecordMetadata{
		Title:           "Published Record",
		Description:     "desc",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	// Publish via the server
	_, err = client.PublishDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	// Now GET should work for published records
	rec, err := client.GetRecord(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}
	if rec.ID != created.ID {
		t.Fatalf("ID = %q, want %q", rec.ID, created.ID)
	}
	if rec.Status != "published" {
		t.Fatalf("Status = %q, want %q", rec.Status, "published")
	}
}

func TestGetRecordNotFound(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	_, err := client.GetRecord(ctx, "999999")
	if err == nil {
		t.Fatal("expected error for non-existent record")
	}
}

// --- GetDraft ---

func TestGetDraft(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "Draft Record",
		Description:     "draft desc",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "publication"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	draft, err := client.GetDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetDraft: %v", err)
	}
	if draft.Metadata.Title != "Draft Record" {
		t.Fatalf("Title = %q, want %q", draft.Metadata.Title, "Draft Record")
	}
	if draft.Status != "draft" {
		t.Fatalf("Status = %q, want %q", draft.Status, "draft")
	}
}

// --- UpdateDraft ---

func TestUpdateDraft(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "Original Title",
		Description:     "Original desc",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	updated, err := client.UpdateDraft(ctx, created.ID, zenodo.RecordMetadata{
		Title:           "Updated Title",
		Description:     "Updated desc",
		PublicationDate: "2026-06-01",
		ResourceType:    zenodo.ResourceType{Type: "publication"},
	})
	if err != nil {
		t.Fatalf("UpdateDraft: %v", err)
	}
	if updated.Metadata.Title != "Updated Title" {
		t.Fatalf("Title = %q, want %q", updated.Metadata.Title, "Updated Title")
	}
	if updated.Metadata.Description != "Updated desc" {
		t.Fatalf("Description = %q, want %q", updated.Metadata.Description, "Updated desc")
	}
}

// --- DeleteDraft ---

func TestDeleteDraft(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "To Delete",
		Description:     "will be deleted",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	err = client.DeleteDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteDraft: %v", err)
	}

	// Should now be gone
	_, err = client.GetDraft(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error after deleting draft")
	}
}

// --- PublishDraft ---

func TestPublishDraft(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "To Publish",
		Description:     "publish me",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	published, err := client.PublishDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}
	if published.Status != "published" {
		t.Fatalf("Status = %q, want %q", published.Status, "published")
	}
}

// --- ListRecords ---

func TestListRecords(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	// Create a couple of records
	for _, title := range []string{"Alpha", "Beta"} {
		_, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
			Title:           title,
			Description:     "desc",
			PublicationDate: "2026-01-01",
			ResourceType:    zenodo.ResourceType{Type: "dataset"},
		})
		if err != nil {
			t.Fatalf("CreateRecord(%s): %v", title, err)
		}
	}

	resp, err := client.ListRecords(ctx)
	if err != nil {
		t.Fatalf("ListRecords: %v", err)
	}
	if resp.Hits.Total < 2 {
		t.Fatalf("expected at least 2 records, got %d", resp.Hits.Total)
	}
}

// --- SearchRecords ---

func TestSearchRecords(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	_, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title:           "Quantum Computing Basics",
		Description:     "An introduction to quantum computing",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	_, err = client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title:           "Machine Learning 101",
		Description:     "Intro to ML",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "publication"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	resp, err := client.SearchRecords(ctx, "quantum")
	if err != nil {
		t.Fatalf("SearchRecords: %v", err)
	}
	if resp.Hits.Total != 1 {
		t.Fatalf("expected 1 result, got %d", resp.Hits.Total)
	}
	if resp.Hits.Hits[0].Metadata.Title != "Quantum Computing Basics" {
		t.Fatalf("Title = %q, want %q", resp.Hits.Hits[0].Metadata.Title, "Quantum Computing Basics")
	}
}

// --- NewVersion ---

func TestNewVersion(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "Versioned Record",
		Description:     "has versions",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	newVer, err := client.NewVersion(ctx, created.ID)
	if err != nil {
		t.Fatalf("NewVersion: %v", err)
	}
	if newVer.ID == created.ID {
		t.Fatalf("expected new ID, got same ID %q", newVer.ID)
	}
	if newVer.Status != "draft" {
		t.Fatalf("Status = %q, want %q", newVer.Status, "draft")
	}
	if newVer.Metadata.Title != "Versioned Record" {
		t.Fatalf("Title = %q, want %q", newVer.Metadata.Title, "Versioned Record")
	}
}

// --- UploadFile ---

func TestUploadFile(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	meta := zenodo.RecordMetadata{
		Title:           "File Upload Test",
		Description:     "test file upload",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "data.csv")
	if err := os.WriteFile(tmpFile, []byte("col1,col2\n1,2\n3,4\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err = client.UploadFile(ctx, created.ID, tmpFile)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
}

// --- Auth errors ---

func TestAuthFailure(t *testing.T) {
	fz := testutil.NewFakeZenodo("correct-token")
	defer fz.Close()

	// Use wrong token
	client := zenodo.NewClient(fz.URL(), "wrong-token")
	ctx := context.Background()

	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected auth error with wrong token")
	}
}

// --- DownloadRecord ---

func TestDownloadRecord(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	// Create a record with a file
	meta := zenodo.RecordMetadata{
		Title:           "Download Test",
		Description:     "test download",
		PublicationDate: "2026-01-01",
		ResourceType:    zenodo.ResourceType{Type: "dataset"},
	}
	created, err := client.CreateRecord(ctx, meta)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	// Upload a file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "upload.txt")
	content := []byte("hello world\n")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err = client.UploadFile(ctx, created.ID, tmpFile)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	// Publish the record
	_, err = client.PublishDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	// Download
	destDir := t.TempDir()
	err = client.DownloadRecord(ctx, created.ID, destDir)
	if err != nil {
		t.Fatalf("DownloadRecord: %v", err)
	}

	// Verify file was downloaded
	downloaded := filepath.Join(destDir, "upload.txt")
	data, err := os.ReadFile(downloaded)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("downloaded content = %q, want %q", data, content)
	}
}
