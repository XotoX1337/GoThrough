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

	hub := newReloadHub()
	go watch(frontendDir, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHarness)
	mux.HandleFunc("/app", serveApp(stepsJSON))
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
func serveApp(stepsJSON []byte) http.HandlerFunc {
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
  let i = 0;
  const at = () => Promise.resolve(steps[i]);
  window.go = { overlay: { App: {
    CurrentStep: at,
    Next: () => { if (i < steps.length - 1) i++; return at(); },
    Prev: () => { if (i > 0) i--; return at(); },
  } } };
  window.runtime = { Quit: () => console.log('[devui] runtime.Quit() (no-op in browser)') };
})();
</script>`
		html := strings.Replace(string(raw), "</body>", inject+"\n</body>", 1)
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
    background: var(--scene), radial-gradient(120% 120% at 30% 20%, #2a3a5e 0%, #141a2e 45%, #06070f 100%);
    background-size: cover;
    background-position: center;
    display: grid;
    place-items: center;
    font-family: system-ui, sans-serif;
  }
  .hud {
    width: 440px;
    height: 220px;
    border-radius: 10px;
    background: transparent;
    box-shadow: 0 20px 60px rgba(0,0,0,0.55);
    overflow: hidden;
  }
  iframe { width: 100%; height: 100%; border: none; background: transparent; color-scheme: light dark; }
  .tag {
    position: fixed; bottom: 12px; left: 12px;
    font-size: 11px; color: rgba(255,255,255,0.4);
    letter-spacing: 0.05em;
  }
</style>
</head>
<body>
  <div class="hud"><iframe src="/app" title="GoThrough HUD"></iframe></div>
  <div class="tag">devui · live reload active · 440×220</div>
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
