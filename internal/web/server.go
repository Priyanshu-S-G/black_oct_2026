// Package web is the HTTP layer: routing, response DTOs, and static assets.
// Handlers stay thin; game rules never live here.
package web

import (
	"encoding/json"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

func init() {
	// Go consults the OS mime database first. On some Windows installs the
	// registry maps .js to text/plain, which makes browsers refuse ES modules.
	// Pin the types the cockpit depends on.
	_ = mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".css", "text/css; charset=utf-8")
}

// Options configures NewHandler.
type Options struct {
	// Assets is the cockpit asset tree, rooted so index.html is at the top.
	Assets fs.FS
	// NoCache marks static responses Cache-Control: no-store (development).
	NoCache bool
}

// NewHandler returns the HTTP handler for the API and the static cockpit.
func NewHandler(opts Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/", handleAPINotFound)
	mux.Handle("/", staticHandler(opts.Assets, opts.NoCache))
	return securityHeaders(mux)
}

// healthResponse is the GET /api/health DTO. It reports process health only;
// it says nothing about game state.
type healthResponse struct {
	Status string `json:"status"`
}

// errorResponse is the provisional structured error shape for the API. The
// final public DTO fields are an open Phase 0 decision (design plan, sec. 10).
type errorResponse struct {
	Success   bool   `json:"success"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			ErrorCode: "METHOD_NOT_ALLOWED",
			Message:   "use GET",
		})
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleAPINotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, errorResponse{
		ErrorCode: "NOT_FOUND",
		Message:   "no such API endpoint",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// staticHandler serves the asset tree read-only, never lists directories, and
// never serves anything outside fsys.
func staticHandler(fsys fs.FS, noCache bool) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if name := strings.TrimPrefix(path.Clean(r.URL.Path), "/"); name != "" {
			if info, err := fs.Stat(fsys, name); err == nil && info.IsDir() {
				http.NotFound(w, r)
				return
			}
		}
		if noCache {
			w.Header().Set("Cache-Control", "no-store")
		}
		files.ServeHTTP(w, r)
	})
}

// securityHeaders applies conservative defaults. The cockpit is same-origin
// only and loads nothing from the network, so the CSP can be strict.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
