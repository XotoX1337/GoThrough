// Command devui serves the overlay frontend in a normal browser with live
// reload, so the HUD can be iterated on without rebuilding the Wails app.
//
// It mocks the Wails bindings (window.go.overlay.App + window.runtime). Rather
// than re-implementing navigation in JavaScript, the mock calls small HTTP
// endpoints backed by the REAL engine.Engine — so `when` gating, choices
// and `next` hand-off behave exactly as in the Wails build and the mock can't
// drift from the engine. Only the StepInfo wire shape is mirrored here (it
// belongs to the overlay package, which is CGo/Wails-only and can't be imported
// by this pure-Go tool); keep it in sync with overlay.StepInfo.
//
// No Node, no npm — pure Go stdlib plus the project's own packages.
//
//	go run ./tools/devui                       # defaults to gothic2/day1
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/tools/frontendbuild"
)

const (
	addr        = "localhost:34116"
	frontendDir = "overlay/frontend"
)

// server holds the live engine the mock bindings drive over HTTP.
type server struct {
	mu     sync.Mutex
	eng    *engine.Engine
	cfgDir string // dir of the loaded config, for resolving `next:`
}

func main() {
	configPath := flag.String("config", "configstore/configs/gothic2/day1.yaml", "walkthrough YAML to preview")
	bgPath := flag.String("bg", "", "optional background image (game screenshot) for the scene")
	flag.Parse()

	wt, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("loading walkthrough: %v", err)
	}
	srv := &server{eng: engine.New(wt), cfgDir: filepath.Dir(*configPath)}

	// The HUD's JS is authored in TypeScript (frontend/src/app.ts) and transpiled
	// to frontend/app.js. Build it once up front so devui serves fresh JS even if
	// app.js is stale; the watcher rebuilds on every src change below.
	if _, err := frontendbuild.Build(); err != nil {
		log.Printf("devui: initial frontend build failed: %v", err)
	}

	hub := newReloadHub()
	go watch(frontendDir, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHarness)
	mux.HandleFunc("/app", serveApp)
	mux.HandleFunc("/app.js", serveAppJS)
	mux.HandleFunc("/__reload", hub.handleSSE)
	srv.routes(mux)
	if *bgPath != "" {
		mux.HandleFunc("/__bg", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, *bgPath)
		})
	}

	url := "http://" + addr
	fmt.Printf("devui: %s — %d items from %s\n", url, len(srv.eng.Items()), *configPath)
	fmt.Println("devui: edit overlay/frontend/index.html and save — the page reloads automatically. Ctrl+C to stop.")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// stepInfo mirrors overlay.StepInfo — the shape the real Wails binding returns.
type stepInfo struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	IsFirst bool   `json:"isFirst"`
	IsLast  bool   `json:"isLast"`
	Section string `json:"section,omitempty"`

	IsChoice  bool               `json:"isChoice,omitempty"`
	ChoiceKey string             `json:"choiceKey,omitempty"`
	Selected  string             `json:"selected,omitempty"`
	Options   []choiceOptionInfo `json:"options,omitempty"`

	ID          int         `json:"id,omitempty"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Tasks       []taskInfo  `json:"tasks,omitempty"`
	Optional    bool        `json:"optional,omitempty"`
	Quests      []questInfo `json:"quests,omitempty"`
	Hints       []string    `json:"hints,omitempty"`
	Warnings    []string    `json:"warnings,omitempty"`
	Infos       []string    `json:"infos,omitempty"`
}

type choiceOptionInfo struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type taskInfo struct {
	Text    string `json:"text"`
	Info    string `json:"info,omitempty"`
	Warning string `json:"warning,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

type questInfo struct {
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
	Note   string `json:"note,omitempty"`
}

// itemInfo mirrors overlay.itemInfo — keep the two in sync.
func itemInfo(it engine.Item, pos, total int, last bool) stepInfo {
	info := stepInfo{Current: pos, Total: total, IsFirst: pos == 1, IsLast: last, Section: it.Section}
	if it.IsChoice() {
		info.IsChoice = true
		info.Title = it.Choice.Prompt
		info.ChoiceKey = it.Choice.Key
		info.Selected = it.Selected
		for _, o := range it.Choice.Options {
			info.Options = append(info.Options, choiceOptionInfo{Value: o.Value, Label: o.Label, Description: o.Description})
		}
		return info
	}
	s := it.Step
	info.ID = s.ID
	info.Title = s.Title
	info.Description = s.Description
	info.Optional = s.Optional
	info.Hints = s.Hints
	info.Warnings = s.Warnings
	info.Infos = s.Infos
	for _, t := range s.Tasks {
		info.Tasks = append(info.Tasks, taskInfo{Text: t.Text, Info: t.Info, Warning: t.Warning, Hint: t.Hint})
	}
	for _, q := range s.Quests {
		info.Quests = append(info.Quests, questInfo{Name: q.Name, Status: q.Status, Note: q.Note})
	}
	return info
}

func (s *server) current() stepInfo {
	cur, total := s.eng.Progress()
	it := s.eng.Current()
	if it == nil {
		return stepInfo{Current: cur, Total: total}
	}
	return itemInfo(*it, cur, total, s.eng.Done())
}

func (s *server) steps() []stepInfo {
	items := s.eng.Items()
	total := len(items)
	out := make([]stepInfo, total)
	for i, it := range items {
		out[i] = itemInfo(it, i+1, total, s.eng.Done() && i == total-1)
	}
	return out
}

func (s *server) routes(mux *http.ServeMux) {
	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, map[string]string{"game": s.eng.Game(), "title": s.eng.Title(), "variant": s.eng.Variant()})
	})
	mux.HandleFunc("/api/steps", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.steps())
	})
	mux.HandleFunc("/api/current", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.current())
	})
	mux.HandleFunc("/api/nextfile", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.eng.NextFile())
	})
	mux.HandleFunc("/api/next", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = s.eng.Next()
		writeJSON(w, s.current())
	})
	mux.HandleFunc("/api/prev", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = s.eng.Prev()
		writeJSON(w, s.current())
	})
	mux.HandleFunc("/api/goto", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		i, _ := strconv.Atoi(r.URL.Query().Get("i"))
		_ = s.eng.Goto(i)
		writeJSON(w, s.current())
	})
	mux.HandleFunc("/api/choose", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = s.eng.Choose(r.URL.Query().Get("key"), r.URL.Query().Get("value"))
		writeJSON(w, s.current())
	})
	mux.HandleFunc("/api/loadnext", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.loadNext()
		writeJSON(w, s.current())
	})
	// hotkeynext mirrors overlay.App.next(): at the end of a walkthrough that has
	// a `next:` file, hand off to it; otherwise advance one step. The response's
	// "swapped" flag tells the mock whether to fire config:changed (full reload)
	// or step:changed (in-place step update), matching the real Wails events.
	mux.HandleFunc("/api/hotkeynext", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		swapped := false
		if s.eng.Done() && s.eng.NextFile() != "" {
			swapped = s.loadNext()
		} else {
			_ = s.eng.Next()
		}
		writeJSON(w, map[string]any{"swapped": swapped, "current": s.current()})
	})
}

// loadNext swaps the engine to the active walkthrough's `next:` file, tracking
// the new config's directory so a further hand-off resolves correctly. Caller
// must hold s.mu. Returns whether a swap actually happened.
func (s *server) loadNext() bool {
	next := s.eng.NextFile()
	if next == "" {
		return false
	}
	path := filepath.Join(s.cfgDir, next)
	wt, err := config.Load(path)
	if err != nil {
		log.Printf("devui: loadnext %q: %v", next, err)
		return false
	}
	s.eng = engine.New(wt)
	s.cfgDir = filepath.Dir(path)
	return true
}

// serveApp serves the real index.html with the Wails bindings mocked, so the
// untouched frontend runs in a plain browser against the live engine.
func serveApp(w http.ResponseWriter, r *http.Request) {
	raw, err := os.ReadFile(filepath.Join(frontendDir, "index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inject := `
<style>html { background: transparent !important; }</style>
<script>
// --- devui: mock of the Wails bindings (engine-backed via /api/*) ------
(function () {
  const listeners = {};
  const emit = (name, data) => (listeners[name] || []).forEach(cb => cb(data));
  const api = (p, opts) => fetch(p, opts).then(r => r.json());
  // Mock settings — mirrors settings.Defaults() (settings/settings.go). Save methods
  // just store and echo back; real registration/persistence only in the Wails build.
  const settings = { version: 1, opacity: 1.0, theme: 'dark', language: 'en', hotkeys: {
    next:         { mods: ['ctrl', 'alt'], key: 'right' },
    prev:         { mods: ['ctrl', 'alt'], key: 'left'  },
    toggleHide:   { mods: ['ctrl', 'alt'], key: 'h'     },
    focusOverlay: { mods: ['ctrl', 'alt'], key: 'm'     },
    quit:         { mods: ['ctrl', 'alt'], key: 'q'     },
  } };
  window.go = { overlay: { App: {
    IsPicker:     () => Promise.resolve(false), // devui always starts in steps view
    ListConfigs:  () => api('/api/meta').then(m => [{ game: m.game, title: m.title, author: '', chapter: 1, path: '(devui)', hash: '' }]),
    DownloadGame: () => Promise.resolve(), // no cache in devui; catalog is the single live config
    OpenBrowse:   () => Promise.resolve(''),
    LoadConfig:   () => Promise.resolve(),
    UnloadConfig: () => Promise.resolve(),
    Meta:         () => api('/api/meta'),
    Steps:        () => api('/api/steps'),
    CurrentStep:  () => api('/api/current'),
    NextFile:     () => api('/api/nextfile'),
    Next:         () => api('/api/next', { method: 'POST' }),
    Prev:         () => api('/api/prev', { method: 'POST' }),
    Goto:         (idx) => api('/api/goto?i=' + idx, { method: 'POST' }),
    Choose:       (key, value) => api('/api/choose?key=' + encodeURIComponent(key) + '&value=' + encodeURIComponent(value), { method: 'POST' }),
    LoadNext:     () => api('/api/loadnext', { method: 'POST' }),
    FitWindow:    () => {}, // no-op: the browser can't resize the OS window
    SaveWindowPos: () => Promise.resolve(), // no-op: the browser can't move the OS window
    Settings:     () => Promise.resolve(settings),
    SaveHotkeys:  (hk) => { settings.hotkeys = hk; return Promise.resolve(settings); },
    SaveOpacity:  (v) => { settings.opacity = v; return Promise.resolve(settings); },
    SaveTheme:    (t) => { settings.theme = t; return Promise.resolve(settings); },
    SaveLanguage: (l) => { settings.language = l; return Promise.resolve(settings); },
    // Clear/reset bindings — no progress store or cache exists in devui (the
    // catalog is the single live config), so these are no-ops; ResetSettings
    // restores the mock settings to defaults like the real binding does.
    ClearChapterProgress: () => Promise.resolve(),
    ClearGameProgress:    () => Promise.resolve(),
    ClearCache:           () => Promise.resolve(),
    ResetSettings:        () => {
      settings.opacity = 1.0; settings.theme = 'dark'; settings.language = 'en';
      settings.hotkeys = {
        next:         { mods: ['ctrl', 'alt'], key: 'right' },
        prev:         { mods: ['ctrl', 'alt'], key: 'left'  },
        toggleHide:   { mods: ['ctrl', 'alt'], key: 'h'     },
        focusOverlay: { mods: ['ctrl', 'alt'], key: 'm'     },
        quit:         { mods: ['ctrl', 'alt'], key: 'q'     },
      };
      return Promise.resolve(settings);
    },
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
    if (e.key === 'ArrowRight') {
      api('/api/hotkeynext', { method: 'POST' })
        .then(res => emit(res.swapped ? 'config:changed' : 'step:changed', res.current));
    } else if (e.key === 'ArrowLeft') {
      api('/api/prev', { method: 'POST' }).then(c => emit('step:changed', c));
    }
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

// serveAppJS serves the transpiled HUD bundle (frontend/app.js), which the
// injected index.html loads via <script src="app.js">. The watcher keeps it in
// sync with frontend/src/app.ts; here we just hand back the file on disk.
func serveAppJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, filepath.Join(frontendDir, "app.js"))
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
			// Re-transpile app.ts → app.js before notifying browsers, so editing
			// the TypeScript source live-reloads the running bundle. Re-baseline
			// the fingerprint AFTER the build so the resulting app.js write doesn't
			// retrigger the watcher (which would loop).
			if _, err := frontendbuild.Build(); err != nil {
				log.Printf("devui: frontend build failed: %v", err)
			}
			last = fingerprint(dir)
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
