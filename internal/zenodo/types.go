package zenodo

import (
	"encoding/json"
	"fmt"
)

// Record represents a Zenodo InvenioRDM record.
type Record struct {
	ID        string         `json:"id"`
	Metadata  RecordMetadata `json:"metadata"`
	Files     []RecordFile   `json:"files,omitempty"`
	CreatedAt string         `json:"created"`
	UpdatedAt string         `json:"updated"`
	Status    string         `json:"status"`
	Links     RecordLinks    `json:"links"`
}

// RecordLinks holds HATEOAS links returned by the API.
type RecordLinks struct {
	Latest string `json:"latest"`
}

// UnmarshalJSON custom unmarshaler to handle ID as both string and number.
func (r *Record) UnmarshalJSON(data []byte) error {
	type Alias Record
	aux := &struct {
		ID any `json:"id"`
		*Alias
	}{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	switch v := aux.ID.(type) {
	case string:
		r.ID = v
	case float64:
		r.ID = fmt.Sprintf("%.0f", v)
	default:
		r.ID = fmt.Sprintf("%v", v)
	}
	return nil
}

// RecordMetadata holds the descriptive metadata for a record.
type RecordMetadata struct {
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Creators        []Creator    `json:"creators"`
	PublicationDate string       `json:"publication_date"`
	ResourceType    ResourceType `json:"resource_type"`
}

// Creator represents a record creator/contributor.
type Creator struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// ResourceType describes the type of resource.
type ResourceType struct {
	Type string `json:"type"`
}

// RecordFile represents a file attached to a record.
type RecordFile struct {
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

// SearchResponse is the top-level response from list/search endpoints.
type SearchResponse struct {
	Hits HitsList `json:"hits"`
}

// HitsList contains the actual record results.
type HitsList struct {
	Hits  []Record `json:"hits"`
	Total int      `json:"total"`
}
