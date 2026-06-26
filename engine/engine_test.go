package engine

import (
	"errors"
	"testing"

	"github.com/XotoX1337/GoThrough/config"
)

func walkthrough(ids ...int) *config.Walkthrough {
	nodes := make([]config.Node, len(ids))
	for i, id := range ids {
		nodes[i] = config.Node{Step: &config.Step{ID: id, Title: "step"}}
	}
	return &config.Walkthrough{Game: "g", Title: "t", Steps: nodes}
}

func step(id int) config.Node { return config.Node{Step: &config.Step{ID: id, Title: "step"}} }

// gated builds a step whose visibility depends on a choice answer.
func gated(id int, key string, values ...string) config.Node {
	return config.Node{Step: &config.Step{
		ID: id, Title: "step",
		When: config.Condition{key: config.StringList(values)},
	}}
}

// branching builds: shared [1,2] -> choice(guild: A / B) -> A-only step 10,11,
// B-only step 20 -> shared [3]. With the flat model the choice never stops the
// sequence; gated steps simply appear once the matching answer is recorded.
func branching() *config.Walkthrough {
	return &config.Walkthrough{
		Game: "g", Title: "t",
		Steps: []config.Node{
			step(1), step(2),
			{Choice: &config.Choice{
				Key: "guild", Prompt: "Pick",
				Options: []config.ChoiceOption{
					{Value: "A", Label: "Alpha"},
					{Value: "B", Label: "Bravo"},
				},
			}},
			gated(10, "guild", "A"),
			gated(11, "guild", "A"),
			gated(20, "guild", "B"),
			step(3),
		},
	}
}

// fakePersister records the most recent save and counts calls.
type fakePersister struct {
	index, stepID, calls int
	choices              map[string]string
}

func (f *fakePersister) Save(index, stepID int, choices map[string]string) error {
	f.index, f.stepID = index, stepID
	f.choices = choices
	f.calls++
	return nil
}

func curID(e *Engine) int {
	c := e.Current()
	if c == nil || c.Step == nil {
		return -1
	}
	return c.Step.ID
}

func TestRestoreByID(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	// Saved index is stale (would point elsewhere) but the ID still matches.
	e.Restore(0, 30, nil)
	if got := curID(e); got != 30 {
		t.Fatalf("restore by ID: current ID = %d, want 30", got)
	}
}

func TestRestoreFallsBackToIndexWhenIDMissing(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	e.Restore(1, 999, nil) // ID gone from config; fall back to index 1
	if got := curID(e); got != 20 {
		t.Fatalf("restore fallback: current ID = %d, want 20", got)
	}
}

func TestRestoreClampsOutOfRange(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	e.Restore(99, 0, nil)
	if cur, _ := e.Progress(); cur != 3 {
		t.Fatalf("restore clamp: current = %d, want 3 (last)", cur)
	}
}

func TestPersistOnNavigation(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	fp := &fakePersister{}
	e.UsePersister(fp)

	if err := e.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if fp.calls != 1 || fp.index != 1 || fp.stepID != 20 {
		t.Fatalf("after Next: calls=%d index=%d stepID=%d, want 1/1/20", fp.calls, fp.index, fp.stepID)
	}

	if err := e.Goto(2); err != nil {
		t.Fatalf("Goto: %v", err)
	}
	if fp.index != 2 || fp.stepID != 30 {
		t.Fatalf("after Goto: index=%d stepID=%d, want 2/30", fp.index, fp.stepID)
	}
}

func TestNoPersistWhenMoveBlocked(t *testing.T) {
	e := New(walkthrough(10))
	fp := &fakePersister{}
	e.UsePersister(fp)

	if err := e.Prev(); !errors.Is(err, ErrAlreadyFirst) {
		t.Fatalf("Prev at first: err = %v, want ErrAlreadyFirst", err)
	}
	if fp.calls != 0 {
		t.Fatalf("blocked move must not persist; calls = %d", fp.calls)
	}
}

// --- choices ---

func TestUnansweredChoiceHidesGatedSteps(t *testing.T) {
	e := New(branching())
	// Sequence is shared prefix [1,2] + the choice + shared suffix [3]; the
	// gated steps (10,11,20) are hidden until the choice is answered.
	if _, total := e.Progress(); total != 4 {
		t.Fatalf("undecided total = %d, want 4 (1,2,choice,3)", total)
	}
	if err := e.Goto(2); err != nil {
		t.Fatalf("Goto choice: %v", err)
	}
	if !e.Current().IsChoice() {
		t.Fatal("item at index 2 should be the choice")
	}
}

func TestChooseRevealsGatedStepsAndAdvances(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2) // onto the choice
	if err := e.Choose("guild", "A"); err != nil {
		t.Fatalf("Choose A: %v", err)
	}
	// Now: [1,2,CHOICE,10,11,3]; the choice stays, we step into the first gated
	// step (10).
	if got := curID(e); got != 10 {
		t.Fatalf("after Choose: current ID = %d, want 10", got)
	}
	if _, total := e.Progress(); total != 6 {
		t.Fatalf("answered total = %d, want 6", total)
	}
	// Walk to the end: 10 -> 11 -> 3 (shared suffix).
	_ = e.Next()
	_ = e.Next()
	if got := curID(e); got != 3 {
		t.Fatalf("suffix: current ID = %d, want 3", got)
	}
	if !e.Done() {
		t.Fatal("should be Done on the shared suffix's last step")
	}
}

func TestReChooseAfterGoingBack(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2)
	_ = e.Choose("guild", "A") // -> [1,2,CHOICE,10,11,3], on step 10 (index 3)

	// Go back onto the choice; it must still be there and report the answer.
	if err := e.Prev(); err != nil {
		t.Fatalf("Prev: %v", err)
	}
	cur := e.Current()
	if !cur.IsChoice() {
		t.Fatal("Prev from the first gated step should land on the choice")
	}
	if cur.Selected != "A" {
		t.Fatalf("choice should remember the answer; Selected = %q, want A", cur.Selected)
	}

	// Re-choose B; the sequence reflows and we step into option B's step (20).
	if err := e.Choose("guild", "B"); err != nil {
		t.Fatalf("re-Choose B: %v", err)
	}
	if got := curID(e); got != 20 {
		t.Fatalf("after re-choose: current ID = %d, want 20", got)
	}
	if _, total := e.Progress(); total != 5 { // [1,2,CHOICE,20,3]
		t.Fatalf("re-chosen total = %d, want 5", total)
	}
}

func TestChooseRejectsBadInput(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2)
	if err := e.Choose("guild", "Z"); !errors.Is(err, ErrBadOption) {
		t.Fatalf("bad option: err = %v, want ErrBadOption", err)
	}
	if err := e.Choose("nope", "A"); !errors.Is(err, ErrNotAChoice) {
		t.Fatalf("wrong key: err = %v, want ErrNotAChoice", err)
	}
	_ = e.Goto(0) // not on a choice
	if err := e.Choose("guild", "A"); !errors.Is(err, ErrNotAChoice) {
		t.Fatalf("not on choice: err = %v, want ErrNotAChoice", err)
	}
}

func TestRestoreReplaysChoice(t *testing.T) {
	e := New(branching())
	e.Restore(4, 11, map[string]string{"guild": "A"})
	// With guild=A the sequence is [1,2,CHOICE,10,11,3]; ID 11 sits at index 4.
	if got := curID(e); got != 11 {
		t.Fatalf("restore with choice: current ID = %d, want 11", got)
	}
	if _, total := e.Progress(); total != 6 {
		t.Fatalf("restored total = %d, want 6", total)
	}
}

func TestRestoreStaleChoiceHidesGatedSteps(t *testing.T) {
	e := New(branching())
	// Saved value no longer matches any option -> gated steps stay hidden.
	e.Restore(2, 0, map[string]string{"guild": "GONE"})
	if _, total := e.Progress(); total != 4 {
		t.Fatalf("stale choice total = %d, want 4 (gated steps hidden)", total)
	}
	if !e.Current().IsChoice() {
		t.Fatal("index 2 should be the choice")
	}
}

func TestChoosePersists(t *testing.T) {
	e := New(branching())
	fp := &fakePersister{}
	e.UsePersister(fp)
	_ = e.Goto(2)
	if err := e.Choose("guild", "B"); err != nil {
		t.Fatalf("Choose: %v", err)
	}
	if fp.choices["guild"] != "B" {
		t.Fatalf("persisted choices = %v, want guild=B", fp.choices)
	}
}
