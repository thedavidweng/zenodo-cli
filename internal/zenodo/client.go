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
func (c *Client) CreateRecord(ctx context.Context, meta RecordMetadata) (*Record, error) {
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
func (c *Client) UpdateDraft(ctx context.Context, id string, meta RecordMetadata) (*Record, error) {
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
func (c *Client) UploadFile(ctx context.Context, id, filePath string) error {
	filename := filepath.Base(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Step 1: Init upload
	initBody := []map[string]any{{"key": filename}}
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/files", initBody, nil); err != nil {
		return fmt.Errorf("init upload: %w", err)
	}

	// Step 2: Upload content
	if err := c.doRaw(ctx, http.MethodPut, "/api/records/"+id+"/draft/files/"+filename+"/content", data, nil); err != nil {
		return fmt.Errorf("upload content: %w", err)
	}

	// Step 3: Commit
	if err := c.do(ctx, http.MethodPost, "/api/records/"+id+"/draft/files/"+filename+"/commit", nil, nil); err != nil {
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
func (c *Client) ResolveLatest(ctx context.Context, id string) (string, error) {
	rec, err := c.GetRecord(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get record: %w", err)
	}
	if rec.Links.Latest == "" {
		return id, nil
	}
	// Follow the latest link (may redirect) to get the actual record.
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
	url := fmt.Sprintf("%s/api/records/%s/files/%s/content", c.BaseURL, id, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request for %s: %w", key, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

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
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write file %s: %w", destPath, err)
	}
	return nil
}

// Do sends an HTTP request with JSON body and decodes JSON response into result.
// This is the public wrapper used by the api command.
func (c *Client) Do(ctx context.Context, method, path string, body any, result any) error {
	return c.do(ctx, method, path, body, result)
}

// do sends an HTTP request with JSON body and decodes JSON response into result.
// It handles auth, retries, and error parsing.
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.RequestInterval):
			}
		}

		url := c.BaseURL + path
		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Rebuild body reader for retries
		if body != nil {
			data, _ := json.Marshal(body)
			reqBody = bytes.NewReader(data)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		err = c.handleResponse(resp, result)
		if err == nil {
			return nil
		}

		// Don't retry client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return err
		}
		lastErr = err
	}

	return lastErr
}

// doRaw sends a request with raw bytes (used for file content upload).
func (c *Client) doRaw(ctx context.Context, method, path string, data []byte, result any) error {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.RequestInterval):
			}
		}

		url := c.BaseURL + path
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Content-Type", "application/octet-stream")

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

// handleResponse reads the response, handles errors, and decodes into result.
func (c *Client) handleResponse(resp *http.Response, result any) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to parse structured error
		var apiErr struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
			Errors  []struct {
				Field    string   `json:"field"`
				Messages []string `json:"messages"`
			} `json:"errors"`
		}
		if json.Unmarshal(bodyBytes, &apiErr) == nil {
			msg := apiErr.Message
			if len(apiErr.Errors) > 0 {
				for _, e := range apiErr.Errors {
					msg += fmt.Sprintf("; %s: %s", e.Field, strings.Join(e.Messages, ", "))
				}
			}
			if msg != "" {
				return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// 204 No Content - nothing to decode
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if result != nil {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
