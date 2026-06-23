// Command devui serves the overlay frontend in a normal browser with live
// reload, so the HUD can be iterated on without rebuilding the Wails app.
//
// It mocks the Wails bindings (window.go.overlay.App + window.runtime) with
// real step data loaded from a walkthrough YAML, renders the HUD inside a
// correctly-sized frame over a faux game scene, and reloads the browser via
// Server-Sent Events whenever a file in overlay/frontend changes.
//
// No Node, no npm — pure Go stdlib plus the project's own config package.
//
//	go run ./tools/devui                       # defaults to gothic2/chapter1
//	go run ./tools/devui -config path/to.yaml  # any walkthrough
//	go run ./tools/devui -bg scene.png         # use a real screenshot as scene
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/XotoX1337/GoThrough/config"
)

const (
	addr        = "localhost:34116"
	frontendDir = "overlay/frontend"
)

func main() {
	configPath := flag.String("config", "configs/gothic2/chapter1.yaml", "walkthrough YAML to preview")
	bgPath := flag.String("bg", "", "optional background image (game screenshot) for the scene")
	flag.Parse()

	wt, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("loading walkthrough: %v", err)
	}
	stepsJSON, err := json.Marshal(toStepInfos(wt))
	if err != nil {
		log.Fatalf("encoding steps: %v", err)
	}
	metaJSON, err := json.Marshal(map[string]string{"game": wt.Game, "title": wt.Title})
	if err != nil {
		log.Fatalf("encoding meta: %v", err)
	}

	hub := newReloadHub()
	go watch(frontendDir, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHarness)
	mux.HandleFunc("/app", serveApp(stepsJSON, metaJSON))
	mux.HandleFunc("/__reload", hub.handleSSE)
	if *bgPath != "" {
		mux.HandleFunc("/__bg", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, *bgPath)
		})
	}

	url := "http://" + addr
	fmt.Printf("devui: %s — %d steps from %s\n", url, len(wt.Steps), *configPath)
	fmt.Println("devui: edit overlay/frontend/index.html and save — the page reloads automatically. Ctrl+C to stop.")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// stepInfo mirrors overlay.StepInfo — the shape the real Wails binding returns.
type stepInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	IsFirst     bool   `json:"isFirst"`
	IsLast      bool   `json:"isLast"`
}

func toStepInfos(wt *config.Walkthrough) []stepInfo {
	total := len(wt.Steps)
	out := make([]stepInfo, total)
	for i, s := range wt.Steps {
		out[i] = stepInfo{
			Title:       s.Title,
			Description: s.Description,
			Current:     i + 1,
			Total:       total,
			IsFirst:     i == 0,
			IsLast:      i == total-1,
		}
	}
	return out
}

// serveApp serves the real index.html with the Wails bindings mocked, so the
// untouched frontend runs in a plain browser against real step data.
func serveApp(stepsJSON, metaJSON []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := os.ReadFile(filepath.Join(frontendDir, "index.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		inject := `
<style>html { background: transparent !important; }</style>
<script>
// --- devui: mock of the Wails bindings ---------------------------------
(function () {
  const steps = ` + string(stepsJSON) + `;
  const meta = ` + string(metaJSON) + `;
  let i = 0;
  const listeners = {};
  const emit = (name, data) => (listeners[name] || []).forEach(cb => cb(data));
  const at = () => Promise.resolve(steps[i]);
  // Mock settings — mirrors settings.Defaults() (settings/settings.go). SaveHotkeys
  // just stores and echoes them back; real registration happens only in the Wails build.
  const settings = { version: 1, hotkeys: {
    next:       { mods: ['ctrl', 'alt'], key: 'right' },
    prev:       { mods: ['ctrl', 'alt'], key: 'left'  },
    toggleHide: { mods: ['ctrl', 'alt'], key: 'h'     },
    quit:       { mods: ['ctrl', 'alt'], key: 'q'     },
  } };
  window.go = { overlay: { App: {
    Meta: () => Promise.resolve(meta),
    Steps: () => Promise.resolve(steps),
    CurrentStep: at,
    Next: () => { if (i < steps.length - 1) i++; return at(); },
    Prev: () => { if (i > 0) i--; return at(); },
    Goto: (idx) => { i = Math.max(0, Math.min(idx, steps.length - 1)); return at(); },
    FitWindow: () => {}, // no-op: the browser can't resize the OS window
    Settings: () => Promise.resolve(settings),
    SaveHotkeys: (hk) => { settings.hotkeys = hk; return Promise.resolve(settings); },
  } } };
  window.runtime = {
    Quit: () => console.log('[devui] runtime.Quit() (no-op in browser)'),
    EventsOn: (name, cb) => { (listeners[name] = listeners[name] || []).push(cb); },
    EventsEmit: (name, data) => emit(name, data),
  };
  // Simulate the real global hotkeys (Ctrl+Alt+Right/Left) with the arrow keys,
  // so the event-driven step:changed path can be exercised in the browser.
  // Click the HUD first to give the iframe keyboard focus.
  window.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowRight') { if (i < steps.length - 1) i++; emit('step:changed', steps[i]); }
    else if (e.key === 'ArrowLeft') { if (i > 0) i--; emit('step:changed', steps[i]); }
  });
})();
</script>`
		// Inject before </head> so window.go / window.runtime exist before the
		// frontend's inline script runs (it resolves `window.go.overlay.App` at
		// parse time) — this mirrors how the real Wails build injects its
		// bindings into the head ahead of app scripts.
		html := strings.Replace(string(raw), "</head>", inject+"\n</head>", 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	}
}

// serveHarness renders the faux game scene with the HUD framed at real size
// (440x220, matching overlay.Run) and wires up live reload.
func serveHarness(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, harnessHTML)
}

const harnessHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<title>GoThrough — devui</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  html, body { height: 100%; }
  body {
    /* Faux "game scene" so the glassmorphism/blur is visible. Override with -bg. */
    background: var(--scene), radial-gradient(120% 90% at 68% 12%, #2b3542 0%, #1a2129 40%, #0c0f13 82%);
    background-size: cover;
    background-position: center;
    font-family: system-ui, sans-serif;
    overflow: hidden;
  }
  /* Subtle scene texture so the frostglass blur reads against something. */
  body::before {
    content: "";
    position: fixed; inset: 0;
    background:
      linear-gradient(180deg, rgba(70,92,112,0.18) 0%, rgba(0,0,0,0) 35%, rgba(8,12,16,0.55) 100%),
      repeating-linear-gradient(54deg, rgba(255,255,255,0.018) 0 2px, transparent 2px 9px);
    pointer-events: none;
  }
  /* The overlay fills the whole "screen", transparent — the HUD positions itself. */
  iframe {
    position: fixed; inset: 0;
    width: 100%; height: 100%;
    border: none; background: transparent;
    color-scheme: light dark;
  }
  .tag {
    position: fixed; bottom: 12px; left: 12px;
    font-size: 11px; color: rgba(255,255,255,0.32);
    letter-spacing: 0.05em; pointer-events: none; z-index: 1;
  }
</style>
</head>
<body>
  <iframe src="/app" title="GoThrough HUD"></iframe>
  <div class="tag">devui · live reload active · fullscreen overlay</div>
  <script>
    // Use the -bg image as the scene background if the server provides one.
    fetch('/__bg', { method: 'HEAD' }).then(res => {
      if (res.ok) document.documentElement.style.setProperty('--scene', "url('/__bg')");
    }).catch(() => {});

    // Live reload via Server-Sent Events.
    const es = new EventSource('/__reload');
    es.onmessage = () => location.reload();
  </script>
</body>
</html>`

// reloadHub fans a single "changed" signal out to every connected browser tab.
type reloadHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newReloadHub() *reloadHub {
	return &reloadHub{clients: make(map[chan struct{}]struct{})}
}

func (h *reloadHub) add() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *reloadHub) remove(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *reloadHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *reloadHub) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.add()
	defer h.remove(ch)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

// watch polls the frontend directory and broadcasts a reload whenever the set
// of files or their modification times changes. Polling keeps this dependency
// -free (no fsnotify) — fine for a handful of files.
func watch(dir string, hub *reloadHub) {
	last := fingerprint(dir)
	for range time.Tick(300 * time.Millisecond) {
		if fp := fingerprint(dir); fp != last {
			last = fp
			hub.broadcast()
		}
	}
}

func fingerprint(dir string) string {
	var b strings.Builder
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(&b, "%s:%d:%d;", path, info.ModTime().UnixNano(), info.Size())
		return nil
	})
	return b.String()
}
