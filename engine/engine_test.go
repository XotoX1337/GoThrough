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

// branching builds: shared [1,2] -> branch(guild: A=[10,11] / B=[20]) -> shared [3].
func branching() *config.Walkthrough {
	return &config.Walkthrough{
		Game: "g", Title: "t",
		Steps: []config.Node{
			step(1), step(2),
			{Branch: &config.Branch{
				PersistKey: "guild", Title: "Pick",
				Options: []config.BranchOption{
					{Label: "A", Steps: []config.Node{step(10), step(11)}},
					{Label: "B", Steps: []config.Node{step(20)}},
				},
			}},
			step(3),
		},
	}
}

// fakePersister records the most recent save and counts calls.
type fakePersister struct {
	index, stepID, calls int
	branches             map[string]string
}

func (f *fakePersister) Save(index, stepID int, branches map[string]string) error {
	f.index, f.stepID = index, stepID
	f.branches = branches
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

// --- branching ---

func TestUndecidedBranchStopsSequence(t *testing.T) {
	e := New(branching())
	// Sequence is shared prefix [1,2] + the decision; suffix is hidden.
	if _, total := e.Progress(); total != 3 {
		t.Fatalf("undecided total = %d, want 3 (2 steps + decision)", total)
	}
	if err := e.Goto(2); err != nil {
		t.Fatalf("Goto decision: %v", err)
	}
	if !e.Current().IsBranch() {
		t.Fatal("item at index 2 should be the branch decision")
	}
	if err := e.Next(); !errors.Is(err, ErrAlreadyLast) {
		t.Fatalf("Next past undecided branch: err = %v, want ErrAlreadyLast", err)
	}
	if e.Done() {
		t.Fatal("an undecided branch must not count as Done")
	}
}

func TestChooseInlinesOptionAndReconverges(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2) // onto the decision
	if err := e.Choose("guild", "A"); err != nil {
		t.Fatalf("Choose A: %v", err)
	}
	// Now: [1,2,BRANCH,10,11,3]; the decision stays, we step into option A (10).
	if got := curID(e); got != 10 {
		t.Fatalf("after Choose: current ID = %d, want 10", got)
	}
	if _, total := e.Progress(); total != 6 {
		t.Fatalf("decided total = %d, want 6", total)
	}
	// Walk to the end: 10 -> 11 -> 3 (shared suffix re-converges).
	_ = e.Next()
	_ = e.Next()
	if got := curID(e); got != 3 {
		t.Fatalf("re-converge: current ID = %d, want 3", got)
	}
	if !e.Done() {
		t.Fatal("should be Done on the shared suffix's last step")
	}
}

func TestReChooseAfterGoingBack(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2)
	_ = e.Choose("guild", "A") // -> [1,2,BRANCH,10,11,3], on step 10 (index 3)

	// Go back onto the decision; it must still be there and report the choice.
	if err := e.Prev(); err != nil {
		t.Fatalf("Prev: %v", err)
	}
	cur := e.Current()
	if !cur.IsBranch() {
		t.Fatal("Prev from the first option step should land on the decision")
	}
	if cur.Selected != "A" {
		t.Fatalf("decision should remember the choice; Selected = %q, want A", cur.Selected)
	}

	// Re-choose B; the sequence reflows and we step into option B (20).
	if err := e.Choose("guild", "B"); err != nil {
		t.Fatalf("re-Choose B: %v", err)
	}
	if got := curID(e); got != 20 {
		t.Fatalf("after re-choose: current ID = %d, want 20", got)
	}
	if _, total := e.Progress(); total != 5 { // [1,2,BRANCH,20,3]
		t.Fatalf("re-chosen total = %d, want 5", total)
	}
}

func TestChooseRejectsBadInput(t *testing.T) {
	e := New(branching())
	_ = e.Goto(2)
	if err := e.Choose("guild", "Z"); !errors.Is(err, ErrBadOption) {
		t.Fatalf("bad option: err = %v, want ErrBadOption", err)
	}
	if err := e.Choose("nope", "A"); !errors.Is(err, ErrNotABranch) {
		t.Fatalf("wrong key: err = %v, want ErrNotABranch", err)
	}
	_ = e.Goto(0) // not on a branch
	if err := e.Choose("guild", "A"); !errors.Is(err, ErrNotABranch) {
		t.Fatalf("not on branch: err = %v, want ErrNotABranch", err)
	}
}

func TestRestoreReplaysBranchChoice(t *testing.T) {
	e := New(branching())
	e.Restore(4, 11, map[string]string{"guild": "A"})
	// With guild=A the sequence is [1,2,BRANCH,10,11,3]; ID 11 sits at index 4.
	if got := curID(e); got != 11 {
		t.Fatalf("restore with branch: current ID = %d, want 11", got)
	}
	if _, total := e.Progress(); total != 6 {
		t.Fatalf("restored total = %d, want 6", total)
	}
}

func TestRestoreStaleBranchFallsBackToDecision(t *testing.T) {
	e := New(branching())
	// Saved label no longer exists -> branch reverts to undecided.
	e.Restore(2, 0, map[string]string{"guild": "GONE"})
	if !e.Current().IsBranch() {
		t.Fatal("stale branch label should leave the decision pending")
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
	if fp.branches["guild"] != "B" {
		t.Fatalf("persisted branches = %v, want guild=B", fp.branches)
	}
}
