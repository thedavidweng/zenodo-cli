package zenodo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// sampleRecordJSON is a representative record payload from the Zenodo API.
var sampleRecordJSON = []byte(`{
  "id": "12345678",
  "created": "2026-01-15T10:30:00.000000+00:00",
  "updated": "2026-01-16T14:20:00.000000+00:00",
  "status": "published",
  "metadata": {
    "title": "Benchmark Test Record: A Study of Performance Characteristics",
    "description": "This is a comprehensive test record with a realistic description that includes multiple sentences to simulate real-world payload sizes.",
    "publication_date": "2026-01-15",
    "resource_type": {"type": "dataset"},
    "creators": [
      {"person_or_org": {"type": "personal", "family_name": "Smith", "given_name": "Alice"}},
      {"person_or_org": {"type": "personal", "family_name": "Jones", "given_name": "Bob"}},
      {"person_or_org": {"type": "personal", "family_name": "Chen", "given_name": "Wei"}}
    ]
  },
  "files": [
    {"key": "data.csv", "size": 1048576, "checksum": "md5:abcdef1234567890abcdef1234567890"},
    {"key": "readme.txt", "size": 2048, "checksum": "md5:1234567890abcdef1234567890abcdef"}
  ],
  "links": {"latest": "https://zenodo.org/api/records/12345679"}
}`)

// sampleSearchResponseJSON is a realistic search/list response with multiple records.
var sampleSearchResponseJSON = func() []byte {
	recs := make([]string, 10)
	for i := range recs {
		recs[i] = fmt.Sprintf(`{
      "id": "%d",
      "created": "2026-01-15T10:30:00.000000+00:00",
      "updated": "2026-01-16T14:20:00.000000+00:00",
      "status": "published",
      "metadata": {
        "title": "Record %d",
        "description": "Description for record %d",
        "publication_date": "2026-01-15",
        "resource_type": {"type": "dataset"},
        "creators": [
          {"person_or_org": {"type": "personal", "family_name": "Author", "given_name": "Test"}}
        ]
      },
      "files": [
        {"key": "file%d.csv", "size": 1024, "checksum": "md5:abcdef1234567890abcdef1234567890"}
      ],
      "links": {"latest": "https://zenodo.org/api/records/%d"}
    }`, 10000000+i, i, i, i, 10000000+i)
	}
	payload := fmt.Sprintf(`{"hits": {"hits": [%s], "total": 10}}`, join(recs))
	return []byte(payload)
}()

func join(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

func BenchmarkRecordUnmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var rec Record
		if err := json.Unmarshal(sampleRecordJSON, &rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecordUnmarshalIDAsNumber(b *testing.B) {
	// Test the custom UnmarshalJSON path where ID arrives as a number.
	payload := []byte(`{
		"id": 12345678,
		"created": "2026-01-15T10:30:00Z",
		"updated": "2026-01-16T14:20:00Z",
		"status": "published",
		"metadata": {
			"title": "Numeric ID Record",
			"description": "desc",
			"publication_date": "2026-01-15",
			"resource_type": {"type": "dataset"}
		},
		"links": {"latest": ""}
	}`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var rec Record
		if err := json.Unmarshal(payload, &rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearchResponseUnmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var resp SearchResponse
		if err := json.Unmarshal(sampleSearchResponseJSON, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleResponse(b *testing.B) {
	// Create a mock httptest server that returns a JSON record.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleRecordJSON)
	}))
	b.Cleanup(srv.Close)

	c := NewClient(srv.URL, "bench-token")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.HTTPClient.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		var rec Record
		if err := c.handleResponse(resp, &rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleResponseSearch(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleSearchResponseJSON)
	}))
	b.Cleanup(srv.Close)

	c := NewClient(srv.URL, "bench-token")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.HTTPClient.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		var sr SearchResponse
		if err := c.handleResponse(resp, &sr); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandleResponseNoContent(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	b.Cleanup(srv.Close)

	c := NewClient(srv.URL, "bench-token")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.HTTPClient.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		if err := c.handleResponse(resp, nil); err != nil {
			b.Fatal(err)
		}
	}
}
