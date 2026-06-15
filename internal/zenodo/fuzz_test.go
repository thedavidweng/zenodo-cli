package zenodo

import (
	"encoding/json"
	"testing"
)

func FuzzRecordUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`{"id":"12345","metadata":{"title":"Test","description":"A test","creators":[{"person_or_org":{"type":"personal","family_name":"Smith","given_name":"Alice"}}],"publication_date":"2026-01-01","resource_type":{"type":"dataset"}},"files":[{"key":"data.csv","size":100,"checksum":"abc123"}],"created":"2026-01-01T00:00:00Z","updated":"2026-01-01T00:00:00Z","status":"draft","links":{"latest":"https://zenodo.org/api/records/12345"}}`))
	f.Add([]byte(`{"id":99999,"metadata":{"title":"Numeric ID","description":"ID as number","creators":[],"publication_date":"2026-01-01","resource_type":{"type":"publication"}},"status":"published","links":{"latest":""}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var r Record
		_ = json.Unmarshal(data, &r)
	})
}

func FuzzSearchResponseUnmarshal(f *testing.F) {
	f.Add([]byte(`{"hits":{"hits":[{"id":"1","metadata":{"title":"T","description":"D","creators":[],"publication_date":"2026-01-01","resource_type":{"type":"dataset"}},"status":"draft","links":{"latest":""}}],"total":1}}`))
	f.Add([]byte(`{"hits":{"hits":[],"total":0}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var sr SearchResponse
		_ = json.Unmarshal(data, &sr)
	})
}
