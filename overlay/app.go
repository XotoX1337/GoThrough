package overlay

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/XotoX1337/GoThrough/config"
	"github.com/XotoX1337/GoThrough/configstore"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/progress"
	"github.com/XotoX1337/GoThrough/settings"
)

// stepChangedEvent is emitted to the frontend whenever the active step changes
// via a global hotkey (button-driven changes return the new state directly).
const stepChangedEvent = "step:changed"

// configChangedEvent is emitted when the active walkthrough is swapped out from
// under the frontend (e.g. a hotkey hand-off to the `next:` file). It carries no
// payload — the frontend re-fetches meta + steps + current, as a full reload.
const configChangedEvent = "config:changed"

// QuestInfo is a quest-log reference sent to the frontend.
type QuestInfo struct {
	Name   string `json:"name"`
	Status string `json:"status,omitempty"` // received | completed | ""
	Note   string `json:"note,omitempty"`
}

// ChoiceOptionInfo is one selectable answer of a choice.
type ChoiceOptionInfo struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// TaskInfo is one actionable sub-step within a step, with optional per-task
// callouts. Text is Markdown.
type TaskInfo struct {
	Text    string `json:"text"`
	Info    string `json:"info,omitempty"`
	Warning string `json:"warning,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

// StepInfo is the data shape sent to the Wails frontend. It represents one
// position in the resolved sequence: a normal step, OR a choice when IsChoice is
// true (in which case Title is the prompt, ChoiceKey identifies it, and Options
// are the answers). Description is Markdown.
type StepInfo struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	IsFirst bool   `json:"isFirst"`
	IsLast  bool   `json:"isLast"`
	Section string `json:"section,omitempty"`

	// Choice (IsChoice == true)
	IsChoice  bool               `json:"isChoice,omitempty"`
	ChoiceKey string             `json:"choiceKey,omitempty"`
	Selected  string             `json:"selected,omitempty"` // chosen option value ("" while undecided)
	Options   []ChoiceOptionInfo `json:"options,omitempty"`

	// Step content (IsChoice == false)
	ID          int         `json:"id,omitempty"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Tasks       []TaskInfo  `json:"tasks,omitempty"`
	Optional    bool        `json:"optional,omitempty"`
	Quests      []QuestInfo `json:"quests,omitempty"`
	Hints       []string    `json:"hints,omitempty"`
	Warnings    []string    `json:"warnings,omitempty"`
	Infos       []string    `json:"infos,omitempty"`
}

// MetaInfo describes the loaded walkthrough for the HUD header.
type MetaInfo struct {
	Game    string `json:"game"`
	Title   string `json:"title"`
	Variant string `json:"variant,omitempty"`
}

// App is the Go backend exposed to the frontend via Wails bindings.
//
// Engine access is guarded by mu because step changes arrive from two
// goroutines: the WebView thread (frontend-bound method calls) and the global
// hotkey listener (see hotkeys.go).
type App struct {
	mu  sync.Mutex
	eng *engine.Engine
	ctx context.Context // set in OnStartup; nil until the window is up

	set     *settings.Store
	hotkeys *hotkeyManager // set in OnStartup, once the window (and ctx) exist

	// Current config reference, kept so a `next:` hand-off can be resolved
	// relative to the file the active walkthrough was loaded from. curEmbedded
	// true means a catalog config (resolved against the on-disk cache); false
	// means an absolute path the user browsed to.
	curPath     string
	curEmbedded bool

	// catalog is the walkthrough catalog, fetched once from the CDN index.json
	// (with the embedded index.json as the offline fallback) and cached for the
	// process lifetime.
	catalogOnce sync.Once
	catalogVal  []configstore.Entry
}

// catalog returns the walkthrough catalog. The first call fetches index.json
// from the CDN with a short timeout; on any failure it falls back to the
// embedded index.json so the picker always has content. The result is cached
// for the process, so repeated picker visits don't re-hit the network.
func (a *App) catalog() []configstore.Entry {
	a.catalogOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
		defer cancel()
		if entries, err := configstore.ListRemote(ctx); err == nil {
			a.catalogVal = entries
			return
		}
		a.catalogVal = configstore.ListEmbedded()
	})
	return a.catalogVal
}

// IsPicker reports whether the app is in config-picker mode (no walkthrough loaded).
func (a *App) IsPicker() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.eng == nil
}

// ListConfigs returns the walkthrough catalog (remote-fetched once at startup,
// embedded fallback when offline).
func (a *App) ListConfigs() []configstore.Entry {
	return a.catalog()
}

// DownloadGame downloads every chapter of the given game into the on-disk cache
// so chapters load instantly and offline — including `next:` hand-offs across
// files. Called when the user picks a game in the two-level picker. Returns an
// error only when a chapter is neither freshly downloaded nor already cached.
func (a *App) DownloadGame(game string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return configstore.DownloadGame(ctx, a.catalog(), game)
}

// refreshUpdates auto-updates already-cached games whose chapters changed
// upstream (hash mismatch) or gained new chapters, then pushes the fresh catalog
// to the frontend so the picker reflects it. Runs in the background at startup.
func (a *App) refreshUpdates() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	catalog := a.catalog()
	configstore.RefreshUpdates(ctx, catalog)

	a.mu.Lock()
	wctx := a.ctx
	a.mu.Unlock()
	if wctx != nil {
		runtime.EventsEmit(wctx, "configs:remote", catalog)
	}
}

// OpenBrowse opens a native file dialog so the user can pick a walkthrough
// YAML outside the bundled set. Returns the selected path, or "" if cancelled.
func (a *App) OpenBrowse() string {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		return ""
	}
	path, _ := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "Walkthrough öffnen",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML Walkthrough (*.yaml, *.yml)", Pattern: "*.yaml;*.yml"},
		},
	})
	return path
}

// LoadConfig loads a walkthrough. When embedded is true, path is a catalog-
// relative path read from the on-disk cache (the game must have been downloaded
// first); when false, path is an absolute file the user browsed to. It wires up
// progress persistence and transitions the app from picker mode into
// walkthrough mode.
func (a *App) LoadConfig(path string, embedded bool) error {
	var data []byte
	var err error
	if embedded {
		data, err = configstore.ReadCached(path)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	wt, err := config.LoadBytes(data)
	if err != nil {
		return err
	}

	eng := engine.New(wt)
	attachProgress(eng, wt)

	a.mu.Lock()
	a.eng = eng
	a.curPath = path
	a.curEmbedded = embedded
	hm := a.hotkeys
	a.mu.Unlock()

	if hm != nil {
		hm.apply(a.set.Get().Hotkeys)
	}
	// Remember this as the walkthrough to reopen on next launch.
	a.saveLastConfig(settings.LastConfig{Path: path, Embedded: embedded})
	return nil
}

// UnloadConfig drops the active walkthrough and returns the app to picker mode.
func (a *App) UnloadConfig() {
	a.mu.Lock()
	a.eng = nil
	a.curPath = ""
	a.curEmbedded = false
	a.mu.Unlock()
	// Returning to the picker is now the remembered state; clear the auto-load.
	a.saveLastConfig(settings.LastConfig{})
}

// progressStore opens the on-disk progress store at its default path.
func progressStore() (*progress.Store, error) {
	path, err := progress.DefaultPath()
	if err != nil {
		return nil, err
	}
	return progress.Open(path)
}

// ClearChapterProgress deletes the saved progress for a single cached chapter,
// identified by its catalog-relative path (as listed in the picker). The
// chapter's YAML is read from the cache to recover its progress key.
func (a *App) ClearChapterProgress(relpath string) error {
	data, err := configstore.ReadCached(relpath)
	if err != nil {
		return fmt.Errorf("reading cached config: %w", err)
	}
	wt, err := config.LoadBytes(data)
	if err != nil {
		return err
	}
	store, err := progressStore()
	if err != nil {
		return err
	}
	return store.Delete(progress.Key(wt))
}

// ClearGameProgress deletes the saved progress for every chapter of a game.
func (a *App) ClearGameProgress(game string) error {
	store, err := progressStore()
	if err != nil {
		return err
	}
	return store.DeleteGame(game)
}

// ClearCache removes every downloaded config from the on-disk cache, so games
// must be re-downloaded from the catalog before their chapters load again.
func (a *App) ClearCache() error {
	return configstore.ClearCache()
}

// ResetSettings restores all user settings to their defaults (also clearing the
// remembered last walkthrough), re-registers the default hotkeys live, and
// returns the fresh settings so the frontend can re-render theme/language/opacity.
func (a *App) ResetSettings() (settings.Settings, error) {
	def := settings.Defaults()
	if err := a.set.Save(def); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}
	a.mu.Lock()
	hm := a.hotkeys
	a.mu.Unlock()
	if hm != nil {
		hm.apply(def.Hotkeys)
	}
	return def, nil
}

// saveLastConfig persists the last-loaded walkthrough reference. It is
// best-effort: a write failure must not stop the walkthrough from loading.
func (a *App) saveLastConfig(lc settings.LastConfig) {
	ns := a.set.Get()
	ns.LastConfig = lc
	_ = a.set.Save(ns)
}

// restoreLastConfig reopens the walkthrough recorded by the previous session,
// turning startup-in-picker-mode into startup-in-the-last-walkthrough. It does
// nothing if a walkthrough is already loaded (e.g. launched via `run <config>`)
// or if none was remembered. A stale reference (config moved, deleted, or no
// longer bundled) is cleared so the app falls back to the picker cleanly.
func (a *App) restoreLastConfig() {
	a.mu.Lock()
	loaded := a.eng != nil
	a.mu.Unlock()
	if loaded {
		return
	}

	lc := a.set.Get().LastConfig
	if lc.Path == "" {
		return
	}
	if err := a.LoadConfig(lc.Path, lc.Embedded); err != nil {
		a.saveLastConfig(settings.LastConfig{})
	}
}

// Meta returns the walkthrough header info (game + title).
func (a *App) Meta() MetaInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return MetaInfo{}
	}
	return MetaInfo{Game: a.eng.Game(), Title: a.eng.Title(), Variant: a.eng.Variant()}
}

// Steps returns every item in the resolved sequence so the HUD can render its
// (section-grouped) checklist. Choices appear as items too.
func (a *App) Steps() []StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return nil
	}
	items := a.eng.Items()
	total := len(items)
	out := make([]StepInfo, total)
	for i, it := range items {
		out[i] = itemInfo(it, i+1, total, a.eng.Done() && i == total-1)
	}
	return out
}

// Choose records the answer for the choice the user is currently facing and
// returns the new active item (the item right after the choice).
func (a *App) Choose(choiceKey, value string) StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return StepInfo{}
	}
	_ = a.eng.Choose(choiceKey, value)
	return a.stepInfo()
}

// NextFile returns the raw `next:` reference of the active walkthrough, or ""
// if there is no follow-up file. The HUD shows a hand-off button when this is
// non-empty and the user reaches the end.
func (a *App) NextFile() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return ""
	}
	return a.eng.NextFile()
}

// LoadNext resolves the active walkthrough's `next:` reference relative to the
// current config and loads it, handing off to the follow-up walkthrough.
func (a *App) LoadNext() error {
	a.mu.Lock()
	next := ""
	if a.eng != nil {
		next = a.eng.NextFile()
	}
	curPath, embedded := a.curPath, a.curEmbedded
	a.mu.Unlock()
	if next == "" {
		return fmt.Errorf("no next file")
	}

	var resolved string
	if embedded {
		// Catalog paths use forward slashes; resolve in that space, then read
		// the follow-up from the cache (DownloadGame fetched all chapters).
		resolved = path.Join(path.Dir(curPath), next)
	} else {
		resolved = filepath.Join(filepath.Dir(curPath), next)
	}
	return a.LoadConfig(resolved, embedded)
}

func (a *App) CurrentStep() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return StepInfo{}
	}
	return a.stepInfo()
}

func (a *App) Next() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return StepInfo{}
	}
	_ = a.eng.Next()
	return a.stepInfo()
}

func (a *App) Prev() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return StepInfo{}
	}
	_ = a.eng.Prev()
	return a.stepInfo()
}

// Goto jumps to a 0-based step index (used by checklist row clicks).
func (a *App) Goto(index int) StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.eng == nil {
		return StepInfo{}
	}
	_ = a.eng.Goto(index)
	return a.stepInfo()
}

// Settings returns the current user settings for the HUD's settings panel.
func (a *App) Settings() settings.Settings {
	return a.set.Get()
}

// SaveHotkeys validates, persists, and re-registers a new set of hotkey
// bindings. Returns the stored settings on success; an invalid binding is
// rejected and the previous bindings stay in effect.
func (a *App) SaveHotkeys(hk settings.Hotkeys) (settings.Settings, error) {
	if err := validateHotkeys(hk); err != nil {
		return a.set.Get(), err
	}

	ns := a.set.Get()
	ns.Hotkeys = hk
	if err := a.set.Save(ns); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}

	a.mu.Lock()
	hm := a.hotkeys
	a.mu.Unlock()
	if hm != nil {
		hm.apply(hk)
	}
	return ns, nil
}

// SaveOpacity persists the panel opacity (0.1–1.0) and returns the updated settings.
func (a *App) SaveOpacity(opacity float64) (settings.Settings, error) {
	if opacity < 0.1 || opacity > 1.0 {
		return a.set.Get(), fmt.Errorf("opacity must be between 0.1 and 1.0")
	}
	ns := a.set.Get()
	ns.Opacity = opacity
	if err := a.set.Save(ns); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}
	return ns, nil
}

// SaveTheme persists the HUD colour theme (dark | light | contrast) and returns
// the updated settings. An unknown theme is rejected so the frontend can't
// store a value its CSS doesn't define.
func (a *App) SaveTheme(theme string) (settings.Settings, error) {
	switch theme {
	case "dark", "light", "contrast":
	default:
		return a.set.Get(), fmt.Errorf("unknown theme %q", theme)
	}
	ns := a.set.Get()
	ns.Theme = theme
	if err := a.set.Save(ns); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}
	return ns, nil
}

// SaveLanguage persists the HUD interface language (en | de) and returns the
// updated settings. An unknown code is rejected so the frontend can't store a
// value its string table doesn't define.
func (a *App) SaveLanguage(lang string) (settings.Settings, error) {
	switch lang {
	case "en", "de":
	default:
		return a.set.Get(), fmt.Errorf("unknown language %q", lang)
	}
	ns := a.set.Get()
	ns.Language = lang
	if err := a.set.Save(ns); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}
	return ns, nil
}

// SaveWindowPos persists the overlay window's on-screen position (logical px,
// top-left corner) so it is restored on the next launch instead of re-anchoring
// to the top-right corner. Called by the frontend at the end of a window drag.
// Best-effort: a write failure is reported but never blocks dragging.
func (a *App) SaveWindowPos(x, y int) error {
	ns := a.set.Get()
	ns.WindowPos = settings.WindowPos{X: x, Y: y, Set: true}
	if err := a.set.Save(ns); err != nil {
		return fmt.Errorf("saving window position: %w", err)
	}
	return nil
}

// validateHotkeys checks that every binding resolves to a real key/modifier
// combination before it is persisted or registered.
func validateHotkeys(hk settings.Hotkeys) error {
	for name, b := range map[string]settings.Binding{
		"next": hk.Next, "prev": hk.Prev, "toggleHide": hk.ToggleHide,
		"focusOverlay": hk.FocusOverlay, "quit": hk.Quit,
	} {
		var err error
		if b.IsMouse() {
			_, _, err = resolveMouse(b)
		} else {
			_, _, err = resolve(b)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

// FitWindow shrink-wraps the OS window to the given content size (logical px),
// keeping the window's current top-left corner fixed — the same corner
// SaveWindowPos/restoreWindowPos persist. Anchoring the top-right instead made
// the boot-time resize (the window is created at overlayWidth, then the frontend
// fits it to its real, wider size) drag the restored window left by the width
// difference. The result is clamped so a fresh-install top-right placement that
// grows wider can't clip off the screen edge.
func (a *App) FitWindow(width, height int) {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil || width < 1 || height < 1 {
		return
	}
	x, y := runtime.WindowGetPosition(ctx)
	runtime.WindowSetSize(ctx, width, height)
	x, y = clampToScreen(ctx, x, y, width, height)
	runtime.WindowSetPosition(ctx, x, y)
}

// next/prev are the hotkey-driven counterparts to Next/Prev. When "next" is
// pressed on the final step of a walkthrough that has a `next:` file, it hands
// off to that file instead of doing nothing — the same action the on-screen
// "Weiter" button performs.
func (a *App) next() {
	a.mu.Lock()
	if a.eng == nil {
		a.mu.Unlock()
		return
	}
	handoff := a.eng.Done() && a.eng.NextFile() != ""
	a.mu.Unlock()
	if handoff {
		a.handOff()
		return
	}
	a.advance((*engine.Engine).Next)
}

func (a *App) prev() { a.advance((*engine.Engine).Prev) }

// handOff loads the active walkthrough's `next:` file and tells the frontend to
// reload from scratch. Used by the hotkey "next" at the end of a walkthrough.
func (a *App) handOff() {
	if err := a.LoadNext(); err != nil {
		return
	}
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx != nil {
		runtime.EventsEmit(ctx, configChangedEvent, nil)
	}
}

func (a *App) advance(move func(*engine.Engine) error) {
	a.mu.Lock()
	if a.eng == nil {
		a.mu.Unlock()
		return
	}
	_ = move(a.eng)
	info := a.stepInfo()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx != nil {
		runtime.EventsEmit(ctx, stepChangedEvent, info)
	}
}

// stepInfo builds the StepInfo for the active item. Caller must hold a.mu.
func (a *App) stepInfo() StepInfo {
	current, total := a.eng.Progress()
	it := a.eng.Current()
	if it == nil {
		return StepInfo{Current: current, Total: total}
	}
	return itemInfo(*it, current, total, a.eng.Done())
}

// itemInfo converts an engine.Item into the frontend StepInfo shape. pos is the
// 1-based position; last marks whether this is the final (completed) item.
func itemInfo(it engine.Item, pos, total int, last bool) StepInfo {
	info := StepInfo{
		Current: pos,
		Total:   total,
		IsFirst: pos == 1,
		IsLast:  last,
		Section: it.Section,
	}
	if it.IsChoice() {
		info.IsChoice = true
		info.Title = it.Choice.Prompt
		info.ChoiceKey = it.Choice.Key
		info.Selected = it.Selected
		for _, o := range it.Choice.Options {
			info.Options = append(info.Options, ChoiceOptionInfo{Value: o.Value, Label: o.Label, Description: o.Description})
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
		info.Tasks = append(info.Tasks, TaskInfo{Text: t.Text, Info: t.Info, Warning: t.Warning, Hint: t.Hint})
	}
	for _, q := range s.Quests {
		info.Quests = append(info.Quests, QuestInfo{Name: q.Name, Status: q.Status, Note: q.Note})
	}
	return info
}

// attachProgress wires the engine to the on-disk progress store, restoring any
// saved position. Progress is reset via the clear bindings, not at load time.
func attachProgress(eng *engine.Engine, wt *config.Walkthrough) {
	path, err := progress.DefaultPath()
	if err != nil {
		return
	}
	store, err := progress.Open(path)
	if err != nil {
		return
	}
	h := store.For(wt)
	if index, stepID, choices, ok := h.Load(); ok {
		eng.Restore(index, stepID, choices)
	}
	eng.UsePersister(h)
}
