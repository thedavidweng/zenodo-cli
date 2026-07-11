package output

import (
	"bytes"
	"testing"
	"time"
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
	var out bytes.Buffer
	r := Renderer{
		Out:     &out,
		Err:     &out,
		JSON:    true,
		Compact: true,
	}
	meta := RuntimeMetaInput{
		Command:   "deposit list",
		Profile:   "default",
		RequestID: "req-123",
		StartedAt: time.Now(),
	}
	data := map[string]any{"id": "12345678", "title": "Test"}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out.Reset()
		if err := r.Success(meta, data, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompactJSONWithWarnings(b *testing.B) {
	var out bytes.Buffer
	r := Renderer{
		Out:     &out,
		Err:     &out,
		JSON:    true,
		Compact: true,
	}
	meta := RuntimeMetaInput{
		Command:   "deposit create",
		RequestID: "req-456",
		StartedAt: time.Now(),
	}
	data := map[string]any{"id": "12345678"}
	warnings := []string{"deprecated", "rate limit"}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out.Reset()
		if err := r.Success(meta, data, warnings); err != nil {
			b.Fatal(err)
		}
	}
}
