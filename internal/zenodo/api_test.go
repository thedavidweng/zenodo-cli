package zenodo

import (
	"context"
	"testing"
)

// fakeAPI is a minimal test adapter that satisfies the API interface
// without any HTTP server. This proves the seam is real — a second
// adapter exists alongside the production Client.
type fakeAPI struct {
	records []Record
}

func (f *fakeAPI) ListRecords(ctx context.Context) (SearchResponse, error) {
	return SearchResponse{Hits: HitsList{Total: len(f.records), Hits: f.records}}, nil
}

func (f *fakeAPI) SearchRecords(ctx context.Context, query string) (SearchResponse, error) {
	return SearchResponse{}, nil
}

func (f *fakeAPI) CreateRecord(ctx context.Context, meta any) (*Record, error) {
	r := Record{ID: "fake-1", Metadata: RecordMetadata{Title: "fake"}}
	f.records = append(f.records, r)
	return &r, nil
}

func (f *fakeAPI) GetRecord(ctx context.Context, id string) (*Record, error) {
	return &Record{}, nil
}
func (f *fakeAPI) GetDraft(ctx context.Context, id string) (*Record, error) {
	return &Record{}, nil
}
func (f *fakeAPI) DeleteDraft(ctx context.Context, id string) error { return nil }
func (f *fakeAPI) PublishDraft(ctx context.Context, id string) (*Record, error) {
	return &Record{}, nil
}
func (f *fakeAPI) NewVersion(ctx context.Context, id string) (*Record, error) {
	return &Record{}, nil
}
func (f *fakeAPI) ListVersions(ctx context.Context, id string) (SearchResponse, error) {
	return SearchResponse{}, nil
}
func (f *fakeAPI) ReserveDOI(ctx context.Context, id string) (*Record, error) {
	return &Record{}, nil
}
func (f *fakeAPI) SubmitToCommunity(ctx context.Context, id, communityID string) error {
	return nil
}
func (f *fakeAPI) ListRequests(ctx context.Context, query string) (SearchResponse, error) {
	return SearchResponse{}, nil
}
func (f *fakeAPI) UploadFile(ctx context.Context, id, filePath string) error { return nil }
func (f *fakeAPI) ListFiles(ctx context.Context, id string) ([]RecordFile, error) {
	return nil, nil
}
func (f *fakeAPI) ListPublishedFiles(ctx context.Context, id string) ([]RecordFile, error) {
	return nil, nil
}
func (f *fakeAPI) DeleteFile(ctx context.Context, id, filename string) error { return nil }
func (f *fakeAPI) DownloadRecord(ctx context.Context, id, destdir string) error {
	return nil
}
func (f *fakeAPI) GetFile(ctx context.Context, id, filename string) (*RecordFile, error) {
	return &RecordFile{}, nil
}
func (f *fakeAPI) ImportFiles(ctx context.Context, id string) error { return nil }
func (f *fakeAPI) ResolveLatest(ctx context.Context, id string) (string, error) {
	return id, nil
}
func (f *fakeAPI) Do(ctx context.Context, method, path string, body, result any) error {
	return nil
}

// TestAPIInterfaceSeal verifies that *Client satisfies the API interface.
func TestAPIInterfaceSeal(t *testing.T) {
	var _ API = (*Client)(nil)
}

// TestFakeAPISatisfiesInterface verifies that a minimal fake can satisfy
// the API interface — the seam is real, not hypothetical.
func TestFakeAPISatisfiesInterface(t *testing.T) {
	var _ API = (*fakeAPI)(nil)

	f := &fakeAPI{}
	resp, err := f.ListRecords(context.Background())
	if err != nil {
		t.Fatalf("ListRecords: %v", err)
	}
	if resp.Hits.Total != 0 {
		t.Errorf("expected 0 records, got %d", resp.Hits.Total)
	}

	_, err = f.CreateRecord(context.Background(), nil)
	if err != nil {
		t.Fatalf("CreateRecord: %v", err)
	}

	resp, _ = f.ListRecords(context.Background())
	if resp.Hits.Total != 1 {
		t.Errorf("expected 1 record after create, got %d", resp.Hits.Total)
	}
}
