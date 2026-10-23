// Package game will own all authoritative game state and rules: state,
// typed actions, simulated time, scheduled events, the scenario fixture, the
// command parser, snapshots, and postmortem data.
//
// NOTHING IN THIS PACKAGE IMPLEMENTS GAMEPLAY YET. This file only fixes the
// shape of the boundary the rest of the program will code against, so the web
// layer can stay independent of rules and the engine independent of HTTP DTOs.
//
// Every name below is provisional. The exact public snapshot fields, run ID
// format, and state-version semantics are Phase 0 decisions that belong to the
// project owner (design plan, sec. 10 and sec. 14); do not treat these types as
// settling them.
package game

// RunID identifies one game run. Responses from a previous run must be
// ignored by the client once a new run starts (design plan, sec. 7.6).
type RunID string

// StateVersion is a monotonically increasing counter within one run, used for
// ordering snapshots. It is not a substitute for engine-side serialisation.
type StateVersion uint64

// Snapshot is the detached, public view of the game for the browser. It must
// contain only what the player is allowed to know, never expose internal
// mutable state, and always pair RunID and StateVersion with data from the
// same point in the simulation. Gameplay fields will be added in later tasks.
type Snapshot struct {
	RunID        RunID
	StateVersion StateVersion
}

// Action is a validated typed player intent. UI controls and the terminal both
// end up here, so there is exactly one rules path (design plan, sec. 7.7).
type Action interface {
	// Kind names the action, e.g. for the action history.
	Kind() string
}

// Result is the engine's outcome for an action or command. On a successful
// mutation Snapshot is generated after the state transition completes; on
// failure it is the unchanged current snapshot and no partial effects apply.
type Result struct {
	Success   bool
	Output    []string
	ErrorCode string
	Snapshot  Snapshot
}

// CommandParser turns a raw simulated-terminal line into a typed Action. It
// only parses. It must never execute anything and never reach a host shell or
// os/exec. How read-only queries (e.g. status) are represented is undecided.
type CommandParser interface {
	Parse(line string) (Action, error)
}

// Engine is the single authoritative game engine. All mutations go through it
// and are serialised; it never reads wall-clock time itself.
type Engine interface {
	// Snapshot returns a detached public view of the current state.
	Snapshot() Snapshot
	// Apply validates and applies one typed action.
	Apply(a Action) Result
	// ExecuteCommand parses a terminal line and applies it through the same
	// validation and state-transition path as Apply.
	ExecuteCommand(line string) Result
}
