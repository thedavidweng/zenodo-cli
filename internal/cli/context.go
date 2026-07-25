package cli

import (
	"context"
	"time"
)

type contextKey string

const appContextKey contextKey = "appContext"

type AppContext struct {
	ConfigFile string
	Profile    string
	Sandbox    bool
	JSON       bool
	Pretty     bool
	Compact    bool
	Full       bool
	Quiet      bool
	ReadOnly   bool
	DryRun     bool
	Confirm    bool
	Timeout    time.Duration
	Retries    int
	RequestID  string
	StartedAt  time.Time
}

func WithAppContext(ctx context.Context, app *AppContext) context.Context {
	return context.WithValue(ctx, appContextKey, app)
}

func GetAppContext(ctx context.Context) *AppContext {
	v, _ := ctx.Value(appContextKey).(*AppContext)
	return v
}
