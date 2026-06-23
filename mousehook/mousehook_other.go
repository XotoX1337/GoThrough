//go:build !windows && !linux

package mousehook

// Manager is the no-op manager for platforms without a global mouse-hotkey
// backend (macOS, and Linux under Wayland builds without X11). Start reports
// ErrUnsupported so the caller can fall back gracefully.
type Manager struct {
	core
}

func New() *Manager { return &Manager{} }

func (m *Manager) Start() error {
	if len(m.bindings) == 0 {
		return nil
	}
	return ErrUnsupported
}

func (m *Manager) Stop() {}
