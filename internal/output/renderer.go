package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

// RuntimeMetaInput holds the runtime metadata for building an Envelope.Meta.
type RuntimeMetaInput struct {
	Command   string
	Profile   string
	RequestID string
	StartedAt time.Time
}

// Renderer writes structured or human-readable output.
type Renderer struct {
	Out     io.Writer
	Err     io.Writer
	JSON    bool
	Pretty  bool
	Compact bool
	Full    bool
	Quiet   bool
}

// Success writes a successful envelope.
func (r *Renderer) Success(metaInput RuntimeMetaInput, data any, warnings []string) error {
	env := model.Envelope{
		OK:   true,
		Data: data,
		Meta: r.buildMeta(metaInput, warnings),
	}
	return r.writeJSON(env)
}

// Failure writes an error envelope and returns a CommandError.
func (r *Renderer) Failure(metaInput RuntimeMetaInput, errBody model.ErrorBody) error {
	if !r.JSON {
		fmt.Fprintf(r.Err, "Error [%s]: %s\n", errBody.Code, errBody.Message)
		return &model.CommandError{
			Code:    errBody.Code,
			Message: errBody.Message,
		}
	}
	env := model.Envelope{
		OK:    false,
		Error: &errBody,
		Meta:  r.buildMeta(metaInput, nil),
	}
	_ = r.writeJSON(env)
	return &model.CommandError{
		Code:    errBody.Code,
		Message: errBody.Message,
	}
}

// Human writes a human-readable message to Out. Suppressed when Quiet is true.
func (r *Renderer) Human(format string, args ...any) {
	if r.Quiet {
		return
	}
	fmt.Fprintf(r.Out, format+"\n", args...)
}

func (r *Renderer) buildMeta(input RuntimeMetaInput, warnings []string) model.Meta {
	duration := time.Since(input.StartedAt)
	return model.Meta{
		Command:       input.Command,
		Profile:       input.Profile,
		DurationMS:    duration.Milliseconds(),
		SchemaVersion: model.SchemaVersion,
		RequestID:     input.RequestID,
		Warnings:      warnings,
	}
}

// fullEnvelope mirrors model.Envelope but without omitempty, so all fields
// (including null/empty) are always present in the JSON output.
type fullEnvelope struct {
	OK    bool             `json:"ok"`
	Data  any              `json:"data"`
	Error *model.ErrorBody `json:"error"`
	Meta  fullMeta         `json:"meta"`
}

type fullMeta struct {
	Command       string   `json:"command"`
	Profile       string   `json:"profile"`
	DurationMS    int64    `json:"duration_ms"`
	SchemaVersion string   `json:"schema_version"`
	RequestID     string   `json:"request_id"`
	Warnings      []string `json:"warnings"`
}

func (r *Renderer) writeJSON(env model.Envelope) error {
	var raw any

	if r.Full {
		warnings := env.Meta.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		raw = fullEnvelope{
			OK:    env.OK,
			Data:  env.Data,
			Error: env.Error,
			Meta: fullMeta{
				Command:       env.Meta.Command,
				Profile:       env.Meta.Profile,
				DurationMS:    env.Meta.DurationMS,
				SchemaVersion: env.Meta.SchemaVersion,
				RequestID:     env.Meta.RequestID,
				Warnings:      warnings,
			},
		}
	} else if r.Compact {
		raw = marshalCompact(env)
	} else {
		raw = env
	}

	var b []byte
	var err error
	if r.Pretty {
		b, err = json.MarshalIndent(raw, "", "  ")
	} else {
		b, err = json.Marshal(raw)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(r.Out, string(b))
	return err
}

// marshalCompact removes zero-value fields from the envelope for compact output.
func marshalCompact(env model.Envelope) model.Envelope {
	if len(env.Meta.Warnings) == 0 {
		env.Meta.Warnings = nil
	}
	if env.Error != nil && env.Error.Details == nil {
		env.Error.Details = nil
	}
	return env
}
