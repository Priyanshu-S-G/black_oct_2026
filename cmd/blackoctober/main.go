// Command blackoctober runs the Black October 2026 simulator: one process that
// serves the HTTP API and the static cockpit.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"blackoctober/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg := app.DefaultConfig()
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr,
		"listen address; default is loopback only, use e.g. 0.0.0.0:8080 deliberately for LAN access")
	flag.BoolVar(&cfg.Dev, "dev", cfg.Dev,
		"development mode: serve cockpit assets from disk instead of the copy embedded in the binary")
	flag.StringVar(&cfg.WebUIDir, "webui-dir", cfg.WebUIDir,
		"directory containing the cockpit assets (only used with -dev)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Error("blackoctober stopped with an error", "err", err)
		return 1
	}
	return 0
}
