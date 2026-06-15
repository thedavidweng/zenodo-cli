package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

func BenchmarkRendererSuccessJSON(b *testing.B) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:  &out,
		Err:  &errBuf,
		JSON: true,
	}
	meta := RuntimeMetaInput{
		Command:   "deposit list",
		Profile:   "default",
		RequestID: "req-bench",
		StartedAt: time.Now(),
	}
	data := map[string]any{
		"id":     "12345678",
		"title":  "Benchmark Record",
		"status": "published",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out.Reset()
		if err := r.Success(meta, data, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRendererSuccessJSONWithWarnings(b *testing.B) {
	var out, errBuf bytes.Buffer
	r := Renderer{
		Out:  &out,
		Err:  &errBuf,
		JSON: true,
	}
	meta := RuntimeMetaInput{
		Command:   "deposit list",
		Profile:   "default",
		RequestID: "req-bench",
		StartedAt: time.Now(),
	}
	data := map[string]any{"id": "12345678"}
	warnings := []string{"deprecated field usage", "rate limit approaching threshold"}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out.Reset()
		if err := r.Success(meta, data, warnings); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompactJSON(b *testing.B) {
	env := model.Envelope{
		OK:   true,
		Data: map[string]any{"id": "12345678", "title": "Test"},
		Meta: model.Meta{
			Command:       "deposit list",
			Profile:       "default",
			DurationMS:    42,
			SchemaVersion: model.SchemaVersion,
			RequestID:     "req-123",
			Warnings:      nil,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = marshalCompact(env)
	}
}

func BenchmarkCompactJSONWithWarnings(b *testing.B) {
	env := model.Envelope{
		OK:   true,
		Data: map[string]any{"id": "12345678"},
		Meta: model.Meta{
			Command:       "deposit create",
			SchemaVersion: model.SchemaVersion,
			Warnings:      []string{"deprecated", "rate limit"},
		},
		Error: &model.ErrorBody{
			Code:    model.ErrValidationFailed,
			Message: "title is required",
			Details: map[string]any{"field": "title"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = marshalCompact(env)
	}
}
