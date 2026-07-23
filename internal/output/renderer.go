package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/thedavidweng/zenodo-cli/internal/model"
)

type RuntimeMetaInput struct {
	Command   string
	Profile   string
	RequestID string
	StartedAt time.Time
}

type Renderer struct {
	Out     io.Writer
	Err     io.Writer
	JSON    bool
	Pretty  bool
	Compact bool
	Full    bool
	Quiet   bool
}

func (r *Renderer) Success(metaInput RuntimeMetaInput, data any, warnings []string) error {
	env := model.Envelope{
		OK:   true,
		Data: data,
		Meta: r.buildMeta(metaInput, warnings),
	}
	return r.writeJSON(env)
}

func (r *Renderer) Failure(metaInput RuntimeMetaInput, errBody model.ErrorBody) error {
	if !r.JSON {
		_, _ = fmt.Fprintf(r.Err, "Error [%s]: %s\n", errBody.Code, errBody.Message)
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

func (r *Renderer) Human(format string, args ...any) {
	if r.Quiet {
		return
	}
	_, _ = fmt.Fprintf(r.Out, format+"\n", args...)
}

// Render emits a JSON success envelope or calls human, depending on mode.
func (r *Renderer) Render(meta RuntimeMetaInput, data any, human func()) error {
	if r.JSON {
		return r.Success(meta, data, nil)
	}
	human()
	return nil
}

func (r *Renderer) buildMeta(input RuntimeMetaInput, warnings []string) model.Meta {
	return model.Meta{
		Command:       input.Command,
		Profile:       input.Profile,
		DurationMS:    time.Since(input.StartedAt).Milliseconds(),
		SchemaVersion: model.SchemaVersion,
		RequestID:     input.RequestID,
		Warnings:      warnings,
	}
}

func (r *Renderer) writeJSON(env model.Envelope) error {
	var raw any

	switch {
	case r.Full:
		warnings := env.Meta.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		raw = map[string]any{
			"ok":    env.OK,
			"data":  env.Data,
			"error": env.Error,
			"meta": map[string]any{
				"command":        env.Meta.Command,
				"profile":        env.Meta.Profile,
				"duration_ms":    env.Meta.DurationMS,
				"schema_version": env.Meta.SchemaVersion,
				"request_id":     env.Meta.RequestID,
				"warnings":       warnings,
			},
		}
	case r.Compact:
		if len(env.Meta.Warnings) == 0 {
			env.Meta.Warnings = nil
		}
		raw = env
	default:
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
	_, err = r.Out.Write(append(b, '\n'))
	return err
}
