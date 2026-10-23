# Black October 2026: The Incident Command Simulator

A deterministic, satirical incident-response simulator. You are the on-call CTO
of the fictional OmniSign Enterprise.

> **Status: scaffold only.** The repository currently contains a runnable Go
> server, a health endpoint, and a static cockpit *shell*. **No gameplay is
> implemented**: no scenario, no simulation clock, no terminal commands, no
> financial rules, no incident logic. HUD values show `-` and every panel is
> labelled "Not implemented" on purpose.

The source of truth for what this will become is the revised game design plan
(`black_october_2026_revised_game_design_plan.txt`). See
[docs/architecture.md](docs/architecture.md) for how the code is organised.

## Requirements

- Go 1.22 or newer (the scaffold was built and checked with Go 1.22.2).
- A web browser.
- Nothing else: no Docker, database, Node.js, or network access.

## Run it

From the repository root.

**Development mode** (assets are read from `webui/` on disk, so editing
HTML/CSS/JS only needs a browser refresh):

```sh
go run ./cmd/blackoctober -dev
```

**Release mode** (assets are embedded in the binary; the working directory does
not matter):

```sh
go build -o blackoctober ./cmd/blackoctober
./blackoctober            # Windows: .\blackoctober.exe
```

Then open <http://127.0.0.1:8080>. You should see the cockpit shell with
"Engine API: reachable" in the top right. Check the API directly:

```sh
curl http://127.0.0.1:8080/api/health
# {"status":"ok"}
```

Stop with Ctrl+C (graceful shutdown).

### Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-addr` | `127.0.0.1:8080` | Listen address. Loopback only by default. |
| `-dev` | `false` | Serve assets from disk instead of the embedded copy. |
| `-webui-dir` | `webui` | Asset directory used with `-dev` (relative to the working directory). |

To reach the game from another machine on your LAN, pass an address explicitly,
for example `-addr 0.0.0.0:8080`. There is no authentication, so do not expose
it to the public internet.

## Checks

```sh
gofmt -l .        # prints nothing when formatted
go vet ./...
go test ./...
```

## What exists today

| Piece | State |
| --- | --- |
| Single Go process serving API + static files | Working |
| `GET /api/health` | Working (process health only) |
| Cockpit shell (HUD, four browser tabs, terminal pane, two modals) | Layout only; placeholders |
| Dev-from-disk and release-embedded asset modes | Working |
| `internal/game` contract types (run ID, state version, snapshot, action, result, parser, engine) | Interfaces only; nothing implements them |
| `/api/game`, `/api/game/new`, `/api/game/actions`, `/api/game/command` | **Not implemented** (return a JSON 404) |
| Scenarios, clock, finance, incidents, postmortem | **Not implemented** |

## Safety

The simulated OmniSign stack is game state, not real services. The terminal,
when it exists, will be parsed by a small Go interpreter and will never reach a
host shell.
