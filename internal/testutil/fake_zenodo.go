package testutil

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"
)

// fakeRecord is the internal storage representation.
type fakeRecord struct {
	ID           string
	Title        string
	Desc         string
	Creators     []map[string]string
	PubDate      string
	ResType      string
	Files        []map[string]any
	FileContents map[string][]byte // key -> content bytes
	Status       string            // draft, published
	CreatedAt    string
	UpdatedAt    string
}

// FakeZenodo is an in-memory HTTP server that simulates the Zenodo InvenioRDM API.
type FakeZenodo struct {
	Server     *httptest.Server
	Token      string
	mu         sync.Mutex
	records    map[string]*fakeRecord
	nextID     int
	ValidToken string
}

// NewFakeZenodo starts a new fake Zenodo server. Requests must carry
// "Authorization: Bearer <token>" where token matches the configured value.
func NewFakeZenodo(token string) *FakeZenodo {
	fz := &FakeZenodo{
		Token:      token,
		ValidToken: token,
		records:    make(map[string]*fakeRecord),
		nextID:     1,
	}
	mux := http.NewServeMux()
	fz.registerRoutes(mux)
	fz.Server = httptest.NewServer(mux)
	return fz
}

// Close shuts down the test server.
func (fz *FakeZenodo) Close() {
	fz.Server.Close()
}

// URL returns the base URL of the fake server.
func (fz *FakeZenodo) URL() string {
	return fz.Server.URL
}

func (fz *FakeZenodo) registerRoutes(mux *http.ServeMux) {
	// List user records
	mux.HandleFunc("/api/user/records", func(w http.ResponseWriter, r *http.Request) {
		if !fz.checkAuth(w, r) {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, `{"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		fz.handleListUserRecords(w, r)
	})

	// Search records
	mux.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		if !fz.checkAuth(w, r) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			fz.handleSearchRecords(w, r)
		case http.MethodPost:
			fz.handleCreateRecord(w, r)
		default:
			http.Error(w, `{"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Draft operations: GET/PUT/DELETE /api/records/{id}/draft
	mux.HandleFunc("/api/records/", func(w http.ResponseWriter, r *http.Request) {
		if !fz.checkAuth(w, r) {
			return
		}
		fz.handleRecordSubpath(w, r)
	})
}

func (fz *FakeZenodo) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"message":"Authentication required"}`)
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if token != fz.ValidToken {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"message":"Invalid token"}`)
		return false
	}
	return true
}

func (fz *FakeZenodo) handleListUserRecords(w http.ResponseWriter, _ *http.Request) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	hits := make([]map[string]any, 0, len(fz.records))
	for _, rec := range fz.records {
		hits = append(hits, fz.recordToJSON(rec))
	}
	resp := map[string]any{
		"hits": map[string]any{
			"hits":  hits,
			"total": len(hits),
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func (fz *FakeZenodo) handleSearchRecords(w http.ResponseWriter, r *http.Request) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	q := strings.ToLower(r.URL.Query().Get("q"))
	var hits []map[string]any
	for _, rec := range fz.records {
		if q == "" || strings.Contains(strings.ToLower(rec.Title), q) || strings.Contains(strings.ToLower(rec.Desc), q) {
			hits = append(hits, fz.recordToJSON(rec))
		}
	}
	resp := map[string]any{
		"hits": map[string]any{
			"hits":  hits,
			"total": len(hits),
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func (fz *FakeZenodo) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"message":"cannot read body"}`, http.StatusBadRequest)
		return
	}

	var input map[string]any
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, `{"message":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	id := strconv.Itoa(fz.nextID)
	fz.nextID++

	now := time.Now().UTC().Format(time.RFC3339)
	rec := &fakeRecord{
		ID:           id,
		Status:       "draft",
		FileContents: make(map[string][]byte),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if meta, ok := input["metadata"].(map[string]any); ok {
		if t, ok := meta["title"].(string); ok {
			rec.Title = t
		}
		if d, ok := meta["description"].(string); ok {
			rec.Desc = d
		}
		if pd, ok := meta["publication_date"].(string); ok {
			rec.PubDate = pd
		}
		if rt, ok := meta["resource_type"].(map[string]any); ok {
			if t, ok := rt["type"].(string); ok {
				rec.ResType = t
			}
		}
		if creators, ok := meta["creators"].([]any); ok {
			for _, c := range creators {
				if cm, ok := c.(map[string]any); ok {
					creator := map[string]string{}
					if n, ok := cm["name"].(string); ok {
						creator["name"] = n
					}
					if t, ok := cm["type"].(string); ok {
						creator["type"] = t
					}
					rec.Creators = append(rec.Creators, creator)
				}
			}
		}
	}

	fz.records[id] = rec
	writeJSON(w, http.StatusCreated, fz.recordToJSON(rec))
}

func (fz *FakeZenodo) handleRecordSubpath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/records/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"message":"missing record ID"}`, http.StatusBadRequest)
		return
	}
	recordID := parts[0]

	if len(parts) == 1 {
		fz.requireMethod(w, r, http.MethodGet, func() { fz.handleGetRecord(w, recordID) })
		return
	}

	rest := parts[1]
	if h, ok := fz.matchRecordRoute(w, r, recordID, rest); ok {
		h()
		return
	}

	http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
}

func (fz *FakeZenodo) requireMethod(w http.ResponseWriter, r *http.Request, method string, handler func()) {
	if r.Method != method {
		http.Error(w, `{"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	handler()
}

func (fz *FakeZenodo) matchRecordRoute(w http.ResponseWriter, r *http.Request, recordID, rest string) (func(), bool) {
	// /api/records/{id}/draft
	if rest == "draft" {
		switch r.Method {
		case http.MethodGet:
			return func() { fz.handleGetDraft(w, recordID) }, true
		case http.MethodPut:
			return func() { fz.handleUpdateDraft(w, r, recordID) }, true
		case http.MethodDelete:
			return func() { fz.handleDeleteDraft(w, recordID) }, true
		default:
			http.Error(w, `{"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return nil, true
		}
	}

	// Exact match routes with method constraint
	type exactRoute struct {
		method  string
		path    string
		handler func()
	}
	for _, route := range []exactRoute{
		{http.MethodPost, "draft/actions/publish", func() { fz.handlePublishDraft(w, recordID) }},
		{http.MethodPost, "versions", func() { fz.handleNewVersion(w, recordID) }},
		{http.MethodPost, "draft/files", func() { fz.handleInitFileUpload(w, r, recordID) }},
		{http.MethodGet, "draft/files", func() { fz.handleListFiles(w, recordID, false) }},
		{http.MethodGet, "files", func() { fz.handleListFiles(w, recordID, true) }},
	} {
		if rest == route.path && r.Method == route.method {
			return route.handler, true
		}
	}

	// Pattern match routes with suffix
	type patternRoute struct {
		method  string
		prefix  string
		suffix  string
		handler func(filename string)
	}
	for _, route := range []patternRoute{
		{http.MethodPut, "draft/files/", "/content", func(fn string) { fz.handleUploadFileContent(w, r, recordID, fn) }},
		{http.MethodPost, "draft/files/", "/commit", func(fn string) { fz.handleCommitFile(w, recordID, fn) }},
		{http.MethodGet, "files/", "/content", func(fn string) { fz.handleDownloadFile(w, recordID, fn) }},
	} {
		if strings.HasPrefix(rest, route.prefix) && strings.HasSuffix(rest, route.suffix) && r.Method == route.method {
			filename := strings.TrimSuffix(strings.TrimPrefix(rest, route.prefix), route.suffix)
			return func() { route.handler(filename) }, true
		}
	}

	// Pattern match routes without suffix (e.g. draft/files/{filename})
	type barePatternRoute struct {
		method  string
		prefix  string
		handler func(filename string)
	}
	for _, route := range []barePatternRoute{
		{http.MethodGet, "draft/files/", func(fn string) { fz.handleGetFile(w, recordID, fn) }},
		{http.MethodDelete, "draft/files/", func(fn string) { fz.handleDeleteFile(w, recordID, fn) }},
	} {
		if strings.HasPrefix(rest, route.prefix) && r.Method == route.method {
			filename := strings.TrimPrefix(rest, route.prefix)
			if filename != "" && !strings.Contains(filename, "/") {
				return func() { route.handler(filename) }, true
			}
		}
	}

	return nil, false
}

func (fz *FakeZenodo) handleGetRecord(w http.ResponseWriter, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}
	if rec.Status != "published" {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, fz.recordToJSON(rec))
}

func (fz *FakeZenodo) handleGetDraft(w http.ResponseWriter, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Draft not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, fz.recordToJSON(rec))
}

func (fz *FakeZenodo) handleUpdateDraft(w http.ResponseWriter, r *http.Request, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Draft not found"}`, http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"message":"cannot read body"}`, http.StatusBadRequest)
		return
	}

	var input map[string]any
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, `{"message":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if meta, ok := input["metadata"].(map[string]any); ok {
		if t, ok := meta["title"].(string); ok {
			rec.Title = t
		}
		if d, ok := meta["description"].(string); ok {
			rec.Desc = d
		}
		if pd, ok := meta["publication_date"].(string); ok {
			rec.PubDate = pd
		}
		if rt, ok := meta["resource_type"].(map[string]any); ok {
			if t, ok := rt["type"].(string); ok {
				rec.ResType = t
			}
		}
		if creators, ok := meta["creators"].([]any); ok {
			rec.Creators = nil
			for _, c := range creators {
				if cm, ok := c.(map[string]any); ok {
					creator := map[string]string{}
					if n, ok := cm["name"].(string); ok {
						creator["name"] = n
					}
					if t, ok := cm["type"].(string); ok {
						creator["type"] = t
					}
					rec.Creators = append(rec.Creators, creator)
				}
			}
		}
	}
	rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	writeJSON(w, http.StatusOK, fz.recordToJSON(rec))
}

func (fz *FakeZenodo) handleDeleteDraft(w http.ResponseWriter, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	if _, ok := fz.records[id]; !ok {
		http.Error(w, `{"message":"Draft not found"}`, http.StatusNotFound)
		return
	}
	delete(fz.records, id)
	w.WriteHeader(http.StatusNoContent)
}

func (fz *FakeZenodo) handlePublishDraft(w http.ResponseWriter, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Draft not found"}`, http.StatusNotFound)
		return
	}
	rec.Status = "published"
	rec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	writeJSON(w, http.StatusOK, fz.recordToJSON(rec))
}

func (fz *FakeZenodo) handleNewVersion(w http.ResponseWriter, id string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	newID := strconv.Itoa(fz.nextID)
	fz.nextID++

	now := time.Now().UTC().Format(time.RFC3339)
	newRec := &fakeRecord{
		ID:           newID,
		Title:        rec.Title,
		Desc:         rec.Desc,
		Creators:     rec.Creators,
		PubDate:      rec.PubDate,
		ResType:      rec.ResType,
		Status:       "draft",
		FileContents: make(map[string][]byte),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	fz.records[newID] = newRec
	writeJSON(w, http.StatusCreated, fz.recordToJSON(newRec))
}

func (fz *FakeZenodo) handleInitFileUpload(w http.ResponseWriter, r *http.Request, recordID string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[recordID]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"message":"cannot read body"}`, http.StatusBadRequest)
		return
	}

	var input []map[string]any
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, `{"message":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	for _, f := range input {
		key, _ := f["key"].(string)
		if key != "" {
			rec.Files = append(rec.Files, map[string]any{
				"key":      key,
				"size":     0,
				"checksum": "",
				"status":   "pending",
			})
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"enabled": true,
		"entries": rec.Files,
	})
}

func (fz *FakeZenodo) handleUploadFileContent(w http.ResponseWriter, r *http.Request, recordID, filename string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[recordID]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"message":"cannot read body"}`, http.StatusBadRequest)
		return
	}

	size := int64(len(body))
	checksum := fmt.Sprintf("md5:%x", md5.Sum(body))

	// Store the actual file content
	rec.FileContents[filename] = body

	found := false
	for _, f := range rec.Files {
		if f["key"] == filename {
			f["size"] = size
			f["checksum"] = checksum
			f["status"] = "uploaded"
			found = true
			break
		}
	}
	if !found {
		rec.Files = append(rec.Files, map[string]any{
			"key":      filename,
			"size":     size,
			"checksum": checksum,
			"status":   "uploaded",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"key":      filename,
		"size":     size,
		"checksum": checksum,
		"status":   "uploaded",
	})
}

func (fz *FakeZenodo) handleCommitFile(w http.ResponseWriter, recordID, filename string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[recordID]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	for _, f := range rec.Files {
		if f["key"] == filename {
			f["status"] = "completed"
			writeJSON(w, http.StatusOK, f)
			return
		}
	}

	http.Error(w, `{"message":"File not found"}`, http.StatusNotFound)
}

func (fz *FakeZenodo) handleDownloadFile(w http.ResponseWriter, recordID, filename string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[recordID]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	content, ok := rec.FileContents[filename]
	if !ok {
		http.Error(w, `{"message":"File not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (fz *FakeZenodo) handleListFiles(w http.ResponseWriter, id string, published bool) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}
	if published && rec.Status != "published" {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	entries := make([]map[string]any, 0, len(rec.Files))
	for _, f := range rec.Files {
		entries = append(entries, map[string]any{
			"key":      f["key"],
			"size":     f["size"],
			"checksum": f["checksum"],
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (fz *FakeZenodo) handleGetFile(w http.ResponseWriter, id, filename string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	for _, f := range rec.Files {
		if f["key"] == filename {
			writeJSON(w, http.StatusOK, map[string]any{
				"key":      f["key"],
				"size":     f["size"],
				"checksum": f["checksum"],
			})
			return
		}
	}
	http.Error(w, `{"message":"File not found"}`, http.StatusNotFound)
}

func (fz *FakeZenodo) handleDeleteFile(w http.ResponseWriter, id, filename string) {
	fz.mu.Lock()
	defer fz.mu.Unlock()

	rec, ok := fz.records[id]
	if !ok {
		http.Error(w, `{"message":"Record not found"}`, http.StatusNotFound)
		return
	}

	for i, f := range rec.Files {
		if f["key"] == filename {
			rec.Files = append(rec.Files[:i], rec.Files[i+1:]...)
			delete(rec.FileContents, filename)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, `{"message":"File not found"}`, http.StatusNotFound)
}

func (fz *FakeZenodo) recordToJSON(rec *fakeRecord) map[string]any {
	files := make([]map[string]any, 0, len(rec.Files))
	for _, f := range rec.Files {
		files = append(files, map[string]any{
			"key":      f["key"],
			"size":     f["size"],
			"checksum": f["checksum"],
		})
	}

	return map[string]any{
		"id":      rec.ID,
		"status":  rec.Status,
		"created": rec.CreatedAt,
		"updated": rec.UpdatedAt,
		"metadata": map[string]any{
			"title":            rec.Title,
			"description":      rec.Desc,
			"creators":         rec.Creators,
			"publication_date": rec.PubDate,
			"resource_type":    map[string]any{"type": rec.ResType},
		},
		"files": files,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
