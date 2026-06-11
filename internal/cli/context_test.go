package cli

import (
	"context"
	"testing"
	"time"
)

func TestAppContextHasSandbox(t *testing.T) {
	app := &AppContext{
		Sandbox: true,
	}
	if !app.Sandbox {
		t.Error("expected Sandbox=true")
	}
}

func TestAppContextSandboxDefaultFalse(t *testing.T) {
	app := &AppContext{}
	if app.Sandbox {
		t.Error("expected Sandbox=false by default")
	}
}

func TestWithAppContextAndGetAppContext(t *testing.T) {
	ctx := context.Background()
	app := &AppContext{
		ConfigFile: "/tmp/test.yaml",
		Profile:    "sandbox",
		Sandbox:    true,
		JSON:       true,
		Timeout:    10 * time.Second,
		Debug:      true,
	}

	ctx = WithAppContext(ctx, app)
	got := GetAppContext(ctx)
	if got == nil {
		t.Fatal("GetAppContext returned nil")
	}
	if got.ConfigFile != "/tmp/test.yaml" {
		t.Errorf("ConfigFile = %q, want %q", got.ConfigFile, "/tmp/test.yaml")
	}
	if got.Profile != "sandbox" {
		t.Errorf("Profile = %q, want sandbox", got.Profile)
	}
	if !got.Sandbox {
		t.Error("expected Sandbox=true")
	}
	if !got.JSON {
		t.Error("expected JSON=true")
	}
	if got.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", got.Timeout)
	}
	if !got.Debug {
		t.Error("expected Debug=true")
	}
}

func TestGetAppContextReturnsNilWhenMissing(t *testing.T) {
	ctx := context.Background()
	got := GetAppContext(ctx)
	if got != nil {
		t.Error("expected nil for missing AppContext")
	}
}

func TestGetAppContextReturnsNilForWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), appContextKey, "not-an-appcontext")
	got := GetAppContext(ctx)
	if got != nil {
		t.Error("expected nil for wrong type")
	}
}
