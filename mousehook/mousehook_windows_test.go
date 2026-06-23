//go:build windows

package mousehook

import "testing"

// TestStartStopInstallsHook exercises the real syscall plumbing: it installs a
// genuine WH_MOUSE_LL hook, confirms Start succeeds (so the callback signature,
// SetWindowsHookEx call and message pump are all wired correctly), then tears it
// down. It does not simulate clicks — that needs real input — but a failure here
// catches a broken hook installation.
func TestStartStopInstallsHook(t *testing.T) {
	m := New()
	m.Add(ModCtrl|ModAlt, ButtonMiddle, func() {})
	if err := m.Start(); err != nil {
		t.Fatalf("Start installed no hook: %v", err)
	}
	m.Stop()
}

func TestStartWithoutBindingsIsNoop(t *testing.T) {
	m := New()
	if err := m.Start(); err != nil {
		t.Fatalf("empty Start should be a no-op, got: %v", err)
	}
	m.Stop() // must be safe even though nothing was installed
}

// TestClassifyButtons checks the message→button mapping (the part that decides
// what gets matched/swallowed) without touching the OS.
func TestClassifyButtons(t *testing.T) {
	cases := []struct {
		wParam uintptr
		btn    Button
		isUp   bool
	}{
		{wmMButtonDown, ButtonMiddle, false},
		{wmMButtonUp, ButtonMiddle, true},
		{wmLButtonDown, ButtonLeft, false},
		{wmRButtonUp, ButtonRight, true},
	}
	for _, c := range cases {
		btn, isUp, ok := classify(c.wParam, nil)
		if !ok || btn != c.btn || isUp != c.isUp {
			t.Fatalf("classify(%#x) = (%v,%v,%v), want (%v,%v,true)", c.wParam, btn, isUp, ok, c.btn, c.isUp)
		}
	}
	if _, _, ok := classify(0x0200 /* WM_MOUSEMOVE */, nil); ok {
		t.Fatal("mouse-move should classify as not-a-button")
	}
}
