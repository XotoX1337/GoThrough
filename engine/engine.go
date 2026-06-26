// Package engine manages step state for an active walkthrough session.
//
// A walkthrough may contain choices (flat decision points that record an answer
// under a key). Steps elsewhere in the file opt into a choice via their `when`
// condition. The engine resolves the document into a flat, linear sequence of
// Items: every choice is emitted (so it stays navigable for re-answering), and a
// step is included only when its `when` condition is satisfied by the recorded
// answers. An unanswered choice simply hides its dependent steps.
package engine

import (
	"errors"

	"github.com/XotoX1337/GoThrough/config"
)

var (
	ErrAlreadyFirst = errors.New("already at first step")
	ErrAlreadyLast  = errors.New("already at last step")
	ErrOutOfRange   = errors.New("step index out of range")
	ErrNotAChoice   = errors.New("current item is not the named choice")
	ErrBadOption    = errors.New("unknown choice option")
)

// Persister records the user's position and choice answers so they can be
// restored in a later session. It is bound to a single walkthrough. Implemented
// by *progress.Handle.
type Persister interface {
	Save(index, stepID int, choices map[string]string) error
}

// Item is one position in the resolved sequence: either a Step or a Choice
// (never both). A choice stays in the sequence even after it is answered — its
// Selected field then holds the chosen option value — so the user can navigate
// back onto it and re-choose. Section is the group title for HUD rendering
// (empty for a flat, sections-less config).
type Item struct {
	Step     *config.Step
	Choice   *config.Choice
	Selected string // chosen option value ("" while undecided)
	Section  string
}

// IsChoice reports whether this item is a choice (answered or not).
func (it Item) IsChoice() bool { return it.Choice != nil }

// Engine tracks which item the user is currently on and the choice answers
// recorded so far (choice key -> option value).
type Engine struct {
	walkthrough *config.Walkthrough
	index       int
	selections  map[string]string
	seq         []Item
	store       Persister
}

// New creates an Engine starting at the first item.
func New(wt *config.Walkthrough) *Engine {
	e := &Engine{walkthrough: wt, index: 0, selections: map[string]string{}}
	e.flatten()
	return e
}

// UsePersister enables autosave: from now on every change is written through p.
// Call Restore first if you want to resume a saved position, so the restore
// itself isn't redundantly written back.
func (e *Engine) UsePersister(p Persister) {
	e.store = p
}

// stepVisible reports whether a step's `when` condition is satisfied by the
// current answers. A step with no condition is always visible; a condition
// referencing an unanswered choice is never satisfied.
func (e *Engine) stepVisible(s *config.Step) bool {
	for key, accepted := range s.When {
		val, ok := e.selections[key]
		if !ok || val == "" {
			return false
		}
		match := false
		for _, a := range accepted {
			if a == val {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

// flatten rebuilds the linear sequence from the walkthrough outline and the
// current choice answers: every choice is emitted, and each step is included
// only if its `when` condition is satisfied.
func (e *Engine) flatten() {
	var seq []Item
	for _, on := range e.walkthrough.Outline() {
		if on.Choice != nil {
			seq = append(seq, Item{Choice: on.Choice, Section: on.Section, Selected: e.selections[on.Choice.Key]})
			continue
		}
		if e.stepVisible(on.Step) {
			seq = append(seq, Item{Step: on.Step, Section: on.Section})
		}
	}
	e.seq = seq
	if e.index >= len(seq) {
		e.index = len(seq) - 1
	}
	if e.index < 0 {
		e.index = 0
	}
}

// Choose records the answer for the choice the user is currently on (which may
// already be answered — this is how re-choosing works) and advances to the item
// right after it. The current item must be that choice (ErrNotAChoice
// otherwise); the value must name a real option (ErrBadOption otherwise). On
// success the sequence is re-flattened and the new position persisted.
func (e *Engine) Choose(key, value string) error {
	cur := e.Current()
	if cur == nil || cur.Choice == nil || cur.Choice.Key != key {
		return ErrNotAChoice
	}
	valid := false
	for _, o := range cur.Choice.Options {
		if o.Value == value {
			valid = true
			break
		}
	}
	if !valid {
		return ErrBadOption
	}
	e.selections[key] = value
	e.flatten()
	// Re-find the choice (its index can shift if answering it changed earlier
	// steps' visibility) and step into the item that follows it.
	for i := range e.seq {
		if c := e.seq[i].Choice; c != nil && c.Key == key {
			if i+1 < len(e.seq) {
				e.index = i + 1
			} else {
				e.index = i
			}
			break
		}
	}
	return e.persist()
}

// Restore positions the engine at a previously saved step and replays the saved
// choice answers. It prefers to match by step ID (resilient to steps being
// inserted or removed) and falls back to the saved index, clamped to range.
func (e *Engine) Restore(index, stepID int, choices map[string]string) {
	e.selections = map[string]string{}
	for k, v := range choices {
		e.selections[k] = v
	}
	e.flatten()

	for i := range e.seq {
		if s := e.seq[i].Step; s != nil && s.ID == stepID {
			e.index = i
			return
		}
	}
	if index < 0 {
		index = 0
	}
	if max := len(e.seq) - 1; index > max {
		index = max
	}
	e.index = index
}

// persist writes the current position and choice answers through the configured
// Persister, if any.
func (e *Engine) persist() error {
	if e.store == nil {
		return nil
	}
	id := 0
	if s := e.seq[e.index].Step; s != nil {
		id = s.ID
	}
	return e.store.Save(e.index, id, e.selections)
}

// Current returns the active item. Never nil while the sequence is non-empty.
func (e *Engine) Current() *Item {
	if len(e.seq) == 0 {
		return nil
	}
	return &e.seq[e.index]
}

// Next advances to the next item. Returns ErrAlreadyLast if at the end.
func (e *Engine) Next() error {
	if e.index >= len(e.seq)-1 {
		return ErrAlreadyLast
	}
	e.index++
	_ = e.persist()
	return nil
}

// Prev goes back one item. Returns ErrAlreadyFirst if at the beginning.
func (e *Engine) Prev() error {
	if e.index == 0 {
		return ErrAlreadyFirst
	}
	e.index--
	_ = e.persist()
	return nil
}

// Goto jumps to a 0-based item index. Returns ErrOutOfRange if out of bounds.
func (e *Engine) Goto(index int) error {
	if index < 0 || index >= len(e.seq) {
		return ErrOutOfRange
	}
	e.index = index
	_ = e.persist()
	return nil
}

// Items returns the resolved sequence for the current path (for the checklist).
func (e *Engine) Items() []Item { return e.seq }

// Index returns the 0-based position of the active item.
func (e *Engine) Index() int { return e.index }

// Game returns the game name from the loaded walkthrough.
func (e *Engine) Game() string { return e.walkthrough.Game }

// Title returns the walkthrough title.
func (e *Engine) Title() string { return e.walkthrough.Title }

// Variant returns the walkthrough variant label (empty if unset).
func (e *Engine) Variant() string { return e.walkthrough.Variant }

// NextFile returns the follow-up file reference (empty if none).
func (e *Engine) NextFile() string { return e.walkthrough.Next }

// Progress returns 1-based current position and total item count.
func (e *Engine) Progress() (current, total int) {
	return e.index + 1, len(e.seq)
}

// Done reports whether the user is on the final item AND it is a real step
// (sitting on a choice is never "done").
func (e *Engine) Done() bool {
	if len(e.seq) == 0 {
		return false
	}
	return e.index == len(e.seq)-1 && !e.seq[e.index].IsChoice()
}
