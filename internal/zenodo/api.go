package zenodo

import "context"

// API is the interface Client satisfies, so tests can inject a fake.
type API interface {
	ListRecords(ctx context.Context) (SearchResponse, error)
	SearchRecords(ctx context.Context, query string) (SearchResponse, error)
	CreateRecord(ctx context.Context, meta any) (*Record, error)
	GetRecord(ctx context.Context, id string) (*Record, error)
	GetDraft(ctx context.Context, id string) (*Record, error)
	DeleteDraft(ctx context.Context, id string) error
	PublishDraft(ctx context.Context, id string) (*Record, error)
	NewVersion(ctx context.Context, id string) (*Record, error)
	ListVersions(ctx context.Context, id string) (SearchResponse, error)
	ReserveDOI(ctx context.Context, id string) (*Record, error)
	SubmitToCommunity(ctx context.Context, id, communityID string) error
	ListRequests(ctx context.Context, query string) (SearchResponse, error)
	UploadFile(ctx context.Context, id, filePath string) error
	ListFiles(ctx context.Context, id string) ([]RecordFile, error)
	ListPublishedFiles(ctx context.Context, id string) ([]RecordFile, error)
	DeleteFile(ctx context.Context, id, filename string) error
	DownloadRecord(ctx context.Context, id, destdir string) error
	GetFile(ctx context.Context, id, filename string) (*RecordFile, error)
	ImportFiles(ctx context.Context, id string) error
	ResolveLatest(ctx context.Context, id string) (string, error)
	Do(ctx context.Context, method, path string, body, result any) error
}
