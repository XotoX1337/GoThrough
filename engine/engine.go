// Package engine manages step state for an active walkthrough session.
package engine

import (
	"errors"

	"github.com/XotoX1337/GoThrough/config"
)

var (
	ErrAlreadyFirst = errors.New("already at first step")
	ErrAlreadyLast  = errors.New("already at last step")
	ErrOutOfRange   = errors.New("step index out of range")
)

// Persister records the user's position so it can be restored in a later
// session. It is bound to a single walkthrough; the engine supplies the current
// step index (0-based) and step ID after every change. Implemented by
// *progress.Handle.
type Persister interface {
	Save(index, stepID int) error
}

// Engine tracks which step the user is currently on.
type Engine struct {
	walkthrough *config.Walkthrough
	index       int
	store       Persister
}

// New creates an Engine starting at the first step.
func New(wt *config.Walkthrough) *Engine {
	return &Engine{walkthrough: wt, index: 0}
}

// UsePersister enables autosave: from now on every step change is written
// through p. Call Restore first if you want to resume a saved position, so the
// restore itself isn't redundantly written back.
func (e *Engine) UsePersister(p Persister) {
	e.store = p
}

// Restore positions the engine at a previously saved step. It prefers to match
// by step ID (resilient to steps being inserted or removed from the config) and
// falls back to the saved index, clamped to the valid range. Out-of-range or
// unmatched state leaves the engine at the first step.
func (e *Engine) Restore(index, stepID int) {
	for i := range e.walkthrough.Steps {
		if e.walkthrough.Steps[i].ID == stepID {
			e.index = i
			return
		}
	}
	if index < 0 {
		index = 0
	}
	if max := len(e.walkthrough.Steps) - 1; index > max {
		index = max
	}
	e.index = index
}

// persist writes the current position through the configured Persister, if any.
// Errors are returned to the caller; persistence is best-effort and a failure
// must not block navigation.
func (e *Engine) persist() error {
	if e.store == nil {
		return nil
	}
	return e.store.Save(e.index, e.walkthrough.Steps[e.index].ID)
}

// Current returns the active step. Never nil.
func (e *Engine) Current() *config.Step {
	return &e.walkthrough.Steps[e.index]
}

// Next advances to the next step. Returns ErrAlreadyLast if at the end.
//
// On a successful move the new position is persisted (best-effort: a save
// failure is ignored rather than reported as a navigation error).
func (e *Engine) Next() error {
	if e.index >= len(e.walkthrough.Steps)-1 {
		return ErrAlreadyLast
	}
	e.index++
	_ = e.persist()
	return nil
}

// Prev goes back one step. Returns ErrAlreadyFirst if at the beginning.
func (e *Engine) Prev() error {
	if e.index == 0 {
		return ErrAlreadyFirst
	}
	e.index--
	_ = e.persist()
	return nil
}

// Goto jumps to a 0-based step index. Returns ErrOutOfRange if out of bounds.
func (e *Engine) Goto(index int) error {
	if index < 0 || index >= len(e.walkthrough.Steps) {
		return ErrOutOfRange
	}
	e.index = index
	_ = e.persist()
	return nil
}

// Steps returns all steps in the walkthrough.
func (e *Engine) Steps() []config.Step {
	return e.walkthrough.Steps
}

// Game returns the game name from the loaded walkthrough.
func (e *Engine) Game() string { return e.walkthrough.Game }

// Title returns the walkthrough title.
func (e *Engine) Title() string { return e.walkthrough.Title }

// Progress returns 1-based current position and total step count.
func (e *Engine) Progress() (current, total int) {
	return e.index + 1, len(e.walkthrough.Steps)
}

// Done reports whether the user is on the final step.
func (e *Engine) Done() bool {
	return e.index == len(e.walkthrough.Steps)-1
}
