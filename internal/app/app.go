// Package app wires configuration and process lifecycle. It contains no game
// rules and no HTTP routing.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"blackoctober/internal/web"
	"blackoctober/webui"
)

// Config is the process configuration.
type Config struct {
	// Addr is the TCP listen address. The default is loopback-only so the game
	// is not exposed beyond the local machine unless the operator asks for it.
	Addr string
	// Dev serves cockpit assets from WebUIDir on disk (refresh the browser to
	// pick up edits). When false, the assets embedded in the binary are used and
	// the working directory is irrelevant.
	Dev bool
	// WebUIDir is the asset directory used when Dev is true.
	WebUIDir string
}

// DefaultConfig returns the default configuration: release mode on loopback.
func DefaultConfig() Config {
	return Config{
		Addr:     "127.0.0.1:8080",
		Dev:      false,
		WebUIDir: "webui",
	}
}

// NewHandler builds the HTTP handler for cfg, choosing the asset source
// explicitly from cfg.Dev.
func NewHandler(cfg Config) (http.Handler, error) {
	assets, err := assetFS(cfg)
	if err != nil {
		return nil, err
	}
	return web.NewHandler(web.Options{
		Assets:  assets,
		NoCache: cfg.Dev,
	}), nil
}

func assetFS(cfg Config) (fs.FS, error) {
	if !cfg.Dev {
		return webui.FS(), nil
	}
	index := filepath.Join(cfg.WebUIDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		return nil, fmt.Errorf("development assets not found at %q (run from the repository root or pass -webui-dir): %w", index, err)
	}
	return os.DirFS(cfg.WebUIDir), nil
}

// Run serves the game until ctx is cancelled, then shuts down gracefully.
func Run(ctx context.Context, cfg Config, log *slog.Logger) error {
	handler, err := NewHandler(cfg)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen on %q: %w", cfg.Addr, err)
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	mode := "release (embedded assets)"
	if cfg.Dev {
		mode = "development (assets from " + cfg.WebUIDir + ")"
	}
	log.Info("Black October 2026 listening", "url", "http://"+ln.Addr().String(), "mode", mode)

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}
