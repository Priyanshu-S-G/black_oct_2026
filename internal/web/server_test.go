package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testAssets() fstest.MapFS {
	return fstest.MapFS{
		"index.html":      {Data: []byte("<!doctype html><title>cockpit</title>")},
		"scripts/app.js":  {Data: []byte("export {};")},
		"styles/main.css": {Data: []byte("body{}")},
		"scripts/api.js":  {Data: []byte("export {};")},
	}
}

func do(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestHealth(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	rec := do(t, h, http.MethodGet, "/api/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body = %v, want status ok", body)
	}
}

func TestHealthRejectsOtherMethods(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	rec := do(t, h, http.MethodPost, "/api/health")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != "GET, HEAD" {
		t.Fatalf("Allow = %q", allow)
	}
}

func TestUnknownAPIPathIsJSON404(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	rec := do(t, h, http.MethodGet, "/api/game")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (gameplay endpoints are not implemented)", rec.Code)
	}
	var body errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.Success || body.ErrorCode != "NOT_FOUND" {
		t.Fatalf("body = %+v", body)
	}
}

func TestStaticServing(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})

	tests := []struct {
		target   string
		wantCode int
		wantType string
	}{
		{"/", http.StatusOK, "text/html"},
		{"/scripts/app.js", http.StatusOK, "text/javascript"},
		{"/styles/main.css", http.StatusOK, "text/css"},
		{"/missing.txt", http.StatusNotFound, ""},
		{"/scripts/", http.StatusNotFound, ""},
		{"/scripts", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		rec := do(t, h, http.MethodGet, tt.target)
		if rec.Code != tt.wantCode {
			t.Errorf("GET %s: status = %d, want %d", tt.target, rec.Code, tt.wantCode)
			continue
		}
		if tt.wantType != "" && !strings.HasPrefix(rec.Header().Get("Content-Type"), tt.wantType) {
			t.Errorf("GET %s: Content-Type = %q, want prefix %q", tt.target, rec.Header().Get("Content-Type"), tt.wantType)
		}
	}
}

func TestStaticRejectsWrites(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	if rec := do(t, h, http.MethodPost, "/"); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / status = %d, want 405", rec.Code)
	}
}

func TestStaticDoesNotEscapeAssetRoot(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	rec := do(t, h, http.MethodGet, "/../go.mod")
	if rec.Code == http.StatusOK {
		t.Fatalf("path traversal returned 200")
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := NewHandler(Options{Assets: testAssets()})
	rec := do(t, h, http.MethodGet, "/")
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options: nosniff")
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "default-src 'self'") {
		t.Error("missing strict Content-Security-Policy")
	}
}

func TestNoCacheOnlyInDevMode(t *testing.T) {
	dev := do(t, NewHandler(Options{Assets: testAssets(), NoCache: true}), http.MethodGet, "/scripts/app.js")
	if dev.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("dev Cache-Control = %q, want no-store", dev.Header().Get("Cache-Control"))
	}
	rel := do(t, NewHandler(Options{Assets: testAssets()}), http.MethodGet, "/scripts/app.js")
	if rel.Header().Get("Cache-Control") == "no-store" {
		t.Error("release mode should not force no-store on static assets")
	}
}
