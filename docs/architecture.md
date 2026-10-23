# Architecture notes

Status: scaffold. This documents what exists now and the boundaries the
approved design plan (sec. 7) asks us to keep. It does not add decisions of its
own beyond those listed under "Choices made in the scaffold".

## Shape

One Go process serves both the HTTP API and the static cockpit. No other
runtime process is required.

```
cmd/blackoctober/main.go   flags, signal handling, exit code
internal/app/app.go        Config, asset-source selection, listener, graceful shutdown
internal/web/server.go     routes, response DTOs, static handler, security headers
internal/game/contract.go  provisional engine-facing types (no behaviour)
webui/                     index.html, styles/, scripts/ (vanilla HTML/CSS/JS)
webui/embed.go             go:embed of the assets for release builds
```

Dependency direction: `cmd -> app -> web, webui`. `internal/game` is not yet
imported by anything. When the engine exists, `web` will depend on
`game.Engine` (an interface) and `app` will construct and inject it. `game`
must never import `web`.

## Assets: dev vs release

`//go:embed` cannot reach outside its own directory, so the embed directive
lives in `webui/embed.go` (package `webui`).

- **Release (default):** `webui.FS()` returns the embedded tree. The working
  directory is irrelevant (covered by a test that runs from an empty temp dir).
- **Development (`-dev`):** `os.DirFS(-webui-dir)` serves files from disk with
  `Cache-Control: no-store`. A browser refresh picks up edits; there is no hot
  reload. If `index.html` is missing from the directory, startup fails with an
  explicit message rather than silently falling back.

The mode is an explicit flag, never inferred.

## HTTP

| Route | Behaviour |
| --- | --- |
| `GET /api/health` | `200 {"status":"ok"}`. Process health only. |
| other methods on `/api/health` | `405` JSON error, `Allow: GET, HEAD` |
| any other `/api/...` | `404` JSON error (`error_code: NOT_FOUND`) |
| everything else | static assets, `GET`/`HEAD` only, no directory listings |

All responses carry `X-Content-Type-Options: nosniff`, `Referrer-Policy:
no-referrer`, and `Content-Security-Policy: default-src 'self'; frame-ancestors
'none'`. The cockpit therefore uses no inline scripts or styles and loads
nothing from the network, so it works offline.

The web package pins `.js` and `.css` MIME types at init, because some Windows
registries map `.js` to `text/plain`, which makes browsers reject ES modules.

## Engine boundary (provisional)

`internal/game/contract.go` declares, with no implementation:

- `RunID`, `StateVersion`: run identity and per-run ordering.
- `Snapshot`: detached public view; currently only carries the two fields above.
- `Action`: typed player intent. UI controls and terminal both end here.
- `Result`: outcome of an action or command, with a post-action snapshot.
- `CommandParser`: line to `Action`; parsing only, never execution.
- `Engine`: `Snapshot()`, `Apply(Action)`, `ExecuteCommand(line)`.

These express plan sec. 7.6-7.10 (one authoritative path, detached snapshots,
engine independent of HTTP DTOs, engine does not read wall-clock time). They
are placeholders for contract discussion, not settled design.

## Choices made in the scaffold (please confirm or overrule)

1. **Module path `blackoctober`**, no hosting domain. Rename if you want a
   `github.com/...` path.
2. **`go 1.22` in `go.mod`.** The plan fixes no Go version for the real build
   (1.25.1 is the *fictional* in-game runtime). 1.22 is what was available to
   verify with; raise it if you want newer language features.
3. **Default bind `127.0.0.1:8080`**, per the plan's "do not expose the HP 240
   G4 to the public internet by default". LAN use needs an explicit `-addr`.
4. **Health payload is `{"status":"ok"}` only**; no version/uptime/run fields,
   to avoid pre-empting DTO decisions.
5. **Provisional error shape** `{"success":false,"error_code":...,"message":...}`
   mirroring plan sec. 7.6/7.8. The final public DTO is a Phase 0 item.
6. **Cockpit tabs:** the four tabs listed in plan sec. 4.4 A-D only. The
   marketing status page (4.4 E) is left out because the plan allows either a
   tab or a feed item and that is undecided.
7. **HUD/Inventory/Ledger show no numbers.** The plan requires them to derive
   from engine state, so the shell shows placeholders rather than invented
   values.
8. **Engine reachability poll** hits `/api/health` every 5 s. The plan's
   ~1 s poll of `GET /api/game` is not implemented because that endpoint
   does not exist yet.

## Deliberately not here

Scenario 1, simulated clock and pacing, scheduled events, financial rules,
Executive Trust, incident declaration, command grammar, CAB logic, postmortem,
persistence, WebSockets/SSE, any database, Docker, or frontend framework. The
open decisions in design plan sec. 14 (time model, pause policy, trust-decay
cadence, SLA rates, runway vs. regulatory penalty, proxy recovery, scope) remain
open and were not guessed.
