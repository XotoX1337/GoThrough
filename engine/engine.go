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

// Engine tracks which step the user is currently on.
type Engine struct {
	walkthrough *config.Walkthrough
	index       int
}

// New creates an Engine starting at the first step.
func New(wt *config.Walkthrough) *Engine {
	return &Engine{walkthrough: wt, index: 0}
}

// Current returns the active step. Never nil.
func (e *Engine) Current() *config.Step {
	return &e.walkthrough.Steps[e.index]
}

// Next advances to the next step. Returns ErrAlreadyLast if at the end.
func (e *Engine) Next() error {
	if e.index >= len(e.walkthrough.Steps)-1 {
		return ErrAlreadyLast
	}
	e.index++
	return nil
}

// Prev goes back one step. Returns ErrAlreadyFirst if at the beginning.
func (e *Engine) Prev() error {
	if e.index == 0 {
		return ErrAlreadyFirst
	}
	e.index--
	return nil
}

// Goto jumps to a 0-based step index. Returns ErrOutOfRange if out of bounds.
func (e *Engine) Goto(index int) error {
	if index < 0 || index >= len(e.walkthrough.Steps) {
		return ErrOutOfRange
	}
	e.index = index
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
