//go:build !windows

package overlay

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// On non-Windows platforms Wails' runtime is the only window API we have, and it
// exposes just the primary screen size — so we keep the original primary-screen
// clamp. Cross-monitor placement is a Windows-only refinement (see
// screen_windows.go); clampToScreen lives in overlay.go.
func clampToWorkArea(ctx context.Context, x, y, w, h int) (int, int) {
	return clampToScreen(ctx, x, y, w, h)
}

func moveOverlayAbs(ctx context.Context, x, y, w, h int) (int, int) {
	x, y = clampToScreen(ctx, x, y, w, h)
	runtime.WindowSetPosition(ctx, x, y)
	return x, y
}
