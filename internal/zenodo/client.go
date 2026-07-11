package zenodo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client communicates with a Zenodo InvenioRDM API.
type Client struct {
	BaseURL         string
	Token           string
	HTTPClient      *http.Client
	Retries         int
	RequestInterval time.Duration
}

// NewClient creates a Client with sensible defaults.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:         strings.TrimRight(baseURL, "/"),
		Token:           token,
		HTTPClient:      &http.Client{Timeout: 30 * time.Second},
		Retries:         3,
		RequestInterval: 500 * time.Millisecond,
	}
}

// ensureLeadingSlash ensures path starts with "/".
func ensureLeadingSlash(path string) string {
	if path == "" || path[0] != '/' {
		return "/" + path
	}
	return path
}

// ListRecords returns records owned by the authenticated user.
func (c *Client) ListRecords(ctx context.Context) (SearchResponse, error) {
	var resp SearchResponse
	err := c.do(ctx, http.MethodGet, "/api/user/records", nil, &resp)
	return resp, err
}

// SearchRecords searches public records with the given query string.
func (c *Client) SearchRecords(ctx context.Context, query string) (SearchResponse, error) {
	var resp SearchResponse
	err := c.do(ctx, http.MethodGet, "/api/records?q="+url.QueryEscape(query), nil, &resp)
	return resp, err
}

// CreateRecord creates a new draft record with the given metadata.
// meta can be a RecordMetadata struct or a raw map[string]any from JSON.
func (c *Client) CreateRecord(ctx context.Context, meta any) (*Record, error) {
	body := map[string]any{"metadata": meta}
	var rec Record
	if err := c.do(ctx, http.MethodPost, "/api/records", body, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetRecord retrieves a published record by ID.
func (c *Client) GetRecord(ctx context.Context, id string) (*Record, error) {
	var rec Record
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id, nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetDraft retrieves a draft record by ID.
func (c *Client) GetDraft(ctx context.Context, id string) (*Record, error) {
	var rec Record
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/draft", nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// UpdateDraft updates the metadata of a draft record.
// meta can be a RecordMetadata struct or a raw map[string]any from JSON.
func (c *Client) UpdateDraft(ctx context.Context, id string, meta any) (*Record, error) {
	body := map[string]any{"metadata": meta}
	var rec Record
	if err := c.do(ctx, http.MethodPut, "/api/records/"+id+"/draft", body, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// DeleteDraft deletes a draft record by ID.
func (c *Client) DeleteDraft(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/records/"+id+"/draft", nil, nil)
}

// PublishDraft publishes a draft record.
func (c *Client) PublishDraft(ctx context.Context, id string) (*Record, error) {
	var rec Record
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/actions/publish", nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// NewVersion creates a new draft version of an existing record.
func (c *Client) NewVersion(ctx context.Context, id string) (*Record, error) {
	var rec Record
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/versions", nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// UploadFile uploads a local file to a draft record.
// It performs the three-step process: init, upload content, commit.
// The file is streamed from disk on each retry attempt.
func (c *Client) UploadFile(ctx context.Context, id, filePath string) error {
	// Validate file exists and is readable before starting the upload process.
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("read file %s: %w", filePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("read file %s: is a directory", filePath)
	}

	filename := filepath.Base(filePath)

	// Step 1: Init upload
	initBody := []map[string]any{{"key": filename}}
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/files", initBody, nil); err != nil {
		return fmt.Errorf("init upload: %w", err)
	}

	// Step 2: Upload content (stream from disk, reopen on retry)
	openFile := func() (io.ReadCloser, error) {
		return os.Open(filePath)
	}
	if err := c.doStreaming(ctx, http.MethodPut, "/api/records/"+id+"/draft/files/"+url.PathEscape(filename)+"/content", openFile, nil); err != nil {
		return fmt.Errorf("upload content: %w", err)
	}

	// Step 3: Commit
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/files/"+url.PathEscape(filename)+"/commit", nil, nil); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	return nil
}

// ListFiles lists files in a draft record.
func (c *Client) ListFiles(ctx context.Context, id string) ([]RecordFile, error) {
	var resp struct {
		Entries []RecordFile `json:"entries"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/draft/files", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Entries, nil
}

// DeleteFile deletes a file from a draft record.
func (c *Client) DeleteFile(ctx context.Context, id, filename string) error {
	return c.do(ctx, http.MethodDelete, "/api/records/"+id+"/draft/files/"+filename, nil, nil)
}

// ListPublishedFiles lists files on a published record.
func (c *Client) ListPublishedFiles(ctx context.Context, id string) ([]RecordFile, error) {
	var resp struct {
		Entries []RecordFile `json:"entries"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/files", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Entries, nil
}

// ListVersions returns all versions of a record.
func (c *Client) ListVersions(ctx context.Context, id string) (SearchResponse, error) {
	var resp SearchResponse
	err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/versions", nil, &resp)
	return resp, err
}

// ReserveDOI reserves a DOI for a draft record.
func (c *Client) ReserveDOI(ctx context.Context, id string) (*Record, error) {
	var rec Record
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/pids/doi", nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// SubmitToCommunity submits a draft record for community review.
func (c *Client) SubmitToCommunity(ctx context.Context, id, communityID string) error {
	body := map[string]any{
		"receiver": map[string]any{
			"community": communityID,
		},
	}
	return c.do(ctx, http.MethodPut, "/api/records/"+id+"/draft/review", body, nil)
}

// GetFile retrieves metadata for a single file in a draft record.
func (c *Client) GetFile(ctx context.Context, id, filename string) (*RecordFile, error) {
	var f RecordFile
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/draft/files/"+filename, nil, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// ImportFiles imports files from the previous version into a new draft.
func (c *Client) ImportFiles(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/actions/files-import", nil, nil)
}

// ListRequests lists review/community-submission requests.
func (c *Client) ListRequests(ctx context.Context, query string) (SearchResponse, error) {
	var resp SearchResponse
	path := "/api/requests"
	if query != "" {
		path += "?q=" + url.QueryEscape(query)
	}
	err := c.do(ctx, http.MethodGet, path, nil, &resp)
	return resp, err
}

// ResolveLatest returns the ID of the latest version of a record.
// If the record has no newer version, the original ID is returned.
func (c *Client) ResolveLatest(ctx context.Context, id string) (string, error) {
	var latestRec Record
	if err := c.do(ctx, http.MethodGet, "/api/records/"+id+"/versions/latest", nil, &latestRec); err != nil {
		return "", fmt.Errorf("resolve latest: %w", err)
	}
	return latestRec.ID, nil
}

// DownloadRecord downloads all files from a published record into destdir.
func (c *Client) DownloadRecord(ctx context.Context, id, destdir string) error {
	rec, err := c.GetRecord(ctx, id)
	if err != nil {
		return fmt.Errorf("get record: %w", err)
	}

	if err := os.MkdirAll(destdir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	for _, f := range rec.Files {
		if err := c.downloadFile(ctx, id, destdir, f.Key); err != nil {
			return err
		}
	}

	return nil
}

// downloadFile downloads a single file from a record.
func (c *Client) downloadFile(ctx context.Context, id, destdir, key string) error {
	downloadURL := fmt.Sprintf("%s/api/records/%s/files/%s/content", c.BaseURL, id, url.PathEscape(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create request for %s: %w", key, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", key, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download %s: HTTP %d", key, resp.StatusCode)
	}

	destPath := filepath.Join(destdir, key)
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", destPath, err)
	}

	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(destPath) // clean up incomplete file
		return fmt.Errorf("write file %s: %w", destPath, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("close file %s: %w", destPath, closeErr)
	}
	return nil
}

// Do sends an HTTP request with JSON body and decodes JSON response into result.
// This is the public wrapper used by the api command.
func (c *Client) Do(ctx context.Context, method, path string, body any, result any) error {
	return c.do(ctx, method, ensureLeadingSlash(path), body, result)
}

// do sends an HTTP request with JSON body and decodes JSON response into result.
// It handles auth, retries, and error parsing.
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	// Marshal body once; reuse bytes for each retry attempt.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
	}

	return c.doWithRetry(ctx, method, path, bodyBytes, "application/json", result)
}

// doWithRetry sends an HTTP request with pre-marshaled body bytes, retrying
// on 5xx and network errors.
func (c *Client) doWithRetry(ctx context.Context, method, path string, bodyBytes []byte, contentType string, result any) error {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.RequestInterval):
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		c.setAuthHeaders(req, contentType)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		err = c.handleResponse(resp, result)
		if err == nil {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return err
		}
		lastErr = err
	}

	return lastErr
}

// doStreaming sends a request with a streaming body. The openBody function is
// called for each attempt to produce a fresh reader, so retries work correctly
// with file handles (which are closed by the HTTP transport after each request).
func (c *Client) doStreaming(ctx context.Context, method, path string, openBody func() (io.ReadCloser, error), result any) error {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.RequestInterval):
			}
		}

		body, err := openBody()
		if err != nil {
			return fmt.Errorf("open body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
		if err != nil {
			_ = body.Close()
			return fmt.Errorf("create request: %w", err)
		}
		c.setAuthHeaders(req, "application/octet-stream")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			_ = body.Close()
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		err = c.handleResponse(resp, result)
		if err == nil {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return err
		}
		lastErr = err
	}

	return lastErr
}

// setAuthHeaders sets the Authorization and Content-Type headers on a request.
func (c *Client) setAuthHeaders(req *http.Request, contentType string) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
}

// handleResponse reads the response, handles errors, and decodes into result.
func (c *Client) handleResponse(resp *http.Response, result any) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		// Read error body for error message construction.
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error response: %w", err)
		}
		return parseAPIError(resp.StatusCode, bodyBytes)
	}

	// 204 No Content — nothing to decode.
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// parseAPIError extracts a human-readable error message from an API error response.
func parseAPIError(statusCode int, bodyBytes []byte) error {
	var apiErr struct {
		Message string `json:"message"`
		Errors  []struct {
			Field    string   `json:"field"`
			Messages []string `json:"messages"`
		} `json:"errors"`
	}
	if json.Unmarshal(bodyBytes, &apiErr) == nil {
		msg := apiErr.Message
		for _, e := range apiErr.Errors {
			msg += fmt.Sprintf("; %s: %s", e.Field, strings.Join(e.Messages, ", "))
		}
		if msg != "" {
			return fmt.Errorf("API error (HTTP %d): %s", statusCode, msg)
		}
	}
	return fmt.Errorf("API error (HTTP %d): %s", statusCode, string(bodyBytes))
}
