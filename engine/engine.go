// Package engine manages step state for an active walkthrough session.
//
// A walkthrough may contain branches (decision points that fork into named
// options and re-converge afterwards). The engine resolves the document into a
// flat, linear sequence of Items for the chosen path: shared steps, then the
// chosen option's steps, then the shared steps that follow the branch. An
// undecided branch appears as a decision Item and the sequence stops there until
// the user picks an option.
package engine

import (
	"errors"

	"github.com/XotoX1337/GoThrough/config"
)

var (
	ErrAlreadyFirst = errors.New("already at first step")
	ErrAlreadyLast  = errors.New("already at last step")
	ErrOutOfRange   = errors.New("step index out of range")
	ErrNotABranch   = errors.New("current item is not the named branch")
	ErrBadOption    = errors.New("unknown branch option")
)

// Persister records the user's position and branch choices so they can be
// restored in a later session. It is bound to a single walkthrough. Implemented
// by *progress.Handle.
type Persister interface {
	Save(index, stepID int, branches map[string]string) error
}

// Item is one position in the resolved sequence: either a Step or a Branch
// decision (never both). A branch decision stays in the sequence even after a
// choice is made — its Selected field then holds the chosen option label — so
// the user can navigate back onto it and re-choose. Section is the group title
// for HUD rendering (empty for a flat, sections-less config).
type Item struct {
	Step     *config.Step
	Branch   *config.Branch
	Selected string // chosen option label ("" while undecided)
	Section  string
}

// IsBranch reports whether this item is a branch decision (decided or not).
func (it Item) IsBranch() bool { return it.Branch != nil }

// Engine tracks which item the user is currently on and the branch options
// chosen so far (persistKey -> option label).
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

// optionIndex returns the chosen option index for a branch, or ok=false if the
// branch has no (valid) recorded choice yet.
func (e *Engine) optionIndex(b *config.Branch) (int, bool) {
	label, ok := e.selections[b.PersistKey]
	if !ok || label == "" {
		return 0, false
	}
	for i, o := range b.Options {
		if o.Label == label {
			return i, true
		}
	}
	return 0, false
}

// flatten rebuilds the linear sequence from the walkthrough outline and the
// current branch selections, stopping at the first undecided branch.
func (e *Engine) flatten() {
	var seq []Item
	for _, on := range e.walkthrough.Outline() {
		if on.Step != nil {
			seq = append(seq, Item{Step: on.Step, Section: on.Section})
			continue
		}
		sub, stopped := e.expandBranch(on.Branch, on.Section)
		seq = append(seq, sub...)
		if stopped {
			break
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

// expandBranch resolves a branch for the current selections. The decision Item
// is always emitted (so it stays navigable for re-choosing). If undecided,
// nothing follows it and stopped=true halts flattening (the shared suffix stays
// hidden until a choice is made). If decided, the chosen option's nodes are
// inlined after the decision (recursively, so nested branches are handled).
func (e *Engine) expandBranch(b *config.Branch, section string) (items []Item, stopped bool) {
	idx, ok := e.optionIndex(b)
	if !ok {
		return []Item{{Branch: b, Section: section}}, true
	}
	decision := Item{Branch: b, Section: section, Selected: b.Options[idx].Label}
	sub, st := e.expandNodes(b.Options[idx].Steps, section)
	return append([]Item{decision}, sub...), st
}

func (e *Engine) expandNodes(nodes []config.Node, section string) (items []Item, stopped bool) {
	var out []Item
	for _, n := range nodes {
		if n.Step != nil {
			out = append(out, Item{Step: n.Step, Section: section})
			continue
		}
		sub, st := e.expandBranch(n.Branch, section)
		out = append(out, sub...)
		if st {
			return out, true
		}
	}
	return out, false
}

// findBranch locates a branch by persistKey anywhere in the document.
func (e *Engine) findBranch(persistKey string) *config.Branch {
	var search func(nodes []config.Node) *config.Branch
	search = func(nodes []config.Node) *config.Branch {
		for _, n := range nodes {
			if n.Branch != nil {
				if n.Branch.PersistKey == persistKey {
					return n.Branch
				}
				for _, o := range n.Branch.Options {
					if b := search(o.Steps); b != nil {
						return b
					}
				}
			}
		}
		return nil
	}
	for _, on := range e.walkthrough.Outline() {
		if on.Branch != nil {
			if on.Branch.PersistKey == persistKey {
				return on.Branch
			}
			for _, o := range on.Branch.Options {
				if b := search(o.Steps); b != nil {
					return b
				}
			}
		}
	}
	return nil
}

// Choose records the option for the branch the user is currently on (which may
// already be decided — this is how re-choosing works) and advances into the
// chosen option. The current item must be that branch (ErrNotABranch
// otherwise); the label must name a real option (ErrBadOption otherwise). On
// success the sequence is re-flattened and the new position persisted.
func (e *Engine) Choose(persistKey, label string) error {
	cur := e.Current()
	if cur == nil || cur.Branch == nil || cur.Branch.PersistKey != persistKey {
		return ErrNotABranch
	}
	valid := false
	for _, o := range cur.Branch.Options {
		if o.Label == label {
			valid = true
			break
		}
	}
	if !valid {
		return ErrBadOption
	}
	pos := e.index // the decision item stays put; we step into the option after it
	e.selections[persistKey] = label
	e.flatten()
	if pos+1 < len(e.seq) {
		e.index = pos + 1
	} else if pos < len(e.seq) {
		e.index = pos
	}
	return e.persist()
}

// Restore positions the engine at a previously saved step and replays the saved
// branch choices. It prefers to match by step ID (resilient to steps being
// inserted or removed) and falls back to the saved index, clamped to range.
func (e *Engine) Restore(index, stepID int, branches map[string]string) {
	e.selections = map[string]string{}
	for k, v := range branches {
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

// persist writes the current position and branch choices through the configured
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

// Next advances to the next item. Returns ErrAlreadyLast if at the end (which
// includes sitting on an undecided branch — the sequence stops there until a
// choice is made).
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

// Next returns the follow-up file reference (empty if none).
func (e *Engine) NextFile() string { return e.walkthrough.Next }

// Progress returns 1-based current position and total item count.
func (e *Engine) Progress() (current, total int) {
	return e.index + 1, len(e.seq)
}

// Done reports whether the user is on the final item AND it is a real step
// (an undecided branch is never "done" — it needs a choice).
func (e *Engine) Done() bool {
	if len(e.seq) == 0 {
		return false
	}
	return e.index == len(e.seq)-1 && !e.seq[e.index].IsBranch()
}
