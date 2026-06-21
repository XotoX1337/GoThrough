package engine

import (
	"errors"
	"testing"

	"github.com/XotoX1337/GoThrough/config"
)

func walkthrough(ids ...int) *config.Walkthrough {
	steps := make([]config.Step, len(ids))
	for i, id := range ids {
		steps[i] = config.Step{ID: id, Title: "step"}
	}
	return &config.Walkthrough{Game: "g", Title: "t", Steps: steps}
}

// fakePersister records the most recent save and counts calls.
type fakePersister struct {
	index, stepID, calls int
}

func (f *fakePersister) Save(index, stepID int) error {
	f.index, f.stepID = index, stepID
	f.calls++
	return nil
}

func TestRestoreByID(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	// Saved index is stale (would point elsewhere) but the ID still matches.
	e.Restore(0, 30)
	if got := e.Current().ID; got != 30 {
		t.Fatalf("restore by ID: current ID = %d, want 30", got)
	}
}

func TestRestoreFallsBackToIndexWhenIDMissing(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	e.Restore(1, 999) // ID gone from config; fall back to index 1
	if got := e.Current().ID; got != 20 {
		t.Fatalf("restore fallback: current ID = %d, want 20", got)
	}
}

func TestRestoreClampsOutOfRange(t *testing.T) {
	e := New(walkthrough(10, 20, 30))
	e.Restore(99, 0)
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
