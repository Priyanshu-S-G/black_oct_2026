package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func get(t *testing.T, cfg Config, target string) *httptest.ResponseRecorder {
	t.Helper()
	h, err := NewHandler(cfg)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

// chdir changes the working directory for one test (t.Chdir needs Go 1.24).
func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

// Release mode must serve the embedded assets without touching the working
// directory, so run it from a directory that has no webui folder.
func TestReleaseModeServesEmbeddedAssets(t *testing.T) {
	chdir(t, t.TempDir())

	cfg := DefaultConfig()
	for _, target := range []string{"/", "/styles/main.css", "/scripts/app.js", "/scripts/api.js"} {
		if rec := get(t, cfg, target); rec.Code != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", target, rec.Code)
		}
	}
	if body := get(t, cfg, "/").Body.String(); !strings.Contains(body, "Black October 2026") {
		t.Error("embedded index.html does not look like the cockpit shell")
	}
}

func TestDevModeServesFromDisk(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("from-disk"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Dev = true
	cfg.WebUIDir = dir

	rec := get(t, cfg, "/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "from-disk") {
		t.Fatalf("status = %d body = %q, want on-disk index", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("dev mode should send Cache-Control: no-store")
	}
}

func TestDevModeFailsClearlyWhenAssetsMissing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dev = true
	cfg.WebUIDir = filepath.Join(t.TempDir(), "nope")

	if _, err := NewHandler(cfg); err == nil {
		t.Fatal("expected an error when dev assets are missing")
	}
}
