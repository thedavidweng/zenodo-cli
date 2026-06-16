package zenodo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

// --- ListFiles ---

func TestListFiles(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "List Files Test", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "listme.txt")
	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(ctx, rec.ID, tmpFile); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	files, err := client.ListFiles(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Key != "listme.txt" {
		t.Fatalf("Key = %q, want %q", files[0].Key, "listme.txt")
	}
}

// --- DeleteFile ---

func TestDeleteFile(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Delete File Test", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "deleteme.txt")
	if err := os.WriteFile(tmpFile, []byte("bye"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(ctx, rec.ID, tmpFile); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	if err := client.DeleteFile(ctx, rec.ID, "deleteme.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	files, err := client.ListFiles(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files after delete, got %d", len(files))
	}
}

// --- ListPublishedFiles ---

func TestListPublishedFiles(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Published Files", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "pub.txt")
	if err := os.WriteFile(tmpFile, []byte("pub"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(ctx, rec.ID, tmpFile); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	_, err = client.PublishDraft(ctx, rec.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	files, err := client.ListPublishedFiles(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListPublishedFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Key != "pub.txt" {
		t.Fatalf("Key = %q, want %q", files[0].Key, "pub.txt")
	}
}

// --- GetFile ---

func TestGetFile(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Get File Test", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "info.txt")
	if err := os.WriteFile(tmpFile, []byte("info data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.UploadFile(ctx, rec.ID, tmpFile); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	f, err := client.GetFile(ctx, rec.ID, "info.txt")
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if f.Key != "info.txt" {
		t.Fatalf("Key = %q, want %q", f.Key, "info.txt")
	}
}

// --- ListVersions ---

func TestListVersions(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Versioned", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	_, err = client.PublishDraft(ctx, rec.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	resp, err := client.ListVersions(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if resp.Hits.Total < 1 {
		t.Fatalf("expected at least 1 version, got %d", resp.Hits.Total)
	}
}

// --- ReserveDOI ---

func TestReserveDOI(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "DOI Record", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	result, err := client.ReserveDOI(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ReserveDOI: %v", err)
	}
	if result.ID != rec.ID {
		t.Fatalf("ID = %q, want %q", result.ID, rec.ID)
	}
}

// --- SubmitToCommunity ---

func TestSubmitToCommunity(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Submit Record", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	err = client.SubmitToCommunity(ctx, rec.ID, "my-community")
	if err != nil {
		t.Fatalf("SubmitToCommunity: %v", err)
	}
}

// --- ImportFiles ---

func TestImportFiles(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Import Record", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	err = client.ImportFiles(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ImportFiles: %v", err)
	}
}

// --- ListRequests ---

func TestListRequests(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	_, err := client.ListRequests(ctx, "")
	if err != nil {
		t.Fatalf("ListRequests: %v", err)
	}

	_, err = client.ListRequests(ctx, "test-query")
	if err != nil {
		t.Fatalf("ListRequests with query: %v", err)
	}
}

// --- ResolveLatest ---

func TestResolveLatestNoNewerVersion(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "No Newer", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	_, err = client.PublishDraft(ctx, rec.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	resolved, err := client.ResolveLatest(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ResolveLatest: %v", err)
	}
	if resolved != rec.ID {
		t.Fatalf("resolved = %q, want %q (no newer version)", resolved, rec.ID)
	}
}

func TestResolveLatestWithNewerVersion(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "Original", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}
	_, err = client.PublishDraft(ctx, rec.ID)
	if err != nil {
		t.Fatalf("PublishDraft: %v", err)
	}

	newVer, err := client.NewVersion(ctx, rec.ID)
	if err != nil {
		t.Fatalf("NewVersion: %v", err)
	}
	_, err = client.PublishDraft(ctx, newVer.ID)
	if err != nil {
		t.Fatalf("PublishDraft new version: %v", err)
	}

	resolved, err := client.ResolveLatest(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ResolveLatest: %v", err)
	}
	if resolved != newVer.ID {
		t.Fatalf("resolved = %q, want %q", resolved, newVer.ID)
	}
}

func TestResolveLatestNotFound(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	_, err := client.ResolveLatest(ctx, "999999")
	if err == nil {
		t.Fatal("expected error for non-existent record")
	}
}

// --- Do (public wrapper) ---

func TestDoPublicWrapper(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	var resp map[string]any
	err := client.Do(ctx, "GET", "/api/user/records", nil, &resp)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
}

// --- handleResponse error paths ---

func TestHandleResponseStructuredError(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	// Try to get a non-existent draft to trigger a 404 with structured error
	_, err := client.GetDraft(ctx, "999999")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("error should contain 404, got: %v", err)
	}
}

func TestHandleResponseNoContent(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	rec, err := client.CreateRecord(ctx, zenodo.RecordMetadata{
		Title: "204 Test", Description: "desc", PublicationDate: "2026-01-01",
		ResourceType: zenodo.ResourceType{Type: "dataset"},
	})
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	// DeleteDraft returns 204 No Content
	err = client.DeleteDraft(ctx, rec.ID)
	if err != nil {
		t.Fatalf("DeleteDraft should succeed with 204: %v", err)
	}
}

// --- NewClient trailing slash ---

func TestNewClientTrailingSlash(t *testing.T) {
	c := zenodo.NewClient("https://zenodo.org/api/", "tok")
	if c.BaseURL != "https://zenodo.org/api" {
		t.Fatalf("BaseURL = %q, want %q", c.BaseURL, "https://zenodo.org/api")
	}
}

// --- Error path tests ---

func newErrorServer(t *testing.T, statusCode int, body string) *zenodo.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return zenodo.NewClient(srv.URL, "tok")
}

func TestPublishDraftError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.PublishDraft(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewVersionError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.NewVersion(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListFilesError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.ListFiles(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListPublishedFilesError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.ListPublishedFiles(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReserveDOIError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.ReserveDOI(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetFileError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.GetFile(ctx, "999", "x.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUploadFileError(t *testing.T) {
	client := newErrorServer(t, 500, `{"message":"internal error"}`)
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "err.txt")
	if err := os.WriteFile(tmpFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := client.UploadFile(ctx, "999", tmpFile)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDownloadRecordError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	err := client.DownloadRecord(ctx, "999", t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListVersionsError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	_, err := client.ListVersions(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSubmitToCommunityError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	err := client.SubmitToCommunity(ctx, "999", "community")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestImportFilesError(t *testing.T) {
	client := newErrorServer(t, 404, `{"message":"not found"}`)
	ctx := context.Background()
	err := client.ImportFiles(ctx, "999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListRequestsError(t *testing.T) {
	client := newErrorServer(t, 500, `{"message":"server error"}`)
	ctx := context.Background()
	_, err := client.ListRequests(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Test context cancellation during retry
func TestDoContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"message":"retry me"}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 5
	client.RequestInterval = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// Test retry exhaustion
func TestDoRetryExhaustion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"message":"always fail"}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 2
	client.RequestInterval = 1 * time.Millisecond

	ctx := context.Background()
	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
}

// Test non-JSON error body
func TestHandleResponseNonJSONError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `plain text error`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	ctx := context.Background()
	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("error should contain 500, got: %v", err)
	}
}

// Test structured API error with field errors
func TestHandleResponseStructuredFieldError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = fmt.Fprint(w, `{"message":"Validation error","status":400,"errors":[{"field":"metadata.title","messages":["Missing data for required field."]}]}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	ctx := context.Background()
	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("error should contain 400, got: %v", err)
	}
}

// Test 4xx error not retried
func TestDoNotRetry4xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(403)
		_, _ = fmt.Fprint(w, `{"message":"forbidden"}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 3
	client.RequestInterval = 1 * time.Millisecond

	ctx := context.Background()
	_, err := client.ListRecords(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry for 4xx), got %d", attempts)
	}
}

// Test doRaw with context cancellation
func TestDoRawContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 5
	client.RequestInterval = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// UploadFile calls doRaw internally
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "cancel.txt")
	_ = os.WriteFile(tmpFile, []byte("x"), 0644)

	err := client.UploadFile(ctx, "1", tmpFile)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// Test doRaw retry exhaustion
func TestDoRawRetryExhaustion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 2
	client.RequestInterval = 1 * time.Millisecond

	ctx := context.Background()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "retry.txt")
	_ = os.WriteFile(tmpFile, []byte("x"), 0644)

	err := client.UploadFile(ctx, "1", tmpFile)
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
}

// Test downloadFile error path
func TestDownloadFileHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	// We need to create the record metadata first via a different server
	// Actually, downloadFile is private and only called via DownloadRecord.
	// DownloadRecord first calls GetRecord. Let's make the server return
	// a valid record for GET /api/records/{id} but 404 for the file content.

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/files/") && strings.HasSuffix(r.URL.Path, "/content") {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{"id":"1","status":"published","metadata":{"title":"t"},"files":[{"key":"missing.txt","size":0}]}`)
	}))
	t.Cleanup(srv2.Close)

	client2 := zenodo.NewClient(srv2.URL, "tok")
	client2.Retries = 0

	err := client2.DownloadRecord(context.Background(), "1", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- UnmarshalJSON edge cases ---

func TestUnmarshalJSONBooleanID(t *testing.T) {
	// JSON boolean ID should hit the default case in UnmarshalJSON
	data := `{"id": true, "metadata": {"title": "test"}, "status": "draft"}`
	var rec zenodo.Record
	if err := json.Unmarshal([]byte(data), &rec); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if rec.ID != "true" {
		t.Errorf("ID = %q, want %q", rec.ID, "true")
	}
}

func TestUnmarshalJSONNullID(t *testing.T) {
	data := `{"id": null, "metadata": {"title": "test"}, "status": "draft"}`
	var rec zenodo.Record
	if err := json.Unmarshal([]byte(data), &rec); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if rec.ID != "<nil>" {
		t.Errorf("ID = %q, want %q", rec.ID, "<nil>")
	}
}

func TestUnmarshalJSONInvalidData(t *testing.T) {
	var rec zenodo.Record
	err := json.Unmarshal([]byte(`{invalid`), &rec)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- doRaw: xx should NOT retry ---

func TestDoRawNotRetry4xx(t *testing.T) {
	var contentAttempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/draft/files") {
			// Init succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/content") {
			// Content upload returns 403
			contentAttempts.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403)
			_, _ = fmt.Fprint(w, `{"message":"forbidden"}`)
			return
		}
		// Commit
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 3
	client.RequestInterval = 1 * time.Millisecond

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "no-retry.txt")
	_ = os.WriteFile(tmpFile, []byte("data"), 0644)

	err := client.UploadFile(context.Background(), "1", tmpFile)
	if err == nil {
		t.Fatal("expected error")
	}
	if n := contentAttempts.Load(); n != 1 {
		t.Fatalf("expected 1 content attempt (no retry for 4xx), got %d", n)
	}
}

// --- doRaw retry exhaustion on content upload ---

func TestDoRawContentRetryExhaustion(t *testing.T) {
	var contentAttempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/draft/files") {
			// Init succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		// Content upload always fails with 500
		contentAttempts.Add(1)
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 2
	client.RequestInterval = 1 * time.Millisecond

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "retry.txt")
	_ = os.WriteFile(tmpFile, []byte("data"), 0644)

	err := client.UploadFile(context.Background(), "1", tmpFile)
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
	if !strings.Contains(err.Error(), "upload content") {
		t.Fatalf("error should mention upload content, got: %v", err)
	}
	if n := contentAttempts.Load(); n != 3 {
		t.Fatalf("expected 3 content attempts (1 initial + 2 retries), got %d", n)
	}
}

// --- handleResponse: invalid JSON body with non-nil result ---

func TestHandleResponseInvalidJSONDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{not valid json}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	_, err := client.ListRecords(context.Background())
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("error should mention decode, got: %v", err)
	}
}

// --- DownloadRecord: destination directory creation failure ---

func TestDownloadRecordDirCreateError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, `{"id":"1","status":"published","metadata":{"title":"t"},"files":[{"key":"f.txt","size":0}]}`)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	// Create a file at the destination path so MkdirAll fails
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "blocked")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Try to use the file as a directory
	err := client.DownloadRecord(context.Background(), "1", filepath.Join(blockingFile, "sub"))
	if err == nil {
		t.Fatal("expected error when directory can't be created")
	}
	if !strings.Contains(err.Error(), "create dir") {
		t.Fatalf("error should mention create dir, got: %v", err)
	}
}

// --- CreateRecord: server error ---

func TestCreateRecordError(t *testing.T) {
	client := newErrorServer(t, 500, `{"message":"internal error"}`)
	ctx := context.Background()
	_, err := client.CreateRecord(ctx, zenodo.RecordMetadata{Title: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- UpdateDraft: server error ---

func TestUpdateDraftError(t *testing.T) {
	client := newErrorServer(t, 500, `{"message":"internal error"}`)
	ctx := context.Background()
	_, err := client.UpdateDraft(ctx, "999", zenodo.RecordMetadata{Title: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- UploadFile: init step fails ---

func TestUploadFileInitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/draft/files") {
			// Init fails
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_, _ = fmt.Fprint(w, `{"message":"init failed"}`)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "init-err.txt")
	_ = os.WriteFile(tmpFile, []byte("data"), 0644)

	err := client.UploadFile(context.Background(), "1", tmpFile)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "init upload") {
		t.Fatalf("error should mention init upload, got: %v", err)
	}
}

// --- UploadFile: commit step fails ---

func TestUploadFileCommitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/draft/files") {
			// Init succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/content") {
			// Content upload succeeds
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		if strings.Contains(r.URL.Path, "/commit") {
			// Commit fails
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_, _ = fmt.Fprint(w, `{"message":"commit failed"}`)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	client := zenodo.NewClient(srv.URL, "tok")
	client.Retries = 0

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "commit-err.txt")
	_ = os.WriteFile(tmpFile, []byte("data"), 0644)

	err := client.UploadFile(context.Background(), "1", tmpFile)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "commit file") {
		t.Fatalf("error should mention commit file, got: %v", err)
	}
}

// --- UploadFile: nonexistent file ---

func TestUploadFileNonexistentFile(t *testing.T) {
	_, client := newClientAndServer(t)
	ctx := context.Background()

	err := client.UploadFile(ctx, "1", "/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read file") {
		t.Fatalf("error should mention read file, got: %v", err)
	}
}
