package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestNewFakeZenodo(t *testing.T) {
	fz := NewFakeZenodo("test-token")
	defer fz.Close()

	if fz.URL() == "" {
		t.Fatal("URL should not be empty")
	}
	if fz.Token != "test-token" {
		t.Fatalf("Token = %q, want %q", fz.Token, "test-token")
	}
}

func TestCheckAuthMissing(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	resp, err := http.Get(fz.URL() + "/api/user/records")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestCheckAuthWrongToken(t *testing.T) {
	fz := NewFakeZenodo("correct")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/user/records", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestCreateAndListRecords(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Test","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}

	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()

	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	// List user records
	req, _ = http.NewRequest("GET", fz.URL()+"/api/user/records", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}

	var listResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&listResp)
	_ = resp.Body.Close()
}

func TestSearchRecords(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	// Create a record first
	body := `{"metadata":{"title":"Quantum Physics","description":"quantum stuff","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Search
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records?q=quantum", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("search status = %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("DELETE", fz.URL()+"/api/user/records", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}

func TestSearchMethodNotAllowed(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("DELETE", fz.URL()+"/api/records", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}

func TestGetPublishedRecord(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	// Create
	body := `{"metadata":{"title":"Pub Test","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Publish
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Get published
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id, nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGetDraft(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Draft Test","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/draft", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get draft: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestUpdateDraft(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Original","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	update := `{"metadata":{"title":"Updated","description":"new desc","publication_date":"2026-06-01","resource_type":{"type":"publication"}}}`
	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/"+id+"/draft", bytes.NewBufferString(update))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDeleteDraft(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Delete Me","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("DELETE", fz.URL()+"/api/records/"+id+"/draft", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestNewVersion(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Versioned","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/versions", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("new version: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestFileUploadAndList(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	// Create record
	body := `{"metadata":{"title":"File Test","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Init upload
	initBody := `[{"key":"test.txt"}]`
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files", bytes.NewBufferString(initBody))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("init status = %d, want 201", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Upload content
	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/"+id+"/draft/files/test.txt/content", bytes.NewBufferString("hello"))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Commit
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files/test.txt/commit", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("commit status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// List draft files
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/draft/files", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Get file info
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/draft/files/test.txt", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get file: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get file status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Publish
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// List published files
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/files", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list published: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list published status = %d, want 200", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Download file
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/files/test.txt/content", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d, want 200", resp.StatusCode)
	}
	content, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(content) != "hello" {
		t.Fatalf("content = %q, want %q", content, "hello")
	}
}

func TestDeleteFile(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Del File","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Init + upload + commit
	initBody := `[{"key":"del.txt"}]`
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files", bytes.NewBufferString(initBody))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/"+id+"/draft/files/del.txt/content", bytes.NewBufferString("x"))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files/del.txt/commit", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Delete file
	req, _ = http.NewRequest("DELETE", fz.URL()+"/api/records/"+id+"/draft/files/del.txt", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestListVersions(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Versions","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Publish
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// List versions
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/versions", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReserveDOI(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"DOI","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/pids/doi", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSubmitToCommunity(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Submit","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	submitBody := `{"receiver":{"community":"test-community"}}`
	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/"+id+"/draft/review", bytes.NewBufferString(submitBody))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestImportFiles(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Import","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/files-import", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestListRequests(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/requests", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestListRequestsMethodNotAllowed(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/requests", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}

func TestResolveLatest(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	// Create and publish
	body := `{"metadata":{"title":"Latest","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Resolve latest (no newer version)
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/versions/latest", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("resolve latest: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRecordNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/records/999999", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDraftNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/records/999999/draft", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPublishDraftNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/records/999999/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestNewVersionNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/records/999999/versions", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestUpdateDraftNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("PUT", fz.URL()+"/api/records/999999/draft", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDeleteDraftNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("DELETE", fz.URL()+"/api/records/999999/draft", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestFileOperationsNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	// Init upload on non-existent record
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records/999999/draft/files", bytes.NewBufferString(`[{"key":"x"}]`))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("init: status = %d, want 404", resp.StatusCode)
	}

	// Upload content on non-existent record
	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/999999/draft/files/x/content", bytes.NewBufferString("x"))
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("upload: status = %d, want 404", resp.StatusCode)
	}

	// Commit on non-existent record
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/999999/draft/files/x/commit", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("commit: status = %d, want 404", resp.StatusCode)
	}

	// Download from non-existent record
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/999999/files/x/content", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("download: status = %d, want 404", resp.StatusCode)
	}

	// List files of non-existent record
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/999999/draft/files", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("list: status = %d, want 404", resp.StatusCode)
	}

	// Get file info of non-existent record
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/999999/draft/files/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get file: status = %d, want 404", resp.StatusCode)
	}

	// Delete file from non-existent record
	req, _ = http.NewRequest("DELETE", fz.URL()+"/api/records/999999/draft/files/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete file: status = %d, want 404", resp.StatusCode)
	}
}

func TestListVersionsNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/records/999999/versions", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestReserveDOINotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/records/999999/draft/pids/doi", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSubmitToCommunityNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("PUT", fz.URL()+"/api/records/999999/draft/review", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestImportFilesNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/records/999999/draft/actions/files-import", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPublishDraftThenGetPublished(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Pub","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Get draft before publish (should work)
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/draft", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Publish
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Get published record
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id, nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPublishedFilesListForDraft(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Draft Pub Files","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// List published files for a draft (should fail)
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/files", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (draft has no published files)", resp.StatusCode)
	}
}

func TestCommitFileNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Commit Missing","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Try to commit a file that was never uploaded
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files/nonexistent.txt/commit", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDownloadFileNotFound(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	body := `{"metadata":{"title":"Download Missing","description":"desc","publication_date":"2026-01-01","resource_type":{"type":"dataset"}}}`
	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&created)
	_ = resp.Body.Close()
	id := created["id"].(string)

	// Init + upload + publish
	initBody := `[{"key":"exists.txt"}]`
	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files", bytes.NewBufferString(initBody))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	req, _ = http.NewRequest("PUT", fz.URL()+"/api/records/"+id+"/draft/files/exists.txt/content", bytes.NewBufferString("x"))
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/files/exists.txt/commit", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	req, _ = http.NewRequest("POST", fz.URL()+"/api/records/"+id+"/draft/actions/publish", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, _ = http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	// Try to download a non-existent file
	req, _ = http.NewRequest("GET", fz.URL()+"/api/records/"+id+"/files/nonexistent.txt/content", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMissingRecordID(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("GET", fz.URL()+"/api/records/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestInvalidJSON(t *testing.T) {
	fz := NewFakeZenodo("tok")
	defer fz.Close()

	req, _ := http.NewRequest("POST", fz.URL()+"/api/records", bytes.NewBufferString("not-json"))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}
